package app

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/OndasAlikhan/tourik/internal/config"
)

func NewDBPool(ctx context.Context, cfg config.AppConfig) (*pgxpool.Pool, error) {
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DatabaseUser, cfg.DatabasePassword),
		Host:     fmt.Sprintf("%s:%s", cfg.DatabaseHost, cfg.DatabasePort),
		Path:     fmt.Sprintf("/%s", cfg.DatabaseName),
		RawQuery: "sslmode=disable",
	}
	pgxConfig, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	pgxConfig.MaxConnLifetime = time.Hour
	pgxConfig.MaxConnLifetimeJitter = 5 * time.Minute
	pgxConfig.MaxConnIdleTime = 30 * time.Minute
	pgxConfig.PingTimeout = 5 * time.Second
	pgxConfig.MaxConns = 10
	pgxConfig.MinConns = 2
	pgxConfig.MinIdleConns = 2
	pgxConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func RunMigrations(pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS("migrations"))
	if err != nil {
		return fmt.Errorf("RunMigrations: create provider: %w", err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		return fmt.Errorf("RunMigrations: up: %w", err)
	}
	return nil
}
