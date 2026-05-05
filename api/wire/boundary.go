package wire

import "github.com/zoobz-io/sum"

// RegisterBoundaries registers all API wire type boundaries with the given key.
func RegisterBoundaries(k sum.Key) {
	sum.NewBoundary[UserResponse](k)
	sum.NewBoundary[TenantResponse](k)
	sum.NewBoundary[SessionListResponse](k)
	sum.NewBoundary[LinkedIdentityListResponse](k)
}
