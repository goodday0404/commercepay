package catalog

import (
	"strings"

	"github.com/google/uuid"
)

type Product struct {
	ID         uuid.UUID
	SKU        string
	Name       string
	PriceMinor int64
	Currency   string
	Available  bool
}

func NewProduct(SKU, name string, priceMinor int64, currency string, available bool) (Product, error) {
	if strings.TrimSpace(SKU) == "" {
		return Product{}, ErrEmptySKU
	}

	if strings.TrimSpace(name) == "" {
		return Product{}, ErrEmptyName
	}

	if priceMinor < 0 {
		return Product{}, ErrNegativePrice
	}

	if strings.TrimSpace(currency) == "" {
		return Product{}, ErrEmptyCurrency
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
