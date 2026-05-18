package entrepreneur

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, e *Entrepreneur) (*Entrepreneur, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Entrepreneur, error)
	GetAll(ctx context.Context, filter EntrepreneurFilter) ([]*Entrepreneur, int, error)
	Update(ctx context.Context, e *Entrepreneur) (*Entrepreneur, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SetSqbApiError(ctx context.Context, id uuid.UUID, sqbErr *string) error
	GetAllWithSqbError(ctx context.Context, limit, offset int) ([]*Entrepreneur, int, error)
}
