package cart

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewCart(t *testing.T) {
	cart := NewCart()

	if cart.ID() == uuid.Nil {
		t.Fatal("expected new cart to have an ID")
	}

	if !cart.IsEmpty() {
		t.Fatal("expected new cart to be empty")
	}
}

func TestNewCartItem(t *testing.T) {
	productID := uuid.New()

	item, err := newCartItem(productID, 2, 1499, "CAD")
	if err != nil {
		t.Fatalf("new cart item: %v", err)
	}

	if item.ID() == uuid.Nil {
		t.Fatal("expected item ID")
	}

	if item.ProductID() != productID {
		t.Fatalf("expected product ID %s, got %s", productID, item.ProductID())
	}

	if item.Quantity() != 2 {
		t.Fatalf("expected quantity 2, got %d", item.Quantity())
	}

	if item.UnitPriceMinor() != 1499 {
		t.Fatalf("expected unit price 1499, got %d", item.UnitPriceMinor())
	}

	if item.Currency() != "CAD" {
		t.Fatalf("expected CAD, got %q", item.Currency())
	}
}

func TestNewCartItem_InvalidQuantity(t *testing.T) {
	_, err := newCartItem(uuid.New(), 0, 1499, "CAD")

	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("expected ErrInvalidQuantity, got %v", err)
	}
}

func TestNewCartItem_InvalidProdutID(t *testing.T) {
	_, err := newCartItem(uuid.Nil, 2, 1499, "CAD")

	if !errors.Is(err, ErrInvalidProductID) {
		t.Fatalf("expected ErrInvalidProductID, got %v", err)
	}
}

func TestNewCartItem_NegativeUnitPrice(t *testing.T) {
	_, err := newCartItem(uuid.New(), 2, -1499, "CAD")

	if !errors.Is(err, ErrNegativeUnitPrice) {
		t.Fatalf("expected ErrNegativeUnitPrice, got %v", err)
	}
}

func TestNewCartItem_EmptyCurrency(t *testing.T) {
	_, err := newCartItem(uuid.New(), 2, 1499, "")

	if !errors.Is(err, ErrEmptyCurrency) {
		t.Fatalf("expected ErrInvalidProductID, got %v", err)
	}
}
