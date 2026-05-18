package city

import (
	"time"

	"github.com/google/uuid"
)

type City struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

type CityFilter struct {
	Name   *string
	Limit  int
	Offset int
}
