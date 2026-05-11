// Package observe provides aperture schema synchronization from postgres via flux.
package observe

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// PollInterval is how often the watcher checks for schema changes.
const PollInterval = 30 * time.Second

// ConfigWatcher implements flux.Watcher by polling a postgres config table.
type ConfigWatcher struct {
	db  *sqlx.DB
	key string
}

// NewConfigWatcher creates a watcher that polls the config table for the given key.
func NewConfigWatcher(db *sqlx.DB, key string) *ConfigWatcher {
	return &ConfigWatcher{db: db, key: key}
}

// Watch returns a channel that emits the current config value immediately,
// then re-emits whenever the value changes.
func (w *ConfigWatcher) Watch(ctx context.Context) (<-chan []byte, error) {
	// Read initial value.
	initial, err := w.read(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading initial config for %q: %w", w.key, err)
	}

	ch := make(chan []byte, 1)
	ch <- initial

	go func() {
		defer close(ch)
		last := string(initial)
		ticker := time.NewTicker(PollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				current, err := w.read(ctx)
				if err != nil {
					continue
				}
				if string(current) != last {
					last = string(current)
					select {
					case ch <- current:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

func (w *ConfigWatcher) read(ctx context.Context) ([]byte, error) {
	var value string
	if err := w.db.GetContext(ctx, &value, "SELECT value FROM config WHERE key = $1", w.key); err != nil {
		return nil, err
	}
	return []byte(value), nil
}
