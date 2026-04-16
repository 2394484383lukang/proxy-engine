package internal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/proxy-engine/internal/hub"
	"github.com/user/proxy-engine/internal/proxy"
	socks5server "github.com/user/proxy-engine/internal/proxy/socks5"
	httpserver "github.com/user/proxy-engine/internal/proxy/http"
)

func TestE2ESOCKS5Proxy(t *testing.T) {
	// Start a backend HTTP server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("backend-response"))
	}))
	defer backend.Close()

	// Start SOCKS5 inbound
	socks5 := socks5server.NewServer("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	socksCh, err := socks5.Listen(ctx)
	require.NoError(t, err)

	h := hub.NewHub()
	h.SetOutbound("DIRECT", proxy.NewDirectProxy())
	go func() {
		for req := range socksCh {
			go h.Dispatch(ctx, req, "DIRECT")
		}
	}()

	// Connect through SOCKS5 proxy to backend
	backendAddr := backend.Listener.Addr().String()
	conn, err := net.Dial("tcp", socks5.Addr())
	require.NoError(t, err)
	defer conn.Close()

	// SOCKS5 handshake
	conn.Write([]byte{0x05, 0x01, 0x00}) // greeting
	buf := make([]byte, 2)
	io.ReadFull(conn, buf)

	// Parse backend host:port
	host, portStr, _ := net.SplitHostPort(backendAddr)
	port := 0
	fmt.Sscanf(portStr, "%d", &port)

	// SOCKS5 connect request (domain type for variety)
	domain := []byte(host)
	req := []byte{0x05, 0x01, 0x00, 0x03, byte(len(domain))}
	req = append(req, domain...)
	req = append(req, byte(port>>8), byte(port&0xff))
	conn.Write(req)

	resp := make([]byte, 10)
	io.ReadFull(conn, resp)
	assert.Equal(t, byte(0x05), resp[0])
	assert.Equal(t, byte(0x00), resp[1]) // success

	// Send HTTP request through the tunnel
	httpReq := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", backendAddr)
	conn.Write([]byte(httpReq))

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	reader := bufio.NewReader(conn)
	respLine, err := reader.ReadString('\n')
	require.NoError(t, err)
	assert.Contains(t, respLine, "200")

	// Read body
	body, _ := io.ReadAll(reader)
	assert.Contains(t, string(body), "backend-response")
}

func TestE2EHTTPProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("http-backend"))
	}))
	defer backend.Close()

	httpProxy := httpserver.NewServer("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpCh, err := httpProxy.Listen(ctx)
	require.NoError(t, err)

	h := hub.NewHub()
	h.SetOutbound("DIRECT", proxy.NewDirectProxy())
	go func() {
		for req := range httpCh {
			go h.Dispatch(ctx, req, "DIRECT")
		}
	}()

	// Use Go's http client with proxy
	proxyURL, _ := url.Parse("http://" + httpProxy.Addr())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(backend.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Logf("Response body: %q", string(body))
	assert.Contains(t, string(body), "http-backend")
}
