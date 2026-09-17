package catalog

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func openTestDB(t *testing.T) *pgxpool.Pool {
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

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	return ctx
}

func resetProducts(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	const query = `TRUNCATE TABLE products`

	ctx := testContext(t)

	_, err := db.Exec(ctx, query)
	if err != nil {
		t.Fatalf("truncate products: %v", err)
	}
}

func insertProductFixture(t *testing.T, db *pgxpool.Pool, id uuid.UUID, sku string, createdAt time.Time) {
	t.Helper()

	const query = `
		INSERT INTO products (
			id,
			sku,
			name,
			price_minor,
			currency,
			available,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	ctx := testContext(t)

	_, err := db.Exec(
		ctx,
		query,
		id,
		sku,
		"Test "+sku,
		int64(1000),
		"CAD",
		true,
		createdAt,
	)
	if err != nil {
		t.Fatalf("insert product fixture %s: %v", sku, err)
	}
}
