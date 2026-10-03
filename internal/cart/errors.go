package cart

import (
	"errors"
	"net/http"

	"github.com/goodday0404/commercepay/internal/platform/httpserver"
)

var (
	ErrCartNotFound       = errors.New("cart not found")
	ErrInvalidProductID   = errors.New("product ID is required")
	ErrInvalidQuantity    = errors.New("quantity must be greater than zero")
	ErrNegativeUnitPrice  = errors.New("unit price cannot be negative")
	ErrEmptyCurrency      = errors.New("currency is required")
	ErrProductUnavailable = errors.New("product is unavailable")
	ErrProductNotFound    = errors.New("product not found")
)

func (h *Handler) handleAddItemErrors(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidProductID),
		errors.Is(err, ErrInvalidQuantity):

		httpserver.WriteJSON(
			w,
			http.StatusBadRequest,
			strToErrMap(err.Error()),
		)

	case errors.Is(err, ErrCartNotFound),
		errors.Is(err, ErrProductNotFound):

		httpserver.WriteJSON(
			w,
			http.StatusNotFound,
			strToErrMap(err.Error()),
		)

	case errors.Is(err, ErrProductUnavailable):

		httpserver.WriteJSON(
			w,
			http.StatusConflict,
			strToErrMap(err.Error()),
		)

	default:
		httpserver.WriteJSON(
			w,
			http.StatusInternalServerError,
			strToErrMap("internal server error"),
		)
	}
}

func (h *Handler) handleGetCartError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrCartNotFound):
		httpserver.WriteJSON(
			w,
			http.StatusNotFound,
			strToErrMap("cart not found"),
		)

	default:
		httpserver.WriteJSON(
			w,
			http.StatusInternalServerError,
			strToErrMap("internal server error"),
		)
	}
}

func strToErrMap(errMsg string) map[string]string {
	return map[string]string{"error": errMsg}
}

// dcda2514-ff27-45fe-b5fa-5e4a01fd44a6

// 42df80ca-c275-4eaf-9522-b63ca572a094

// curl -i \
//   -X POST \
//   "http://localhost:8080/carts/42df80ca-c275-4eaf-9522-b63ca572a094/items" \
//   -H 'Content-Type: application/json' \
//   -d "{
//     \"product_id\": \"dcda2514-ff27-45fe-b5fa-5e4a01fd44a6\",
//     \"quantity\": 1
//   }"

//   {"id":"9bbfe236-788f-4f5d-b2ea-1f6549bf2d4e","product_id":"dcda2514-ff27-45fe-b5fa-5e4a01fd44a6","quantity":2,"unit_price_minor":1499,"currency":"CAD"}

//   SELECT
//     id,
//     cart_id,
//     product_id,
//     quantity,
//     unit_price_minor,
//     currency
// FROM cart_items
// WHERE cart_id = '42df80ca-c275-4eaf-9522-b63ca572a094'
//   AND product_id = 'dcda2514-ff27-45fe-b5fa-5e4a01fd44a6';

//   SELECT COUNT(*)
// FROM cart_items
// WHERE cart_id = '42df80ca-c275-4eaf-9522-b63ca572a094'
//   AND product_id = 'dcda2514-ff27-45fe-b5fa-5e4a01fd44a6';
