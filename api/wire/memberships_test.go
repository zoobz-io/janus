package wire

import "testing"

// Validate() must end in .Err(): a bare check.All(...) or .V() return hits the
// typed-nil trap (non-nil error even when valid), so rocco 422s every request.
func TestAddMemberRequestValidate(t *testing.T) {
	if err := (AddMemberRequest{UserID: "u-1", Role: "viewer"}).Validate(); err != nil {
		t.Errorf("valid request = %v, want nil", err)
	}
	if (AddMemberRequest{UserID: "u-1", Role: "bogus"}).Validate() == nil {
		t.Error("unknown role should fail validation")
	}
}

func TestUpdateMemberRoleRequestValidate(t *testing.T) {
	if err := (UpdateMemberRoleRequest{Role: "admin"}).Validate(); err != nil {
		t.Errorf("valid request = %v, want nil", err)
	}
	if (UpdateMemberRoleRequest{Role: "bogus"}).Validate() == nil {
		t.Error("unknown role should fail validation")
	}
}
