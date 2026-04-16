package http

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/user/proxy-engine/internal/proxy"
)

// Server implements an HTTP proxy inbound supporting CONNECT method.
type Server struct {
	addr     string
	listener net.Listener
	mu       sync.Mutex
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

func (s *Server) Listen(ctx context.Context) (<-chan *proxy.ConnRequest, error) {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return nil, fmt.Errorf("http proxy listen: %w", err)
	}
	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()

	ch := make(chan *proxy.ConnRequest, 128)

	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				ln.Close()
				return
			default:
			}
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go s.handleConn(ctx, conn, ch)
		}
	}()

	return ch, nil
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn, ch chan<- *proxy.ConnRequest) {
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	reader := bufio.NewReader(conn)

	// Read first line: METHOD URI HTTP/version
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 2 {
		return
	}
	method := parts[0]
	target := parts[1]

	// Drain remaining headers
	for {
		headerLine, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		if strings.TrimSpace(headerLine) == "" {
			break
		}
	}

	if method != "CONNECT" {
		// Only CONNECT is supported in phase 1
		_, _ = conn.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
		return
	}

	// Respond with 200 Connection Established
	if _, err := conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n")); err != nil {
		return
	}

	// Hand off to hub
	req := &proxy.ConnRequest{
		Target: target,
		Conn:   conn,
	}
	c := conn
	conn = nil // prevent double close

	select {
	case ch <- req:
		_ = c
	case <-ctx.Done():
		c.Close()
	}
}
