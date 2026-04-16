package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port      int    `yaml:"port"`
	SocksPort int    `yaml:"socks-port"`
	MixedPort int    `yaml:"mixed-port"`
	Mode      string `yaml:"mode"`
	LogLevel  string `yaml:"log-level"`

	// Future phases will add: dns, tun, proxies, proxy-groups, rules, rule-providers
}

func defaultConfig() *Config {
	return &Config{
		LogLevel: "info",
	}
}

// Parse reads YAML config bytes and returns a Config.
func Parse(data []byte) (*Config, error) {
	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// Load reads a config file from disk.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return Parse(data)
}
