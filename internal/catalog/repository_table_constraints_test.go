package catalog

import (
	"errors"
	"testing"

	"github.com/goodday0404/commercepay/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestRepositoryProductsTableRejectsBlankRequiredText(t *testing.T) {
	tests := []struct {
		name     string
		sku      string
		product  string
		currency string
	}{
		{
			name:     "blank SKU",
			sku:      "   ",
			product:  "Valid Product",
			currency: "CAD",
		},
		{
			name:     "blank name",
			sku:      "VALID-SKU",
			product:  "   ",
			currency: "CAD",
		},
		{
			name:     "blank currency",
			sku:      "VALID-SKU",
			product:  "Valid Product",
			currency: "   ",
		},
	}

	db := testutil.OpenTestDB(t)
	ctx := testutil.TestContext(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ResetCatalogAndCart(t, db)

			_, err := db.Exec(
				ctx,
				`
					INSERT INTO products (
						id,
						sku,
						name,
						price_minor,
						currency,
						available
					)
					VALUES ($1, $2, $3, $4, $5, $6)
				`,
				uuid.New(),
				tt.sku,
				tt.product,
				int64(1000),
				tt.currency,
				true,
			)

			if err == nil {
				t.Fatalf("expected database to reject %s", tt.name)
			}

			var pgErr *pgconn.PgError

			if !errors.As(err, &pgErr) {
				t.Fatalf("expected PostgreSQL error, got %T: %v", err, err)
			}

			if pgErr.Code != SQLSTATE_CHECK_VIOLATION {
				t.Fatalf("expected SQLSTATE 23514, got %s", pgErr.Code)
			}
		})
	}
}

func TestRepositoryInsertMapsBlankTextConstraintViolations(t *testing.T) {
	tests := []struct {
		name        string
		sku         string
		productName string
		currency    string
		wantErr     error
	}{
		{
			name:        "blank SKU",
			sku:         "   ",
			productName: "Valid Product",
			currency:    "CAD",
			wantErr:     ErrEmptySKU,
		},
		{
			name:        "blank name",
			sku:         "VALID-SKU",
			productName: "   ",
			currency:    "CAD",
			wantErr:     ErrEmptyName,
		},
		{
			name:        "blank currency",
			sku:         "VALID-SKU",
			productName: "Valid Product",
			currency:    "   ",
			wantErr:     ErrEmptyCurrency,
		},
	}

	db := testutil.OpenTestDB(t)
	repo := NewRepository(db)
	ctx := testutil.TestContext(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ResetCatalogAndCart(t, db)

			product := Product{
				ID:         uuid.New(),
				SKU:        tt.sku,
				Name:       tt.productName,
				PriceMinor: 1000,
				Currency:   tt.currency,
				Available:  true,
			}

			err := repo.Insert(ctx, product)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantErr,
					err,
				)
			}
		}) // Run
	} // for
}

func TestRepositoryProductsTableRejectsNegativePrice(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)

	ctx := testutil.TestContext(t)

	const query = `
			INSERT INTO products (
				id,
				sku,
				name,
				price_minor,
				currency,
				available
			)
			VALUES ($1, $2, $3, $4, $5, $6)
		`

	_, err := db.Exec(
		ctx,
		query,
		uuid.New(),
		"NEGATIVE-PRICE",
		"Invalid Product",
		int64(-1),
		"CAD",
		true,
	)

	if err == nil {
		t.Fatal("expected database to reject negative price")
	}

	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		t.Fatalf("expected PostgreSQL error, got %T: %v", err, err)
	}

	if pgErr.Code != "23514" {
		t.Fatalf("expected SQLSTATE 23514, got %s", pgErr.Code)
	}

	if pgErr.ConstraintName != PRODUCTS_PRICE_MINOR_CHECK {
		t.Fatalf("expected price constraint, got %q", pgErr.ConstraintName)
	}
}
