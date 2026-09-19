package catalog

import (
	"context"

	"github.com/google/uuid"
)

type fakeProductRepository struct {
	insertFn  func(context.Context, Product) error
	getByIDFn func(context.Context, uuid.UUID) (Product, error)
	listFn    func(context.Context, int, *ProductCursor) (ProductPage, error)
}

func (f *fakeProductRepository) Insert(ctx context.Context, product Product) error {
	if f.insertFn == nil {
		return nil
	}

	return f.insertFn(ctx, product)
}

func (f *fakeProductRepository) GetByID(ctx context.Context, id uuid.UUID) (Product, error) {
	if f.getByIDFn == nil {
		return Product{}, nil
	}

	return f.getByIDFn(ctx, id)
}

func (f *fakeProductRepository) List(ctx context.Context, limit int, after *ProductCursor) (ProductPage, error) {
	if f.listFn == nil {
		return ProductPage{}, nil
	}

	return f.listFn(ctx, limit, after)
}
