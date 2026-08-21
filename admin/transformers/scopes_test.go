package transformers

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/janus/admin/wire"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// fakeResolver implements contracts.ApplicationLabels for transformer tests.
// names is the id->name direction; ids is the name->id direction.
type fakeResolver struct {
	names map[string]string
	ids   map[string]string
	err   error
}

func (f fakeResolver) ResolveNames(_ context.Context, ids []string) (map[string]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make(map[string]string, len(ids))
	for _, id := range ids {
		if n, ok := f.names[id]; ok {
			out[id] = n
		}
	}
	return out, nil
}

func (f fakeResolver) ResolveID(_ context.Context, name string) (string, bool, error) {
	if f.err != nil {
		return "", false, f.err
	}
	id, ok := f.ids[name]
	return id, ok, nil
}

func labelResolver() fakeResolver {
	return fakeResolver{
		names: map[string]string{"app-1": "Nexus", "app-2": "Globex"},
		ids:   map[string]string{"Nexus": "app-1", "Globex": "app-2"},
	}
}

func errResolver() fakeResolver { return fakeResolver{err: errors.New("redis down")} }

func TestScopeToResponse(t *testing.T) {
	ctx := context.Background()
	scopes := []*models.Scope{
		{ID: "s1", ApplicationID: "app-1", Name: "read"},
		{ID: "s2", ApplicationID: "app-2", Name: "write"},
	}

	single, err := ScopeToResponse(ctx, scopes[0], labelResolver())
	if err != nil || single.Application != "Nexus" {
		t.Fatalf("ScopeToResponse = %q,%v; want Nexus,nil", single.Application, err)
	}

	list, err := ScopesToResponse(ctx, scopes, labelResolver())
	if err != nil {
		t.Fatalf("ScopesToResponse: %v", err)
	}
	if list[0].Application != "Nexus" || list[1].Application != "Globex" {
		t.Fatalf("list labels = %q,%q", list[0].Application, list[1].Application)
	}

	// Unresolved id leaves the label empty, not an error.
	unresolved, err := ScopeToResponse(ctx, &models.Scope{ID: "s", ApplicationID: "unknown"}, labelResolver())
	if err != nil || unresolved.Application != "" {
		t.Fatalf("unresolved = %q,%v; want empty,nil", unresolved.Application, err)
	}

	if _, err := ScopeToResponse(ctx, scopes[0], errResolver()); err == nil {
		t.Fatal("ScopeToResponse should propagate resolver error")
	}
	if _, err := ScopesToResponse(ctx, scopes, errResolver()); err == nil {
		t.Fatal("ScopesToResponse should propagate resolver error")
	}
}

func TestResolveScopeSearchDefaults(t *testing.T) {
	ctx := context.Background()
	params, number, size, err := ResolveScopeSearch(ctx, wire.SearchScopesRequest{}, labelResolver())
	if err != nil {
		t.Fatalf("ResolveScopeSearch: %v", err)
	}
	if number != 1 || size != 25 {
		t.Fatalf("defaults number/size = %d/%d, want 1/25", number, size)
	}
	if params.Sort.Field != "updated_at" || params.Sort.Order != stores.SortDesc {
		t.Fatalf("default sort = %+v", params.Sort)
	}
	if params.ApplicationIDs != nil {
		t.Fatalf("no application facet should yield nil ApplicationIDs, got %v", params.ApplicationIDs)
	}
}

func TestResolveScopeSearchApplicationFacet(t *testing.T) {
	ctx := context.Background()

	// Known names resolve to ids.
	body := wire.SearchScopesRequest{Facets: map[string][]string{"application": {"Nexus", "Globex"}}}
	params, _, _, err := ResolveScopeSearch(ctx, body, labelResolver())
	if err != nil {
		t.Fatalf("ResolveScopeSearch: %v", err)
	}
	if len(params.ApplicationIDs) != 2 {
		t.Fatalf("resolved ids = %v, want [app-1 app-2]", params.ApplicationIDs)
	}

	// Present facet with only unknown names -> non-nil EMPTY (matches nothing),
	// not nil (which would mean no filter).
	body = wire.SearchScopesRequest{Facets: map[string][]string{"application": {"Ghost"}}}
	params, _, _, err = ResolveScopeSearch(ctx, body, labelResolver())
	if err != nil {
		t.Fatalf("ResolveScopeSearch: %v", err)
	}
	if params.ApplicationIDs == nil || len(params.ApplicationIDs) != 0 {
		t.Fatalf("all-unknown facet = %v, want non-nil empty", params.ApplicationIDs)
	}

	// Resolver error propagates.
	if _, _, _, err := ResolveScopeSearch(ctx, body, errResolver()); err == nil {
		t.Fatal("ResolveScopeSearch should propagate resolver error")
	}
}

func TestScopeSearchToResponse(t *testing.T) {
	ctx := context.Background()
	result := &stores.ScopeSearchResult{
		Items: []*models.Scope{
			{ID: "s1", ApplicationID: "app-2", Name: "read"},
			{ID: "s2", ApplicationID: "app-1", Name: "write"},
		},
		ApplicationIDs: []string{"app-2", "app-1"}, // store order (by id)
		TotalItems:     2,
	}

	resp, err := ScopeSearchToResponse(ctx, result, 1, 25, labelResolver())
	if err != nil {
		t.Fatalf("ScopeSearchToResponse: %v", err)
	}
	if resp.Scopes[0].Application != "Globex" || resp.Scopes[1].Application != "Nexus" {
		t.Fatalf("row labels = %q,%q", resp.Scopes[0].Application, resp.Scopes[1].Application)
	}
	// Facet values are names, sorted (resolution destroys the store's id order).
	if len(resp.Facets["application"]) != 2 || resp.Facets["application"][0] != "Globex" || resp.Facets["application"][1] != "Nexus" {
		t.Fatalf("facet = %v, want [Globex Nexus]", resp.Facets["application"])
	}
	if resp.Page.TotalItems != 2 || resp.Page.TotalPages != 1 {
		t.Fatalf("page = %+v", resp.Page)
	}

	if _, err := ScopeSearchToResponse(ctx, result, 1, 25, errResolver()); err == nil {
		t.Fatal("ScopeSearchToResponse should propagate resolver error")
	}
}
