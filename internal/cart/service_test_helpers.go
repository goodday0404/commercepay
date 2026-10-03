package cart

import (
	"context"

	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/google/uuid"
)

type fakeCartRepository struct {
	insertFn      func(context.Context, Cart) error
	insertItemsFn func(context.Context, uuid.UUID, CartItem) (CartItem, error)
	getByIDFn     func(context.Context, uuid.UUID) (Cart, error)
}

func (f *fakeCartRepository) Insert(ctx context.Context, cart Cart) error {
	if f == nil {
		return nil
	}

	return f.insertFn(ctx, cart)
}

func (f *fakeCartRepository) AddItem(ctx context.Context, cartID uuid.UUID, item CartItem) (CartItem, error) {
	if f.insertItemsFn == nil {
		return CartItem{}, nil
	}

	return f.insertItemsFn(ctx, cartID, item)
}

func (f *fakeCartRepository) GetByID(ctx context.Context, id uuid.UUID) (Cart, error) {
	if f.getByIDFn == nil {
		return Cart{}, nil
	}

	return f.getByIDFn(ctx, id)
}

type fakeProductCatalog struct {
	getProductFn  func(context.Context, uuid.UUID) (catalog.Product, error)
	getProductsFn func(context.Context, []uuid.UUID) ([]catalog.Product, error)
}

func (f *fakeProductCatalog) GetProduct(ctx context.Context, id uuid.UUID) (catalog.Product, error) {
	if f == nil {
		return catalog.Product{}, nil
	}

	return f.getProductFn(ctx, id)
}

func (f *fakeProductCatalog) GetProducts(ctx context.Context, ids []uuid.UUID) ([]catalog.Product, error) {
	if f.getProductsFn == nil {
		return []catalog.Product{}, nil
	}

	return f.getProductsFn(ctx, ids)
}
