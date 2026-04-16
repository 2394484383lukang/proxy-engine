package hub

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/proxy-engine/internal/proxy"
)

// mockProxy records what target was dialed
type mockProxy struct {
	dialErr error
	target  string
}

func (m *mockProxy) Dial(ctx context.Context, target string) (net.Conn, error) {
	if m.dialErr != nil {
		return nil, m.dialErr
	}
	m.target = target
	// Return a pair of connected pipes
	serverConn, clientConn := net.Pipe()
	// Echo data back
	go func() {
		io.Copy(serverConn, serverConn)
		serverConn.Close()
	}()
	return clientConn, nil
}

func (m *mockProxy) Type() proxy.ProxyType {
	return proxy.ProxyTypeDirect
}

func TestHubDispatchDirect(t *testing.T) {
	h := NewHub()
	mp := &mockProxy{}
	h.SetOutbound("default", mp)

	// Simulate an inbound connection via pipe
	clientConn, serverConn := net.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Dispatch in background
	go h.Dispatch(ctx, &proxy.ConnRequest{
		Target: "example.com:80",
		Conn:   serverConn,
	}, "default")

	// Write something on client side
	clientConn.Write([]byte("hello"))
	buf := make([]byte, 5)
	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := clientConn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(buf[:n]))

	clientConn.Close()
	assert.Equal(t, "example.com:80", mp.target)
}

func TestHubDispatchReject(t *testing.T) {
	h := NewHub()
	rejectProxy := &proxy.RejectProxy{}
	h.SetOutbound("reject", rejectProxy)

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	h.Dispatch(ctx, &proxy.ConnRequest{
		Target: "example.com:80",
		Conn:   serverConn,
	}, "reject")

	// Server conn should be closed by hub
	serverConn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err := serverConn.Read(make([]byte, 1))
	assert.Error(t, err) // connection should be closed
}

func TestHubSetOutbound(t *testing.T) {
	h := NewHub()
	mp := &mockProxy{}
	h.SetOutbound("test", mp)
	assert.NotNil(t, h.GetOutbound("test"))
	assert.Nil(t, h.GetOutbound("nonexistent"))
}
