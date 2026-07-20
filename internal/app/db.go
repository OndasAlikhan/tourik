package app

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/OndasAlikhan/tourik/internal/config"
)

func NewDBPool(ctx context.Context, cfg config.AppConfig) (*pgxpool.Pool, error) {
	dbConfig := pgxpool.Config{
		ConnConfig: &pgx.ConnConfig{
			Config: pgconn.Config{
				Host:     cfg.DatabaseHost,
				Port:     cfg.DatabasePort,
				Database: cfg.DatabaseName,
				User:     cfg.DatabaseUser,
				Password: cfg.DatabasePassword,
			},
		},
		MaxConnLifetime:       time.Hour,
		MaxConnLifetimeJitter: 5 * time.Minute,
		MaxConnIdleTime:       30 * time.Minute,
		PingTimeout:           5 * time.Second,
		MaxConns:              10,
		MinConns:              2,
		MinIdleConns:          2,
		HealthCheckPeriod:     time.Minute,
	}

	pool, err := pgxpool.NewWithConfig(ctx, &dbConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
