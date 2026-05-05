package config

import (
	"errors"
	"fmt"
)

// Mesh holds aegis mesh node configuration.
type Mesh struct {
	ID      string `env:"APP_MESH_ID" default:"janus"`
	Name    string `env:"APP_MESH_NAME" default:"janus"`
	Host    string `env:"APP_MESH_HOST" default:"0.0.0.0"`
	CertDir string `env:"APP_MESH_CERT_DIR" default:"./certs"`
	Port    int    `env:"APP_MESH_PORT" default:"9090"`
}

// Addr returns the host:port address for the mesh node.
func (c Mesh) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Validate checks that the configuration is valid.
func (c Mesh) Validate() error {
	if c.ID == "" {
		return errors.New("mesh node ID is required")
	}
	if c.Name == "" {
		return errors.New("mesh node name is required")
	}
	if c.Port <= 0 {
		return errors.New("mesh port must be positive")
	}
	return nil
}
