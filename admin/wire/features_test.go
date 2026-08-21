package wire

import "testing"

func TestAddFeatureRequestValidate(t *testing.T) {
	if err := (AddFeatureRequest{ScopeID: "s-1"}).Validate(); err != nil {
		t.Errorf("valid request = %v, want nil", err)
	}
	if (AddFeatureRequest{}).Validate() == nil {
		t.Error("empty scope_id should fail")
	}
}
