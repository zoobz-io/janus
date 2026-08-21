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

var resolver = fakeResolver{names: map[string]string{"app-1": "Nexus", "app-2": "Globex"}}

func errResolver() fakeResolver { return fakeResolver{err: errors.New("redis down")} }

func TestScopeLabelResolution(t *testing.T) {
	ctx := context.Background()
	scopes := []*models.Scope{
		{ID: "s1", ApplicationID: "app-1", Name: "read"},
		{ID: "s2", ApplicationID: "app-2", Name: "write"},
	}

	single, err := ScopeToResponse(ctx, scopes[0], resolver)
	if err != nil || single.Application != "Nexus" {
		t.Fatalf("ScopeToResponse = %q,%v; want Nexus,nil", single.Application, err)
	}

	list, err := ScopesToResponse(ctx, scopes, resolver)
	if err != nil {
		t.Fatalf("ScopesToResponse: %v", err)
	}
	if list[0].Application != "Nexus" || list[1].Application != "Globex" {
		t.Fatalf("list labels = %q,%q", list[0].Application, list[1].Application)
	}

	if _, err := ScopeToResponse(ctx, scopes[0], errResolver()); err == nil {
		t.Fatal("ScopeToResponse should propagate resolver error")
	}
	if _, err := ScopesToResponse(ctx, scopes, errResolver()); err == nil {
		t.Fatal("ScopesToResponse should propagate resolver error")
	}
}

func TestLicenseLabelResolution(t *testing.T) {
	ctx := context.Background()
	licenses := []*models.License{
		{ID: "l1", TenantID: "t1", ApplicationID: "app-1"},
		{ID: "l2", TenantID: "t2", ApplicationID: "app-2"},
	}

	single, err := LicenseToResponse(ctx, licenses[0], resolver)
	if err != nil || single.Application != "Nexus" || single.TenantID != "t1" {
		t.Fatalf("LicenseToResponse = %+v,%v", single, err)
	}
	list, err := LicensesToResponse(ctx, licenses, resolver)
	if err != nil || list[1].Application != "Globex" {
		t.Fatalf("LicensesToResponse = %+v,%v", list, err)
	}
	if _, err := LicensesToResponse(ctx, licenses, errResolver()); err == nil {
		t.Fatal("LicensesToResponse should propagate resolver error")
	}
}

func TestTierLabelResolution(t *testing.T) {
	ctx := context.Background()
	tiers := []*models.Tier{
		{ID: "ti1", ApplicationID: "app-1", Slug: "pro", Name: "Pro", Rank: 1},
		{ID: "ti2", ApplicationID: "app-2", Slug: "free", Name: "Free", Rank: 0},
	}

	single, err := TierToResponse(ctx, tiers[0], resolver)
	if err != nil || single.Application != "Nexus" || single.Slug != "pro" {
		t.Fatalf("TierToResponse = %+v,%v", single, err)
	}
	list, err := TiersToResponse(ctx, tiers, resolver)
	if err != nil || list[1].Application != "Globex" {
		t.Fatalf("TiersToResponse = %+v,%v", list, err)
	}
	if _, err := TiersToResponse(ctx, tiers, errResolver()); err == nil {
		t.Fatal("TiersToResponse should propagate resolver error")
	}
}

func TestGrantLabelResolution(t *testing.T) {
	ctx := context.Background()
	tier := "ti1"
	grants := []*models.Grant{
		{ID: "g1", UserID: "u1", TenantID: "t1", ApplicationID: "app-1", TierID: &tier, Roles: []string{"admin"}},
		{ID: "g2", UserID: "u2", TenantID: "t2", ApplicationID: "app-2"},
	}

	single, err := GrantToResponse(ctx, grants[0], resolver)
	if err != nil || single.Application != "Nexus" || single.TierID != "ti1" {
		t.Fatalf("GrantToResponse = %+v,%v", single, err)
	}
	// Nil tier pointer resolves to empty string.
	second, err := GrantToResponse(ctx, grants[1], resolver)
	if err != nil || second.Application != "Globex" || second.TierID != "" {
		t.Fatalf("GrantToResponse (no tier) = %+v,%v", second, err)
	}
	list, err := GrantsToResponse(ctx, grants, resolver)
	if err != nil || list[0].Application != "Nexus" {
		t.Fatalf("GrantsToResponse = %+v,%v", list, err)
	}
	if _, err := GrantsToResponse(ctx, grants, errResolver()); err == nil {
		t.Fatal("GrantsToResponse should propagate resolver error")
	}
}

// Unresolved ids leave the label empty rather than erroring.
func TestUnresolvedLabelIsEmpty(t *testing.T) {
	ctx := context.Background()
	scope := &models.Scope{ID: "s", ApplicationID: "unknown", Name: "x"}
	resp, err := ScopeToResponse(ctx, scope, resolver)
	if err != nil {
		t.Fatalf("ScopeToResponse: %v", err)
	}
	if resp.Application != "" {
		t.Fatalf("unresolved label = %q, want empty", resp.Application)
	}
}
