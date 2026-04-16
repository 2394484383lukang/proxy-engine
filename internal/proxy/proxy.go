package proxy

import (
	"context"
	"net"
)

type ProxyType string

const (
	ProxyTypeDirect ProxyType = "DIRECT"
	ProxyTypeReject ProxyType = "REJECT"
	ProxyTypeSOCKS5 ProxyType = "socks5"
	ProxyTypeHTTP   ProxyType = "http"
	ProxyTypeSS     ProxyType = "ss"
	ProxyTypeVMess  ProxyType = "vmess"
	ProxyTypeVLESS  ProxyType = "vless"
	ProxyTypeTrojan ProxyType = "trojan"
)

// Proxy is the unified outbound interface. Implementations establish a
// connection to the target address through their respective protocol.
type Proxy interface {
	// Dial establishes a connection to target (host:port) and returns
	// a net.Conn that the caller can read/write directly.
	Dial(ctx context.Context, target string) (net.Conn, error)
	// Type returns the protocol type identifier.
	Type() ProxyType
}

// Inbound accepts incoming connections from clients.
type Inbound interface {
	// Listen starts accepting connections. It returns a channel of
	// ConnRequest that the Hub consumes.
	Listen(ctx context.Context) (<-chan *ConnRequest, error)
	// Addr returns the listen address.
	Addr() string
}

// ConnRequest represents an inbound connection waiting to be dispatched.
type ConnRequest struct {
	// Target is the destination address (host:port) the client wants to reach.
	Target string
	// Conn is the client-side connection. The Hub will pipe data between
	// this and the outbound connection.
	Conn net.Conn
}
