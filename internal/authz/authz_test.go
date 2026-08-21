package authz

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/janus/database/models"
)

// stubMemberships satisfies the Memberships interface for driving the authz
// helpers without a database.
type stubMemberships struct {
	membership *models.Membership
	getErr     error
	ownerCount int64
	countErr   error
}

func (s *stubMemberships) GetByUserAndTenant(_ context.Context, _, _ string) (*models.Membership, error) {
	return s.membership, s.getErr
}

func (s *stubMemberships) CountOtherOwners(_ context.Context, _, _ string) (int64, error) {
	return s.ownerCount, s.countErr
}

func TestRequireRole(t *testing.T) {
	ctx := context.Background()

	t.Run("MatchingRole", func(t *testing.T) {
		m := &stubMemberships{membership: &models.Membership{Role: models.UserRoleAdmin}}
		mem, err := RequireRole(ctx, m, "u", "t", models.UserRoleAdmin, models.UserRoleOwner)
		if err != nil || mem == nil {
			t.Fatalf("RequireRole = %v, %v; want membership, nil", mem, err)
		}
	})

	t.Run("InsufficientRole", func(t *testing.T) {
		m := &stubMemberships{membership: &models.Membership{Role: models.UserRoleViewer}}
		if _, err := RequireRole(ctx, m, "u", "t", models.UserRoleAdmin); !errors.Is(err, ErrInsufficientRole) {
			t.Fatalf("RequireRole = %v, want ErrInsufficientRole", err)
		}
	})

	t.Run("NotMember", func(t *testing.T) {
		m := &stubMemberships{}
		if _, err := RequireRole(ctx, m, "u", "t", models.UserRoleAdmin); !errors.Is(err, ErrNotMember) {
			t.Fatalf("RequireRole = %v, want ErrNotMember", err)
		}
	})

	t.Run("StoreErrorIsNotADenial", func(t *testing.T) {
		storeErr := errors.New("connection refused")
		m := &stubMemberships{getErr: storeErr}
		_, err := RequireRole(ctx, m, "u", "t", models.UserRoleAdmin)
		if !errors.Is(err, storeErr) {
			t.Fatalf("RequireRole = %v, want wrapped store error", err)
		}
		if errors.Is(err, ErrNotMember) {
			t.Fatal("store failure must not read as ErrNotMember")
		}
	})
}

func TestRequireOwnerExists(t *testing.T) {
	ctx := context.Background()

	t.Run("CoOwnerExists", func(t *testing.T) {
		m := &stubMemberships{ownerCount: 1}
		if err := RequireOwnerExists(ctx, m, "t", "u"); err != nil {
			t.Fatalf("RequireOwnerExists = %v, want nil", err)
		}
	})

	t.Run("LastOwner", func(t *testing.T) {
		m := &stubMemberships{ownerCount: 0}
		if err := RequireOwnerExists(ctx, m, "t", "u"); !errors.Is(err, ErrLastOwner) {
			t.Fatalf("RequireOwnerExists = %v, want ErrLastOwner", err)
		}
	})

	t.Run("StoreErrorPropagates", func(t *testing.T) {
		storeErr := errors.New("connection refused")
		m := &stubMemberships{countErr: storeErr}
		if err := RequireOwnerExists(ctx, m, "t", "u"); !errors.Is(err, storeErr) {
			t.Fatalf("RequireOwnerExists = %v, want store error", err)
		}
	})
}
