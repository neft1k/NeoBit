package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPool(ctx context.Context) (*pgxpool.Pool, error) {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	db := os.Getenv("POSTGRES_DB")
	if host == "" || port == "" || user == "" || pass == "" || db == "" {
		return nil, fmt.Errorf("database environment variables not set")
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	connAttempts := 5
	for i := 1; i <= connAttempts; i++ {
		err = pool.Ping(ctx)
		if err == nil {
			return pool, nil
		}
		if i < connAttempts {
			time.Sleep(2 * time.Second)
			continue
		}
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
