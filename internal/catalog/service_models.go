package catalog

import (
	"time"

	"github.com/google/uuid"
)

type CreateProductInput struct {
	SKU        string
	Name       string
	PriceMinor int64
	Currency   string
	Available  bool
}

type ListProductsInput struct {
	Limit int
	After *ProductCursor
}

type ProductCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type ProductPage struct {
	Products   []Product
	NextCursor *ProductCursor
}
