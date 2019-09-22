package proxy

import "time"

type Config struct {
	cacheSize  int
	cacheTTL   time.Duration
	listenPort int
}

type Option func(*Config)

func WithCacheSize(size int) Option {
	return func(c *Config) {
		c.cacheSize = size
	}
}

func WithCacheTTL(duration time.Duration) Option {
	return func(c *Config) {
		c.cacheTTL = duration
	}
}

func ListenOn(portNum int) Option {
	return func(c *Config) {
		c.listenPort = portNum
	}
}

func defaultConfig() *Config {
	return &Config{
		cacheSize:  1e6,
		cacheTTL:   time.Minute,
		listenPort: 24689,
	}
}
