package config

import (
	"os"
)

// Config holds configuration for NXLaunch
type Config struct {
	PayloadDir string
	Timeout    int
	Verbose    bool
	RetryCount int
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		PayloadDir: "payloads",
		Timeout:    5000,
		Verbose:    false,
		RetryCount: 3,
	}
}

// GetPayloadDir returns the payload directory
func (c *Config) GetPayloadDir() string {
	return c.PayloadDir
}

// EnsurePayloadDir creates the payload directory if it doesn't exist
func (c *Config) EnsurePayloadDir() error {
	return os.MkdirAll(c.PayloadDir, 0755)
}
