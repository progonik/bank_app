package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	domain "github.com/prodonik/bank_app/internal/domain/integration"
)

type IntegrationRepository struct {
	db *sql.DB
}

func NewIntegrationRepository(db *sql.DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

func (r *IntegrationRepository) GetAll(ctx context.Context) ([]*domain.Integration, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, active, active_until, created_at, updated_at
		FROM integrations
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query integrations: %w", err)
	}
	defer rows.Close()

	var integrations []*domain.Integration
	for rows.Next() {
		integration, err := scanIntegration(rows)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, integration)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate integrations: %w", err)
	}

	if err := r.attachUsers(ctx, integrations); err != nil {
		return nil, err
	}
	return integrations, nil
}

func (r *IntegrationRepository) GetByCode(ctx context.Context, code string) (*domain.Integration, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, active, active_until, created_at, updated_at
		FROM integrations
		WHERE code = $1
	`, code)

	integration, err := scanIntegration(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIntegrationNotFound
		}
		return nil, err
	}

	if err := r.attachUsers(ctx, []*domain.Integration{integration}); err != nil {
		return nil, err
	}
	return integration, nil
}

func (r *IntegrationRepository) UpdateState(ctx context.Context, code string, active bool, activeUntil *time.Time) (*domain.Integration, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE integrations
		SET active = $2, active_until = $3, updated_at = NOW()
		WHERE code = $1
		RETURNING id, code, name, active, active_until, created_at, updated_at
	`, code, active, activeUntil)

	integration, err := scanIntegration(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIntegrationNotFound
		}
		return nil, err
	}
	if err := r.attachUsers(ctx, []*domain.Integration{integration}); err != nil {
		return nil, err
	}
	return integration, nil
}

func (r *IntegrationRepository) IsUsable(ctx context.Context, code string, now time.Time) (bool, error) {
	var usable bool
	err := r.db.QueryRowContext(ctx, `
		SELECT active AND (active_until IS NULL OR active_until >= $2)
		FROM integrations
		WHERE code = $1
	`, code, now).Scan(&usable)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return usable, nil
}

func (r *IntegrationRepository) attachUsers(ctx context.Context, integrations []*domain.Integration) error {
	if len(integrations) == 0 {
		return nil
	}

	byID := make(map[uuid.UUID]*domain.Integration, len(integrations))
	ids := make([]uuid.UUID, 0, len(integrations))
	for _, integration := range integrations {
		byID[integration.ID] = integration
		ids = append(ids, integration.ID)
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT iu.id, iu.integration_id, iu.user_id, iu.role, u.full_name, u.login, u.role, u.status, iu.created_at
		FROM integration_users iu
		JOIN users u ON u.id = iu.user_id
		WHERE iu.integration_id IN (`+strings.Join(placeholders, ", ")+`)
		ORDER BY u.full_name ASC
	`, args...)
	if err != nil {
		return fmt.Errorf("query integration users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user domain.IntegrationUser
		if err := rows.Scan(
			&user.ID,
			&user.IntegrationID,
			&user.UserID,
			&user.Role,
			&user.UserFullName,
			&user.UserLogin,
			&user.UserRole,
			&user.UserStatus,
			&user.CreatedAt,
		); err != nil {
			return fmt.Errorf("scan integration user: %w", err)
		}
		if integration := byID[user.IntegrationID]; integration != nil {
			integration.Users = append(integration.Users, user)
		}
	}
	return rows.Err()
}

type integrationScanner interface {
	Scan(dest ...any) error
}

func scanIntegration(scanner integrationScanner) (*domain.Integration, error) {
	var integration domain.Integration
	if err := scanner.Scan(
		&integration.ID,
		&integration.Code,
		&integration.Name,
		&integration.Active,
		&integration.ActiveUntil,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &integration, nil
}
