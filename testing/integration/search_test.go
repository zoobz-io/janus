//go:build testing

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// searchParams returns a params value with the contract defaults applied, so
// each test only sets the fields it exercises.
func searchParams() stores.ApplicationSearchParams {
	return stores.ApplicationSearchParams{
		Sort: stores.SearchSort{Field: "updated_at", Order: stores.SortDesc},
		Page: stores.SearchPage{Offset: 0, Limit: 25},
	}
}

// setTimestamps forces created_at/updated_at for an application row so date and
// sort behavior can be tested deterministically.
func setTimestamps(t *testing.T, id string, created, updated time.Time) {
	t.Helper()
	if _, err := testDB.Exec(
		`UPDATE applications SET created_at = $1, updated_at = $2 WHERE id = $3`,
		created, updated, id,
	); err != nil {
		t.Fatalf("setTimestamps: %v", err)
	}
}

func TestApplicationSearch(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	apps := testStores.Applications

	acme, _ := apps.CreateApplication(ctx, "Acme Portal", "acme")
	acmeAPI, _ := apps.CreateApplication(ctx, "Acme API", "acme-api")
	globex, _ := apps.CreateApplication(ctx, "Globex", "globex")
	initech, _ := apps.CreateApplication(ctx, "Initech", "initech")

	// Two of the four are inactive.
	if _, err := apps.Update(ctx, globex.ID, "Globex", models.ApplicationStatusInactive); err != nil {
		t.Fatalf("Update globex: %v", err)
	}
	if _, err := apps.Update(ctx, initech.ID, "Initech", models.ApplicationStatusInactive); err != nil {
		t.Fatalf("Update initech: %v", err)
	}

	t.Run("EmptyMatchesAll", func(t *testing.T) {
		res, err := apps.Search(ctx, searchParams())
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 4 {
			t.Fatalf("total = %d, want 4", res.TotalItems)
		}
		if len(res.Items) != 4 {
			t.Fatalf("items = %d, want 4", len(res.Items))
		}
	})

	t.Run("TextMatchesNameCaseInsensitive", func(t *testing.T) {
		p := searchParams()
		p.Query = "ACME"
		res, err := apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 2 {
			t.Fatalf("total = %d, want 2 (Acme Portal, Acme API)", res.TotalItems)
		}
	})

	t.Run("TextMatchesSlugOnly", func(t *testing.T) {
		// "-api" appears in slug "acme-api" but not in any name.
		p := searchParams()
		p.Query = "-api"
		res, err := apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 1 || (len(res.Items) == 1 && res.Items[0].ID != acmeAPI.ID) {
			t.Fatalf("expected only acme-api, got total=%d", res.TotalItems)
		}
	})

	t.Run("StatusFacetSingle", func(t *testing.T) {
		p := searchParams()
		p.Statuses = []string{"active"}
		res, err := apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 2 {
			t.Fatalf("active total = %d, want 2", res.TotalItems)
		}
	})

	t.Run("StatusFacetOrWithin", func(t *testing.T) {
		p := searchParams()
		p.Statuses = []string{"active", "inactive"}
		res, err := apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 4 {
			t.Fatalf("active|inactive total = %d, want 4", res.TotalItems)
		}
	})

	t.Run("QueryAndFacetAndedTogether", func(t *testing.T) {
		// Acme apps are active, so query=acme AND status=inactive yields nothing.
		p := searchParams()
		p.Query = "acme"
		p.Statuses = []string{"inactive"}
		res, err := apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 0 {
			t.Fatalf("total = %d, want 0", res.TotalItems)
		}
		if len(res.Items) != 0 {
			t.Fatalf("items = %d, want 0", len(res.Items))
		}
	})

	t.Run("FacetValuesReflectFilteredSet", func(t *testing.T) {
		// Unfiltered: both statuses present, sorted.
		res, err := apps.Search(ctx, searchParams())
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(res.Statuses) != 2 || res.Statuses[0] != "active" || res.Statuses[1] != "inactive" {
			t.Fatalf("facet statuses = %v, want [active inactive]", res.Statuses)
		}

		// Filtered to active only: facet collapses to [active] (shared WHERE).
		p := searchParams()
		p.Statuses = []string{"active"}
		res, err = apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(res.Statuses) != 1 || res.Statuses[0] != "active" {
			t.Fatalf("filtered facet statuses = %v, want [active]", res.Statuses)
		}
	})

	_ = acme
	_ = initech
}

func TestApplicationSearchLikeEscaping(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })
	apps := testStores.Applications

	// "50%" must be matched literally: the % is escaped, so it does not act as a
	// wildcard that would also match "500".
	if _, err := apps.CreateApplication(ctx, "Discount 50% Off", "discount-half"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := apps.CreateApplication(ctx, "Discount 500 Off", "discount-500"); err != nil {
		t.Fatalf("create: %v", err)
	}

	p := searchParams()
	p.Query = "50%"
	res, err := apps.Search(ctx, p)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if res.TotalItems != 1 {
		t.Fatalf("query %q matched %d rows, want 1 (escaping failed)", p.Query, res.TotalItems)
	}

	// Underscore is likewise literal: "_" must not match an arbitrary character.
	if _, err := apps.CreateApplication(ctx, "a_b", "underscore-lit"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := apps.CreateApplication(ctx, "axb", "underscore-wild"); err != nil {
		t.Fatalf("create: %v", err)
	}
	p = searchParams()
	p.Query = "a_b"
	res, err = apps.Search(ctx, p)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if res.TotalItems != 1 {
		t.Fatalf("query %q matched %d rows, want 1 (underscore not escaped)", p.Query, res.TotalItems)
	}
}

func TestApplicationSearchDateBounds(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })
	apps := testStores.Applications

	jan, _ := apps.CreateApplication(ctx, "Jan App", "jan")
	jun, _ := apps.CreateApplication(ctx, "Jun App", "jun")
	dec, _ := apps.CreateApplication(ctx, "Dec App", "dec")

	t2026 := func(month time.Month) time.Time {
		return time.Date(2026, month, 15, 12, 0, 0, 0, time.UTC)
	}
	setTimestamps(t, jan.ID, t2026(time.January), t2026(time.January))
	setTimestamps(t, jun.ID, t2026(time.June), t2026(time.June))
	setTimestamps(t, dec.ID, t2026(time.December), t2026(time.December))

	from := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

	t.Run("FromAndTo", func(t *testing.T) {
		p := searchParams()
		p.Dates = map[string]stores.DateBound{"created_at": {From: &from, To: &to}}
		res, err := apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 1 || (len(res.Items) == 1 && res.Items[0].ID != jun.ID) {
			t.Fatalf("Mar-Sep window: total=%d, want only Jun", res.TotalItems)
		}
	})

	t.Run("FromOnly", func(t *testing.T) {
		p := searchParams()
		p.Dates = map[string]stores.DateBound{"created_at": {From: &from}}
		res, err := apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 2 { // Jun, Dec
			t.Fatalf("from Mar: total=%d, want 2", res.TotalItems)
		}
	})

	t.Run("ToOnly", func(t *testing.T) {
		p := searchParams()
		p.Dates = map[string]stores.DateBound{"created_at": {To: &to}}
		res, err := apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 2 { // Jan, Jun
			t.Fatalf("to Sep: total=%d, want 2", res.TotalItems)
		}
	})
}

func TestApplicationSearchPaging(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })
	apps := testStores.Applications

	// Create 5 apps with a stable created_at ordering.
	created := make([]string, 0, 5)
	for i, name := range []string{"App A", "App B", "App C", "App D", "App E"} {
		a, err := apps.CreateApplication(ctx, name, "app-"+string(rune('a'+i)))
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		ts := time.Date(2026, time.January, 1+i, 0, 0, 0, 0, time.UTC)
		setTimestamps(t, a.ID, ts, ts)
		created = append(created, a.ID)
	}

	t.Run("SecondPage", func(t *testing.T) {
		p := searchParams()
		p.Sort = stores.SearchSort{Field: "created_at", Order: stores.SortAsc}
		p.Page = stores.SearchPage{Offset: 2, Limit: 2}
		res, err := apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 5 {
			t.Fatalf("total = %d, want 5 (count ignores window)", res.TotalItems)
		}
		if len(res.Items) != 2 {
			t.Fatalf("page items = %d, want 2", len(res.Items))
		}
		// created ascending: page offset 2 limit 2 => 3rd and 4th created.
		if res.Items[0].ID != created[2] || res.Items[1].ID != created[3] {
			t.Fatalf("unexpected page contents")
		}
	})

	t.Run("OutOfRangePageIsEmptyButCounted", func(t *testing.T) {
		p := searchParams()
		p.Page = stores.SearchPage{Offset: 100, Limit: 25}
		res, err := apps.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 5 {
			t.Fatalf("total = %d, want 5", res.TotalItems)
		}
		if len(res.Items) != 0 {
			t.Fatalf("items = %d, want 0", len(res.Items))
		}
	})
}
