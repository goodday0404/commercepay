package catalog

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const defaultPageSize = 20

type fakeCatalogService struct {
	createProductFn func(context.Context, CreateProductInput) (Product, error)
	getProductFn    func(context.Context, uuid.UUID) (Product, error)
	listProductsFn  func(context.Context, ListProductsInput) (ProductPage, error)
}

func (f *fakeCatalogService) CreateProduct(ctx context.Context, input CreateProductInput) (Product, error) {
	if f.createProductFn == nil {
		return Product{}, nil
	}

	return f.createProductFn(ctx, input)
}

func (f *fakeCatalogService) GetProduct(ctx context.Context, id uuid.UUID) (Product, error) {
	if f.getProductFn == nil {
		return Product{}, nil
	}

	return f.getProductFn(ctx, id)
}

func (f *fakeCatalogService) ListProducts(ctx context.Context, input ListProductsInput) (ProductPage, error) {
	if f.listProductsFn == nil {
		return ProductPage{}, nil
	}

	return f.listProductsFn(ctx, input)
}

func newTestRouter(service catalogService) http.Handler {
	handler := NewHandler(service, defaultPageSize)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	return router
}
