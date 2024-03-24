package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// Config is the configuration for the worker
type Config struct {
	RedisHostPort string `envconfig:"REDIS_HOSTPORT" default:"redis:6379"`
	RedisPrefix   string `envconfig:"REDIS_PREFIX" default:""`

	// Fetcher
	UseFetcher          string  `envconfig:"USE_FETCHER" default:"mock"`
	FetchTimeoutSeconds float32 `envconfig:"FETCH_TIMEOUT_SECONDS" default:"4"`

	// RateLimiting
	SleepoverDurationSeconds int `envconfig:"SLEEPOVER_DURATION_SECONDS" default:"4"`

	// Mock API
	MockFetcherAvgRequestSeconds float32 `envconfig:"MOCK_FETCHER_AVG_REQUEST_SECONDS" default:"2.5"`
	MockRateLimit                int     `envconfig:"MOCK_RATE_LIMIT" default:"20"`
	MockRateLimitWindowSeconds   int     `envconfig:"MOCK_RATE_LIMIT_WINDOW_SECONDS" default:"60"`
}

// New creates a new Config from the environment
func New() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to build config from env")
	}
	return &cfg, nil
}
