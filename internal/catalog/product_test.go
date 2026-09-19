package catalog

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewProductCreatesProduct(t *testing.T) {
	product, err := NewProduct(
		"TSHIRT-BLK-M",
		"CommercePay T-Shirt",
		2999,
		"CAD",
		true,
	)
	if product.ID == uuid.Nil {
		t.Fatal("expected product ID to be generated")
	}

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if product.SKU != "TSHIRT-BLK-M" {
		t.Fatalf("expected SKU TSHIRT-BLK-M, got %s", product.SKU)
	}

	if product.Name != "CommercePay T-Shirt" {
		t.Fatalf("expected CommercePay T-Shirt, got %s", product.Name)
	}

	if product.PriceMinor != 2999 {
		t.Fatalf("expected price 2999, got %d", product.PriceMinor)
	}

	if product.Currency != "CAD" {
		t.Fatalf("expected CAD, got %s", product.Currency)
	}

	if !product.Available {
		t.Fatalf("expected true, got %v", product.Available)
	}
}

func TestNewProductRejectsNegativePrice(t *testing.T) {
	_, err := NewProduct(
		"TSHIRT-BLK-M",
		"CommercePay T-Shirt",
		-2999,
		"CAD",
		true,
	)

	if !errors.Is(err, ErrNegativePrice) {
		t.Fatalf("expected ErrNegativePrice, got %v", err)
	}
}
