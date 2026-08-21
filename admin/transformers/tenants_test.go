package transformers

import (
	"testing"
	"time"

	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

func TestResolveTenantSearchDefaults(t *testing.T) {
	params, number, size := ResolveTenantSearch(wire.SearchTenantsRequest{})
	if number != 1 || size != 25 {
		t.Fatalf("defaults number/size = %d/%d, want 1/25", number, size)
	}
	if params.Sort.Field != "updated_at" || params.Sort.Order != stores.SortDesc {
		t.Fatalf("default sort = %+v", params.Sort)
	}
	if params.Query != "" || params.Statuses != nil || params.Dates != nil {
		t.Fatalf("expected empty filters, got %+v", params)
	}
}

func TestResolveTenantSearchMapping(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	body := wire.SearchTenantsRequest{
		Query:  "acme",
		Facets: map[string][]string{"status": {"suspended"}},
		Dates:  map[string]wire.DateRange{"updated_at": {From: &from}},
		Sort:   &wire.SortSpec{Field: "created_at", Order: "asc"},
		Page:   &wire.PageRequest{Number: intPtr(3), Size: intPtr(10)},
	}
	params, number, size := ResolveTenantSearch(body)
	if number != 3 || size != 10 || params.Page.Offset != 20 {
		t.Fatalf("page = %d/%d offset=%d, want 3/10 offset 20", number, size, params.Page.Offset)
	}
	if params.Query != "acme" || len(params.Statuses) != 1 || params.Statuses[0] != "suspended" {
		t.Fatalf("filters = %+v", params)
	}
	if params.Sort.Field != "created_at" || params.Sort.Order != stores.SortAsc {
		t.Fatalf("sort = %+v", params.Sort)
	}
}

func TestTenantSearchToResponse(t *testing.T) {
	result := &stores.TenantSearchResult{
		Items: []*models.Tenant{
			{ID: "t1", Name: "Acme", Slug: "acme", Status: "active"},
			{ID: "t2", Name: "Globex", Slug: "globex", Status: "suspended"},
		},
		Statuses:   []string{"active", "suspended"},
		TotalItems: 2,
	}
	resp := TenantSearchToResponse(result, 1, 25)
	if len(resp.Tenants) != 2 || resp.Tenants[0].Slug != "acme" {
		t.Fatalf("tenants = %+v", resp.Tenants)
	}
	if resp.Page.TotalItems != 2 || resp.Page.TotalPages != 1 {
		t.Fatalf("page = %+v", resp.Page)
	}
	if len(resp.Facets["status"]) != 2 {
		t.Fatalf("facet = %v", resp.Facets["status"])
	}
}
