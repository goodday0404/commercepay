package cart

import (
	"testing"

	"github.com/google/uuid"

	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/goodday0404/commercepay/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

func insertProduct(t *testing.T, repo *catalog.Repository,
	SKU string, name string, priceMinor int64, currency string, available bool,
) catalog.Product {
	product, err := catalog.NewProduct(
		SKU,
		name,
		priceMinor,
		currency,
		available,
	)
	if err != nil {
		t.Fatalf("new product: %v", err)
	}

	ctx := testutil.TestContext(t)

	if err := repo.Insert(ctx, product); err != nil {
		t.Fatalf("insert product fixture: %v", err)
	}

	return product
}

func insertCartItem(t *testing.T, repo *Repository, cartID uuid.UUID, item CartItem) CartItem {
	ctx := testutil.TestContext(t)

	persisted, err := repo.AddItem(ctx, cartID, item)
	if err != nil {
		t.Fatalf("add item 1: %v", err)
	}

	return persisted
}

// func insertProduct(t *testing.T, repo *catalog.Repository) catalog.Product {
// 	product, err := catalog.NewProduct(
// 		"CART-ITEM-TEST",
// 		"Cart Item Test Product",
// 		1499,
// 		"CAD",
// 		true,
// 	)
// 	if err != nil {
// 		t.Fatalf("new product: %v", err)
// 	}

// 	ctx := testutil.TestContext(t)

// 	if err := repo.Insert(ctx, product); err != nil {
// 		t.Fatalf("insert product fixture: %v", err)
// 	}

// 	return product
// }

func setTestCartandProduct(t *testing.T, db *pgxpool.Pool) (*Repository, *catalog.Repository, Cart, catalog.Product) {
	cartRepo := NewRepository(db)
	catalogRepo := catalog.NewRepository(db)

	ctx := testutil.TestContext(t)

	// ---------------------------------------------------------
	// Arrange: create a real Product.
	// ---------------------------------------------------------

	product := insertProduct(
		t,
		catalogRepo,
		"CART-ITEM-TEST",
		"Cart Item Test Product",
		1499,
		"CAD",
		true,
	)

	// ---------------------------------------------------------
	// Arrange: create a real Cart.
	// ---------------------------------------------------------

	cart := NewCart()

	if err := cartRepo.Insert(ctx, cart); err != nil {
		t.Fatalf("insert cart fixture: %v", err)
	}

	return cartRepo, catalogRepo, cart, product
}
