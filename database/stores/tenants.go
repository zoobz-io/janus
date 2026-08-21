package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/soy"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/database/models"
)

// tenantDateFields is the set of timestamp columns the tenant search contract
// exposes for date filtering, in a fixed order for stable SQL generation.
var tenantDateFields = []string{"created_at", "updated_at"}

// applyTenantSearch adds the shared tenant-search WHERE clause: escaped infix
// ILIKE over name/slug, the status facet as an IN set, and inclusive >=/<= date
// bounds.
func applyTenantSearch[B searchFilterable[B]](b B, p TenantSearchParams, params map[string]any) B {
	b = applyTextSearch(b, p.Query, params, "name", "slug")
	b = applyStatusFacet(b, p.Statuses, params)
	return applyDateBounds(b, p.Dates, tenantDateFields, params)
}

// Search runs the admin search over tenants: one filtered page, the total count,
// and the distinct status values present in that set. All three share one WHERE.
func (s *Tenants) Search(ctx context.Context, p TenantSearchParams) (*TenantSearchResult, error) {
	params := map[string]any{}
	items, total, statuses, err := runSearch(ctx, params,
		func() *soy.Query[models.Tenant] { return applyTenantSearch(s.Query(), p, params) },
		applyTenantSearch(s.Count(), p, params),
		p.Sort, p.Page.Offset, p.Page.Limit,
		"status", func(t *models.Tenant) string { return t.Status })
	if err != nil {
		return nil, err
	}
	return &TenantSearchResult{Items: items, TotalItems: total, Statuses: statuses}, nil
}

// Tenants provides database access for tenants.
type Tenants struct {
	*sum.Database[models.Tenant]
}

// NewTenants creates a new tenants store.
func NewTenants(db *sqlx.DB, renderer astql.Renderer) *Tenants {
	return &Tenants{
		Database: sum.NewDatabase[models.Tenant](db, "tenants", renderer),
	}
}

// GetTenant retrieves a tenant by ID.
func (s *Tenants) GetTenant(ctx context.Context, id string) (*models.Tenant, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// CreateTenant creates a new tenant.
func (s *Tenants) CreateTenant(ctx context.Context, name, slug string) (*models.Tenant, error) {
	t := &models.Tenant{
		ID:     uuid.New().String(),
		Name:   name,
		Slug:   slug,
		Status: models.TenantStatusActive,
	}
	if err := s.Set(ctx, "", t); err != nil {
		return nil, fmt.Errorf("creating tenant: %w", err)
	}
	return t, nil
}

// CreateTenantTx creates a new tenant inside an existing transaction.
func (s *Tenants) CreateTenantTx(ctx context.Context, tx *sqlx.Tx, name, slug string) (*models.Tenant, error) {
	t := &models.Tenant{
		ID:     uuid.New().String(),
		Name:   name,
		Slug:   slug,
		Status: models.TenantStatusActive,
	}
	if err := s.SetTx(ctx, tx, "", t); err != nil {
		return nil, fmt.Errorf("creating tenant: %w", err)
	}
	return t, nil
}

// UpdateTenant updates an existing tenant.
func (s *Tenants) UpdateTenant(ctx context.Context, id, name string, status models.TenantStatus) (*models.Tenant, error) {
	t, err := s.GetTenant(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Name = name
	t.Status = status
	if err := s.Set(ctx, id, t); err != nil {
		return nil, fmt.Errorf("updating tenant: %w", err)
	}
	return t, nil
}

// ListTenants retrieves a paginated list of tenants.
func (s *Tenants) ListTenants(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Tenant], error) {
	items, err := s.Query().
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, nil)
	if err != nil {
		return nil, err
	}
	total, countErr := s.Count().Exec(ctx, nil)
	if countErr != nil {
		return nil, countErr
	}
	return &models.OffsetResult[models.Tenant]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}
