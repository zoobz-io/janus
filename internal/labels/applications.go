package labels

import (
	"context"
	"errors"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// applicationSource is the read side of the applications store the label cache
// needs: single lookups for read-repair and a full scan for reconciliation.
// *stores.Applications satisfies it; tests inject a fake.
type applicationSource interface {
	GetApplication(ctx context.Context, id string) (*models.Application, error)
	ListAll(ctx context.Context) ([]*models.Application, error)
}

// labelEntry is the stored value for a mapping. sum.NewStore requires a struct
// value type — atomization rejects bare strings — so the string is wrapped.
type labelEntry struct {
	Value string `json:"value"`
}

// Key prefixes for the two mapping directions in Redis.
const (
	appIDPrefix   = "app:id:"   // app:id:<id>     -> current name
	appNamePrefix = "app:name:" // app:name:<name> -> id
)

// ApplicationLabels maintains the bidirectional application id<->name mapping in
// Redis. id->name is always current (overwritten on rename); name->id is
// append-only across renames (we only ever Set, never Delete) so legacy names
// keep resolving, while a name reused by a new application repoints to the new id.
type ApplicationLabels struct {
	store *sum.Store[labelEntry]
	apps  applicationSource
}

// NewApplicationLabels builds the label store over the shared Redis client,
// with the applications store as the read-repair source on cache misses.
func NewApplicationLabels(client *goredis.Client, apps *stores.Applications) *ApplicationLabels {
	return newApplicationLabels(newRedisProvider(client), apps, "application-labels")
}

// newApplicationLabels is the injectable constructor: it takes the raw KV
// provider and applications source so tests can drive the mapping logic against
// an in-memory provider and a fake source. name is the catalog store name.
func newApplicationLabels(provider grub.StoreProvider, apps applicationSource, name string) *ApplicationLabels {
	return &ApplicationLabels{
		store: sum.NewStore[labelEntry](provider, name),
		apps:  apps,
	}
}

// Put writes both mapping directions for one application. Mappings never expire.
func (l *ApplicationLabels) Put(ctx context.Context, id, name string) error {
	if err := l.store.Set(ctx, appIDPrefix+id, &labelEntry{Value: name}, 0); err != nil {
		return err
	}
	return l.store.Set(ctx, appNamePrefix+name, &labelEntry{Value: id}, 0)
}

// ResolveID maps an application name (label) to its id for inbound filtering.
// The bool is false when the name is unknown.
func (l *ApplicationLabels) ResolveID(ctx context.Context, name string) (string, bool, error) {
	entry, err := l.store.Get(ctx, appNamePrefix+name)
	if errors.Is(err, grub.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return entry.Value, true, nil
}

// ResolveNames maps application ids to their current names for outbound
// responses, in one batch. Cache misses are read-repaired: looked up in the
// applications store and written back. Ids that resolve to no application are
// omitted from the result.
func (l *ApplicationLabels) ResolveNames(ctx context.Context, ids []string) (map[string]string, error) {
	names := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return names, nil
	}

	seen := make(map[string]struct{}, len(ids))
	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		keys = append(keys, appIDPrefix+id)
	}

	found, err := l.store.GetBatch(ctx, keys)
	if err != nil {
		return nil, err
	}

	for id := range seen {
		if entry, ok := found[appIDPrefix+id]; ok && entry != nil {
			names[id] = entry.Value
			continue
		}
		// Read-repair: miss -> load from the applications store and heal the cache.
		app, lookupErr := l.apps.GetApplication(ctx, id)
		if lookupErr != nil {
			continue // unknown id -> omit (label resolution is best-effort)
		}
		names[id] = app.Name
		if err := l.Put(ctx, app.ID, app.Name); err != nil {
			return nil, err
		}
	}
	return names, nil
}

// Reconcile upserts the mapping for every application. Idempotent; run on boot
// to cover missed events, cold start, and out-of-band changes.
func (l *ApplicationLabels) Reconcile(ctx context.Context) error {
	apps, err := l.apps.ListAll(ctx)
	if err != nil {
		return err
	}
	for _, app := range apps {
		if err := l.Put(ctx, app.ID, app.Name); err != nil {
			return err
		}
	}
	return nil
}
