package catalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type CreateProductInput struct {
	SKU        string
	Name       string
	PriceMinor int64
	Currency   string
	Available  bool
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
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
