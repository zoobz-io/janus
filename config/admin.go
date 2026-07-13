package config

import "errors"

// Admin holds configuration for the admin HTTP API binary.
type Admin struct {
	Port int `env:"APP_ADMIN_PORT" default:"8081"`
}

// Validate checks that the configuration is valid.
func (c Admin) Validate() error {
	if c.Port <= 0 {
		return errors.New("admin port must be positive")
	}
	return nil
}
