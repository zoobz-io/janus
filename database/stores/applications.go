package stores

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/soy"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/database/models"
)

// Applications provides database access for applications.
type Applications struct {
	*sum.Database[models.Application]
}

// NewApplications creates a new applications store.
func NewApplications(db *sqlx.DB, renderer astql.Renderer) *Applications {
	return &Applications{
		Database: sum.NewDatabase[models.Application](db, "applications", renderer),
	}
}

// GetApplication retrieves an application by ID.
func (s *Applications) GetApplication(ctx context.Context, id string) (*models.Application, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// GetBySlug retrieves an application by slug.
func (s *Applications) GetBySlug(ctx context.Context, slug string) (*models.Application, error) {
	return s.Select().
		Where("slug", "=", "slug").
		Exec(ctx, map[string]any{"slug": slug})
}

// ListApplications retrieves all active applications.
func (s *Applications) ListApplications(ctx context.Context) ([]*models.Application, error) {
	return s.Query().
		Where("status", "=", "status").
		OrderBy("name", "ASC").
		Exec(ctx, map[string]any{"status": models.ApplicationStatusActive})
}

// ListAll retrieves every application regardless of status (admin API).
func (s *Applications) ListAll(ctx context.Context) ([]*models.Application, error) {
	return s.Query().
		OrderBy("name", "ASC").
		Exec(ctx, nil)
}

// Update updates an application's name and status (admin API).
func (s *Applications) Update(ctx context.Context, id, name string, status models.ApplicationStatus) (*models.Application, error) {
	a, err := s.GetApplication(ctx, id)
	if err != nil {
		return nil, err
	}
	a.Name = name
	a.Status = status
	if err := s.Set(ctx, id, a); err != nil {
		return nil, fmt.Errorf("updating application: %w", err)
	}
	return a, nil
}

// applicationDateFields is the set of timestamp columns the search contract
// exposes for date filtering, in a fixed order for stable SQL generation.
var applicationDateFields = []string{"created_at", "updated_at"}

// searchFilterable is satisfied by both *soy.Query[T] and *soy.Aggregate[T]:
// their Where/WhereOr methods return the concrete builder type, so a
// self-referential type parameter lets one helper assemble the shared WHERE
// clause once and apply it to the page query, the count, and the facet query.
type searchFilterable[Self any] interface {
	Where(field, operator, param string) Self
	WhereOr(conditions ...soy.Condition) Self
}

// escapeLike escapes LIKE/ILIKE metacharacters so user input is matched
// literally. Postgres ILIKE treats backslash as the default escape character;
// the backslash replacement is listed first so NewReplacer's single left-to-right
// pass does not re-escape the backslashes it inserts.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// applyApplicationSearch adds the shared search WHERE clause to a query or
// aggregate builder and records the bound parameters in params. Text search ORs
// an escaped infix ILIKE across name/slug; the status facet is an OR-set via
// IN (rendered as = ANY); date bounds are inclusive >=/<=, each optional.
func applyApplicationSearch[B searchFilterable[B]](b B, p ApplicationSearchParams, params map[string]any) B {
	b = applyTextSearch(b, p.Query, params, "name", "slug")
	b = applyStatusFacet(b, p.Statuses, params)
	return applyDateBounds(b, p.Dates, applicationDateFields, params)
}

// Search runs the admin search contract over applications: one filtered page,
// the total count across the full filtered set, and the distinct status values
// present in that set. All three share the same WHERE assembly.
func (s *Applications) Search(ctx context.Context, p ApplicationSearchParams) (*ApplicationSearchResult, error) {
	params := map[string]any{}
	items, total, statuses, err := runSearch(ctx, params,
		func() *soy.Query[models.Application] { return applyApplicationSearch(s.Query(), p, params) },
		applyApplicationSearch(s.Count(), p, params),
		p.Sort, p.Page.Offset, p.Page.Limit,
		"status", func(a *models.Application) string { return a.Status })
	if err != nil {
		return nil, err
	}
	return &ApplicationSearchResult{Items: items, TotalItems: total, Statuses: statuses}, nil
}

// CreateApplication creates a new application.
func (s *Applications) CreateApplication(ctx context.Context, name, slug string) (*models.Application, error) {
	a := &models.Application{
		ID:     uuid.New().String(),
		Name:   name,
		Slug:   slug,
		Status: models.ApplicationStatusActive,
	}
	if err := s.Set(ctx, "", a); err != nil {
		return nil, fmt.Errorf("creating application: %w", err)
	}
	return a, nil
}
