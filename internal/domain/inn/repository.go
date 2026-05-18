package inn

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, i *Inn) (*Inn, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Inn, error)
	GetByName(ctx context.Context, name string) (*Inn, error)
	GetAll(ctx context.Context, filter InnFilter) ([]*Inn, int, error)
}
