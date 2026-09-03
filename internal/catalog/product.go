package catalog

import (
	"errors"

	"github.com/google/uuid"
)

var ErrNegativePrice = errors.New("product price cannot be negative")

type Product struct {
	ID         uuid.UUID
	SKU        string
	Name       string
	PriceMinor int64
	Currency   string
	Available  bool
}

func NewProduct(SKU, name string, priceMinor int64, currency string, available bool) (Product, error) {
	if priceMinor < 0 {
		return Product{}, ErrNegativePrice
	}

	return Product{
		ID:         uuid.New(),
		SKU:        SKU,
		Name:       name,
		PriceMinor: priceMinor,
		Currency:   currency,
		Available:  available,
	}, nil
}
