package city

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, c *City) (*City, error)
	GetByID(ctx context.Context, id uuid.UUID) (*City, error)
	GetAll(ctx context.Context, filter CityFilter) ([]*City, int, error)
	Update(ctx context.Context, c *City) (*City, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
