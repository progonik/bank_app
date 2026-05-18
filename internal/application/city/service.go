package city

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domain "github.com/prodonik/bank_app/internal/domain/city"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	Name string
}

type UpdateInput struct {
	Name *string
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.City, error) {
	c := &domain.City{Name: input.Name}
	created, err := s.repo.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to create city: %w", err)
	}
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.City, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context, filter domain.CityFilter) ([]*domain.City, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.repo.GetAll(ctx, filter)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.City, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		existing.Name = *input.Name
	}

	return s.repo.Update(ctx, existing)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
