package transformers

import (
	"context"
	"testing"

	"github.com/zoobz-io/janus/database/models"
)

func TestGrantToResponse(t *testing.T) {
	ctx := context.Background()
	tier := "ti1"
	grants := []*models.Grant{
		{ID: "g1", UserID: "u1", TenantID: "t1", ApplicationID: "app-1", TierID: &tier, Roles: []string{"admin"}},
		{ID: "g2", UserID: "u2", TenantID: "t2", ApplicationID: "app-2"},
	}

	single, err := GrantToResponse(ctx, grants[0], labelResolver())
	if err != nil || single.Application != "Nexus" || single.TierID != "ti1" {
		t.Fatalf("GrantToResponse = %+v,%v", single, err)
	}
	// Nil tier pointer resolves to empty string.
	second, err := GrantToResponse(ctx, grants[1], labelResolver())
	if err != nil || second.Application != "Globex" || second.TierID != "" {
		t.Fatalf("GrantToResponse (no tier) = %+v,%v", second, err)
	}
	list, err := GrantsToResponse(ctx, grants, labelResolver())
	if err != nil || list[0].Application != "Nexus" {
		t.Fatalf("GrantsToResponse = %+v,%v", list, err)
	}
	if _, err := GrantsToResponse(ctx, grants, errResolver()); err == nil {
		t.Fatal("GrantsToResponse should propagate resolver error")
	}
}
