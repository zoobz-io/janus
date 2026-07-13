package observe

import (
	"context"
	"fmt"

	"github.com/zoobz-io/aperture"
	"github.com/zoobz-io/flux"
)

// StartSchemaSync creates a flux capacitor that watches the aperture schema
// in the config service and applies changes to the aperture instance. The
// capacitor blocks until the initial schema is loaded.
func StartSchemaSync(ctx context.Context, ap *aperture.Aperture) (*flux.Capacitor[aperture.Schema], error) {
	watcher := NewConfigWatcher("aperture_schema")

	capacitor := flux.New[aperture.Schema](watcher, func(_ context.Context, _, curr aperture.Schema) error {
		return ap.Apply(curr)
	}).Codec(flux.YAMLCodec{})

	if err := capacitor.Start(ctx); err != nil {
		return nil, fmt.Errorf("starting aperture schema sync: %w", err)
	}

	return capacitor, nil
}
