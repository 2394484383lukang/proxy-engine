package proxy

import (
	"context"
	"errors"
	"net"
)

// RejectProxy immediately rejects all connections.
type RejectProxy struct{}

var ErrRejected = errors.New("connection rejected by rule")

func (r *RejectProxy) Dial(ctx context.Context, target string) (net.Conn, error) {
	return nil, ErrRejected
}

func (r *RejectProxy) Type() ProxyType {
	return ProxyTypeReject
}
