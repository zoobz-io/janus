//go:build testing

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

func userParams() stores.UserSearchParams {
	return stores.UserSearchParams{
		Sort: stores.SearchSort{Field: "updated_at", Order: stores.SortDesc},
		Page: stores.SearchPage{Offset: 0, Limit: 25},
	}
}

func setUserTimestamps(t *testing.T, id string, created, updated time.Time) {
	t.Helper()
	if _, err := testDB.Exec(
		`UPDATE users SET created_at = $1, updated_at = $2 WHERE id = $3`, created, updated, id,
	); err != nil {
		t.Fatalf("setUserTimestamps: %v", err)
	}
}

func TestUserSearch(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	alice, _ := testStores.Users.CreateUser(ctx, "alice@acme.test", "Alice Anderson")
	bob, _ := testStores.Users.CreateUser(ctx, "bob@acme.test", "Bob Brown")
	carol, _ := testStores.Users.CreateUser(ctx, "carol@globex.test", "Carol Clark")
	// Two of three inactive.
	testStores.Users.Update(ctx, bob.ID, "Bob Brown", models.UserStatusInactive)
	testStores.Users.Update(ctx, carol.ID, "Carol Clark", models.UserStatusInactive)

	t.Run("EmptyMatchesAll", func(t *testing.T) {
		res, err := testStores.Users.Search(ctx, userParams())
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 3 {
			t.Fatalf("total = %d, want 3", res.TotalItems)
		}
	})

	t.Run("TextMatchesEmailFragment", func(t *testing.T) {
		p := userParams()
		p.Query = "globex" // only carol's email domain
		res, err := testStores.Users.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 1 || (len(res.Items) == 1 && res.Items[0].ID != carol.ID) {
			t.Fatalf("email fragment total = %d, want only carol", res.TotalItems)
		}
	})

	t.Run("TextMatchesDisplayNameCaseInsensitive", func(t *testing.T) {
		p := userParams()
		p.Query = "anderson" // in Alice's display name; different case
		res, err := testStores.Users.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 1 || (len(res.Items) == 1 && res.Items[0].ID != alice.ID) {
			t.Fatalf("name fragment total = %d, want only alice", res.TotalItems)
		}
	})

	t.Run("StatusFacetSingle", func(t *testing.T) {
		p := userParams()
		p.Statuses = []string{"active"}
		res, err := testStores.Users.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 1 { // only alice
			t.Fatalf("active total = %d, want 1", res.TotalItems)
		}
	})

	t.Run("StatusFacetOrWithin", func(t *testing.T) {
		p := userParams()
		p.Statuses = []string{"active", "inactive"}
		res, err := testStores.Users.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 3 {
			t.Fatalf("active|inactive total = %d, want 3", res.TotalItems)
		}
	})

	t.Run("FacetValuesReflectFilteredSet", func(t *testing.T) {
		res, err := testStores.Users.Search(ctx, userParams())
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(res.Statuses) != 2 || res.Statuses[0] != "active" || res.Statuses[1] != "inactive" {
			t.Fatalf("facet statuses = %v, want [active inactive]", res.Statuses)
		}
	})
}

func TestUserSearchLikeEscaping(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	testStores.Users.CreateUser(ctx, "a@x.test", "Save 50% Now")
	testStores.Users.CreateUser(ctx, "b@x.test", "Save 500 Now")

	p := userParams()
	p.Query = "50%"
	res, err := testStores.Users.Search(ctx, p)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if res.TotalItems != 1 {
		t.Fatalf("query %q matched %d, want 1 (escaping failed)", p.Query, res.TotalItems)
	}
}

func TestUserSearchDateBoundsAndPaging(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	ids := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		u, err := testStores.Users.CreateUser(ctx, string(rune('a'+i))+"@d.test", "User")
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		ts := time.Date(2026, time.January, 1+i, 0, 0, 0, 0, time.UTC)
		setUserTimestamps(t, u.ID, ts, ts)
		ids = append(ids, u.ID)
	}

	t.Run("DateWindow", func(t *testing.T) {
		from := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, time.January, 4, 0, 0, 0, 0, time.UTC)
		p := userParams()
		p.Dates = map[string]stores.DateBound{"created_at": {From: &from, To: &to}}
		res, err := testStores.Users.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 3 { // Jan 2,3,4
			t.Fatalf("window total = %d, want 3", res.TotalItems)
		}
	})

	t.Run("SecondPageAndCount", func(t *testing.T) {
		p := userParams()
		p.Sort = stores.SearchSort{Field: "created_at", Order: stores.SortAsc}
		p.Page = stores.SearchPage{Offset: 2, Limit: 2}
		res, err := testStores.Users.Search(ctx, p)
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
		p := userParams()
		p.Page = stores.SearchPage{Offset: 100, Limit: 25}
		res, err := testStores.Users.Search(ctx, p)
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if res.TotalItems != 5 || len(res.Items) != 0 {
			t.Fatalf("out-of-range total=%d items=%d, want 5/0", res.TotalItems, len(res.Items))
		}
	})
}
