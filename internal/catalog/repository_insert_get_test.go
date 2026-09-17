package catalog

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestRepositoryInsertAndGetByID(t *testing.T) {
	db := openTestDB(t)
	resetProducts(t, db)

	repo := NewRepository(db)
	product, err := NewProduct(
		"TEST-MUG",
		"Test Mug",
		1499,
		"CAD",
		true,
	)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	ctx := testContext(t)

	if err := repo.Insert(ctx, product); err != nil {
		t.Fatalf("insert product: %v", err)
	}

	got, err := repo.GetByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}

	if product.ID != got.ID {
		t.Fatalf("expected ID %s, got %s", product.ID, got.ID)
	}

	if product.SKU != got.SKU {
		t.Fatalf("expected SKU %s, got %s", product.SKU, got.SKU)
	}

	if product.Name != got.Name {
		t.Fatalf("expected name %s, got %s", product.Name, got.Name)
	}

	if product.PriceMinor != 1499 {
		t.Fatalf("expected price minor %d, got %d", product.PriceMinor, got.PriceMinor)
	}

	if product.Currency != "CAD" {
		t.Fatalf("expected currency %s, got %s", product.Currency, got.Currency)
	}

	if !product.Available {
		t.Fatalf("expected available %t, got %t", product.Available, got.Available)
	}
}

func TestRepositoryInsertReturnsSKUAlreadyExists(t *testing.T) {
	db := openTestDB(t)
	resetProducts(t, db)
	repo := NewRepository(db)

	product, err := NewProduct(
		"TEST-MUG",
		"Test Mug",
		1499,
		"CAD",
		true,
	)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	dupSKU, err := NewProduct(
		"TEST-MUG",
		"Test Mug",
		1499,
		"CAD",
		true,
	)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	ctx := testContext(t)

	if err := repo.Insert(ctx, product); err != nil {
		t.Fatalf("insert product: %v", err)
	}

	if err := repo.Insert(ctx, dupSKU); !errors.Is(err, ErrSKUAlreadyExists) {
		t.Fatalf("expected %v, got %v", ErrSKUAlreadyExists, err)
	}
}

func TestRepositoryGetByIDReturnsProductNotFound(t *testing.T) {
	db := openTestDB(t)
	resetProducts(t, db)
	repo := NewRepository(db)

	ctx := testContext(t)

	_, err := repo.GetByID(ctx, uuid.New())
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}
