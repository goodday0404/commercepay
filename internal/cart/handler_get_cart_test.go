package cart

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestHandlerGetCart(t *testing.T) {
	cartID := uuid.New()

	item1 := CartItem{
		id:             uuid.New(),
		productID:      uuid.New(),
		quantity:       2,
		unitPriceMinor: 1499,
		currency:       "CAD",
	}

	item2 := CartItem{
		id:             uuid.New(),
		productID:      uuid.New(),
		quantity:       1,
		unitPriceMinor: 2599,
		currency:       "CAD",
	}

	getCartResult := Cart{
		id:    cartID,
		items: []CartItem{item1, item2},
	}

	var (
		getCartCalls int
		gotGetCartID uuid.UUID
	)

	service := &fakeCartService{
		getCartFn: func(ctx context.Context, id uuid.UUID) (Cart, error) {
			getCartCalls++
			gotGetCartID = id
			return getCartResult, nil
		},
	}

	router := newTestRouter(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/carts/"+cartID.String(),
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	if getCartCalls != 1 {
		t.Fatalf("expected GetCart once, got %d", getCartCalls)
	}

	if gotGetCartID != cartID {
		t.Fatalf("expected cart ID %s, got %s", cartID, gotGetCartID)
	}

	var got cartResponse

	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != cartID {
		t.Fatalf("expected cart ID %s, got %s", cartID, got.ID)
	}

	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got.Items))
	}

	gotByID := make(map[uuid.UUID]cartItemResponse, len(got.Items))

	for _, item := range got.Items {
		gotByID[item.ID] = item
	}

	got1, ok := gotByID[item1.ID()]
	if !ok {
		t.Fatalf("expected item %s", item1.ID())
	}

	if got1.Quantity != 2 {
		t.Fatalf("expected quantity 2, got %d", got1.Quantity)
	}

	if got1.UnitPriceMinor != 1499 {
		t.Fatalf("expected price 1499, got %d", got1.UnitPriceMinor)
	}

	got2, ok := gotByID[item2.ID()]
	if !ok {
		t.Fatalf("expected item %s", item2.ID())
	}

	if got2.Quantity != 1 {
		t.Fatalf("expected quantity 1, got %d", got2.Quantity)
	}

	if got2.UnitPriceMinor != 2599 {
		t.Fatalf("expected price 2599, got %d", got2.UnitPriceMinor)
	}
}

func TestHandlerGetCart_EmptyCart(t *testing.T) {
	cartID := uuid.New()

	getCartResult := Cart{
		id:    cartID,
		items: nil,
	}

	service := &fakeCartService{
		getCartFn: func(ctx context.Context, id uuid.UUID) (Cart, error) {
			return getCartResult, nil
		},
	}

	router := newTestRouter(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/carts/"+cartID.String(),
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusOK,
			rec.Code,
			rec.Body.String(),
		)
	}

	var got cartResponse

	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Items == nil {
		t.Fatal("expected empty items array, got nil")
	}

	if len(got.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(got.Items))
	}
}

func TestHandlerGetCart_InvalidCartID(t *testing.T) {
	var getCartCalls int

	service := &fakeCartService{
		getCartFn: func(ctx context.Context, id uuid.UUID) (Cart, error) {
			getCartCalls++
			return Cart{}, nil
		},
	}

	router := newTestRouter(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/carts/not-a-uuid",
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if getCartCalls != 0 {
		t.Fatalf("expected GetCart not to be called, got %d calls", getCartCalls)
	}
}

func TestHandlerGetCart_NotFound(t *testing.T) {
	cartID := uuid.New()

	var getCartCalls int

	service := &fakeCartService{
		getCartFn: func(ctx context.Context, id uuid.UUID) (Cart, error) {
			getCartCalls++
			return Cart{}, fmt.Errorf("get cart: %w", ErrCartNotFound)
		},
	}

	router := newTestRouter(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/carts/"+cartID.String(),
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusNotFound,
			rec.Code,
			rec.Body.String(),
		)
	}

	if getCartCalls != 1 {
		t.Fatalf("expected GetCart once, got %d", getCartCalls)
	}
}

func TestHandlerGetCart_InternalError(t *testing.T) {
	cartID := uuid.New()

	internalErr := errors.New("database connection reset")

	service := &fakeCartService{
		getCartFn: func(ctx context.Context, id uuid.UUID) (Cart, error) {
			return Cart{}, errors.New("internal error")
		},
	}

	router := newTestRouter(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/carts/"+cartID.String(),
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}

	if strings.Contains(rec.Body.String(), internalErr.Error()) {
		t.Fatalf("response exposed internal error: %s", rec.Body.String())
	}
}
