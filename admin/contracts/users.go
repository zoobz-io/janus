package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// Users defines the admin API's capability boundary over the users store.
type Users interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.User], error)
	// Search runs the admin search over users.
	Search(ctx context.Context, params stores.UserSearchParams) (*stores.UserSearchResult, error)
	CreateUser(ctx context.Context, email, displayName string) (*models.User, error)
	Update(ctx context.Context, id, displayName string, status models.UserStatus) (*models.User, error)
}
