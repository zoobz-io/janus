// Package config provides typed configuration structs loaded from environment variables.
package config

import "errors"

// App holds application-level configuration.
type App struct {
	Port int `env:"APP_PORT" default:"8080"`
}

// Validate checks that the configuration is valid.
func (c App) Validate() error {
	if c.Port <= 0 {
		return errors.New("port must be positive")
	}
	return nil
}
