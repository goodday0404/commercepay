package catalog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestServiceCreateProductRejectsInvalidProductBeforePersistence(t *testing.T) {
	insertCalled := false

	repo := &fakeProductRepository{
		insertFn: func(ctx context.Context, product Product) error {
			insertCalled = true
			return nil
		},
	}

	service := NewService(repo, 2)

	_, err := service.CreateProduct(
		context.Background(),
		CreateProductInput{
			SKU:        "BAD-PRICE",
			Name:       "Bad Product",
			PriceMinor: -100,
			Currency:   "CAD",
			Available:  true,
		},
	)

	if !errors.Is(err, ErrNegativePrice) {
		t.Fatalf("expected ErrNegativePrice, got %v", err)
	}

	if insertCalled {
		t.Fatal("expected repository Insert not to be called")
	}
}

func TestServiceCreateProductPersistsAndReturnsProduct(t *testing.T) {
	var inserted Product

	repo := &fakeProductRepository{
		insertFn: func(ctx context.Context, product Product) error {
			inserted = product
			return nil
		},
	}

	service := NewService(repo, 2)

	got, err := service.CreateProduct(
		context.Background(),
		CreateProductInput{
			SKU:        "TEST-MUG",
			Name:       "Test Mug",
			PriceMinor: 1499,
			Currency:   "CAD",
			Available:  true,
		},
	)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	if got.ID == uuid.Nil {
		t.Fatal("expected generated product ID")
	}

	if inserted.ID != got.ID {
		t.Fatalf("expected inserted ID %s, got %s", got.ID, inserted.ID)
	}

	if inserted.SKU != "TEST-MUG" {
		t.Fatalf("expected inserted SKU %q, got %q", "TEST-MUG", inserted.SKU)
	}

	if inserted.Name != "Test Mug" {
		t.Fatalf("expected inserted name %q, got %q", "Test Mug", inserted.Name)
	}

	if inserted.PriceMinor != 1499 {
		t.Fatalf("expected inserted price 1499, got %d", inserted.PriceMinor)
	}

	if inserted.Currency != "CAD" {
		t.Fatalf("expected inserted currency CAD, got %q", inserted.Currency)
	}

	if !inserted.Available {
		t.Fatal("expected inserted Product to be available")
	}
}

func TestServiceCreateProductPreservesRepositoryError(t *testing.T) {
	repo := &fakeProductRepository{
		insertFn: func(ctx context.Context, product Product) error {
			return ErrSKUAlreadyExists
		},
	}

	service := NewService(repo, 2)

	_, err := service.CreateProduct(
		context.Background(),
		CreateProductInput{
			SKU:        "EXISTING-SKU",
			Name:       "Another Product",
			PriceMinor: 2000,
			Currency:   "CAD",
			Available:  true,
		},
	)

	if !errors.Is(err, ErrSKUAlreadyExists) {
		t.Fatalf("expected ErrSKUAlreadyExists, got %v", err)
	}
}

func TestServiceGetProductReturnsRepositoryProduct(t *testing.T) {
	expectedID := uuid.New()

	expected := Product{
		ID:         expectedID,
		SKU:        "TEST-SKU",
		Name:       "Test Product",
		PriceMinor: 1000,
		Currency:   "CAD",
		Available:  true,
	}

	var receivedID uuid.UUID

	repo := &fakeProductRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (Product, error) {
			receivedID = id
			return expected, nil
		},
	}

	service := NewService(repo, 2)

	got, err := service.GetProduct(context.Background(), expectedID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}

	if receivedID != expectedID {
		t.Fatalf("expected repository ID %s, got %s", expectedID, receivedID)
	}

	if got.ID != expected.ID {
		t.Fatalf("expected returned ID %s, got %s", expected.ID, got.ID)
	}
}

func TestServiceGetProductPreservesProductNotFound(t *testing.T) {
	repo := &fakeProductRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (Product, error) {
			return Product{}, ErrProductNotFound
		},
	}

	service := NewService(repo, 2)

	_, err := service.GetProduct(context.Background(), uuid.New())

	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestServiceListProductsRejectsInvalidLimitBeforeRepository(t *testing.T) {
	t.Helper()

	tests := []struct {
		name  string
		limit int
	}{
		{
			name:  "zero",
			limit: 0,
		},
		{
			name:  "negative",
			limit: -1,
		},
		{
			name:  "above maximum",
			limit: 101,
		},
	}

	var listCalled bool

	repo := &fakeProductRepository{
		listFn: func(ctx context.Context, limit int, after *ProductCursor) (ProductPage, error) {
			listCalled = true
			return ProductPage{}, nil
		},
	}

	service := NewService(repo, 2)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listCalled = false

			_, err := service.ListProducts(
				context.Background(),
				ListProductsInput{
					Limit: tt.limit,
				},
			)

			if !errors.Is(err, ErrInvalidLimit) {
				t.Fatalf("expected ErrInvalidLimit, got %v", err)
			}

			if listCalled {
				t.Fatal("expected repository List not to be called")
			}
		})
	}
}

func TestServiceListProductsPassesPaginationToRepository(t *testing.T) {
	t.Helper()

	cursor := &ProductCursor{
		CreatedAt: time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC),
		ID:        uuid.New(),
	}

	expectedPage := ProductPage{
		Products: []Product{
			{
				ID:         uuid.New(),
				SKU:        "P1",
				Name:       "Product 1",
				PriceMinor: 1000,
				Currency:   "CAD",
				Available:  true,
			},
		},
	}

	var (
		receivedLimit int
		receivedAfter *ProductCursor
	)

	repo := &fakeProductRepository{
		listFn: func(ctx context.Context, limit int, after *ProductCursor) (ProductPage, error) {
			receivedLimit = limit
			receivedAfter = after

			return expectedPage, nil
		},
	}

	const (
		expectLimit        = 2
		expectedNumProduct = 1
	)

	service := NewService(repo, 2)

	got, err := service.ListProducts(
		context.Background(),
		ListProductsInput{
			Limit: expectLimit,
			After: cursor,
		},
	)
	if err != nil {
		t.Fatalf("list products: %v", err)
	}

	if receivedLimit != expectLimit {
		t.Fatalf("expected limit %d, got %d", expectLimit, receivedLimit)
	}

	if receivedAfter == nil {
		t.Fatal("expected repository to receive cursor")
	}

	if receivedAfter.ID != cursor.ID {
		t.Fatalf("expected cursor ID %s, got %s", cursor.ID, receivedAfter.ID)
	}

	if !receivedAfter.CreatedAt.Equal(cursor.CreatedAt) {
		t.Fatalf("expected cursor time %s, got %s", cursor.CreatedAt, receivedAfter.CreatedAt)
	}

	if len(got.Products) != expectedNumProduct {
		t.Fatalf("expected %d product, got %d", expectedNumProduct, len(got.Products))
	}

	if got.Products[0].ID != expectedPage.Products[0].ID {
		t.Fatalf(
			"expected returned Product ID %s, got %s",
			expectedPage.Products[0].ID,
			got.Products[0].ID,
		)
	}
}

func TestServiceListProductsPreservesRepositoryError(t *testing.T) {
	repositoryErr := errors.New("repository unavailable")

	repo := &fakeProductRepository{
		listFn: func(ctx context.Context, limit int, after *ProductCursor) (ProductPage, error) {
			return ProductPage{}, repositoryErr
		},
	}

	service := NewService(repo, 2)

	_, err := service.ListProducts(context.Background(), ListProductsInput{Limit: 2})

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error to be preserved, got %v", err)
	}
}
