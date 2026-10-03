package catalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type productRepository interface {
	Insert(context.Context, Product) error
	GetByID(context.Context, uuid.UUID) (Product, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]Product, error)
	List(context.Context, int, *ProductCursor) (ProductPage, error)
}

type Service struct {
	repo        productRepository
	maxPageSize int
}

func NewService(repo productRepository, maxPageSize int) *Service {
	return &Service{
		repo:        repo,
		maxPageSize: maxPageSize,
	}
}

func (s *Service) CreateProduct(ctx context.Context, input CreateProductInput) (Product, error) {
	product, err := NewProduct(
		input.SKU,
		input.Name,
		input.PriceMinor,
		input.Currency,
		input.Available,
	)
	if err != nil {
		return Product{}, fmt.Errorf("create product: %w", err)
	}

	err = s.repo.Insert(ctx, product)
	if err != nil {
		return Product{}, fmt.Errorf("create product: %w", err)
	}

	return product, nil
}

func (s *Service) GetProduct(ctx context.Context, id uuid.UUID) (Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Product{}, fmt.Errorf("get product: %w", err)
	}
	return product, nil
}

func (s *Service) GetProducts(ctx context.Context, ids []uuid.UUID) ([]Product, error) {
	if len(ids) == 0 {
		return []Product{}, nil
	}

	products, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}

	return products, nil
}

func (s *Service) ListProducts(ctx context.Context, input ListProductsInput) (ProductPage, error) {
	if input.Limit <= 0 || input.Limit > s.maxPageSize {
		return ProductPage{}, ErrInvalidLimit
	}

	page, err := s.repo.List(ctx, input.Limit, input.After)
	if err != nil {
		return ProductPage{}, fmt.Errorf("list product: %w", err)
	}

	return page, nil
}
