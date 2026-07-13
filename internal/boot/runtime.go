package boot

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/jmoiron/sqlx"
	goredis "github.com/redis/go-redis/v9"
	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/cereal"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/config"
	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/janus/stores"
)

// Runtime holds the pieces built by Init and shared by every janus binary.
// Contract registration, wire boundaries and Freeze are the caller's
// responsibility — they differ per surface (public API, admin API, mesh).
type Runtime struct {
	Svc    *sum.Service
	K      sum.Key
	DB     *sqlx.DB
	Redis  *goredis.Client
	Stores *stores.Stores
}

// Init performs the setup common to every janus binary: sum init, shared config
// load (database/redis/encryption/otel), encryption, infra connections, store
// construction and model boundary registration. The registry is left unfrozen
// so each binary can register its own contracts before freezing.
//
// The caller owns the returned DB and Redis clients — defer Close on both.
func Init(ctx context.Context) (*Runtime, error) {
	svc := sum.New()
	k := sum.Start()

	if err := sum.Config[config.Database](ctx, k, nil); err != nil {
		return nil, fmt.Errorf("loading database config: %w", err)
	}
	if err := sum.Config[config.Redis](ctx, k, nil); err != nil {
		return nil, fmt.Errorf("loading redis config: %w", err)
	}
	if err := sum.Config[config.Encryption](ctx, k, nil); err != nil {
		return nil, fmt.Errorf("loading encryption config: %w", err)
	}
	if err := sum.Config[config.OTEL](ctx, k, nil); err != nil {
		return nil, fmt.Errorf("loading otel config: %w", err)
	}

	encCfg := sum.MustUse[config.Encryption](ctx)
	encKey, err := hex.DecodeString(encCfg.Key)
	if err != nil {
		return nil, fmt.Errorf("decoding encryption key: %w", err)
	}
	aesEncryptor, err := cereal.AES(encKey)
	if err != nil {
		return nil, fmt.Errorf("creating AES encryptor: %w", err)
	}
	svc.WithEncryptor(cereal.EncryptAES, aesEncryptor)

	db, err := Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	redisClient, err := Redis(ctx)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}

	allStores := stores.New(db, astqlpg.New())

	// Model boundaries are the same for every binary that touches the stores.
	sum.NewBoundary[models.Tenant](k)
	sum.NewBoundary[models.User](k)
	sum.NewBoundary[models.Membership](k)
	sum.NewBoundary[models.Account](k)
	sum.NewBoundary[models.Session](k)
	sum.NewBoundary[models.Application](k)
	sum.NewBoundary[models.License](k)
	sum.NewBoundary[models.Grant](k)

	return &Runtime{Svc: svc, K: k, DB: db, Redis: redisClient, Stores: allStores}, nil
}
