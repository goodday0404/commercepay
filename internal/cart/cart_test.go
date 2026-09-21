package cart

import (
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
