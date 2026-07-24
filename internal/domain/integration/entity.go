package integration

import (
	"time"

	"github.com/google/uuid"
)

type Integration struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Active      bool
	ActiveUntil *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Users       []IntegrationUser
}

type IntegrationUser struct {
	ID            uuid.UUID
	IntegrationID uuid.UUID
	UserID        uuid.UUID
	Role          string
	UserFullName  string
	UserLogin     string
	UserRole      string
	UserStatus    bool
	CreatedAt     time.Time
}
