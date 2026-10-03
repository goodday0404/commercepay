package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/google/uuid"
)

func TestServiceCreateCart(t *testing.T) {
	var (
		insertedCart Cart
		productCalls int
	)

	repo := &fakeCartRepository{
		insertFn: func(ctx context.Context, cart Cart) error {
			insertedCart = cart
			return nil
		},
	}

	products := &fakeProductCatalog{
		getProductFn: func(context.Context, uuid.UUID) (catalog.Product, error) {
			productCalls++
			return catalog.Product{}, nil
		},
	}

	service := NewService(repo, products)

	got, err := service.CreateCart(context.Background())
	if err != nil {
		t.Fatalf("create cart: %v", err)
	}

	if got.ID() == uuid.Nil {
		t.Fatal("expected created cart to have an ID")
	}

	if !got.IsEmpty() {
		t.Fatal("expected created cart to be empty")
	}

	if insertedCart.ID() != got.ID() {
		t.Fatalf(
			"expected inserted cart ID %s, got %s",
			got.ID(),
			insertedCart.ID(),
		)
	}

	if productCalls != 0 {
		t.Fatalf("expected Catalog not to be called, got %d calls", productCalls)
	}
}

func TestServiceCreateCart_InsertFails(t *testing.T) {
	insertErr := errors.New("database unavailable")
	var productCalls int

	repo := &fakeCartRepository{
		insertFn: func(ctx context.Context, cart Cart) error {
			return insertErr
		},
	}

	products := &fakeProductCatalog{
		getProductFn: func(context.Context, uuid.UUID) (catalog.Product, error) {
			productCalls++
			return catalog.Product{}, nil
		},
	}

	service := NewService(repo, products)

	got, err := service.CreateCart(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, insertErr) {
		t.Fatalf("expected error to wrap insert error, got %v", err)
	}

	if got.ID() != uuid.Nil {
		t.Fatalf("expected zero Cart on failure, got ID %s", got.ID())
	}

	if productCalls != 0 {
		t.Fatalf("expected Catalog not to be called, got %d calls", productCalls)
	}
}
