//go:build testing

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

func tenantParams() stores.TenantSearchParams {
	return stores.TenantSearchParams{
		Sort: stores.SearchSort{Field: "updated_at", Order: stores.SortDesc},
		Page: stores.SearchPage{Offset: 0, Limit: 25},
	}
}

func setTenantTimestamps(t *testing.T, id string, created, updated time.Time) {
	t.Helper()
	if _, err := testDB.Exec(
		`UPDATE tenants SET created_at = $1, updated_at = $2 WHERE id = $3`, created, updated, id,
	); err != nil {
		t.Fatalf("setTenantTimestamps: %v", err)
	}
}

func TestTenantSearch(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	acme, _ := testStores.Tenants.CreateTenant(ctx, "Acme Corp", "acme-corp")
	globex, _ := testStores.Tenants.CreateTenant(ctx, "Globex", "globex")
	initech, _ := testStores.Tenants.CreateTenant(ctx, "Initech", "initech")
	// Two suspended.
	testStores.Tenants.UpdateTenant(ctx, globex.ID, "Globex", models.TenantStatusSuspended)
	testStores.Tenants.UpdateTenant(ctx, initech.ID, "Initech", models.TenantStatusSuspended)

	t.Run("EmptyMatchesAll", func(t *testing.T) {
		res, err := testStores.Tenants.Search(ctx, tenantParams())
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 3 {
			t.Fatalf("total = %d, want 3", res.TotalItems)
		}
	})

	t.Run("TextMatchesNameFragment", func(t *testing.T) {
		p := tenantParams()
		p.Query = "acme"
		res, err := testStores.Tenants.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 1 || (len(res.Items) == 1 && res.Items[0].ID != acme.ID) {
			t.Fatalf("name fragment total = %d, want only acme", res.TotalItems)
		}
	})

	t.Run("TextMatchesSlugFragment", func(t *testing.T) {
		p := tenantParams()
		p.Query = "-corp" // only in acme's slug "acme-corp"
		res, err := testStores.Tenants.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 1 || (len(res.Items) == 1 && res.Items[0].ID != acme.ID) {
			t.Fatalf("slug fragment total = %d, want only acme", res.TotalItems)
		}
	})

	t.Run("StatusFacet", func(t *testing.T) {
		p := tenantParams()
		p.Statuses = []string{"suspended"}
		res, err := testStores.Tenants.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 2 {
			t.Fatalf("suspended total = %d, want 2", res.TotalItems)
		}
	})

	t.Run("FacetValues", func(t *testing.T) {
		res, err := testStores.Tenants.Search(ctx, tenantParams())
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(res.Statuses) != 2 || res.Statuses[0] != "active" || res.Statuses[1] != "suspended" {
			t.Fatalf("facet statuses = %v, want [active suspended]", res.Statuses)
		}
	})
}

func TestTenantSearchLikeEscaping(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	testStores.Tenants.CreateTenant(ctx, "Save 50% Co", "save-half")
	testStores.Tenants.CreateTenant(ctx, "Save 500 Co", "save-500")

	p := tenantParams()
	p.Query = "50%"
	res, err := testStores.Tenants.Search(ctx, p)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if res.TotalItems != 1 {
		t.Fatalf("query %q matched %d, want 1 (escaping failed)", p.Query, res.TotalItems)
	}
}

func TestTenantSearchDateBoundsAndPaging(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	ids := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		tn, err := testStores.Tenants.CreateTenant(ctx, "Tenant", "tenant-"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		ts := time.Date(2026, time.January, 1+i, 0, 0, 0, 0, time.UTC)
		setTenantTimestamps(t, tn.ID, ts, ts)
		ids = append(ids, tn.ID)
	}

	t.Run("DateWindow", func(t *testing.T) {
		from := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, time.January, 4, 0, 0, 0, 0, time.UTC)
		p := tenantParams()
		p.Dates = map[string]stores.DateBound{"created_at": {From: &from, To: &to}}
		res, err := testStores.Tenants.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 3 {
			t.Fatalf("window total = %d, want 3", res.TotalItems)
		}
	})

	t.Run("SecondPageAndCount", func(t *testing.T) {
		p := tenantParams()
		p.Sort = stores.SearchSort{Field: "created_at", Order: stores.SortAsc}
		p.Page = stores.SearchPage{Offset: 2, Limit: 2}
		res, err := testStores.Tenants.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 5 {
			t.Fatalf("total = %d, want 5 (count ignores window)", res.TotalItems)
		}
		if len(res.Items) != 2 || res.Items[0].ID != ids[2] || res.Items[1].ID != ids[3] {
			t.Fatalf("unexpected page contents")
		}
	})

	t.Run("OutOfRangePageEmptyButCounted", func(t *testing.T) {
		p := tenantParams()
		p.Page = stores.SearchPage{Offset: 100, Limit: 25}
		res, err := testStores.Tenants.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 5 || len(res.Items) != 0 {
			t.Fatalf("out-of-range total=%d items=%d, want 5/0", res.TotalItems, len(res.Items))
		}
	})
}
