package models

import "time"

// Config is a key/value runtime configuration row. It backs operational
// settings that can change without a redeploy — e.g. the aperture schema
// polled by internal/observe.
type Config struct {
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" default:"now()"`
	Key       string    `json:"key" db:"key" constraints:"primarykey"`
	Value     string    `json:"value" db:"value" constraints:"notnull"`
}

// GetID returns the config's primary key.
func (c Config) GetID() string {
	return c.Key
}

// GetCreatedAt returns the last update timestamp.
func (c Config) GetCreatedAt() time.Time {
	return c.UpdatedAt
}

// Clone returns a copy of the config.
func (c Config) Clone() Config {
	return c
}
