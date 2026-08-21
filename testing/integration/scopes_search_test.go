//go:build testing

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

func scopeParams() stores.ScopeSearchParams {
	return stores.ScopeSearchParams{
		Sort: stores.SearchSort{Field: "updated_at", Order: stores.SortDesc},
		Page: stores.SearchPage{Offset: 0, Limit: 25},
	}
}

func setScopeTimestamps(t *testing.T, id string, created, updated time.Time) {
	t.Helper()
	if _, err := testDB.Exec(
		`UPDATE scopes SET created_at = $1, updated_at = $2 WHERE id = $3`, created, updated, id,
	); err != nil {
		t.Fatalf("setScopeTimestamps: %v", err)
	}
}

func TestScopeSearchStore(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t); flushLabels(t) })

	nexus, _ := testStores.Applications.CreateApplication(ctx, "Nexus", "nexus")
	globex, _ := testStores.Applications.CreateApplication(ctx, "Globex", "globex")

	read, _ := testStores.Scopes.Define(ctx, nexus.ID, "projects:read", "Read projects")
	write, _ := testStores.Scopes.Define(ctx, nexus.ID, "projects:write", "Write projects")
	billing, _ := testStores.Scopes.Define(ctx, globex.ID, "billing:read", "Read invoices")
	_ = write

	t.Run("EmptyMatchesAll", func(t *testing.T) {
		res, err := testStores.Scopes.Search(ctx, scopeParams())
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 3 {
			t.Fatalf("total = %d, want 3", res.TotalItems)
		}
	})

	t.Run("TextMatchesNameAndDescription", func(t *testing.T) {
		p := scopeParams()
		p.Query = "read" // matches "projects:read" name + "Read projects"/"Read invoices" descriptions
		res, err := testStores.Scopes.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 2 { // read (name+desc), billing (desc "Read invoices")
			t.Fatalf("total = %d, want 2", res.TotalItems)
		}
	})

	t.Run("TextMatchesDescriptionOnly", func(t *testing.T) {
		p := scopeParams()
		p.Query = "invoices" // only in billing's description
		res, err := testStores.Scopes.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 1 || (len(res.Items) == 1 && res.Items[0].ID != billing.ID) {
			t.Fatalf("total = %d, want only billing", res.TotalItems)
		}
	})

	t.Run("ApplicationFilterByIDs", func(t *testing.T) {
		p := scopeParams()
		p.ApplicationIDs = []string{nexus.ID}
		res, err := testStores.Scopes.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 2 {
			t.Fatalf("nexus scopes = %d, want 2", res.TotalItems)
		}
	})

	t.Run("EmptyApplicationFilterMatchesNothing", func(t *testing.T) {
		p := scopeParams()
		p.ApplicationIDs = []string{} // non-nil empty: filter requested, no ids
		res, err := testStores.Scopes.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 0 || len(res.Items) != 0 {
			t.Fatalf("empty filter total = %d, want 0", res.TotalItems)
		}
	})

	t.Run("FacetListsDistinctApplicationIDs", func(t *testing.T) {
		res, err := testStores.Scopes.Search(ctx, scopeParams())
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(res.ApplicationIDs) != 2 {
			t.Fatalf("facet ids = %v, want 2 distinct", res.ApplicationIDs)
		}
	})

	_ = read
}

func TestScopeSearchLikeEscaping(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t); flushLabels(t) })
	app, _ := testStores.Applications.CreateApplication(ctx, "Disco", "disco")

	testStores.Scopes.Define(ctx, app.ID, "discount:50%", "50% off")
	testStores.Scopes.Define(ctx, app.ID, "discount:500", "flat 500")

	p := scopeParams()
	p.Query = "50%"
	res, err := testStores.Scopes.Search(ctx, p)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if res.TotalItems != 1 {
		t.Fatalf("query %q matched %d, want 1 (escaping failed)", p.Query, res.TotalItems)
	}
}

func TestScopeSearchNullDescriptionDropsFromTextMatch(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t); flushLabels(t) })
	app, _ := testStores.Applications.CreateApplication(ctx, "NullApp", "nullapp")

	// A scope with a NULL description and a name that does NOT contain the query.
	if _, err := testDB.Exec(
		`INSERT INTO scopes (id, application_id, name, description) VALUES ($1, $2, $3, NULL)`,
		"null-scope", app.ID, "alpha",
	); err != nil {
		t.Fatalf("insert null-desc scope: %v", err)
	}
	// A scope whose name matches the query.
	testStores.Scopes.Define(ctx, app.ID, "beta-read", "")

	p := scopeParams()
	p.Query = "read"
	res, err := testStores.Scopes.Search(ctx, p)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// Only beta-read matches; the NULL-description "alpha" cannot match text
	// (name FALSE OR description NULL = NULL, excluded). This is correct SQL.
	if res.TotalItems != 1 || (len(res.Items) == 1 && res.Items[0].Name != "beta-read") {
		t.Fatalf("null-desc row leaked into text match: total=%d", res.TotalItems)
	}

	// Unfiltered, the NULL-description row is present.
	all, _ := testStores.Scopes.Search(ctx, scopeParams())
	if all.TotalItems != 2 {
		t.Fatalf("unfiltered total = %d, want 2 (incl null-desc)", all.TotalItems)
	}
}

func TestScopeSearchDateBounds(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t); flushLabels(t) })
	app, _ := testStores.Applications.CreateApplication(ctx, "Timed", "timed")

	jan, _ := testStores.Scopes.Define(ctx, app.ID, "jan", "")
	jun, _ := testStores.Scopes.Define(ctx, app.ID, "jun", "")
	dec, _ := testStores.Scopes.Define(ctx, app.ID, "dec", "")
	ts := func(m time.Month) time.Time { return time.Date(2026, m, 15, 12, 0, 0, 0, time.UTC) }
	setScopeTimestamps(t, jan.ID, ts(time.January), ts(time.January))
	setScopeTimestamps(t, jun.ID, ts(time.June), ts(time.June))
	setScopeTimestamps(t, dec.ID, ts(time.December), ts(time.December))

	from := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	p := scopeParams()
	p.Dates = map[string]stores.DateBound{"created_at": {From: &from, To: &to}}
	res, err := testStores.Scopes.Search(ctx, p)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if res.TotalItems != 1 || (len(res.Items) == 1 && res.Items[0].ID != jun.ID) {
		t.Fatalf("Mar-Sep window: total=%d, want only Jun", res.TotalItems)
	}
}

// End-to-end through the transformers, exercising label resolution: filter by
// current name, by legacy name after a rename, and by an unknown name.
func TestScopeSearchLabelFacetEndToEnd(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t); flushLabels(t) })

	nexus, _ := testStores.Applications.CreateApplication(ctx, "Nexus", "nexus")
	globex, _ := testStores.Applications.CreateApplication(ctx, "Globex", "globex")
	testStores.Scopes.Define(ctx, nexus.ID, "projects:read", "")
	testStores.Scopes.Define(ctx, globex.ID, "billing:read", "")
	if err := testAppLabels.Reconcile(ctx); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	search := func(t *testing.T, appNames ...string) wire.ScopeSearchResponse {
		t.Helper()
		body := wire.SearchScopesRequest{Facets: map[string][]string{"application": appNames}}
		params, number, size, err := transformers.ResolveScopeSearch(ctx, body, testAppLabels)
		if err != nil {
			t.Fatalf("ResolveScopeSearch: %v", err)
		}
		result, err := testStores.Scopes.Search(ctx, params)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		resp, err := transformers.ScopeSearchToResponse(ctx, result, number, size, testAppLabels)
		if err != nil {
			t.Fatalf("ScopeSearchToResponse: %v", err)
		}
		return resp
	}

	t.Run("FilterByCurrentName", func(t *testing.T) {
		if r := search(t, "Nexus"); r.Page.TotalItems != 1 {
			t.Fatalf("by current name total = %d, want 1", r.Page.TotalItems)
		}
	})

	t.Run("FilterByLegacyNameAfterRename", func(t *testing.T) {
		// Rename Nexus -> NexusCorp and push the new mapping (as the event would).
		if _, err := testStores.Applications.Update(ctx, nexus.ID, "NexusCorp", models.ApplicationStatusActive); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if err := testAppLabels.Put(ctx, nexus.ID, "NexusCorp"); err != nil {
			t.Fatalf("Put: %v", err)
		}
		// The old name still resolves (append-only name->id) and finds its scopes.
		if legacy := search(t, "Nexus"); legacy.Page.TotalItems != 1 {
			t.Fatalf("legacy name total = %d, want 1", legacy.Page.TotalItems)
		}
		// Outbound facet shows the CURRENT name.
		current := search(t, "NexusCorp")
		if len(current.Facets["application"]) != 1 || current.Facets["application"][0] != "NexusCorp" {
			t.Fatalf("facet = %v, want [NexusCorp]", current.Facets["application"])
		}
	})

	t.Run("FilterByUnknownNameIsEmpty", func(t *testing.T) {
		r := search(t, "DoesNotExist")
		if r.Page.TotalItems != 0 || len(r.Scopes) != 0 {
			t.Fatalf("unknown name total = %d, want 0 (empty page, 200)", r.Page.TotalItems)
		}
	})
}
