package stores

import (
	"strings"
	"testing"

	"github.com/zoobz-io/soy"

	"github.com/zoobz-io/janus/database/models"
)

func TestApplyTenantSearch(t *testing.T) {
	params := map[string]any{}
	p := TenantSearchParams{
		Query:    "acme",
		Statuses: []string{"active", "suspended"},
		Dates:    map[string]DateBound{"created_at": {From: timePtr(2026, 1)}},
	}
	sql := renderQuery(t, "tenants", func(q *soy.Query[models.Tenant]) *soy.Query[models.Tenant] {
		return applyTenantSearch(q, p, params)
	})

	// Text search is over name/slug.
	if !strings.Contains(sql, `"name" ILIKE :query`) || !strings.Contains(sql, `"slug" ILIKE :query`) {
		t.Fatalf("expected name/slug ILIKE: %s", sql)
	}
	if !strings.Contains(sql, `"status" = ANY(:statuses)`) {
		t.Fatalf("expected status = ANY: %s", sql)
	}
	if !strings.Contains(sql, `"created_at" >= :created_at_from`) {
		t.Fatalf("expected created_at lower bound: %s", sql)
	}
	if params["query"] != "%acme%" {
		t.Fatalf("query param = %q, want %%acme%%", params["query"])
	}
}
