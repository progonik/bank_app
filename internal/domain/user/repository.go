package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) (*User, error)
	GetByLogin(ctx context.Context, login string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetAll(ctx context.Context, filter UserFilter) ([]*User, int, error)
	Update(ctx context.Context, u *User) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) (*Session, error)
	GetByRefreshTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	GetByUserAndDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*Session, error)
	UpdateToken(ctx context.Context, sessionID uuid.UUID, refreshTokenHash string) (*Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
