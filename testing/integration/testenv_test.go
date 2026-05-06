//go:build testing

package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	goredis "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/cereal"
	"github.com/zoobz-io/sentinel"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/janus/stores"
)

var (
	testDB     *sqlx.DB
	testRedis  *goredis.Client
	testStores *stores.Stores
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Register struct tags so sentinel can extract them.
	sentinel.Tag("db")
	sentinel.Tag("type")
	sentinel.Tag("constraints")
	sentinel.Tag("default")
	sentinel.Tag("store.encrypt")
	sentinel.Tag("load.decrypt")

	// Initialize sum service — required by stores and model hooks.
	svc := sum.New()
	k := sum.Start()

	// Session model's BeforeSave/AfterLoad hooks need an encryptor and boundaries.
	testEncKey := make([]byte, 32) // zeroed key for testing
	aesEnc, err := cereal.AES(testEncKey)
	if err != nil {
		log.Fatalf("failed to create test encryptor: %v", err)
	}
	svc.WithEncryptor(cereal.EncryptAES, aesEnc)

	sum.NewBoundary[models.Tenant](k)
	sum.NewBoundary[models.User](k)
	sum.NewBoundary[models.Membership](k)
	sum.NewBoundary[models.LinkedIdentity](k)
	sum.NewBoundary[models.Session](k)
	sum.NewBoundary[models.Application](k)
	sum.NewBoundary[models.TenantApplication](k)
	sum.NewBoundary[models.UserApplication](k)
	sum.Freeze(k)

	pgContainer, err := postgres.Run(ctx, "postgres:16",
		postgres.WithDatabase("janus_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres: %v", err)
	}

	redisContainer, err := redis.Run(ctx, "redis:7",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(15*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start redis: %v", err)
	}

	pgDSN, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres dsn: %v", err)
	}

	testDB, err = sqlx.Connect("postgres", pgDSN)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	if err := runMigrations(testDB); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	redisEndpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		log.Fatalf("failed to get redis endpoint: %v", err)
	}

	testRedis = goredis.NewClient(&goredis.Options{Addr: redisEndpoint})
	if err := testRedis.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to ping redis: %v", err)
	}

	renderer := astqlpg.New()
	testStores = stores.New(testDB, renderer)

	code := m.Run()

	_ = testDB.Close()
	_ = testRedis.Close()
	_ = pgContainer.Terminate(ctx)
	_ = redisContainer.Terminate(ctx)

	os.Exit(code)
}

func runMigrations(db *sqlx.DB) error {
	migration, err := os.ReadFile("../../migrations/001_initial.sql")
	if err != nil {
		return fmt.Errorf("reading migration file: %w", err)
	}
	if _, err := db.Exec(string(migration)); err != nil {
		return fmt.Errorf("executing migration: %w", err)
	}
	return nil
}

// cleanTable truncates a table, cascading to dependent tables.
func cleanTable(t *testing.T, tables ...string) {
	t.Helper()
	for _, table := range tables {
		if _, err := testDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			t.Fatalf("failed to clean table %s: %v", table, err)
		}
	}
}

// cleanAll truncates all tables in dependency order.
func cleanAll(t *testing.T) {
	t.Helper()
	cleanTable(t,
		"user_applications",
		"tenant_applications",
		"sessions",
		"linked_identities",
		"memberships",
		"applications",
		"tenants",
		"users",
	)
}
