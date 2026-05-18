package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/prodonik/bank_app/internal/domain/inn"
	"github.com/prodonik/bank_app/internal/infrastructure/database/sqlc"
)

type InnRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewInnRepository(db *sql.DB) *InnRepository {
	return &InnRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *InnRepository) Create(ctx context.Context, i *inn.Inn) (*inn.Inn, error) {
	row, err := r.queries.CreateINN(ctx, i.Name)
	if err != nil {
		return nil, err
	}
	return mapSqlcInn(row), nil
}

func (r *InnRepository) GetByID(ctx context.Context, id uuid.UUID) (*inn.Inn, error) {
	row, err := r.queries.GetINNByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, inn.ErrInnNotFound
		}
		return nil, err
	}
	return mapSqlcInn(row), nil
}

func (r *InnRepository) GetByName(ctx context.Context, name string) (*inn.Inn, error) {
	row, err := r.queries.GetINNByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, inn.ErrInnNotFound
		}
		return nil, err
	}
	return mapSqlcInn(row), nil
}

func (r *InnRepository) GetAll(ctx context.Context, filter inn.InnFilter) ([]*inn.Inn, int, error) {
	var where string
	var args []any
	paramIdx := 1

	if filter.Name != nil {
		where = fmt.Sprintf(" WHERE name ILIKE $%d", paramIdx)
		args = append(args, "%"+*filter.Name+"%")
		paramIdx++
	}

	countQuery := "SELECT COUNT(*) FROM inns" + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count inns: %w", err)
	}

	dataQuery := fmt.Sprintf(
		"SELECT id, name, created_at FROM inns%s ORDER BY name ASC LIMIT $%d OFFSET $%d",
		where, paramIdx, paramIdx+1,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query inns: %w", err)
	}
	defer rows.Close()

	var inns []*inn.Inn
	for rows.Next() {
		var i inn.Inn
		if err := rows.Scan(&i.ID, &i.Name, &i.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan inn: %w", err)
		}
		inns = append(inns, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return inns, total, nil
}

func mapSqlcInn(i sqlc.Inn) *inn.Inn {
	return &inn.Inn{
		ID:        i.ID,
		Name:      i.Name,
		CreatedAt: i.CreatedAt,
	}
}
