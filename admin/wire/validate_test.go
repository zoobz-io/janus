package wire

import "testing"

// These validators return a single check builder result. A bare `.V()` return
// hits the typed-nil trap (non-nil error even when valid); the fix is `.V().Err()`.
// This guards that a valid body validates to a genuine nil error.
func TestSingleValidatorsReturnNilWhenValid(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"AuthorizeLicenseRequest", AuthorizeLicenseRequest{TenantID: "t-1"}.Validate()},
		{"CreateScopeRequest", CreateScopeRequest{Name: "read"}.Validate()},
		{"AddFeatureRequest", AddFeatureRequest{ScopeID: "s-1"}.Validate()},
		{"UpdateMemberRoleRequest", UpdateMemberRoleRequest{Role: "admin"}.Validate()},
	}
	for _, c := range cases {
		if c.err != nil {
			t.Errorf("%s.Validate() on valid input = %v, want nil", c.name, c.err)
		}
	}
}

func TestSingleValidatorsRejectInvalid(t *testing.T) {
	if (AuthorizeLicenseRequest{}).Validate() == nil {
		t.Error("AuthorizeLicenseRequest with empty tenant_id should fail")
	}
	if (CreateScopeRequest{}).Validate() == nil {
		t.Error("CreateScopeRequest with empty name should fail")
	}
	if (UpdateMemberRoleRequest{Role: "bogus"}).Validate() == nil {
		t.Error("UpdateMemberRoleRequest with unknown role should fail")
	}
}
