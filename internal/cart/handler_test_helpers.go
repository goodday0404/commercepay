package cart

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type fakeCartService struct {
	createCartFn func(context.Context) (Cart, error)
	addItemFn    func(context.Context, uuid.UUID, uuid.UUID, int) (CartItem, error)
	getCartFn    func(context.Context, uuid.UUID) (Cart, error)
}

func (f *fakeCartService) CreateCart(ctx context.Context) (Cart, error) {
	if f.createCartFn == nil {
		return Cart{}, nil
	}

	return f.createCartFn(ctx)
}

func (f *fakeCartService) AddItem(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
	if f.addItemFn == nil {
		return CartItem{}, nil
	}

	return f.addItemFn(ctx, cartID, productID, quantity)
}

func (f *fakeCartService) GetCart(ctx context.Context, id uuid.UUID) (Cart, error) {
	if f.getCartFn == nil {
		return Cart{}, nil
	}

	return f.getCartFn(ctx, id)
}

func newTestRouter(service cartService) chi.Router {
	router := chi.NewRouter()
	handler := NewHandler(service)
	handler.RegisterRoutes(router)

	return router
}
