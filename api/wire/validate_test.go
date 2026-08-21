package wire

import "testing"

// Every Validate() in this package must end in .Err(): a bare check.All(...)
// or .V() return hits the typed-nil trap (non-nil error even when valid), so
// rocco 422s every request. The admin surface was swept in two passes (#7,
// #12); this guards the public surface the same way.
func TestValidatorsReturnNilWhenValid(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"CreateTenantRequest", CreateTenantRequest{Name: "Acme Corp", Slug: "acme-corp"}.Validate()},
		{"AddMemberRequest", AddMemberRequest{UserID: "u-1", Role: "viewer"}.Validate()},
		{"UpdateMemberRoleRequest", UpdateMemberRoleRequest{Role: "admin"}.Validate()},
		{"AuthorizeApplicationRequest", AuthorizeApplicationRequest{ApplicationID: "a-1"}.Validate()},
		{"GrantApplicationRequest", GrantApplicationRequest{ApplicationID: "a-1", UserID: "u-1"}.Validate()},
		{"UpdateProfileRequest", UpdateProfileRequest{DisplayName: "Jane Doe"}.Validate()},
	}
	for _, c := range cases {
		if c.err != nil {
			t.Errorf("%s.Validate() on valid input = %v, want nil", c.name, c.err)
		}
	}
}

func TestValidatorsRejectInvalid(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"CreateTenantRequest empty", CreateTenantRequest{}.Validate()},
		{"AddMemberRequest bad role", AddMemberRequest{UserID: "u-1", Role: "bogus"}.Validate()},
		{"UpdateMemberRoleRequest bad role", UpdateMemberRoleRequest{Role: "bogus"}.Validate()},
		{"AuthorizeApplicationRequest empty", AuthorizeApplicationRequest{}.Validate()},
		{"GrantApplicationRequest empty", GrantApplicationRequest{}.Validate()},
		{"UpdateProfileRequest empty", UpdateProfileRequest{}.Validate()},
	}
	for _, c := range cases {
		if c.err == nil {
			t.Errorf("%s should fail validation, got nil", c.name)
		}
	}
}
