package wire

import "testing"

func TestUpdateMemberRoleRequestValidate(t *testing.T) {
	if err := (UpdateMemberRoleRequest{Role: "admin"}).Validate(); err != nil {
		t.Errorf("valid request = %v, want nil", err)
	}
	if (UpdateMemberRoleRequest{Role: "bogus"}).Validate() == nil {
		t.Error("unknown role should fail")
	}
}
