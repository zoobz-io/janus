package stores

import (
	"strings"
	"testing"

	"github.com/zoobz-io/soy"

	"github.com/zoobz-io/janus/database/models"
)

func TestEscapeLike(t *testing.T) {
	cases := map[string]string{
		"plain": "plain",
		`50%`:   `50\%`,
		`a_b`:   `a\_b`,
		`a\b`:   `a\\b`,
		`%_\`:   `\%\_\\`,
		"":      "",
	}
	for in, want := range cases {
		if got := escapeLike(in); got != want {
			t.Errorf("escapeLike(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestApplyApplicationSearch(t *testing.T) {
	params := map[string]any{}
	p := ApplicationSearchParams{
		Query:    "acme",
		Statuses: []string{"active"},
		Dates:    map[string]DateBound{"created_at": {From: timePtr(2026, 1), To: timePtr(2026, 6)}},
	}
	sql := renderQuery(t, "applications", func(q *soy.Query[models.Application]) *soy.Query[models.Application] {
		return applyApplicationSearch(q, p, params)
	})

	// Text search is over name/slug.
	if !strings.Contains(sql, `"name" ILIKE :query`) || !strings.Contains(sql, `"slug" ILIKE :query`) {
		t.Fatalf("expected name/slug ILIKE: %s", sql)
	}
	if !strings.Contains(sql, `"status" = ANY(:statuses)`) {
		t.Fatalf("expected status = ANY: %s", sql)
	}
	if !strings.Contains(sql, `"created_at" >= :created_at_from`) || !strings.Contains(sql, `"created_at" <= :created_at_to`) {
		t.Fatalf("expected date bounds: %s", sql)
	}
	if params["query"] != "%acme%" {
		t.Fatalf("query param = %q, want %%acme%%", params["query"])
	}
}
