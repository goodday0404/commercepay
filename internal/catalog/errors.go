package catalog

import (
	"errors"
	"net/http"
)

var (
	ErrNegativePrice    = errors.New("product price cannot be negative")
	ErrProductNotFound  = errors.New("product not found")
	ErrSKUAlreadyExists = errors.New("product SKU already exists")
	ErrInvalidLimit     = errors.New("invalid product list limit")

	ErrEmptySKU      = errors.New("product SKU cannot be empty")
	ErrEmptyName     = errors.New("product name cannot be empty")
	ErrEmptyCurrency = errors.New("product currency cannot be empty")
)

func (h *Handler) handleCreateProductError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNegativePrice):
		http.Error(
			w,
			"product price cannot be negative",
			http.StatusBadRequest,
		)

	case errors.Is(err, ErrEmptySKU):
		http.Error(
			w,
			"product SKU cannot be empty string",
			http.StatusBadRequest,
		)

	case errors.Is(err, ErrEmptyName):
		http.Error(
			w,
			"product name cannot be empty string",
			http.StatusBadRequest,
		)

	case errors.Is(err, ErrEmptyCurrency):
		http.Error(
			w,
			"product currency cannot be empty string",
			http.StatusBadRequest,
		)

	case errors.Is(err, ErrSKUAlreadyExists):
		http.Error(
			w,
			"product SKU already exists",
			http.StatusConflict,
		)

	default:
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
	}
}

func (h *Handler) handleGetProductError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrProductNotFound):
		http.Error(
			w,
			"product not found",
			http.StatusNotFound,
		)

	default:
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
	}
}

func (h *Handler) handleListProductsError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, ErrInvalidLimit):
		http.Error(
			w,
			"limit must be between 1 and 100",
			http.StatusBadRequest,
		)

	default:
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
	}
}
