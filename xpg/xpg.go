package xpg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sptGabriel/defaults/config"
	"go.opentelemetry.io/otel/trace"
)

// NewConnectionPool creates a pgxpool with tracing, connection limits, and health check.
func NewConnectionPool(ctx context.Context, cfg *config.PGDB, tracer trace.Tracer) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("parsing pgxpool config: %w", err)
	}

	// setting pool size
	config.MaxConns = int32(cfg.PoolMaxSize)
	config.MinConns = int32(cfg.PoolMinSize)
	config.MaxConnIdleTime = cfg.PoolMaxConnIdleTime
	config.MaxConnLifetime = cfg.PoolMaxConnLifetime

	// setting tracer
	config.ConnConfig.Tracer = NewQuerierTracer(tracer)

	pgxConn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("creating new pgxpool: %w", err)
	}

	if err := pgxConn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging pgxpool on %s: %w", cfg.Host, err)
	}

	return pgxConn, nil
}
