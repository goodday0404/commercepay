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

func TestHandlerAddItem(t *testing.T) {
	cartID := uuid.New()
	productID := uuid.New()

	result, err := newCartItem(
		productID,
		3,
		1599,
		"CAD",
	)
	if err != nil {
		t.Fatalf("new cart item: %v", err)
	}

	var (
		addItemCalls int
		gotCartID    uuid.UUID
		gotProductID uuid.UUID
		gotQuantity  int
	)

	service := &fakeCartService{
		addItemFn: func(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
			addItemCalls++
			gotCartID = cartID
			gotProductID = productID
			gotQuantity += quantity
			return result, nil
		},
	}

	router := newTestRouter(service)

	body := `{
		"product_id": "` + productID.String() + `",
		"quantity": 1
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/carts/"+cartID.String()+"/items",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
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

	if addItemCalls != 1 {
		t.Fatalf("expected AddItem to be called once, got %d", addItemCalls)
	}

	if gotCartID != cartID {
		t.Fatalf("expected cart ID %s, got %s", cartID, gotCartID)
	}

	if gotProductID != productID {
		t.Fatalf(
			"expected product ID %s, got %s",
			productID,
			gotProductID,
		)
	}

	if gotQuantity != 1 {
		t.Fatalf("expected requested quantity 1, got %d", gotQuantity)
	}

	var got cartItemResponse

	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != result.ID() {
		t.Fatalf("expected item ID %s, got %s", result.ID(), got.ID)
	}

	if got.ProductID != productID {
		t.Fatalf(
			"expected product ID %s, got %s", productID, got.ProductID,
		)
	}

	// ---------------------------------------------------------
	// Assert: udating correct quantity.
	//
	// Request quantity was 1.
	// Service result quantity is 3.
	//
	// Handler must return the resulting CartItem,
	// not echo the request.
	// ---------------------------------------------------------

	if got.Quantity != 3 {
		t.Fatalf("expected resulting quantity 3, got %d", got.Quantity)
	}

	if got.UnitPriceMinor != 1599 {
		t.Fatalf("expected unit price 1599, got %d", got.UnitPriceMinor)
	}

	if got.Currency != "CAD" {
		t.Fatalf("expected currency CAD, got %q", got.Currency)
	}

	contentType := rec.Header().Get("Content-Type")

	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}
}

func TestHandlerAddItem_InvalidCartID(t *testing.T) {
	var addItemCalls int

	service := &fakeCartService{
		addItemFn: func(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
			addItemCalls++
			return CartItem{}, nil
		},
	}

	router := newTestRouter(service)

	productID := uuid.New()

	body := `{
		"product_id": "` + productID.String() + `",
		"quantity": 1
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/carts/not-a-uuid/items",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if addItemCalls != 0 {
		t.Fatalf("expected AddItem not to be called, got %d calls", addItemCalls)
	}
}

func TestHandlerAddItem_InvalidJSON(t *testing.T) {
	var addItemCalls int

	service := &fakeCartService{
		addItemFn: func(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
			addItemCalls++
			return CartItem{}, nil
		},
	}

	router := newTestRouter(service)

	cartID := uuid.New()

	req := httptest.NewRequest(
		http.MethodPost,
		"/carts/"+cartID.String()+"/items",
		strings.NewReader(`{
			"product_id":
		`),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if addItemCalls != 0 {
		t.Fatalf("expected AddItem not to be called, got %d calls", addItemCalls)
	}
}

func TestHandlerAddItem_InvalidQuantity(t *testing.T) {
	cartID := uuid.New()
	productID := uuid.New()

	var (
		addItemCalls int
		gotQuantity  int
	)

	service := &fakeCartService{
		addItemFn: func(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
			addItemCalls++
			return CartItem{}, ErrInvalidQuantity
		},
	}

	router := newTestRouter(service)

	body := `{
		"product_id": "` + productID.String() + `",
		"quantity": 0
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/carts/"+cartID.String()+"/items",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if addItemCalls != 1 {
		t.Fatalf("expected AddItem to be called once, got %d", addItemCalls)
	}

	if gotQuantity != 0 {
		t.Fatalf("expected quantity 0 to reach Service, got %d", gotQuantity)
	}
}

func TestHandlerAddItem_InvalidProductID(t *testing.T) {
	cartID := uuid.New()

	var (
		addItemCalls int
		gotProductID uuid.UUID
	)

	service := &fakeCartService{
		addItemFn: func(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
			addItemCalls++
			return CartItem{}, ErrInvalidProductID
		},
	}

	router := newTestRouter(service)

	body := `{
		"product_id": "00000000-0000-0000-0000-000000000000",
		"quantity": 1
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/carts/"+cartID.String()+"/items",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if addItemCalls != 1 {
		t.Fatalf("expected AddItem to be called once, got %d", addItemCalls)
	}

	if gotProductID != uuid.Nil {
		t.Fatalf("expected nil Product ID, got %s", gotProductID)
	}
}

func TestHandlerAddItem_CartNotFound(t *testing.T) {
	cartID := uuid.New()
	productID := uuid.New()

	var addItemCalls int

	service := &fakeCartService{
		addItemFn: func(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
			addItemCalls++
			return CartItem{}, fmt.Errorf("add cart item: %w", ErrCartNotFound)
		},
	}

	router := newTestRouter(service)

	body := `{
		"product_id": "` + productID.String() + `",
		"quantity": 1
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/carts/"+cartID.String()+"/items",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
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

	if addItemCalls != 1 {
		t.Fatalf("expected AddItem to be called once, got %d", addItemCalls)
	}
}

func TestHandlerAddItem_ProductNotFound(t *testing.T) {
	cartID := uuid.New()
	productID := uuid.New()

	var addItemCalls int

	service := &fakeCartService{
		addItemFn: func(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
			addItemCalls++
			return CartItem{}, fmt.Errorf("get product: %w", ErrProductNotFound)
		},
	}

	router := newTestRouter(service)

	body := `{
		"product_id": "` + productID.String() + `",
		"quantity": 1
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/carts/"+cartID.String()+"/items",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
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

	if addItemCalls != 1 {
		t.Fatalf("expected AddItem to be called once, got %d", addItemCalls)
	}
}

func TestHandlerAddItem_ProductUnavailable(t *testing.T) {
	cartID := uuid.New()
	productID := uuid.New()

	var addItemCalls int

	service := &fakeCartService{
		addItemFn: func(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
			addItemCalls++
			return CartItem{}, fmt.Errorf("validate product: %w", ErrProductUnavailable)
		},
	}

	router := newTestRouter(service)

	body := `{
		"product_id": "` + productID.String() + `",
		"quantity": 1
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/carts/"+cartID.String()+"/items",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusConflict,
			rec.Code,
			rec.Body.String(),
		)
	}

	if addItemCalls != 1 {
		t.Fatalf("expected AddItem to be called once, got %d", addItemCalls)
	}
}

func TestHandlerAddItem_InternalError(t *testing.T) {
	cartID := uuid.New()
	productID := uuid.New()

	internalErr := errors.New("database unavailable: dial tcp 10.0.0.7:5432")

	var addItemCalls int

	service := &fakeCartService{
		addItemFn: func(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
			addItemCalls++
			return CartItem{}, fmt.Errorf("validate product: %w", errors.New("internal server error"))
		},
	}

	router := newTestRouter(service)

	body := `{
		"product_id": "` + productID.String() + `",
		"quantity": 1
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/carts/"+cartID.String()+"/items",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}

	if addItemCalls != 1 {
		t.Fatalf("expected AddItem to be called once, got %d", addItemCalls)
	}

	responseBody := rec.Body.String()

	if strings.Contains(responseBody, internalErr.Error()) {
		t.Fatalf("response exposed internal error: %s", responseBody)
	}
}
