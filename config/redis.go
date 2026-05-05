package config

import "errors"

// Redis holds Redis connection configuration.
type Redis struct {
	Addr string `env:"APP_REDIS_ADDR" default:"localhost:6379"`
}

// Validate checks that the configuration is valid.
func (c Redis) Validate() error {
	if c.Addr == "" {
		return errors.New("redis address is required")
	}
	return nil
}
