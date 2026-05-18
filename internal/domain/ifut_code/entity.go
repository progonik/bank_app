package ifutcode

import (
	"time"

	"github.com/google/uuid"
)

type IfutCode struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

type IfutCodeFilter struct {
	Name   *string
	Limit  int
	Offset int
}
