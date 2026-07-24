package integration

import (
	"context"
	"time"
)

type Repository interface {
	GetAll(ctx context.Context) ([]*Integration, error)
	GetByCode(ctx context.Context, code string) (*Integration, error)
	UpdateState(ctx context.Context, code string, active bool, activeUntil *time.Time) (*Integration, error)
	IsUsable(ctx context.Context, code string, now time.Time) (bool, error)
}
