// Package stores provides data access layer implementations for all janus domain entities.
package stores

import "github.com/zoobz-io/grub"

// ErrNotFound is returned when a queried entity does not exist.
var ErrNotFound = grub.ErrNotFound
