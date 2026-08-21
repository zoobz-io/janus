package wire

import "testing"

func TestAuthorizeLicenseRequestValidate(t *testing.T) {
	// Valid body must validate to a genuine nil error (the bare `.V()` return hit
	// the typed-nil trap; the fix is `.V().Err()`).
	if err := (AuthorizeLicenseRequest{TenantID: "t-1"}).Validate(); err != nil {
		t.Errorf("valid request = %v, want nil", err)
	}
	if (AuthorizeLicenseRequest{}).Validate() == nil {
		t.Error("empty tenant_id should fail")
	}
}
