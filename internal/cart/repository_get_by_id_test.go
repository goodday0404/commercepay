package cart

import (
	"errors"
	"testing"

	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/goodday0404/commercepay/internal/testutil"
	"github.com/google/uuid"
)

func TestRepositoryGetByID_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)
	repo := NewRepository(db)

	ctx := testutil.TestContext(t)

	missingCartID := uuid.New()

	_, err := repo.GetByID(ctx, missingCartID)

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrCartNotFound) {
		t.Fatalf("expected ErrCartNotFound, got %v", err)
	}
}

func TestRepositoryGetByID_EmptyCart(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)
	repo := NewRepository(db)

	cart := NewCart()

	ctx := testutil.TestContext(t)

	if err := repo.Insert(ctx, cart); err != nil {
		t.Fatalf("insert cart fixture: %v", err)
	}

	got, err := repo.GetByID(ctx, cart.ID())
	if err != nil {
		t.Fatalf("get cart: %v", err)
	}

	if got.ID() != cart.ID() {
		t.Fatalf("expected cart ID %s, got %s", cart.ID(), got.ID())
	}

	if !got.IsEmpty() {
		t.Fatal("expected loaded cart to be empty")
	}

	items := got.Items()

	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestRepositoryGetByID_WithItems(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)

	cartRepo := NewRepository(db)
	catalogRepo := catalog.NewRepository(db)

	// ---------------------------------------------------------
	// Product 1 & 2
	// ---------------------------------------------------------

	product1 := insertProduct(
		t,
		catalogRepo,
		"CART-LOAD-1-"+uuid.NewString(),
		"Cart Load Product 1",
		1499,
		"CAD",
		true,
	)

	product2 := insertProduct(
		t,
		catalogRepo,
		"CART-LOAD-2-"+uuid.NewString(),
		"Cart Load Product 2",
		2599,
		"CAD",
		true,
	)

	// ---------------------------------------------------------
	// Cart
	// ---------------------------------------------------------

	cart := NewCart()
	ctx := testutil.TestContext(t)

	if err := cartRepo.Insert(ctx, cart); err != nil {
		t.Fatalf("insert cart fixture: %v", err)
	}

	// ---------------------------------------------------------
	// CartItem 1 & 2
	// ---------------------------------------------------------

	item1, err := newCartItem(
		product1.ID,
		2,
		1499,
		"CAD",
	)
	if err != nil {
		t.Fatalf("new cart item 1: %v", err)
	}

	item2, err := newCartItem(
		product2.ID,
		1,
		2599,
		"CAD",
	)
	if err != nil {
		t.Fatalf(
			"new cart item 2: %v",
			err,
		)
	}

	persisted1 := insertCartItem(t, cartRepo, cart.ID(), item1)
	persisted2 := insertCartItem(t, cartRepo, cart.ID(), item2)

	// ---------------------------------------------------------
	// Act
	// ---------------------------------------------------------

	got, err := cartRepo.GetByID(ctx, cart.ID())
	if err != nil {
		t.Fatalf("get cart: %v", err)
	}

	// ---------------------------------------------------------
	// Assert root
	// ---------------------------------------------------------

	if got.ID() != cart.ID() {
		t.Fatalf("expected cart ID %s, got %s", cart.ID(), got.ID())
	}

	items := got.Items()

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Do not make the test depend on UUID ordering.
	itemsByID := make(map[uuid.UUID]CartItem, len(items))

	for _, item := range items {
		itemsByID[item.ID()] = item
	}

	// ---------------------------------------------------------
	// Assert Item 1
	// ---------------------------------------------------------

	got1, ok := itemsByID[persisted1.ID()]
	if !ok {
		t.Fatalf("expected item %s", persisted1.ID())
	}

	if got1.ProductID() != product1.ID {
		t.Fatalf("expected product ID %s, got %s", product1.ID, got1.ProductID())
	}

	if got1.Quantity() != 2 {
		t.Fatalf("expected quantity 2, got %d", got1.Quantity())
	}

	if got1.UnitPriceMinor() != 1499 {
		t.Fatalf("expected price 1499, got %d", got1.UnitPriceMinor())
	}

	if got1.Currency() != "CAD" {
		t.Fatalf("expected CAD, got %q", got1.Currency())
	}

	// ---------------------------------------------------------
	// Assert Item 2
	// ---------------------------------------------------------

	got2, ok := itemsByID[persisted2.ID()]
	if !ok {
		t.Fatalf("expected item %s", persisted2.ID())
	}

	if got2.ProductID() != product2.ID {
		t.Fatalf("expected product ID %s, got %s", product2.ID, got2.ProductID())
	}

	if got2.Quantity() != 1 {
		t.Fatalf("expected quantity 1, got %d", got2.Quantity())
	}

	if got2.UnitPriceMinor() != 2599 {
		t.Fatalf("expected price 2599, got %d", got2.UnitPriceMinor())
	}

	if got2.Currency() != "CAD" {
		t.Fatalf("expected CAD, got %q", got2.Currency())
	}
}
