package socks5

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/user/proxy-engine/internal/proxy"
)

// Server implements a SOCKS5 inbound that accepts connections and
// emits ConnRequest through a channel.
type Server struct {
	addr     string
	listener *net.TCPListener
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
		return nil, fmt.Errorf("socks5 listen: %w", err)
	}
	tcpLn := ln.(*net.TCPListener)
	s.mu.Lock()
	s.listener = tcpLn
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
			// Set deadline to allow context cancellation
			tcpLn.SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := ln.Accept()
			if err != nil {
				// Check if it's a timeout (expected) or real error
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout is expected, continue loop
				}
				return // Real error, exit
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

	// Phase 1: Method negotiation
	if err := s.negotiate(conn); err != nil {
		return
	}

	// Phase 2: Read connect request
	target, err := s.readRequest(conn)
	if err != nil {
		return
	}

	// Send success reply (VER=5, REP=0, RSV=0, ATYP=1, ADDR=0.0.0.0, PORT=0)
	if _, err := conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}); err != nil {
		return
	}

	// Hand off to hub — the hub owns this conn now
	req := &proxy.ConnRequest{
		Target: target,
		Conn:   conn,
	}
	// Prevent double close: set local conn to nil since hub owns it
	c := conn
	conn = nil

	select {
	case ch <- req:
		// Hub will close c
		_ = c
	case <-ctx.Done():
		c.Close()
	}
}

func (s *Server) negotiate(conn net.Conn) error {
	// Client sends: VER (1 byte) + NMETHODS (1 byte) + METHODS (NMETHODS bytes)
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}
	version := buf[0]
	nMethods := buf[1]
	if version != 0x05 {
		return fmt.Errorf("unsupported socks version: %d", version)
	}
	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return err
	}
	// Check if no-auth (0x00) is offered
	hasNoAuth := false
	for _, m := range methods {
		if m == 0x00 {
			hasNoAuth = true
			break
		}
	}
	if !hasNoAuth {
		// No acceptable method
		conn.Write([]byte{0x05, 0xFF})
		return fmt.Errorf("no acceptable method")
	}
	// Reply with NO AUTH (0x00)
	if _, err := conn.Write([]byte{0x05, 0x00}); err != nil {
		return err
	}
	return nil
}

func (s *Server) readRequest(conn net.Conn) (string, error) {
	// VER(1) + CMD(1) + RSV(1) + ATYP(1) = 4 bytes header
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", err
	}
	if header[0] != 0x05 {
		return "", fmt.Errorf("unsupported socks version: %d", header[0])
	}
	if header[1] != 0x01 {
		return "", fmt.Errorf("unsupported command: %d (only CONNECT)", header[1])
	}
	if header[2] != 0x00 {
		return "", fmt.Errorf("reserved field must be zero")
	}

	var host string
	switch header[3] {
	case 0x01: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return "", err
		}
		host = net.IP(ip).String()
	case 0x03: // Domain
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return "", err
		}
		domain := make([]byte, lenBuf[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return "", err
		}
		host = string(domain)
	case 0x04: // IPv6
		ip := make([]byte, 16)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return "", err
		}
		host = net.IP(ip).String()
	default:
		return "", fmt.Errorf("unsupported address type: %d", header[3])
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return "", err
	}
	port := binary.BigEndian.Uint16(portBuf)

	return fmt.Sprintf("%s:%d", host, port), nil
}
