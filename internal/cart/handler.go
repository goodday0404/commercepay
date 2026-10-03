package cart

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/goodday0404/commercepay/internal/platform/httpserver"
	"github.com/google/uuid"
)

type cartService interface {
	CreateCart(ctx context.Context) (Cart, error)
	AddItem(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error)
	GetCart(ctx context.Context, id uuid.UUID) (Cart, error)
}

type Handler struct {
	service cartService
}

func NewHandler(service cartService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/carts", h.CreateCart)
	router.Post("/carts/{cart_id}/items", h.AddItem)
	router.Get("/carts/{cart_id}", h.GetCart)
}

func (h *Handler) CreateCart(w http.ResponseWriter, r *http.Request) {
	cart, err := h.service.CreateCart(r.Context())
	if err != nil {
		httpserver.WriteJSON(
			w,
			http.StatusInternalServerError,
			strToErrMap("internal server error"),
		)
		return
	}

	response := createCartResponse{ID: cart.id}
	httpserver.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	cartID, err := uuid.Parse(chi.URLParam(r, "cart_id"))
	if err != nil {
		httpserver.WriteJSON(
			w,
			http.StatusBadRequest,
			strToErrMap("invalid cart ID"),
		)
		return
	}

	var req addItemRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteJSON(
			w,
			http.StatusBadRequest,
			strToErrMap("invalid request body"),
		)
		return
	}

	item, err := h.service.AddItem(r.Context(), cartID, req.ProductID, req.Quantity)
	if err != nil {
		h.handleAddItemErrors(w, err)
		return
	}

	httpserver.WriteJSON(
		w,
		http.StatusOK,
		toCartItemResponse(item),
	)
}

func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	cartID, err := uuid.Parse(chi.URLParam(r, "cart_id"))
	if err != nil {
		httpserver.WriteJSON(
			w,
			http.StatusBadRequest,
			strToErrMap("invalid cart ID"),
		)
		return
	}

	cart, err := h.service.GetCart(r.Context(), cartID)
	if err != nil {
		h.handleGetCartError(w, err)
		return
	}

	httpserver.WriteJSON(
		w,
		http.StatusOK,
		toCartResponse(cart),
	)
}
