package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/proxy-engine/internal/config"
)

// Server provides a RESTful API for controlling the proxy engine.
type Server struct {
	config  *config.Config
	handler http.Handler
}

func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()
	s := &Server{
		config:  cfg,
		handler: mux,
	}

	mux.HandleFunc("/api/configs", s.handleConfigs)
	mux.HandleFunc("/api/traffic", s.handleTraffic)
	mux.HandleFunc("/api/health", s.handleHealth)

	return s
}

// Handler returns the http.Handler for use with http.Server.
func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) handleConfigs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}

func (s *Server) handleTraffic(w http.ResponseWriter, r *http.Request) {
	// WebSocket upgrade placeholder — returns empty JSON for now.
	// Full WebSocket implementation in Phase 7.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"up": 0, "down": 0})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
