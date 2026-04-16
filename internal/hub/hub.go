package hub

import (
	"context"
	"io"
	"log"
	"sync"

	"github.com/user/proxy-engine/internal/proxy"
)

// Hub is the central connection dispatcher. It receives ConnRequests from
// inbound servers and dispatches them to the appropriate outbound proxy.
type Hub struct {
	mu       sync.RWMutex
	outbound map[string]proxy.Proxy
}

func NewHub() *Hub {
	return &Hub{
		outbound: make(map[string]proxy.Proxy),
	}
}

// SetOutbound registers an outbound proxy by name.
func (h *Hub) SetOutbound(name string, p proxy.Proxy) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.outbound[name] = p
}

// GetOutbound returns the outbound proxy by name, or nil if not found.
func (h *Hub) GetOutbound(name string) proxy.Proxy {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.outbound[name]
}

// Dispatch connects the inbound request to the specified outbound proxy
// and bidirectionally copies data until one side closes.
func (h *Hub) Dispatch(ctx context.Context, req *proxy.ConnRequest, outboundName string) {
	p := h.GetOutbound(outboundName)
	if p == nil {
		log.Printf("hub: outbound %q not found, closing connection", outboundName)
		req.Conn.Close()
		return
	}

	remote, err := p.Dial(ctx, req.Target)
	if err != nil {
		log.Printf("hub: dial %s via %s failed: %v", req.Target, outboundName, err)
		req.Conn.Close()
		return
	}

	// Bidirectional copy
	go func() {
		io.Copy(remote, req.Conn)
		remote.Close()
	}()
	io.Copy(req.Conn, remote)
	req.Conn.Close()
}
