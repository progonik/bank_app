package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/prodonik/bank_app/internal/domain/user"
	"github.com/prodonik/bank_app/internal/infrastructure/database/sqlc"
)

type SessionRepository struct {
	queries *sqlc.Queries
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		queries: sqlc.New(db),
	}
}

func (r *SessionRepository) Create(ctx context.Context, session *user.Session) (*user.Session, error) {
	row, err := r.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		UserID:           session.UserID,
		DeviceID:         session.DeviceID,
		RefreshTokenHash: session.RefreshTokenHash,
	})
	if err != nil {
		return nil, err
	}
	return mapSqlcSession(row), nil
}

func (r *SessionRepository) GetByRefreshTokenHash(ctx context.Context, tokenHash string) (*user.Session, error) {
	row, err := r.queries.GetSessionByRefreshTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrSessionNotFound
		}
		return nil, err
	}
	return mapSqlcSession(row), nil
}

func (r *SessionRepository) GetByUserAndDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*user.Session, error) {
	row, err := r.queries.GetSessionByUserAndDevice(ctx, sqlc.GetSessionByUserAndDeviceParams{
		UserID:   userID,
		DeviceID: deviceID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrSessionNotFound
		}
		return nil, err
	}
	return mapSqlcSession(row), nil
}

func (r *SessionRepository) UpdateToken(ctx context.Context, sessionID uuid.UUID, refreshTokenHash string) (*user.Session, error) {
	row, err := r.queries.UpdateSessionToken(ctx, sqlc.UpdateSessionTokenParams{
		ID:               sessionID,
		RefreshTokenHash: refreshTokenHash,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrSessionNotFound
		}
		return nil, err
	}
	return mapSqlcSession(row), nil
}

func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteSession(ctx, id)
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteSessionsByUserID(ctx, userID)
}

func mapSqlcSession(s sqlc.UserActiveSession) *user.Session {
	return &user.Session{
		ID:               s.ID,
		UserID:           s.UserID,
		DeviceID:         s.DeviceID,
		RefreshTokenHash: s.RefreshTokenHash,
		LastLoginAt:      s.LastLoginAt,
		CreatedAt:        s.CreatedAt,
	}
}
