package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestHandlerCreateProductReturnsBadRequestForMalformedJSON(t *testing.T) {
	serviceCalled := false

	service := &fakeCatalogService{
		createProductFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			serviceCalled = true
			return Product{}, nil
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(http.MethodPost, "/admin/products", strings.NewReader(`{`))

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if serviceCalled {
		t.Fatal("expected service not to be called for malformed JSON")
	}
}

func TestHandlerCreateProductRequiresAvailable(t *testing.T) {
	serviceCalled := false

	service := &fakeCatalogService{
		createProductFn: func(tx context.Context, input CreateProductInput) (Product, error) {
			serviceCalled = true
			return Product{}, nil
		},
	}

	router := newTestRouter(service)

	body := `{
		"sku": "TEST-MUG",
		"name": "Test Mug",
		"price_minor": 1499,
		"currency": "CAD"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/products",
		strings.NewReader(body),
	)

	request.Header.Set("Content-Type", "application/json")

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
		t.Fatal("expected service not to be called when available is omitted")
	}
}

func TestHandlerCreateProductAcceptsExplicitFalseAvailability(t *testing.T) {
	var received CreateProductInput

	service := &fakeCatalogService{
		createProductFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			received = input

			return Product{
				ID:         uuid.New(),
				SKU:        input.SKU,
				Name:       input.Name,
				PriceMinor: input.PriceMinor,
				Currency:   input.Currency,
				Available:  input.Available,
			}, nil
		},
	}

	router := newTestRouter(service)

	body := `{
		"sku": "TEST-MUG",
		"name": "Test Mug",
		"price_minor": 1499,
		"currency": "CAD",
		"available": false
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/products",
		strings.NewReader(body),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	if received.Available {
		t.Fatal("expected service input Available to be false")
	}
}

func TestHandlerCreateProductMapsDuplicateSKUToConflict(t *testing.T) {
	service := &fakeCatalogService{
		createProductFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			return Product{}, ErrSKUAlreadyExists
		},
	}

	router := newTestRouter(service)

	body := `{
		"sku": "EXISTING-SKU",
		"name": "Another Product",
		"price_minor": 2000,
		"currency": "CAD",
		"available": true
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/products",
		strings.NewReader(body),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusConflict,
			recorder.Code,
		)
	}
}

func TestHandlerCreateProductMapsUnexpectedErrorToInternalServerError(t *testing.T) {
	internalErr := errors.New("database unavailable")

	service := &fakeCatalogService{
		createProductFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			return Product{}, internalErr
		},
	}

	router := newTestRouter(service)

	body := `{
		"sku": "TEST-MUG",
		"name": "Test Mug",
		"price_minor": 1499,
		"currency": "CAD",
		"available": true
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/products",
		strings.NewReader(body),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}

	if strings.Contains(recorder.Body.String(), internalErr.Error()) {
		t.Fatalf(
			"response exposed internal error: %q",
			recorder.Body.String(),
		)
	}
}

func TestHandlerCreateProductReturnsCreatedProduct(t *testing.T) {
	productID := uuid.New()

	service := &fakeCatalogService{
		createProductFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
			if input.SKU != "TEST-MUG" {
				t.Fatalf("expected SKU TEST-MUG, got %q", input.SKU)
			}

			if input.PriceMinor != 1499 {
				t.Fatalf("expected price 1499, got %d", input.PriceMinor)
			}

			return Product{
				ID:         productID,
				SKU:        input.SKU,
				Name:       input.Name,
				PriceMinor: input.PriceMinor,
				Currency:   input.Currency,
				Available:  input.Available,
			}, nil
		},
	}

	router := newTestRouter(service)

	body := `{
		"sku": "TEST-MUG",
		"name": "Test Mug",
		"price_minor": 1499,
		"currency": "CAD",
		"available": true
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/products",
		strings.NewReader(body),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	contentType := recorder.Header().Get("Content-Type")

	if contentType != "application/json" {
		t.Fatalf("expected application/json, got %q", contentType)
	}

	var response productResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.ID != productID.String() {
		t.Fatalf("expected ID %s, got %s", productID, response.ID)
	}

	if response.SKU != "TEST-MUG" {
		t.Fatalf("expected SKU TEST-MUG, got %q", response.SKU)
	}

	if response.PriceMinor != 1499 {
		t.Fatalf("expected price 1499, got %d", response.PriceMinor)
	}
}

func TestHandlerCreateProductMapsValidationErrorsToBadRequest(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "negative price",
			err:  ErrNegativePrice,
		},
		{
			name: "empty SKU",
			err:  ErrEmptySKU,
		},
		{
			name: "empty name",
			err:  ErrEmptyName,
		},
		{
			name: "empty currency",
			err:  ErrEmptyCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeCatalogService{
				createProductFn: func(ctx context.Context, input CreateProductInput) (Product, error) {
					return Product{}, tt.err
				},
			}

			router := newTestRouter(service)

			body := `{
				"sku": "TEST",
				"name": "Test",
				"price_minor": 100,
				"currency": "CAD",
				"available": true
			}`

			request := httptest.NewRequest(
				http.MethodPost,
				"/admin/products",
				strings.NewReader(body),
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
		})
	}
}
