package stores

import (
	"strings"
	"testing"
	"time"

	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/soy"

	"github.com/zoobz-io/janus/database/models"
)

// renderQuery builds a DB-less soy query for the table, applies the given filter,
// and returns the rendered SQL. Rendering never touches the database, so a nil
// connection is fine — this lets the WHERE-assembly be unit-tested without one.
func renderQuery[M any](t *testing.T, table string, apply func(*soy.Query[M]) *soy.Query[M]) string {
	t.Helper()
	s, err := soy.New[M](nil, table, astqlpg.New())
	if err != nil {
		t.Fatalf("soy.New: %v", err)
	}
	res, err := apply(s.Query()).Render()
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	return res.SQL
}

// timePtr returns a pointer to the first day of the given year/month (UTC).
func timePtr(year, month int) *time.Time {
	tm := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return &tm
}

func userQuery(t *testing.T) *soy.Query[models.User] {
	t.Helper()
	s, err := soy.New[models.User](nil, "users", astqlpg.New())
	if err != nil {
		t.Fatalf("soy.New: %v", err)
	}
	return s.Query()
}

func TestApplyTextSearch(t *testing.T) {
	// Empty query adds nothing.
	params := map[string]any{}
	sql := renderQuery(t, "users", func(q *soy.Query[models.User]) *soy.Query[models.User] {
		return applyTextSearch(q, "", params, "email", "display_name")
	})
	if _, ok := params["query"]; ok {
		t.Fatal("empty query should not bind a param")
	}
	if strings.Contains(sql, "ILIKE") {
		t.Fatalf("empty query should add no ILIKE: %s", sql)
	}

	// Non-empty query: metacharacters escaped, wrapped in %...%, ORed over fields.
	params = map[string]any{}
	sql = renderQuery(t, "users", func(q *soy.Query[models.User]) *soy.Query[models.User] {
		return applyTextSearch(q, "ac_me%", params, "email", "display_name")
	})
	if params["query"] != `%ac\_me\%%` {
		t.Fatalf("escaped param = %q, want %%ac\\_me\\%%%%", params["query"])
	}
	if strings.Count(sql, "ILIKE") != 2 || !strings.Contains(sql, "OR") {
		t.Fatalf("expected two ORed ILIKE clauses: %s", sql)
	}
}

func TestApplyStatusFacet(t *testing.T) {
	// Empty set adds nothing.
	params := map[string]any{}
	_ = applyStatusFacet(userQuery(t), nil, params)
	if _, ok := params["statuses"]; ok {
		t.Fatal("empty status set should not bind a param")
	}

	// Non-empty: bound as an array param, rendered = ANY(...).
	params = map[string]any{}
	sql := renderQuery(t, "users", func(q *soy.Query[models.User]) *soy.Query[models.User] {
		return applyStatusFacet(q, []string{"active", "inactive"}, params)
	})
	if _, ok := params["statuses"]; !ok {
		t.Fatal("status set should bind the statuses param")
	}
	if !strings.Contains(sql, `"status" = ANY(:statuses)`) {
		t.Fatalf("expected = ANY status clause: %s", sql)
	}
}

func TestApplyDateBounds(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	// Both bounds on created_at, only From on updated_at; a field absent from the
	// map contributes nothing.
	params := map[string]any{}
	sql := renderQuery(t, "users", func(q *soy.Query[models.User]) *soy.Query[models.User] {
		return applyDateBounds(q, map[string]DateBound{
			"created_at": {From: &from, To: &to},
			"updated_at": {From: &from},
		}, []string{"created_at", "updated_at"}, params)
	})
	if params["created_at_from"] != from || params["created_at_to"] != to || params["updated_at_from"] != from {
		t.Fatalf("date params = %+v", params)
	}
	if _, ok := params["updated_at_to"]; ok {
		t.Fatal("absent To bound should not bind a param")
	}
	if !strings.Contains(sql, `"created_at" >= :created_at_from`) || !strings.Contains(sql, `"created_at" <= :created_at_to`) {
		t.Fatalf("expected inclusive >=/<= bounds: %s", sql)
	}

	// A field with no entry in the dates map is skipped entirely.
	params = map[string]any{}
	sql = renderQuery(t, "users", func(q *soy.Query[models.User]) *soy.Query[models.User] {
		return applyDateBounds(q, map[string]DateBound{}, []string{"created_at"}, params)
	})
	if len(params) != 0 || strings.Contains(sql, ">=") {
		t.Fatalf("no bounds should add nothing: params=%v sql=%s", params, sql)
	}
}
