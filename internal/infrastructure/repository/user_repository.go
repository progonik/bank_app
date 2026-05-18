package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/prodonik/bank_app/internal/domain/user"
	"github.com/prodonik/bank_app/internal/infrastructure/database/sqlc"
)

type UserRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	row, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:     u.FullName,
		Role:         u.Role,
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		Status:       u.Status,
	})
	if err != nil {
		return nil, err
	}
	return mapSqlcUser(row), nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	row, err := r.queries.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return mapSqlcUser(row), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return mapSqlcUser(row), nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) (*user.User, error) {
	row, err := r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:       u.ID,
		FullName: u.FullName,
		Role:     u.Role,
		Status:   u.Status,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return mapSqlcUser(row), nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteUser(ctx, id)
}

func (r *UserRepository) GetAll(ctx context.Context, filter user.UserFilter) ([]*user.User, int, error) {
	where, args := buildWhereClause(filter)

	// Count query
	countQuery := "SELECT COUNT(*) FROM users" + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(
		"SELECT id, full_name, role, login, password_hash, status, created_at FROM users%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		where, len(args)+1, len(args)+2,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var u user.User
		if err := rows.Scan(&u.ID, &u.FullName, &u.Role, &u.Login, &u.PasswordHash, &u.Status, &u.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, total, nil
}

func buildWhereClause(filter user.UserFilter) (string, []any) {
	var conditions []string
	var args []any
	paramIdx := 1

	if filter.FullName != nil {
		conditions = append(conditions, fmt.Sprintf("full_name ILIKE $%d", paramIdx))
		args = append(args, "%"+*filter.FullName+"%")
		paramIdx++
	}
	if filter.Role != nil {
		conditions = append(conditions, fmt.Sprintf("role = $%d", paramIdx))
		args = append(args, *filter.Role)
		paramIdx++
	}
	if filter.Login != nil {
		conditions = append(conditions, fmt.Sprintf("login ILIKE $%d", paramIdx))
		args = append(args, "%"+*filter.Login+"%")
		paramIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", paramIdx))
		args = append(args, *filter.Status)
		paramIdx++
	}
	if filter.CreatedFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", paramIdx))
		args = append(args, *filter.CreatedFrom)
		paramIdx++
	}
	if filter.CreatedTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", paramIdx))
		args = append(args, *filter.CreatedTo)
		paramIdx++
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func mapSqlcUser(u sqlc.User) *user.User {
	return &user.User{
		ID:           u.ID,
		FullName:     u.FullName,
		Role:         u.Role,
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
	}
}
