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

func (r *Repository) List(ctx context.Context, limit int, cursor *ProductCursor) (ProductPage, error) {
	const firstPageQuery = `
		SELECT
			id,
			sku,
			name,
			price_minor,
			currency,
			available,
			created_at
		FROM products
		ORDER BY created_at DESC, id DESC
		LIMIT $1
	`

	const nextPageQuery = `
		SELECT
			id,
			sku,
			name,
			price_minor,
			currency,
			available,
			created_at
		FROM products
		WHERE (created_at, id) < ($1, $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3
	`
	fetchLimit := limit + 1
	var rows pgx.Rows
	var err error

	if cursor == nil {
		rows, err = r.db.Query(ctx, firstPageQuery, fetchLimit)
	} else {
		rows, err = r.db.Query(ctx, nextPageQuery, cursor.CreatedAt, cursor.ID, fetchLimit)
	}

	defer rows.Close()

	if err != nil {
		return ProductPage{}, fmt.Errorf("list products: %w", err)
	}

	resultRows := make([]productListRow, 0, fetchLimit)

	for rows.Next() {
		var row productListRow

		if err := rows.Scan(
			&row.Product.ID,
			&row.Product.SKU,
			&row.Product.Name,
			&row.Product.PriceMinor,
			&row.Product.Currency,
			&row.Product.Available,
			&row.CreatedAt,
		); err != nil {
			return ProductPage{}, fmt.Errorf("scan product: %w", err)
		}

		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return ProductPage{}, fmt.Errorf("iterate products: %w", err)
	}

	hasMore := len(resultRows) > limit && limit > 0

	if hasMore {
		resultRows = resultRows[:limit]
	}

	products := make([]Product, 0, len(resultRows))

	for _, resultRow := range resultRows {
		products = append(products, resultRow.Product)
	}

	var nextProduct *ProductCursor

	if hasMore {
		lastProduct := resultRows[len(resultRows)-1]
		nextProduct = &ProductCursor{
			CreatedAt: lastProduct.CreatedAt,
			ID:        lastProduct.Product.ID,
		}
	}

	return ProductPage{
		Products:   products,
		NextCursor: nextProduct,
	}, nil
}
