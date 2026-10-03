package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pgxForeignKeyViolation = "23503"
	cartItemsCartIDFkey    = "cart_items_cart_id_fkey"
	cartItemsProductIDFkey = "cart_items_product_id_fkey"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Insert(ctx context.Context, cart Cart) error {
	const query = `
		INSERT INTO carts (id)
		VALUES ($1)
	`

	if _, err := r.db.Exec(ctx, query, cart.id); err != nil {
		return fmt.Errorf("insert cart: %w", err)
	}

	return nil
}

func (r *Repository) AddItem(ctx context.Context, cartID uuid.UUID, item CartItem) (CartItem, error) {
	const query = `
		INSERT INTO cart_items (
			id,
			cart_id,
			product_id,
			quantity,
			unit_price_minor,
			currency
		)
		VALUES ($1, $2, $3, $4, $5, $6)

		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET
			quantity = cart_items.quantity + EXCLUDED.quantity,

			unit_price_minor = EXCLUDED.unit_price_minor,

			currency = EXCLUDED.currency

		RETURNING
			id,
			product_id,
			quantity,
			unit_price_minor,
			currency
	`

	var persisted CartItem

	err := r.db.QueryRow(
		ctx,
		query,
		item.id,
		cartID,
		item.productID,
		item.quantity,
		item.unitPriceMinor,
		item.currency,
	).Scan(
		&persisted.id,
		&persisted.productID,
		&persisted.quantity,
		&persisted.unitPriceMinor,
		&persisted.currency,
	)

	if err == nil {
		return persisted, nil
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == pgxForeignKeyViolation {
		switch pgErr.ConstraintName {
		case cartItemsCartIDFkey:
			return CartItem{}, fmt.Errorf("%w: %s", ErrCartNotFound, cartID)

		case cartItemsProductIDFkey:
			// Product existed when Catalog was read,
			// but disappeared before this INSERT.
			//
			// Keep this as an infrastructure-level
			// persistence failure for now. Sprint 2.14
			// will finish cross-module error vocabulary.
		}
	}

	return CartItem{}, fmt.Errorf("add cart item: %w", err)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Cart, error) {
	const query = `
		SELECT
			c.id,
			ci.id,
			ci.product_id,
			ci.quantity,
			ci.unit_price_minor,
			ci.currency
		FROM carts AS c
		LEFT JOIN cart_items AS ci
			ON ci.cart_id = c.id
		WHERE c.id = $1
		ORDER BY ci.id
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return Cart{}, fmt.Errorf("query cart: %w", err)
	}

	defer rows.Close()

	var (
		cart  Cart
		found bool
	)

	for rows.Next() {
		var (
			cartID         uuid.UUID
			itemID         pgtype.UUID
			productID      pgtype.UUID
			quantity       pgtype.Int4
			unitPriceMinor pgtype.Int8
			currency       pgtype.Text
		)

		if err := rows.Scan(
			&cartID,
			&itemID,
			&productID,
			&quantity,
			&unitPriceMinor,
			&currency,
		); err != nil {
			return Cart{}, fmt.Errorf("scan cart row: %w", err)
		}

		if !found {
			cart = Cart{
				id:    cartID,
				items: make([]CartItem, 0),
			}

			found = true
		}

		if !itemID.Valid {
			continue
		}

		if !productID.Valid ||
			!quantity.Valid ||
			!unitPriceMinor.Valid ||
			!currency.Valid {

			return Cart{}, fmt.Errorf("scan cart row: incomplete cart item")
		}

		item := CartItem{
			id:             uuid.UUID(itemID.Bytes),
			productID:      uuid.UUID(productID.Bytes),
			quantity:       int(quantity.Int32),
			unitPriceMinor: unitPriceMinor.Int64,
			currency:       currency.String,
		}

		cart.items = append(cart.items, item)
	}

	if err := rows.Err(); err != nil {
		return Cart{}, fmt.Errorf("iterate cart rows: %w", err)
	}

	if !found {
		return Cart{}, fmt.Errorf("%w: %s", ErrCartNotFound, id)
	}

	return cart, nil
}
