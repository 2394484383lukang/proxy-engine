package http

import (
	"bufio"
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPConnectListen(t *testing.T) {
	srv := NewServer("127.0.0.1:0")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := srv.Listen(ctx)
	require.NoError(t, err)
	require.NotNil(t, ch)

	addr := srv.Addr()
	assert.NotEmpty(t, addr)
	t.Logf("HTTP proxy listening on %s", addr)

	// Connect as an HTTP CONNECT client
	go func() {
		time.Sleep(100 * time.Millisecond)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send HTTP CONNECT request
		connectReq := "CONNECT 127.0.0.1:80 HTTP/1.1\r\nHost: 127.0.0.1:80\r\n\r\n"
		conn.Write([]byte(connectReq))

		// Read response
		reader := bufio.NewReader(conn)
		reader.ReadString('\n') // drain the response line
	}()

	select {
	case req := <-ch:
		assert.NotNil(t, req)
		assert.Equal(t, "127.0.0.1:80", req.Target)
		assert.NotNil(t, req.Conn)
		req.Conn.Close()
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for HTTP CONNECT")
	}
}

func TestHTTPServerAddr(t *testing.T) {
	srv := NewServer("127.0.0.1:0")
	assert.Equal(t, "127.0.0.1:0", srv.Addr())
}
