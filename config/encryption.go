package config

import "errors"

// Encryption holds encryption key configuration.
type Encryption struct {
	Key string `env:"APP_ENCRYPTION_KEY" secret:"app/encryption-key"`
}

// Validate checks that the configuration is valid.
func (c Encryption) Validate() error {
	if c.Key == "" {
		return errors.New("encryption key is required")
	}
	if len(c.Key) != 64 {
		return errors.New("encryption key must be 64 hex characters (32 bytes)")
	}
	return nil
}
