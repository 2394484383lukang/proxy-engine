package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/proxy-engine/internal/config"
)

func TestAPIGetConfigs(t *testing.T) {
	cfg := &config.Config{
		Port:      7890,
		SocksPort: 7891,
		Mode:      "rule",
		LogLevel:  "info",
	}
	srv := New(cfg)

	req := httptest.NewRequest("GET", "/api/configs", nil)
	w := httptest.NewRecorder()
	srv.handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "rule", result["Mode"])
}

func TestAPIGetTraffic(t *testing.T) {
	cfg := &config.Config{Mode: "rule"}
	srv := New(cfg)

	req := httptest.NewRequest("GET", "/api/traffic", nil)
	w := httptest.NewRecorder()
	srv.handler.ServeHTTP(w, req)

	// Traffic endpoint upgrades to WebSocket; without WS client we just
	// verify the route exists (returns 400 bad request upgrade)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK)
}

func TestAPIHealthCheck(t *testing.T) {
	cfg := &config.Config{Mode: "rule"}
	srv := New(cfg)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	srv.handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}
