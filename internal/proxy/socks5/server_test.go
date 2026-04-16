package socks5

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSocks5Listen(t *testing.T) {
	srv := NewServer("127.0.0.1:0") // port 0 = OS assigns

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := srv.Listen(ctx)
	require.NoError(t, err)
	require.NotNil(t, ch)

	// Read the allocated address
	addr := srv.Addr()
	assert.NotEmpty(t, addr)
	t.Logf("SOCKS5 listening on %s", addr)

	// Connect as a SOCKS5 client in background
	go func() {
		time.Sleep(100 * time.Millisecond)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return
		}
		defer conn.Close()

		// SOCKS5 greeting: version 5, 1 auth method (no auth)
		conn.Write([]byte{0x05, 0x01, 0x00})

		// Read server choice
		buf := make([]byte, 2)
		io.ReadFull(conn, buf)

		// SOCKS5 connect request to 127.0.0.1:80
		// VER=5, CMD=1(CONNECT), RSV=0, ATYP=1(IPv4), ADDR=127.0.0.1, PORT=0x0050
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x7f, 0x00, 0x00, 0x01, 0x00, 0x50})

		// Read response (just drain it)
		resp := make([]byte, 10)
		io.ReadFull(conn, resp)

		// Send some data
		conn.Write([]byte("test"))
	}()

	// Accept the request from channel
	select {
	case req := <-ch:
		assert.NotNil(t, req)
		assert.Equal(t, "127.0.0.1:80", req.Target)
		assert.NotNil(t, req.Conn)
		req.Conn.Close()
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SOCKS5 connection")
	}
}

func TestSocks5ServerAddr(t *testing.T) {
	srv := NewServer("127.0.0.1:0")
	assert.Equal(t, "127.0.0.1:0", srv.Addr())
}
