package config

import "errors"

// OTEL holds OpenTelemetry exporter configuration.
type OTEL struct {
	Endpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4318"`
}

// Validate checks that the configuration is valid.
func (c OTEL) Validate() error {
	if c.Endpoint == "" {
		return errors.New("otel endpoint is required")
	}
	return nil
}
