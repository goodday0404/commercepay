package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHandlerListProductsReturnsBadRequestForMalformedLimit(t *testing.T) {
	serviceCalled := false

	service := &fakeCatalogService{
		listProductsFn: func(ctx context.Context, input ListProductsInput) (ProductPage, error) {
			serviceCalled = true
			return ProductPage{}, nil
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products?limit=abc",
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
		t.Fatal("expected service not to be called for malformed limit")
	}
}

func TestHandlerListProductsReturnsBadRequestForInvalidCursor(t *testing.T) {
	serviceCalled := false

	service := &fakeCatalogService{
		listProductsFn: func(ctx context.Context, input ListProductsInput) (ProductPage, error) {
			serviceCalled = true
			return ProductPage{}, nil
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products?cursor=not-valid-base64",
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
		t.Fatal("expected service not to be called for invalid cursor")
	}
}

func TestHandlerListProductsUsesDefaultLimit(t *testing.T) {
	var received ListProductsInput

	service := &fakeCatalogService{
		listProductsFn: func(ctx context.Context, input ListProductsInput) (ProductPage, error) {
			received = input

			return ProductPage{
				Products: []Product{},
			}, nil
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products",
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

	if received.Limit != 20 {
		t.Fatalf("expected default limit 20, got %d", received.Limit)
	}

	if received.After != nil {
		t.Fatal("expected nil cursor for first page")
	}
}

func TestHandlerListProductsMapsInvalidLimitToBadRequest(t *testing.T) {
	const invalidLimitInput = 101

	service := &fakeCatalogService{
		listProductsFn: func(ctx context.Context, input ListProductsInput) (ProductPage, error) {
			if input.Limit != invalidLimitInput {
				t.Fatalf("expected limit 101, got %d", input.Limit)
			}

			return ProductPage{}, ErrInvalidLimit
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/products?limit=%d", invalidLimitInput),
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
}

func TestHandlerListProductsDecodesCursor(t *testing.T) {
	expectedCursor := ProductCursor{
		CreatedAt: time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC),
		ID:        uuid.New(),
	}

	encodedCursor, err := encodeProductCursor(expectedCursor)
	if err != nil {
		t.Fatalf("encode test cursor: %v", err)
	}

	var received ListProductsInput

	service := &fakeCatalogService{
		listProductsFn: func(ctx context.Context, input ListProductsInput) (ProductPage, error) {
			received = input

			return ProductPage{
				Products: []Product{},
			}, nil
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products?limit=10&cursor="+encodedCursor,
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

	if received.Limit != 10 {
		t.Fatalf("expected limit 10, got %d", received.Limit)
	}

	if received.After == nil {
		t.Fatal("expected decoded cursor")
	}

	if received.After.ID != expectedCursor.ID {
		t.Fatalf(
			"expected cursor ID %s, got %s",
			expectedCursor.ID,
			received.After.ID,
		)
	}

	if !received.After.CreatedAt.Equal(expectedCursor.CreatedAt) {
		t.Fatalf(
			"expected cursor time %s, got %s",
			expectedCursor.CreatedAt,
			received.After.CreatedAt,
		)
	}
}

func TestHandlerListProductsReturnsNextCursor(t *testing.T) {
	productID := uuid.New()

	next := &ProductCursor{
		CreatedAt: time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC),
		ID:        productID,
	}

	service := &fakeCatalogService{
		listProductsFn: func(ctx context.Context, input ListProductsInput) (ProductPage, error) {
			return ProductPage{
				Products: []Product{
					{
						ID:         productID,
						SKU:        "TEST-MUG",
						Name:       "Test Mug",
						PriceMinor: 1499,
						Currency:   "CAD",
						Available:  true,
					},
				},
				NextCursor: next,
			}, nil
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products?limit=1",
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

	var response listProductsResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode list response: %v", err)
	}

	if len(response.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(response.Items))
	}

	if response.NextCursor == nil {
		t.Fatal("expected next_cursor in response")
	}

	decoded, err := decodeProductCursor(*response.NextCursor)
	if err != nil {
		t.Fatalf("decode response cursor: %v", err)
	}

	if decoded.ID != next.ID {
		t.Fatalf(
			"expected cursor ID %s, got %s",
			next.ID,
			decoded.ID,
		)
	}

	if !decoded.CreatedAt.Equal(next.CreatedAt) {
		t.Fatalf(
			"expected cursor time %s, got %s",
			next.CreatedAt,
			decoded.CreatedAt,
		)
	}
}

func TestHandlerListProductsOmitsNextCursorOnLastPageAndEmptyItems(t *testing.T) {
	service := &fakeCatalogService{
		listProductsFn: func(ctx context.Context, input ListProductsInput) (ProductPage, error) {
			return ProductPage{
				Products:   []Product{},
				NextCursor: nil,
			}, nil
		},
	}

	router := newTestRouter(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	var response map[string]any

	body := recorder.Body.String()

	if !strings.Contains(body, `"items":[]`) {
		t.Fatalf(
			"expected empty items array, got %s",
			body,
		)
	}

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if _, exists := response["next_cursor"]; exists {
		t.Fatal("expected next_cursor to be omitted")
	}
}

// go test ./internal/catalog \
//   -run TestHandlerCreateProductMapsValidationErrorsToBadRequest \
//   -v
