package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/soy"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/database/models"
)

// Scopes provides database access for application-defined scopes.
type Scopes struct {
	*sum.Database[models.Scope]
}

// NewScopes creates a new scopes store.
func NewScopes(db *sqlx.DB, renderer astql.Renderer) *Scopes {
	return &Scopes{
		Database: sum.NewDatabase[models.Scope](db, "scopes", renderer),
	}
}

// Define creates a new scope for an application.
func (s *Scopes) Define(ctx context.Context, applicationID, name, description string) (*models.Scope, error) {
	sc := &models.Scope{
		ID:            uuid.New().String(),
		ApplicationID: applicationID,
		Name:          name,
		Description:   description,
	}
	if err := s.Set(ctx, "", sc); err != nil {
		return nil, fmt.Errorf("defining scope: %w", err)
	}
	return sc, nil
}

// GetByID retrieves a scope by ID. Returns (nil, nil) if absent.
func (s *Scopes) GetByID(ctx context.Context, id string) (*models.Scope, error) {
	results, err := s.Query().
		Where("id", "=", "id").
		Limit(1).
		Exec(ctx, map[string]any{"id": id})
	if err != nil {
		return nil, fmt.Errorf("looking up scope: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// GetByName retrieves a scope by application and name. Returns (nil, nil) if absent.
func (s *Scopes) GetByName(ctx context.Context, applicationID, name string) (*models.Scope, error) {
	results, err := s.Query().
		Where("application_id", "=", "application_id").
		Where("name", "=", "name").
		Limit(1).
		Exec(ctx, map[string]any{"application_id": applicationID, "name": name})
	if err != nil {
		return nil, fmt.Errorf("looking up scope by name: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// ListByApplication retrieves all scopes defined by an application.
func (s *Scopes) ListByApplication(ctx context.Context, applicationID string) ([]*models.Scope, error) {
	return s.Query().
		Where("application_id", "=", "application_id").
		OrderBy("name", "ASC").
		Exec(ctx, map[string]any{"application_id": applicationID})
}

// scopeDateFields is the set of timestamp columns the scope search contract
// exposes for date filtering, in a fixed order for stable SQL generation.
var scopeDateFields = []string{"created_at", "updated_at"}

// applyScopeSearch adds the shared scope-search WHERE clause to a query or
// aggregate builder (see searchFilterable). Text search ORs an escaped infix
// ILIKE across name/description; the application facet is an id OR-set via IN
// (rendered as = ANY); date bounds are inclusive >=/<=, each optional.
func applyScopeSearch[B searchFilterable[B]](b B, p ScopeSearchParams, params map[string]any) B {
	if p.Query != "" {
		params["query"] = "%" + escapeLike(p.Query) + "%"
		b = b.WhereOr(
			soy.C("name", "ILIKE", "query"),
			soy.C("description", "ILIKE", "query"),
		)
	}
	// A non-nil ApplicationIDs (even empty) means "filter to these ids"; an empty
	// set renders = ANY('{}') and correctly matches nothing.
	if p.ApplicationIDs != nil {
		params["app_ids"] = pq.Array(p.ApplicationIDs)
		b = b.Where("application_id", "IN", "app_ids")
	}
	for _, field := range scopeDateFields {
		bound, ok := p.Dates[field]
		if !ok {
			continue
		}
		if bound.From != nil {
			param := field + "_from"
			params[param] = *bound.From
			b = b.Where(field, ">=", param)
		}
		if bound.To != nil {
			param := field + "_to"
			params[param] = *bound.To
			b = b.Where(field, "<=", param)
		}
	}
	return b
}

// ListAll retrieves every scope across all applications (admin API).
func (s *Scopes) ListAll(ctx context.Context) ([]*models.Scope, error) {
	return s.Query().
		OrderBy("name", "ASC").
		OrderBy("id", "ASC").
		Exec(ctx, nil)
}

// Search runs the admin cross-application search over scopes: one filtered page,
// the total count across the full filtered set, and the distinct application ids
// present in that set (the application facet). All three share one WHERE assembly.
func (s *Scopes) Search(ctx context.Context, p ScopeSearchParams) (*ScopeSearchResult, error) {
	params := map[string]any{}
	// Facet is application_id; the transformer resolves the ids to names.
	items, total, appIDs, err := runSearch(ctx, params,
		func() *soy.Query[models.Scope] { return applyScopeSearch(s.Query(), p, params) },
		applyScopeSearch(s.Count(), p, params),
		p.Sort, p.Page.Offset, p.Page.Limit,
		"application_id", func(sc *models.Scope) string { return sc.ApplicationID })
	if err != nil {
		return nil, err
	}
	return &ScopeSearchResult{Items: items, TotalItems: total, ApplicationIDs: appIDs}, nil
}

// Delete removes a scope by ID. Features referencing it are cascaded by the database.
func (s *Scopes) Delete(ctx context.Context, id string) error {
	if err := s.Database.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting scope: %w", err)
	}
	return nil
}
