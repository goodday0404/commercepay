package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/goodday0404/commercepay/internal/catalog"
	"github.com/google/uuid"
)

type cartRepository interface {
	Insert(context.Context, Cart) error
	AddItem(ctx context.Context, cartID uuid.UUID, item CartItem) (CartItem, error)
	GetByID(ctx context.Context, id uuid.UUID) (Cart, error)
}

type productCatalog interface {
	GetProduct(context.Context, uuid.UUID) (catalog.Product, error)
	GetProducts(ctx context.Context, ids []uuid.UUID) ([]catalog.Product, error)
}

type Service struct {
	repo     cartRepository
	products productCatalog
}

func NewService(repo cartRepository, products productCatalog) *Service {
	return &Service{
		repo:     repo,
		products: products,
	}
}

func (s *Service) CreateCart(ctx context.Context) (Cart, error) {
	cart := NewCart()

	if err := s.repo.Insert(ctx, cart); err != nil {
		return Cart{}, fmt.Errorf("create cart: %w", err)
	}

	return cart, nil
}

func (s *Service) AddItem(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, quantity int) (CartItem, error) {
	if productID == uuid.Nil {
		return CartItem{}, ErrInvalidProductID
	}

	if quantity <= 0 {
		return CartItem{}, ErrInvalidQuantity
	}

	product, err := s.products.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, catalog.ErrProductNotFound) {
			return CartItem{}, fmt.Errorf("%w: %s", ErrProductNotFound, productID)
		}

		return CartItem{}, fmt.Errorf("get product: %w", err)
	}

	if !product.Available {
		return CartItem{}, fmt.Errorf("%w: %s", ErrProductUnavailable, productID)
	}

	candidate, err := newCartItem(
		product.ID,
		quantity,
		product.PriceMinor,
		product.Currency,
	)
	if err != nil {
		return CartItem{}, fmt.Errorf("create cart item: %w", err)
	}

	persisted, err := s.repo.AddItem(ctx, cartID, candidate)
	if err != nil {
		return CartItem{}, fmt.Errorf("add cart item: %w", err)
	}

	return persisted, nil
}

func (s *Service) GetCart(ctx context.Context, id uuid.UUID) (Cart, error) {
	cart, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Cart{}, fmt.Errorf("get cart: %w", err)
	}

	return cart, nil
}

// func (s *Service) GetCart(ctx context.Context, id uuid.UUID) (CartDetails, error) {
// 	cart, err := s.repo.GetByID(ctx, id)
// 	if err != nil {
// 		return CartDetails{}, fmt.Errorf("get cart: %w", err)
// 	}

// 	items := cart.Items()

// 	details := CartDetails{
// 		ID:    cart.ID(),
// 		Items: make([]CartItemDetails, 0, len(items)),
// 	}

// 	if len(items) == 0 {
// 		return details, nil
// 	}

// 	productIDs := make([]uuid.UUID, 0, len(items))

// 	for _, item := range items {
// 		productIDs = append(productIDs, item.ProductID())
// 	}

// 	products, err := s.products.GetProducts(ctx, productIDs)
// 	if err != nil {
// 		return CartDetails{}, fmt.Errorf("get cart products: %w", err)
// 	}

// 	productsByID := make(map[uuid.UUID]catalog.Product, len(products))

// 	for _, product := range products {
// 		productsByID[product.ID] = product
// 	}

// 	for _, item := range items {
// 		product, ok := productsByID[item.ProductID()]

// 		if !ok {
// 			return CartDetails{}, fmt.Errorf(
// 				"get cart: product %s referenced by cart item %s was not returned by catalog",
// 				item.ProductID(),
// 				item.ID(),
// 			)
// 		}

// 		details.Items = append(details.Items, CartItemDetails{
// 			Item:    item,
// 			Product: product,
// 		})
// 	}

// 	return details, nil
// }
