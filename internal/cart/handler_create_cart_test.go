package cart

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerCreateCart(t *testing.T) {
	var created Cart
	var calls int

	service := &fakeCartService{
		createCartFn: func(ctx context.Context) (Cart, error) {
			created = NewCart()
			calls++
			return created, nil
		},
	}

	router := newTestRouter(service)

	req := httptest.NewRequest(http.MethodPost, "/carts", nil)

	recoder := httptest.NewRecorder()

	router.ServeHTTP(recoder, req)

	if recoder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recoder.Code,
		)
	}

	if calls != 1 {
		t.Fatalf("expected service to be called once, got %d", calls)
	}

	var got createCartResponse

	if err := json.NewDecoder(recoder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != created.ID() {
		t.Fatalf(
			"expected cart ID %s, got %s",
			created.ID(),
			got.ID,
		)
	}
}

func TestHandlerCreateCart_ServiceFails(t *testing.T) {
	var calls int
	internalErr := errors.New("database unavailable")

	service := &fakeCartService{
		createCartFn: func(ctx context.Context) (Cart, error) {
			calls++
			return Cart{}, internalErr
		},
	}

	router := newTestRouter(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/carts",
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

	if calls != 1 {
		t.Fatalf("expected service to be called once, got %d", calls)
	}

	if strings.Contains(rec.Body.String(), "database unavailable") {
		t.Fatal("response exposed internal error")
	}
}

func TestHandlerCreateCart_WrongMethod(t *testing.T) {
	service := &fakeCartService{}
	router := newTestRouter(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/carts",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}
