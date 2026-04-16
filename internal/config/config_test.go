package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	raw := `
port: 7890
socks-port: 7891
mixed-port: 7892
mode: rule
log-level: info
`

	cfg, err := Parse([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, 7890, cfg.Port)
	assert.Equal(t, 7891, cfg.SocksPort)
	assert.Equal(t, 7892, cfg.MixedPort)
	assert.Equal(t, "rule", cfg.Mode)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestParseConfigDefaults(t *testing.T) {
	raw := `mode: direct`
	cfg, err := Parse([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, 0, cfg.Port)
	assert.Equal(t, 0, cfg.SocksPort)
	assert.Equal(t, "direct", cfg.Mode)
	assert.Equal(t, "info", cfg.LogLevel) // default
}

func TestParseConfigInvalid(t *testing.T) {
	_, err := Parse([]byte("not: valid: yaml: ["))
	// Actually YAML can parse a lot of things, so test something truly broken
	_, err = Parse([]byte("port: not_a_number"))
	assert.Error(t, err)
}

func TestLoadFromFile(t *testing.T) {
	// This test will be implemented when we have a temp file helper
	t.Skip("file loading tested in integration")
}
