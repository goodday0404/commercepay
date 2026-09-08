package catalog

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/goodday0404/commercepay/internal/platform/httpserver"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

type createProductRequest struct {
	SKU        string `json:"sku"`
	Name       string `json:"name"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
	Available  *bool  `json:"available"`
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

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/admin/products", h.CreateProduct)
	r.Get("/products/{id}", h.GetProduct)
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if req.Available == nil {
		http.Error(
			w,
			"available is required",
			http.StatusBadRequest,
		)
		return
	}

	product, err := h.service.CreateProduct(
		r.Context(),
		CreateProductInput{
			SKU:        req.SKU,
			Name:       req.Name,
			PriceMinor: req.PriceMinor,
			Currency:   req.Currency,
			Available:  *req.Available,
		},
	)
	if err != nil {
		h.handleCreateProductError(w, err)
		return
	}

	httpserver.WriteJSON(
		w,
		http.StatusCreated,
		newProductResponse(product),
	)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idText := chi.URLParam(r, "id")
	id, err := uuid.Parse(idText)
	if err != nil {
		http.Error(
			w,
			"invalid product id",
			http.StatusBadRequest,
		)
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		h.handleGetProductError(w, err)
		return
	}

	httpserver.WriteJSON(
		w,
		http.StatusOK,
		newProductResponse(product),
	)
}

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
