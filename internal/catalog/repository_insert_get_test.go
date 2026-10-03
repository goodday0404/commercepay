package catalog

import (
	"errors"
	"testing"

	"github.com/goodday0404/commercepay/internal/testutil"
	"github.com/google/uuid"
)

func TestRepositoryInsertAndGetByID(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)

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

	ctx := testutil.TestContext(t)

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
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)
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

	ctx := testutil.TestContext(t)

	if err := repo.Insert(ctx, product); err != nil {
		t.Fatalf("insert product: %v", err)
	}

	if err := repo.Insert(ctx, dupSKU); !errors.Is(err, ErrSKUAlreadyExists) {
		t.Fatalf("expected %v, got %v", ErrSKUAlreadyExists, err)
	}
}

func TestRepositoryGetByIDReturnsProductNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)
	repo := NewRepository(db)

	ctx := testutil.TestContext(t)

	_, err := repo.GetByID(ctx, uuid.New())
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestRepositoryGetByIDs(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.ResetCatalogAndCart(t, db)
	repo := NewRepository(db)

	ctx := testutil.TestContext(t)

	product1, err := NewProduct(
		"BATCH-1-"+uuid.NewString(),
		"Batch Product 1",
		1499,
		"CAD",
		true,
	)
	if err != nil {
		t.Fatalf("new product 1: %v", err)
	}

	product2, err := NewProduct(
		"BATCH-2-"+uuid.NewString(),
		"Batch Product 2",
		2599,
		"CAD",
		true,
	)
	if err != nil {
		t.Fatalf("new product 2: %v", err)
	}

	product3, err := NewProduct(
		"BATCH-3-"+uuid.NewString(),
		"Batch Product 3",
		3999,
		"CAD",
		false,
	)
	if err != nil {
		t.Fatalf("new product 3: %v", err)
	}

	for _, product := range []Product{product1, product2, product3} {
		if err := repo.Insert(ctx, product); err != nil {
			t.Fatalf("insert product %s: %v", product.ID, err)
		}
	}

	got, err := repo.GetByIDs(
		ctx, []uuid.UUID{product1.ID, product2.ID},
	)
	if err != nil {
		t.Fatalf("get products by IDs: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 products, got %d", len(got))
	}

	gotByID := make(map[uuid.UUID]Product, len(got))

	for _, product := range got {
		gotByID[product.ID] = product
	}

	got1, ok := gotByID[product1.ID]
	if !ok {
		t.Fatalf("expected product %s", product1.ID)
	}

	if got1.Name != product1.Name {
		t.Fatalf("expected name %q, got %q", product1.Name, got1.Name)
	}

	if got1.PriceMinor != 1499 {
		t.Fatalf("expected price 1499, got %d", got1.PriceMinor)
	}

	got2, ok := gotByID[product2.ID]
	if !ok {
		t.Fatalf("expected product %s", product2.ID)
	}

	if got2.Name != product2.Name {
		t.Fatalf("expected name %q, got %q", product2.Name, got2.Name)
	}

	if _, ok := gotByID[product3.ID]; ok {
		t.Fatalf("did not expect unrelated product %s", product3.ID)
	}
}
