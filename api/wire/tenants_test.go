package wire

import "testing"

// Validate() must end in .Err(): a bare check.All(...) or .V() return hits the
// typed-nil trap (non-nil error even when valid), so rocco 422s every request.
func TestCreateTenantRequestValidate(t *testing.T) {
	if err := (CreateTenantRequest{Name: "Acme Corp", Slug: "acme-corp"}).Validate(); err != nil {
		t.Errorf("valid request = %v, want nil", err)
	}
	if (CreateTenantRequest{}).Validate() == nil {
		t.Error("empty request should fail validation")
	}
}
