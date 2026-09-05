package catalog

import "errors"

var (
	ErrNegativePrice    = errors.New("product price cannot be negative")
	ErrProductNotFound  = errors.New("product not found")
	ErrSKUAlreadyExists = errors.New("product SKU already exists")
)
