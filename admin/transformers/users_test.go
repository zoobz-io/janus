package transformers

import (
	"testing"
	"time"

	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

func TestResolveUserSearchDefaults(t *testing.T) {
	params, number, size := ResolveUserSearch(wire.SearchUsersRequest{})
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

func TestResolveUserSearchMapping(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	body := wire.SearchUsersRequest{
		Query:  "jane",
		Facets: map[string][]string{"status": {"active"}},
		Dates:  map[string]wire.DateRange{"created_at": {From: &from}},
		Sort:   &wire.SortSpec{Field: "created_at", Order: "asc"},
		Page:   &wire.PageRequest{Number: intPtr(2), Size: intPtr(10)},
	}
	params, number, size := ResolveUserSearch(body)
	if number != 2 || size != 10 || params.Page.Offset != 10 {
		t.Fatalf("page = %d/%d offset=%d, want 2/10 offset 10", number, size, params.Page.Offset)
	}
	if params.Query != "jane" || len(params.Statuses) != 1 || params.Statuses[0] != "active" {
		t.Fatalf("filters = %+v", params)
	}
	if params.Sort.Field != "created_at" || params.Sort.Order != stores.SortAsc {
		t.Fatalf("sort = %+v", params.Sort)
	}
	if b, ok := params.Dates["created_at"]; !ok || b.From == nil || !b.From.Equal(from) {
		t.Fatalf("dates = %+v", params.Dates)
	}
}

func TestUserSearchToResponse(t *testing.T) {
	result := &stores.UserSearchResult{
		Items: []*models.User{
			{ID: "u1", Email: "jane@example.com", DisplayName: "Jane", Status: "active"},
			{ID: "u2", Email: "joe@example.com", DisplayName: "Joe", Status: "inactive"},
		},
		Statuses:   []string{"active", "inactive"},
		TotalItems: 2,
	}
	resp := UserSearchToResponse(result, 1, 25)
	if len(resp.Users) != 2 || resp.Users[0].Email != "jane@example.com" {
		t.Fatalf("users = %+v", resp.Users)
	}
	if resp.Page.TotalItems != 2 || resp.Page.TotalPages != 1 {
		t.Fatalf("page = %+v", resp.Page)
	}
	if len(resp.Facets["status"]) != 2 {
		t.Fatalf("facet = %v", resp.Facets["status"])
	}
}
