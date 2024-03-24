package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type Config struct {
	APIServerPort int    `envconfig:"PORT" default:"5000"`
	RedisHostPort string `envconfig:"REDIS_HOSTPORT" default:"redis:6379"`
	RedisPrefix   string `envconfig:"REDIS_PREFIX" default:""`
}

func New() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to build config from env")
	}
	return &cfg, nil
}
