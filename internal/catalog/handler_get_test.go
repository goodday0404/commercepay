package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestHandlerGetProductReturnsBadRequestForInvalidID(t *testing.T) {
	serviceCalled := false

	service := &fakeCatalogService{
		getProductFn: func(ctx context.Context, id uuid.UUID) (Product, error) {
			serviceCalled = true
			return Product{}, nil
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products/not-a-uuid",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if serviceCalled {
		t.Fatal("expected service not to be called for invalid UUID")
	}
}

func TestHandlerGetProductMapsNotFoundTo404(t *testing.T) {
	productID := uuid.New()

	service := &fakeCatalogService{
		getProductFn: func(ctx context.Context, id uuid.UUID) (Product, error) {
			if id != productID {
				t.Fatalf("expected ID %s, got %s", productID, id)
			}

			return Product{}, ErrProductNotFound
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products/"+productID.String(),
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}

func TestHandlerGetProductReturnsProduct(t *testing.T) {
	productID := uuid.New()

	expected := Product{
		ID:         productID,
		SKU:        "TEST-MUG",
		Name:       "Test Mug",
		PriceMinor: 1499,
		Currency:   "CAD",
		Available:  true,
	}

	service := &fakeCatalogService{
		getProductFn: func(ctx context.Context, id uuid.UUID) (Product, error) {
			return expected, nil
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products/"+productID.String(),
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var response productResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.ID != productID.String() {
		t.Fatalf("expected ID %s, got %s", productID, response.ID)
	}
}
