package ifutcode

import (
	"context"

	domain "github.com/prodonik/bank_app/internal/domain/ifut_code"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll(ctx context.Context, filter domain.IfutCodeFilter) ([]*domain.IfutCode, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.repo.GetAll(ctx, filter)
}
