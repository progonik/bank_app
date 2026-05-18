package ifutcode

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, i *IfutCode) (*IfutCode, error)
	GetByID(ctx context.Context, id uuid.UUID) (*IfutCode, error)
	GetByName(ctx context.Context, name string) (*IfutCode, error)
	GetAll(ctx context.Context, filter IfutCodeFilter) ([]*IfutCode, int, error)
}
