package wire

import "testing"

// Validate() must end in .Err(): a bare check.All(...) or .V() return hits the
// typed-nil trap (non-nil error even when valid), so rocco 422s every request.
func TestUpdateProfileRequestValidate(t *testing.T) {
	if err := (UpdateProfileRequest{DisplayName: "Jane Doe"}).Validate(); err != nil {
		t.Errorf("valid request = %v, want nil", err)
	}
	if (UpdateProfileRequest{}).Validate() == nil {
		t.Error("empty request should fail validation")
	}
}
