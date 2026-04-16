package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/proxy-engine/internal/api"
	"github.com/user/proxy-engine/internal/config"
	"github.com/user/proxy-engine/internal/hub"
	"github.com/user/proxy-engine/internal/proxy"
	socks5server "github.com/user/proxy-engine/internal/proxy/socks5"
	httpserver "github.com/user/proxy-engine/internal/proxy/http"
)

var (
	version   = "0.1.0"
	buildDate = "unknown"
)

func main() {
	configPath := flag.String("c", "config.yaml", "path to config file")
	showVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("proxy-engine %s (built %s)\n", version, buildDate)
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("proxy-engine v%s starting", version)
	log.Printf("config: mode=%s port=%d socks-port=%d", cfg.Mode, cfg.Port, cfg.SocksPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup Hub with DIRECT as default outbound
	h := hub.NewHub()
	h.SetOutbound("DIRECT", proxy.NewDirectProxy())

	// Start SOCKS5 inbound
	if cfg.SocksPort > 0 {
		socks5 := socks5server.NewServer(fmt.Sprintf("127.0.0.1:%d", cfg.SocksPort))
		socksCh, err := socks5.Listen(ctx)
		if err != nil {
			log.Fatalf("failed to start SOCKS5: %v", err)
		}
		go consumeInbound(ctx, h, socksCh, "DIRECT")
		log.Printf("SOCKS5 listening on %s", socks5.Addr())
	}

	// Start HTTP proxy inbound
	if cfg.Port > 0 {
		httpProxy := httpserver.NewServer(fmt.Sprintf("127.0.0.1:%d", cfg.Port))
		httpCh, err := httpProxy.Listen(ctx)
		if err != nil {
			log.Fatalf("failed to start HTTP proxy: %v", err)
		}
		go consumeInbound(ctx, h, httpCh, "DIRECT")
		log.Printf("HTTP proxy listening on %s", httpProxy.Addr())
	}

	// Start API server
	apiSrv := api.New(cfg)
	httpServer := &http.Server{
		Addr:    "127.0.0.1:9090",
		Handler: apiSrv.Handler(),
	}
	go func() {
		log.Printf("API server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down...")
	cancel()
	httpServer.Close()
	log.Println("bye")
}

func consumeInbound(ctx context.Context, h *hub.Hub, ch <-chan *proxy.ConnRequest, outbound string) {
	for {
		select {
		case req, ok := <-ch:
			if !ok {
				return
			}
			go h.Dispatch(ctx, req, outbound)
		case <-ctx.Done():
			return
		}
	}
}
