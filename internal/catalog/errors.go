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
)

func (h *Handler) handleCreateProductError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNegativePrice):
		http.Error(
			w,
			"product price cannot be negative",
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
