package db

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"url-shortener/internal/shared/config"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func NewPostgres(ctx context.Context, cfg config.Config) (*pgxpool.Pool, *sqlx.DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse database url: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("create pgxpool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("ping postgres: %w", err)
	}

	sqlxDB, err := sqlx.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("open sqlx db: %w", err)
	}

	if err := sqlxDB.PingContext(ctx); err != nil {
		sqlxDB.Close()
		pool.Close()
		return nil, nil, fmt.Errorf("ping sqlx db: %w", err)
	}

	return pool, sqlxDB, nil
}

func RunMigration(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		filenames = append(filenames, entry.Name())
	}

	sort.Strings(filenames)
	for _, filename := range filenames {
		content, err := migrationFiles.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("execute migration %s: %w", filename, err)
		}
	}

	return nil
}
