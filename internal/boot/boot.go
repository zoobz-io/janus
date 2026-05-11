// Package boot provides per-concern infrastructure connection functions.
//
// Each function reads config via sum.MustUse, builds the client, and returns it.
// Callers own lifecycle — defer Close on returned clients.
// Callers emit startup events after successful connection.
package boot

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	goredis "github.com/redis/go-redis/v9"
	"github.com/zoobz-io/aperture"
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/config"
	intotel "github.com/zoobz-io/janus/internal/otel"
)

// Database creates a PostgreSQL connection from config.
func Database(ctx context.Context) (*sqlx.DB, error) {
	cfg := sum.MustUse[config.Database](ctx)
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	return db, nil
}

// Redis creates a Redis client from config and verifies connectivity.
func Redis(ctx context.Context) (*goredis.Client, error) {
	cfg := sum.MustUse[config.Redis](ctx)
	client := goredis.NewClient(&goredis.Options{
		Addr: cfg.Addr,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}
	return client, nil
}

// OTEL creates OpenTelemetry providers. serviceName identifies the process
// in traces and metrics.
func OTEL(ctx context.Context, serviceName string) (*intotel.Providers, error) {
	cfg := sum.MustUse[config.OTEL](ctx)
	providers, err := intotel.New(ctx, intotel.Config{
		Endpoint:    cfg.Endpoint,
		ServiceName: serviceName,
	})
	if err != nil {
		return nil, fmt.Errorf("creating otel providers: %w", err)
	}
	return providers, nil
}

// Aperture creates an aperture bridge from capitan events to OTEL providers.
func Aperture(_ context.Context, providers *intotel.Providers) (*aperture.Aperture, error) {
	ap, err := aperture.New(
		capitan.Default(),
		providers.Log,
		providers.Metric,
		providers.Trace,
	)
	if err != nil {
		return nil, fmt.Errorf("creating aperture: %w", err)
	}
	return ap, nil
}
