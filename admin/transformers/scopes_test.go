package transformers

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/janus/database/models"
)

// fakeResolver implements contracts.ApplicationLabels for transformer tests.
type fakeResolver struct {
	names map[string]string
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

func (f fakeResolver) ResolveID(_ context.Context, _ string) (string, bool, error) {
	return "", false, nil
}

func labelResolver() fakeResolver {
	return fakeResolver{names: map[string]string{"app-1": "Nexus", "app-2": "Globex"}}
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
