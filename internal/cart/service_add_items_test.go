package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/goodday0404/commercepay/internal/testutil"
	"github.com/google/uuid"
)

func TestServiceAddFirstItem(t *testing.T) {
	cartID := uuid.New()

	product, err := catalog.NewProduct(
		"CART-PRODUCT",
		"Cart Product",
		1499,
		"CAD",
		true,
	)
	if err != nil {
		t.Fatalf("new product: %v", err)
	}

	var (
		insertedItem       CartItem
		insertedItemCartID uuid.UUID
	)

	repo := &fakeCartRepository{
		insertFn: func(ctx context.Context, cart Cart) error {
			return nil
		},

		insertItemsFn: func(ctx context.Context, cartID uuid.UUID, item CartItem) (CartItem, error) {
			insertedItemCartID = cartID
			insertedItem = item
			return item, nil
		},
	}

	products := &fakeProductCatalog{
		getProductFn: func(context.Context, uuid.UUID) (catalog.Product, error) {
			return product, nil
		},
	}

	service := NewService(repo, products)
	ctx := testutil.TestContext(t)

	item, err := service.AddItem(ctx, cartID, product.ID, 2)
	if err != nil {
		t.Fatalf("add first item: %v", err)
	}

	if item.ProductID() != product.ID {
		t.Fatalf("expected product ID %s, got %s", product.ID, item.ProductID())
	}

	if item.Quantity() != 2 {
		t.Fatalf("expected quantity 2, got %d", item.Quantity())
	}

	if item.UnitPriceMinor() != 1499 {
		t.Fatalf(
			"expected snapshot price 1499, got %d",
			item.UnitPriceMinor(),
		)
	}

	if item.Currency() != "CAD" {
		t.Fatalf(
			"expected CAD, got %s",
			item.Currency(),
		)
	}

	if insertedItemCartID != cartID {
		t.Fatalf(
			"expected cart ID %s, got %s",
			cartID,
			insertedItemCartID,
		)
	}

	if insertedItem.ID() != item.ID() {
		t.Fatalf(
			"expected persisted item %s, got %s",
			item.ID(),
			insertedItem.ID(),
		)
	}
}

func TestRepositoryInsertItem_CartNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)

	repo := NewRepository(db)
	ctx := testutil.TestContext(t)
	product, err := catalog.NewProduct(
		"CART-FK-TEST",
		"Cart FK Test Product",
		1499,
		"CAD",
		true,
	)
	if err != nil {
		t.Fatalf("new product: %v", err)
	}

	catalogRepo := catalog.NewRepository(db)

	if err := catalogRepo.Insert(ctx, product); err != nil {
		t.Fatalf("insert product fixture: %v", err)
	}

	item, err := newCartItem(
		product.ID,
		1,
		product.PriceMinor,
		product.Currency,
	)
	if err != nil {
		t.Fatalf("new cart item: %v", err)
	}

	missingCartID := uuid.New()

	_, err = repo.AddItem(ctx, missingCartID, item)

	if !errors.Is(err, ErrCartNotFound) {
		t.Fatalf("expected ErrCartNotFound, got %v", err)
	}
}

func TestServiceAddFirstItem_ProductUnavailable(t *testing.T) {
	var insertCall int

	product, err := catalog.NewProduct(
		"UNAVAILABLE",
		"Unavailable Product",
		1499,
		"CAD",
		false,
	)
	if err != nil {
		t.Fatalf("new product: %v", err)
	}

	repo := &fakeCartRepository{
		insertFn: func(ctx context.Context, cart Cart) error {
			insertCall++
			return nil
		},

		insertItemsFn: func(ctx context.Context, cartID uuid.UUID, item CartItem) (CartItem, error) {
			return CartItem{}, ErrProductUnavailable
		},
	}

	var productsCalls int

	products := &fakeProductCatalog{
		getProductFn: func(context.Context, uuid.UUID) (catalog.Product, error) {
			productsCalls++
			return catalog.Product{}, nil
		},
	}

	service := NewService(repo, products)

	_, err = service.AddItem(
		context.Background(),
		uuid.New(),
		product.ID,
		1,
	)

	if !errors.Is(err, ErrProductUnavailable) {
		t.Fatalf("expected ErrProductUnavailable, got %v", err)
	}

	if insertCall != 0 {
		t.Fatal(
			"expected unavailable product not to be inserted",
		)
	}
}

func TestServiceAddItem_ReturnsPersistedItem(t *testing.T) {
	cartID := uuid.New()

	product, err := catalog.NewProduct(
		"ADD-ITEM",
		"Add Item Product",
		1599,
		"CAD",
		true,
	)
	if err != nil {
		t.Fatalf("new product: %v", err)
	}

	existingID := uuid.New()

	persisted := CartItem{
		id:             existingID,
		productID:      product.ID,
		quantity:       3,
		unitPriceMinor: 1599,
		currency:       "CAD",
	}

	repo := &fakeCartRepository{
		insertItemsFn: func(ctx context.Context, cartID uuid.UUID, item CartItem) (CartItem, error) {
			return persisted, nil
		},
	}

	products := &fakeProductCatalog{
		getProductFn: func(ctx context.Context, id uuid.UUID) (catalog.Product, error) {
			return product, nil
		},
	}

	service := NewService(repo, products)

	got, err := service.AddItem(testutil.TestContext(t), cartID, product.ID, 1)
	if err != nil {
		t.Fatalf("add item: %v", err)
	}

	if got.ID() != existingID {
		t.Fatalf("expected existing item ID %s, got %s", existingID, got.ID())
	}

	if got.Quantity() != 3 {
		t.Fatalf("expected quantity 3, got %d", got.Quantity())
	}
}

// go test ./internal/cart -run TestRepositoryInsert -v
