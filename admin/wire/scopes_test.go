package wire

import "testing"

func TestCreateScopeRequestValidate(t *testing.T) {
	if err := (CreateScopeRequest{Name: "read"}).Validate(); err != nil {
		t.Errorf("valid request = %v, want nil", err)
	}
	if (CreateScopeRequest{}).Validate() == nil {
		t.Error("empty name should fail")
	}
}
