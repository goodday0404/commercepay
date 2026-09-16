package catalog

import (
	"strings"
	"time"

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

type createProductRequest struct {
	SKU        string `json:"sku"`
	Name       string `json:"name"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
	Available  *bool  `json:"available"`
}

type CreateProductInput struct {
	SKU        string
	Name       string
	PriceMinor int64
	Currency   string
	Available  bool
}

type productResponse struct {
	ID         string `json:"id"`
	SKU        string `json:"sku"`
	Name       string `json:"name"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
	Available  bool   `json:"available"`
}

func newProductResponse(product Product) productResponse {
	return productResponse{
		ID:         product.ID.String(),
		SKU:        product.SKU,
		Name:       product.Name,
		PriceMinor: product.PriceMinor,
		Currency:   product.Currency,
		Available:  product.Available,
	}
}

func newProductResponses(products []Product) []productResponse {
	productResponses := make([]productResponse, 0, len(products))

	for _, product := range products {
		productResponses = append(productResponses, newProductResponse(product))
	}

	return productResponses
}

type productCursorPayload struct {
	Version   int       `json:"v"`
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

type listProductsResponse struct {
	Items      []productResponse `json:"items"`
	NextCursor *string           `json:"next_cursor,omitempty"`
}

type ProductCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type ListProductsInput struct {
	Limit int
	After *ProductCursor
}

type ProductPage struct {
	Products   []Product
	NextCursor *ProductCursor
}

type productListRow struct {
	Product   Product
	CreatedAt time.Time
}
