package catalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const SQLSTATE_UNIQUE_CONSTRANT = "23505"

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Insert(ctx context.Context, product Product) error {
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

	_, err := r.db.Exec(
		ctx,
		query,
		product.ID,
		product.SKU,
		product.Name,
		product.PriceMinor,
		product.Currency,
		product.Available,
	)

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) &&
		pgErr.Code == SQLSTATE_UNIQUE_CONSTRANT &&
		pgErr.ConstraintName == "products_sku_unique" {

		return fmt.Errorf(
			"insert product with SKU %q: %w",
			product.SKU,
			ErrSKUAlreadyExists,
		)
	}

	return fmt.Errorf("insert product: %w", err)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Product, error) {
	const query = `
		SELECT
			id,
			sku,
			name,
			price_minor,
			currency,
			available
		FROM products
		WHERE id = $1
	`

	var product Product

	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.PriceMinor,
		&product.Currency,
		&product.Available,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, fmt.Errorf("get product %s: %w", id, ErrProductNotFound)
		}

		return Product{}, fmt.Errorf("get product by id: %w", err)
	}

	return product, nil
}
