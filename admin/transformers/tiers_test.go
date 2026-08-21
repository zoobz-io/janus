package transformers

import (
	"context"
	"testing"

	"github.com/zoobz-io/janus/database/models"
)

func TestTierToResponse(t *testing.T) {
	ctx := context.Background()
	tiers := []*models.Tier{
		{ID: "ti1", ApplicationID: "app-1", Slug: "pro", Name: "Pro", Rank: 1},
		{ID: "ti2", ApplicationID: "app-2", Slug: "free", Name: "Free", Rank: 0},
	}

	single, err := TierToResponse(ctx, tiers[0], labelResolver())
	if err != nil || single.Application != "Nexus" || single.Slug != "pro" {
		t.Fatalf("TierToResponse = %+v,%v", single, err)
	}
	list, err := TiersToResponse(ctx, tiers, labelResolver())
	if err != nil || list[1].Application != "Globex" {
		t.Fatalf("TiersToResponse = %+v,%v", list, err)
	}
	if _, err := TierToResponse(ctx, tiers[0], errResolver()); err == nil {
		t.Fatal("TierToResponse should propagate resolver error")
	}
	if _, err := TiersToResponse(ctx, tiers, errResolver()); err == nil {
		t.Fatal("TiersToResponse should propagate resolver error")
	}
}
