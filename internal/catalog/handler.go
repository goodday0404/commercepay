package catalog

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/goodday0404/commercepay/internal/platform/httpserver"
	"github.com/google/uuid"
)

const currentCursorVersion = 2

type Handler struct {
	service         *Service
	DefaultPageSize int
	MaxPageSize     int
}

func NewHandler(service *Service, defaultPageSize, maxPageSize int) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/admin/products", h.CreateProduct)
	r.Get("/products/{id}", h.GetProduct)
	r.Get("/products", h.ListProducts)
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

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, err := parseIntQueryParameter(r, "limit", 20)
	if err != nil {
		http.Error(
			w,
			"invalid limit",
			http.StatusBadRequest,
		)
		return
	}

	var after *ProductCursor

	cursorValue := r.URL.Query().Get("cursor")

	if cursorValue != "" {
		cursor, err := decodeProductCursor(cursorValue)
		if err != nil {
			http.Error(
				w,
				"invalid cursor",
				http.StatusBadRequest,
			)
			return
		}

		after = &cursor
	}

	input := ListProductsInput{
		Limit: limit,
		After: after,
	}

	page, err := h.service.ListProduct(r.Context(), input)
	if err != nil {
		h.handleListProductsError(w, err)
		return
	}

	response := listProductsResponse{
		Items: newProductResponses(page.Products),
	}

	if page.NextCursor != nil {
		encoded, err := encodeProductCursor(*page.NextCursor)
		if err != nil {
			http.Error(
				w,
				"internal server error",
				http.StatusInternalServerError,
			)
			return
		}

		response.NextCursor = &encoded
	}

	httpserver.WriteJSON(
		w,
		http.StatusOK,
		response,
	)
}

func encodeProductCursor(cursor ProductCursor) (string, error) {
	payload := productCursorPayload{
		Version:   currentCursorVersion,
		CreatedAt: cursor.CreatedAt,
		ID:        cursor.ID.String(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeProductCursor(value string) (ProductCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return ProductCursor{}, fmt.Errorf("decoe product cursor: %w", err)
	}

	var payload productCursorPayload

	if err := json.Unmarshal(data, &payload); err != nil {
		return ProductCursor{}, fmt.Errorf("unmarshal product cursor payload %w", err)
	}

	if payload.Version != currentCursorVersion {
		return ProductCursor{}, errors.New("unsupported cursor version")
	}

	id, err := uuid.Parse(payload.ID)
	if err != nil {
		return ProductCursor{}, fmt.Errorf("parse product cursor id: %w", err)
	}

	if payload.CreatedAt.IsZero() {
		return ProductCursor{}, errors.New("invalid cursor timestamp")
	}

	return ProductCursor{
		CreatedAt: payload.CreatedAt,
		ID:        id,
	}, nil
}

func parseIntQueryParameter(r *http.Request, name string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(name)

	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse query parameter: %w", err)
	}

	return parsed, nil
}

// total_count
// page_number
// previous_cursor
// page_count
