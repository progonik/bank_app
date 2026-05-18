package inn

import (
	"time"

	"github.com/google/uuid"
)

type Inn struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

type InnFilter struct {
	Name   *string
	Limit  int
	Offset int
}
