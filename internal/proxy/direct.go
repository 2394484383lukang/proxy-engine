package proxy

import (
	"context"
	"net"
)

// DirectProxy connects directly to the target without any proxy protocol.
type DirectProxy struct {
	dialer *net.Dialer
}

// NewDirectProxy creates a new DirectProxy with a reusable dialer.
func NewDirectProxy() *DirectProxy {
	return &DirectProxy{dialer: &net.Dialer{}}
}

func (d *DirectProxy) Dial(ctx context.Context, target string) (net.Conn, error) {
	return d.dialer.DialContext(ctx, "tcp", target)
}

func (d *DirectProxy) Type() ProxyType {
	return ProxyTypeDirect
}
