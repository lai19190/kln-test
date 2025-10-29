package config

import (
	"encoding/json"
	"os"
	"sync"
)

// Config holds all configuration settings
type Config struct {
	mu     sync.RWMutex
	path   string
	Worker WorkerConfig `json:"worker"`
	Auth   AuthConfig   `json:"auth"`
}

type WorkerConfig struct {
	PoolSize  int               `json:"poolSize"`
	QueueSize int               `json:"queueSize"`
	Retry     WorkerRetryConfig `json:"retry"`
}

type WorkerRetryConfig struct {
	MaxAttempts    int `json:"maxAttempts"`
	InitialTimeout int `json:"initialTimeout"`
	MaxTimeout     int `json:"maxTimeout"`
}

type AuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Load reads the configuration file and returns a new Config instance
func Load(path string) (*Config, error) {
	cfg := &Config{path: path}
	if err := cfg.Reload(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Reload reloads the configuration from disk
func (c *Config) Reload() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}

	// Create a temporary config to parse into
	var temp Config
	if err := json.Unmarshal(file, &temp); err != nil {
		return err
	}

	// Copy the loaded values to the current config
	c.Worker = temp.Worker
	c.Auth = temp.Auth

	return nil
}

func (c *Config) GetAuthConfig() AuthConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Auth
}

func (c *Config) GetWorkerConfig() WorkerConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Worker
}
