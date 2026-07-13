package wire

import "github.com/zoobz-io/sum"

// RegisterBoundaries registers all API wire type boundaries with the given key.
func RegisterBoundaries(k sum.Key) {
	sum.NewBoundary[UserResponse](k)
	sum.NewBoundary[TenantResponse](k)
	sum.NewBoundary[SessionListResponse](k)
	sum.NewBoundary[AccountListResponse](k)
	sum.NewBoundary[ApplicationListResponse](k)
	sum.NewBoundary[LicenseListResponse](k)
	sum.NewBoundary[GrantListResponse](k)
	sum.NewBoundary[MemberListResponse](k)
}
