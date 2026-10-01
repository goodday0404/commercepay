package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create test database pool: %v", err)
	}

	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	return pool
}

func TestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	return ctx
}

func ResetCatalogAndCart(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	_, err := db.Exec(ctx, `
		TRUNCATE TABLE
			cart_items,
			carts,
			products
	`)
	if err != nil {
		t.Fatalf("reset test database: %v", err)
	}
}
