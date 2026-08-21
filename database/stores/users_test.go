package stores

import (
	"strings"
	"testing"

	"github.com/zoobz-io/soy"

	"github.com/zoobz-io/janus/database/models"
)

func TestApplyUserSearch(t *testing.T) {
	params := map[string]any{}
	p := UserSearchParams{
		Query:    "jane",
		Statuses: []string{"active", "inactive"},
		Dates:    map[string]DateBound{"updated_at": {From: timePtr(2026, 1)}},
	}
	sql := renderQuery(t, "users", func(q *soy.Query[models.User]) *soy.Query[models.User] {
		return applyUserSearch(q, p, params)
	})

	// Text search is over email/display_name.
	if !strings.Contains(sql, `"email" ILIKE :query`) || !strings.Contains(sql, `"display_name" ILIKE :query`) {
		t.Fatalf("expected email/display_name ILIKE: %s", sql)
	}
	if !strings.Contains(sql, `"status" = ANY(:statuses)`) {
		t.Fatalf("expected status = ANY: %s", sql)
	}
	if !strings.Contains(sql, `"updated_at" >= :updated_at_from`) {
		t.Fatalf("expected updated_at lower bound: %s", sql)
	}
	if strings.Contains(sql, "<=") {
		t.Fatalf("no upper bound was set, should be absent: %s", sql)
	}
	if params["query"] != "%jane%" {
		t.Fatalf("query param = %q, want %%jane%%", params["query"])
	}
}
