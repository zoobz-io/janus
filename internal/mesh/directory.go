// Package mesh implements the aegis gRPC service contracts for identity,
// session management, and directory lookups.
package mesh

import (
	"context"
	"fmt"

	directorypb "github.com/zoobz-io/aegis/proto/directory/v1"

	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// DirectoryServer implements directorypb.DirectoryServiceServer.
type DirectoryServer struct {
	directorypb.UnimplementedDirectoryServiceServer
	users       *stores.Users
	tenants     *stores.Tenants
	memberships *stores.Memberships
}

// NewDirectoryServer creates a new DirectoryServer.
func NewDirectoryServer(users *stores.Users, tenants *stores.Tenants, memberships *stores.Memberships) *DirectoryServer {
	return &DirectoryServer{
		users:       users,
		tenants:     tenants,
		memberships: memberships,
	}
}

// GetUser retrieves a user by ID.
func (s *DirectoryServer) GetUser(ctx context.Context, req *directorypb.GetUserRequest) (*directorypb.GetUserResponse, error) {
	user, err := s.users.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("looking up user: %w", err)
	}
	return userToProto(user), nil
}

// GetUserByEmail retrieves a user by email address.
func (s *DirectoryServer) GetUserByEmail(ctx context.Context, req *directorypb.GetUserByEmailRequest) (*directorypb.GetUserResponse, error) {
	user, err := s.users.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("looking up user by email: %w", err)
	}
	return userToProto(user), nil
}

// GetTenant retrieves a tenant by ID.
func (s *DirectoryServer) GetTenant(ctx context.Context, req *directorypb.GetTenantRequest) (*directorypb.GetTenantResponse, error) {
	tenant, err := s.tenants.GetTenant(ctx, req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("looking up tenant: %w", err)
	}
	return tenantToProto(tenant), nil
}

// CreateTenant creates a new tenant with an owner.
func (s *DirectoryServer) CreateTenant(ctx context.Context, req *directorypb.CreateTenantRequest) (*directorypb.CreateTenantResponse, error) {
	tenant, err := s.tenants.CreateTenant(ctx, req.Name, req.Slug)
	if err != nil {
		return nil, fmt.Errorf("creating tenant: %w", err)
	}

	if req.OwnerUserId != "" {
		if _, err := s.memberships.Create(ctx, req.OwnerUserId, tenant.ID, models.UserRoleOwner); err != nil {
			return nil, fmt.Errorf("creating owner membership: %w", err)
		}
	}

	events.TenantCreated.Emit(ctx, events.TenantEvent{TenantID: tenant.ID})

	return &directorypb.CreateTenantResponse{
		TenantId: tenant.ID,
	}, nil
}

// UpdateTenant updates tenant details.
func (s *DirectoryServer) UpdateTenant(ctx context.Context, req *directorypb.UpdateTenantRequest) (*directorypb.UpdateTenantResponse, error) {
	// Fetch current tenant to preserve status (proto doesn't expose status updates).
	tenant, err := s.tenants.GetTenant(ctx, req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("looking up tenant: %w", err)
	}

	name := tenant.Name
	if req.Name != "" {
		name = req.Name
	}

	if _, err := s.tenants.UpdateTenant(ctx, req.TenantId, name, tenant.Status); err != nil {
		return nil, fmt.Errorf("updating tenant: %w", err)
	}

	events.TenantUpdated.Emit(ctx, events.TenantEvent{TenantID: req.TenantId})

	return &directorypb.UpdateTenantResponse{}, nil
}

func userToProto(u *models.User) *directorypb.GetUserResponse {
	return &directorypb.GetUserResponse{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.DisplayName,
		CreatedAt: u.CreatedAt.Unix(),
	}
}

func tenantToProto(t *models.Tenant) *directorypb.GetTenantResponse {
	return &directorypb.GetTenantResponse{
		Id:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt.Unix(),
		UpdatedAt: t.UpdatedAt.Unix(),
	}
}
