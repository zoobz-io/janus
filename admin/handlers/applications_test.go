package handlers

import (
	"testing"
	"time"

	"github.com/zoobz-io/janus/admin/transformers"
	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
)

func intPtr(v int) *int { return &v }

func TestResolveApplicationSearchDefaults(t *testing.T) {
	params, number, size := resolveApplicationSearch(wire.SearchApplicationsRequest{})

	if number != 1 {
		t.Errorf("default number = %d, want 1", number)
	}
	if size != 25 {
		t.Errorf("default size = %d, want 25", size)
	}
	if params.Sort.Field != "updated_at" || params.Sort.Order != models.SortDesc {
		t.Errorf("default sort = %+v, want {updated_at DESC}", params.Sort)
	}
	if params.Page.Offset != 0 || params.Page.Limit != 25 {
		t.Errorf("default page = %+v, want {Offset:0 Limit:25}", params.Page)
	}
	if params.Query != "" || params.Statuses != nil || params.Dates != nil {
		t.Errorf("expected empty filters, got %+v", params)
	}
}

func TestResolveApplicationSearchMapping(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	body := wire.SearchApplicationsRequest{
		Query:  "acme",
		Facets: map[string][]string{"status": {"active"}},
		Dates:  map[string]wire.DateRange{"created_at": {From: &from}},
		Sort:   &wire.SortSpec{Field: "created_at", Order: "asc"},
		Page:   &wire.PageRequest{Number: intPtr(3), Size: intPtr(10)},
	}

	params, number, size := resolveApplicationSearch(body)

	if number != 3 || size != 10 {
		t.Errorf("number/size = %d/%d, want 3/10", number, size)
	}
	// Page 3 of size 10 => offset 20.
	if params.Page.Offset != 20 || params.Page.Limit != 10 {
		t.Errorf("page = %+v, want {Offset:20 Limit:10}", params.Page)
	}
	if params.Query != "acme" {
		t.Errorf("query = %q, want acme", params.Query)
	}
	if len(params.Statuses) != 1 || params.Statuses[0] != "active" {
		t.Errorf("statuses = %v, want [active]", params.Statuses)
	}
	if params.Sort.Field != "created_at" || params.Sort.Order != models.SortAsc {
		t.Errorf("sort = %+v, want {created_at ASC}", params.Sort)
	}
	bound, ok := params.Dates["created_at"]
	if !ok || bound.From == nil || !bound.From.Equal(from) || bound.To != nil {
		t.Errorf("dates[created_at] = %+v, want From=%v To=nil", bound, from)
	}
}

func TestSortOrderToSQL(t *testing.T) {
	cases := map[string]models.SortOrder{
		"asc":  models.SortAsc,
		"ASC":  models.SortAsc,
		"desc": models.SortDesc,
		"DESC": models.SortDesc,
		"":     models.SortDesc, // anything non-asc defaults to descending
	}
	for in, want := range cases {
		if got := sortOrderToSQL(in); got != want {
			t.Errorf("sortOrderToSQL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestApplicationSearchToResponsePageMath(t *testing.T) {
	tests := []struct {
		total    int64
		size     int
		wantPage int
	}{
		{total: 0, size: 25, wantPage: 0},
		{total: 1, size: 25, wantPage: 1},
		{total: 25, size: 25, wantPage: 1},
		{total: 26, size: 25, wantPage: 2},
		{total: 57, size: 25, wantPage: 3},
		{total: 100, size: 10, wantPage: 10},
		{total: 101, size: 10, wantPage: 11},
	}
	for _, tt := range tests {
		result := &models.ApplicationSearchResult{TotalItems: tt.total, Statuses: []string{}}
		resp := transformers.ApplicationSearchToResponse(result, 1, tt.size)
		if resp.Page.TotalPages != tt.wantPage {
			t.Errorf("total=%d size=%d: total_pages=%d, want %d", tt.total, tt.size, resp.Page.TotalPages, tt.wantPage)
		}
		if resp.Page.TotalItems != tt.total {
			t.Errorf("total_items=%d, want %d", resp.Page.TotalItems, tt.total)
		}
		if resp.Applications == nil {
			t.Error("applications should be non-nil empty slice, got nil")
		}
	}
}
