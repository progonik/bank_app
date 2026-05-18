package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/prodonik/bank_app/internal/domain/city"
	"github.com/prodonik/bank_app/internal/infrastructure/database/sqlc"
)

type CityRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewCityRepository(db *sql.DB) *CityRepository {
	return &CityRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *CityRepository) Create(ctx context.Context, c *city.City) (*city.City, error) {
	row, err := r.queries.CreateCity(ctx, c.Name)
	if err != nil {
		return nil, err
	}
	return mapSqlcCity(row), nil
}

func (r *CityRepository) GetByID(ctx context.Context, id uuid.UUID) (*city.City, error) {
	row, err := r.queries.GetCityByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, city.ErrCityNotFound
		}
		return nil, err
	}
	return mapSqlcCity(row), nil
}

func (r *CityRepository) Update(ctx context.Context, c *city.City) (*city.City, error) {
	row, err := r.queries.UpdateCity(ctx, sqlc.UpdateCityParams{
		ID:   c.ID,
		Name: c.Name,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, city.ErrCityNotFound
		}
		return nil, err
	}
	return mapSqlcCity(row), nil
}

func (r *CityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteCity(ctx, id)
}

func (r *CityRepository) GetAll(ctx context.Context, filter city.CityFilter) ([]*city.City, int, error) {
	var where string
	var args []any
	paramIdx := 1

	if filter.Name != nil {
		where = fmt.Sprintf(" WHERE name ILIKE $%d", paramIdx)
		args = append(args, "%"+*filter.Name+"%")
		paramIdx++
	}

	countQuery := "SELECT COUNT(*) FROM cities" + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count cities: %w", err)
	}

	dataQuery := fmt.Sprintf(
		"SELECT id, name, created_at FROM cities%s ORDER BY name ASC LIMIT $%d OFFSET $%d",
		where, paramIdx, paramIdx+1,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query cities: %w", err)
	}
	defer rows.Close()

	var cities []*city.City
	for rows.Next() {
		var c city.City
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return cities, total, nil
}

func mapSqlcCity(c sqlc.City) *city.City {
	return &city.City{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
	}
}
