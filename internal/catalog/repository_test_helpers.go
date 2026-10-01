package catalog

import (
	"testing"
	"time"

	"github.com/goodday0404/commercepay/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

	ctx := testutil.TestContext(t)

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
