package integration

import (
	"context"
	"strings"
	"time"

	domain "github.com/prodonik/bank_app/internal/domain/integration"
)

type Service struct {
	repo domain.Repository
}

type UpdateStateInput struct {
	Active      bool
	ActiveUntil *time.Time
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll(ctx context.Context) ([]*domain.Integration, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) UpdateState(ctx context.Context, code string, input UpdateStateInput) (*domain.Integration, error) {
	return s.repo.UpdateState(ctx, normalizeCode(code), input.Active, input.ActiveUntil)
}

func (s *Service) IsUsable(ctx context.Context, code string) (bool, error) {
	return s.repo.IsUsable(ctx, normalizeCode(code), time.Now())
}

func normalizeCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}
