package stores

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
// IN (rendered as = ANY); date bounds are inclusive GE/LE, each optional.
func applyApplicationSearch[B searchFilterable[B]](b B, p ApplicationSearchParams, params map[string]any) B {
	if p.Query != "" {
		params["query"] = "%" + escapeLike(p.Query) + "%"
		b = b.WhereOr(
			soy.C("name", "ILIKE", "query"),
			soy.C("slug", "ILIKE", "query"),
		)
	}
	if len(p.Statuses) > 0 {
		params["statuses"] = pq.Array(p.Statuses)
		b = b.Where("status", "IN", "statuses")
	}
	for _, field := range applicationDateFields {
		bound, ok := p.Dates[field]
		if !ok {
			continue
		}
		if bound.From != nil {
			param := field + "_from"
			params[param] = *bound.From
			b = b.Where(field, "GE", param)
		}
		if bound.To != nil {
			param := field + "_to"
			params[param] = *bound.To
			b = b.Where(field, "LE", param)
		}
	}
	return b
}

// Search runs the admin search contract over applications: one filtered page,
// the total count across the full filtered set, and the distinct status values
// present in that set. All three share the same WHERE assembly.
func (s *Applications) Search(ctx context.Context, p ApplicationSearchParams) (*ApplicationSearchResult, error) {
	params := map[string]any{}

	// Page: filtered, sorted with an id tiebreak for stable paging, windowed.
	items, err := applyApplicationSearch(s.Query(), p, params).
		OrderBy(p.Sort.Field, p.Sort.Order).
		OrderBy("id", SortAsc).
		Limit(p.Page.Limit).
		Offset(p.Page.Offset).
		Exec(ctx, params)
	if err != nil {
		return nil, err
	}

	// Total across the full filtered set (same WHERE, no window).
	total, err := applyApplicationSearch(s.Count(), p, params).Exec(ctx, params)
	if err != nil {
		return nil, err
	}

	// Facet values: distinct status present in the full filtered set.
	statusRows, err := applyApplicationSearch(s.Query().Fields("status").Distinct(), p, params).Exec(ctx, params)
	if err != nil {
		return nil, err
	}
	statuses := make([]string, 0, len(statusRows))
	for _, row := range statusRows {
		statuses = append(statuses, row.Status)
	}
	sort.Strings(statuses)

	return &ApplicationSearchResult{
		Items:      items,
		TotalItems: int64(total),
		Statuses:   statuses,
	}, nil
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
