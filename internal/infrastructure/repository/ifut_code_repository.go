package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	ifutcode "github.com/prodonik/bank_app/internal/domain/ifut_code"
	"github.com/prodonik/bank_app/internal/infrastructure/database/sqlc"
)

type IfutCodeRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewIfutCodeRepository(db *sql.DB) *IfutCodeRepository {
	return &IfutCodeRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *IfutCodeRepository) Create(ctx context.Context, i *ifutcode.IfutCode) (*ifutcode.IfutCode, error) {
	row, err := r.queries.CreateIfutCode(ctx, i.Name)
	if err != nil {
		return nil, err
	}
	return mapSqlcIfutCode(row), nil
}

func (r *IfutCodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*ifutcode.IfutCode, error) {
	row, err := r.queries.GetIfutCodeByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ifutcode.ErrIfutCodeNotFound
		}
		return nil, err
	}
	return mapSqlcIfutCode(row), nil
}

func (r *IfutCodeRepository) GetByName(ctx context.Context, name string) (*ifutcode.IfutCode, error) {
	row, err := r.queries.GetIfutCodeByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ifutcode.ErrIfutCodeNotFound
		}
		return nil, err
	}
	return mapSqlcIfutCode(row), nil
}

func (r *IfutCodeRepository) GetAll(ctx context.Context, filter ifutcode.IfutCodeFilter) ([]*ifutcode.IfutCode, int, error) {
	var where string
	var args []any
	paramIdx := 1

	if filter.Name != nil {
		where = fmt.Sprintf(" WHERE name ILIKE $%d", paramIdx)
		args = append(args, "%"+*filter.Name+"%")
		paramIdx++
	}

	countQuery := "SELECT COUNT(*) FROM ifut_codes" + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count ifut_codes: %w", err)
	}

	dataQuery := fmt.Sprintf(
		"SELECT id, name, created_at FROM ifut_codes%s ORDER BY name ASC LIMIT $%d OFFSET $%d",
		where, paramIdx, paramIdx+1,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query ifut_codes: %w", err)
	}
	defer rows.Close()

	var codes []*ifutcode.IfutCode
	for rows.Next() {
		var i ifutcode.IfutCode
		if err := rows.Scan(&i.ID, &i.Name, &i.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan ifut_code: %w", err)
		}
		codes = append(codes, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return codes, total, nil
}

func mapSqlcIfutCode(i sqlc.IfutCode) *ifutcode.IfutCode {
	return &ifutcode.IfutCode{
		ID:        i.ID,
		Name:      i.Name,
		CreatedAt: i.CreatedAt,
	}
}
