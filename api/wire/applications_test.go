package wire

import "testing"

// Validate() must end in .Err(): a bare check.All(...) or .V() return hits the
// typed-nil trap (non-nil error even when valid), so rocco 422s every request.
func TestAuthorizeApplicationRequestValidate(t *testing.T) {
	if err := (AuthorizeApplicationRequest{ApplicationID: "a-1"}).Validate(); err != nil {
		t.Errorf("valid request = %v, want nil", err)
	}
	if (AuthorizeApplicationRequest{}).Validate() == nil {
		t.Error("empty request should fail validation")
	}
}

func TestGrantApplicationRequestValidate(t *testing.T) {
	if err := (GrantApplicationRequest{ApplicationID: "a-1", UserID: "u-1"}).Validate(); err != nil {
		t.Errorf("valid request = %v, want nil", err)
	}
	if (GrantApplicationRequest{}).Validate() == nil {
		t.Error("empty request should fail validation")
	}
}
