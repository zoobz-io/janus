package transformers

import (
	"context"
	"testing"

	"github.com/zoobz-io/janus/database/models"
)

func TestLicenseToResponse(t *testing.T) {
	ctx := context.Background()
	licenses := []*models.License{
		{ID: "l1", TenantID: "t1", ApplicationID: "app-1"},
		{ID: "l2", TenantID: "t2", ApplicationID: "app-2"},
	}

	single, err := LicenseToResponse(ctx, licenses[0], labelResolver())
	if err != nil || single.Application != "Nexus" || single.TenantID != "t1" {
		t.Fatalf("LicenseToResponse = %+v,%v", single, err)
	}
	list, err := LicensesToResponse(ctx, licenses, labelResolver())
	if err != nil || list[1].Application != "Globex" {
		t.Fatalf("LicensesToResponse = %+v,%v", list, err)
	}
	if _, err := LicenseToResponse(ctx, licenses[0], errResolver()); err == nil {
		t.Fatal("LicenseToResponse should propagate resolver error")
	}
	if _, err := LicensesToResponse(ctx, licenses, errResolver()); err == nil {
		t.Fatal("LicensesToResponse should propagate resolver error")
	}
}
