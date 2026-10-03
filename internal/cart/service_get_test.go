package cart

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/goodday0404/commercepay/internal/testutil"
	"github.com/google/uuid"
)

func TestServiceGetCart(t *testing.T) {
	cartID := uuid.New()

	item := CartItem{
		id:             uuid.New(),
		productID:      uuid.New(),
		quantity:       2,
		unitPriceMinor: 1499,
		currency:       "CAD",
	}

	expected := Cart{
		id:    cartID,
		items: []CartItem{item},
	}

	var (
		productsCalls int
		getByIDCalls  int
		gotGetByID    uuid.UUID
	)

	repo := &fakeCartRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (Cart, error) {
			getByIDCalls++
			gotGetByID = id
			return expected, nil
		},
	}

	products := &fakeProductCatalog{
		getProductFn: func(ctx context.Context, id uuid.UUID) (catalog.Product, error) {
			productsCalls++
			return catalog.Product{}, nil
		},
	}
	service := NewService(repo, products)
	ctx := testutil.TestContext(t)

	got, err := service.GetCart(ctx, cartID)
	if err != nil {
		t.Fatalf("get cart: %v", err)
	}

	if getByIDCalls != 1 {
		t.Fatalf("expected GetByID once, got %d", getByIDCalls)
	}

	if gotGetByID != cartID {
		t.Fatalf("expected cart ID %s, got %s", cartID, gotGetByID)
	}

	if got.ID() != cartID {
		t.Fatalf("expected cart ID %s, got %s", cartID, got.ID())
	}

	if len(got.Items()) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items()))
	}

	if productsCalls != 0 {
		t.Fatalf("expected Catalog not to be called, got %d", productsCalls)
	}
}

func TestServiceGetCart_NotFound(t *testing.T) {
	cartID := uuid.New()

	var getByIDCalls int

	repo := &fakeCartRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (Cart, error) {
			getByIDCalls++
			return Cart{}, fmt.Errorf("load cart: %w", ErrCartNotFound)
		},
	}

	products := &fakeProductCatalog{}
	service := NewService(repo, products)
	ctx := testutil.TestContext(t)

	_, err := service.GetCart(ctx, cartID)

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrCartNotFound) {
		t.Fatalf("expected ErrCartNotFound, got %v", err)
	}

	if getByIDCalls != 1 {
		t.Fatalf("expected GetByID once, got %d", getByIDCalls)
	}
}
