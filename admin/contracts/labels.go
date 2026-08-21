package contracts

import "context"

// ApplicationLabels is the admin API's boundary over the application label
// mapping: outbound id->name resolution for responses, inbound name->id for
// filtering. Backed by the Redis mapping, never by SQL joins.
type ApplicationLabels interface {
	// ResolveNames maps application ids to their current display names in one
	// batch. Unknown ids are omitted from the result.
	ResolveNames(ctx context.Context, ids []string) (map[string]string, error)
	// ResolveID maps an application name (label) to its id; ok is false when the
	// name is unknown.
	ResolveID(ctx context.Context, name string) (id string, ok bool, err error)
}
