package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	domain "github.com/prodonik/bank_app/internal/domain/entrepreneur"
	"github.com/prodonik/bank_app/internal/infrastructure/database/sqlc"
)

type EntrepreneurRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewEntrepreneurRepository(db *sql.DB) *EntrepreneurRepository {
	return &EntrepreneurRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *EntrepreneurRepository) Create(ctx context.Context, e *domain.Entrepreneur) (*domain.Entrepreneur, error) {
	row, err := r.queries.CreateEntrepreneur(ctx, sqlc.CreateEntrepreneurParams{
		InnID:                 e.InnID,
		LegalName:             e.LegalName,
		RegistrationAuthority: e.RegistrationAuthority,
		RegistrationDate:      e.RegistrationDate,
		RegistrationNumber:    e.RegistrationNumber,
		LegalForm:             e.LegalForm,
		IfutCodeID:            toNullUUID(e.IfutCodeID),
		DbibtCode:             e.DbibtCode,
		ActivityStatus:        e.ActivityStatus,
		CharterFund:           e.CharterFund,
		Founders:              e.Founders,
		Email:                 e.Email,
		Phone:                 e.Phone,
		MhobtCode:             e.MhobtCode,
		Address:               e.Address,
		DirectorName:          e.DirectorName,
		SqbApiError:           toNullString(e.SqbApiError),
	})
	if err != nil {
		return nil, err
	}
	return mapCreateRow(row, e.InnName, e.IfutCodeName), nil
}

func (r *EntrepreneurRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Entrepreneur, error) {
	row, err := r.queries.GetEntrepreneurByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrEntrepreneurNotFound
		}
		return nil, err
	}
	return mapGetByIDRow(row), nil
}

func (r *EntrepreneurRepository) Update(ctx context.Context, e *domain.Entrepreneur) (*domain.Entrepreneur, error) {
	row, err := r.queries.UpdateEntrepreneur(ctx, sqlc.UpdateEntrepreneurParams{
		ID:                    e.ID,
		LegalName:             e.LegalName,
		RegistrationAuthority: e.RegistrationAuthority,
		RegistrationDate:      e.RegistrationDate,
		RegistrationNumber:    e.RegistrationNumber,
		LegalForm:             e.LegalForm,
		IfutCodeID:            toNullUUID(e.IfutCodeID),
		DbibtCode:             e.DbibtCode,
		ActivityStatus:        e.ActivityStatus,
		CharterFund:           e.CharterFund,
		Founders:              e.Founders,
		Email:                 e.Email,
		Phone:                 e.Phone,
		MhobtCode:             e.MhobtCode,
		Address:               e.Address,
		DirectorName:          e.DirectorName,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrEntrepreneurNotFound
		}
		return nil, err
	}
	return mapUpdateRow(row, e.InnName, e.IfutCodeName), nil
}

func (r *EntrepreneurRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteEntrepreneur(ctx, id)
}

func (r *EntrepreneurRepository) SetSqbApiError(ctx context.Context, id uuid.UUID, sqbErr *string) error {
	return r.queries.SetSqbApiError(ctx, sqlc.SetSqbApiErrorParams{
		ID:          id,
		SqbApiError: toNullString(sqbErr),
	})
}

func (r *EntrepreneurRepository) GetAllWithSqbError(ctx context.Context, limit, offset int) ([]*domain.Entrepreneur, int, error) {
	countQuery := `SELECT COUNT(*) FROM entrepreneurs WHERE sqb_api_error IS NOT NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count failed entrepreneurs: %w", err)
	}

	dataQuery := `SELECT e.id, e.inn_id, i.name, e.legal_name, e.registration_authority,
		e.registration_date, e.registration_number, e.legal_form, e.ifut_code_id,
		ic.name, e.dbibt_code, e.activity_status, e.charter_fund, e.founders, e.email,
		e.phone, e.mhobt_code, e.address, e.director_name, e.sqb_api_error, e.created_at
	FROM entrepreneurs e
	JOIN inns i ON e.inn_id = i.id
	LEFT JOIN ifut_codes ic ON e.ifut_code_id = ic.id
	WHERE e.sqb_api_error IS NOT NULL
	ORDER BY e.created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, dataQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query failed entrepreneurs: %w", err)
	}
	defer rows.Close()

	var entrepreneurs []*domain.Entrepreneur
	for rows.Next() {
		var e domain.Entrepreneur
		var ifutCodeID uuid.NullUUID
		var ifutCodeName sql.NullString
		var sqbApiError sql.NullString
		if err := rows.Scan(
			&e.ID, &e.InnID, &e.InnName, &e.LegalName, &e.RegistrationAuthority,
			&e.RegistrationDate, &e.RegistrationNumber, &e.LegalForm, &ifutCodeID,
			&ifutCodeName, &e.DbibtCode, &e.ActivityStatus, &e.CharterFund, &e.Founders,
			&e.Email, &e.Phone, &e.MhobtCode, &e.Address, &e.DirectorName, &sqbApiError, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan entrepreneur: %w", err)
		}
		if ifutCodeID.Valid {
			e.IfutCodeID = &ifutCodeID.UUID
		}
		if ifutCodeName.Valid {
			e.IfutCodeName = ifutCodeName.String
		}
		if sqbApiError.Valid {
			e.SqbApiError = &sqbApiError.String
		}
		entrepreneurs = append(entrepreneurs, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return entrepreneurs, total, nil
}

func (r *EntrepreneurRepository) GetAll(ctx context.Context, filter domain.EntrepreneurFilter) ([]*domain.Entrepreneur, int, error) {
	var conditions []string
	var args []any
	paramIdx := 1

	if filter.LegalName != nil {
		conditions = append(conditions, fmt.Sprintf("e.legal_name ILIKE $%d", paramIdx))
		args = append(args, "%"+*filter.LegalName+"%")
		paramIdx++
	}
	if filter.InnName != nil {
		conditions = append(conditions, fmt.Sprintf("i.name ILIKE $%d", paramIdx))
		args = append(args, "%"+*filter.InnName+"%")
		paramIdx++
	}
	if filter.ActivityStatus != nil {
		conditions = append(conditions, fmt.Sprintf("e.activity_status = $%d", paramIdx))
		args = append(args, *filter.ActivityStatus)
		paramIdx++
	}
	if filter.DirectorName != nil {
		conditions = append(conditions, fmt.Sprintf("e.director_name ILIKE $%d", paramIdx))
		args = append(args, "%"+*filter.DirectorName+"%")
		paramIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("e.created_at >= $%d", paramIdx))
		args = append(args, *filter.DateFrom)
		paramIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("e.created_at <= $%d", paramIdx))
		args = append(args, *filter.DateTo)
		paramIdx++
	}

	var where string
	if len(conditions) > 0 {
		where = " WHERE "
		for i, cond := range conditions {
			if i > 0 {
				where += " AND "
			}
			where += cond
		}
	}

	countQuery := "SELECT COUNT(*) FROM entrepreneurs e JOIN inns i ON e.inn_id = i.id LEFT JOIN ifut_codes ic ON e.ifut_code_id = ic.id" + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count entrepreneurs: %w", err)
	}

	dataQuery := fmt.Sprintf(
		`SELECT e.id, e.inn_id, i.name, e.legal_name, e.registration_authority,
			e.registration_date, e.registration_number, e.legal_form, e.ifut_code_id,
			ic.name, e.dbibt_code, e.activity_status, e.charter_fund, e.founders, e.email,
			e.phone, e.mhobt_code, e.address, e.director_name, e.sqb_api_error, e.created_at
		FROM entrepreneurs e
		JOIN inns i ON e.inn_id = i.id
		LEFT JOIN ifut_codes ic ON e.ifut_code_id = ic.id%s
		ORDER BY e.created_at DESC LIMIT $%d OFFSET $%d`,
		where, paramIdx, paramIdx+1,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query entrepreneurs: %w", err)
	}
	defer rows.Close()

	var entrepreneurs []*domain.Entrepreneur
	for rows.Next() {
		var e domain.Entrepreneur
		var ifutCodeID uuid.NullUUID
		var ifutCodeName sql.NullString
		var sqbApiError sql.NullString
		if err := rows.Scan(
			&e.ID, &e.InnID, &e.InnName, &e.LegalName, &e.RegistrationAuthority,
			&e.RegistrationDate, &e.RegistrationNumber, &e.LegalForm, &ifutCodeID,
			&ifutCodeName, &e.DbibtCode, &e.ActivityStatus, &e.CharterFund, &e.Founders,
			&e.Email, &e.Phone, &e.MhobtCode, &e.Address, &e.DirectorName, &sqbApiError, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan entrepreneur: %w", err)
		}
		if ifutCodeID.Valid {
			e.IfutCodeID = &ifutCodeID.UUID
		}
		if ifutCodeName.Valid {
			e.IfutCodeName = ifutCodeName.String
		}
		if sqbApiError.Valid {
			e.SqbApiError = &sqbApiError.String
		}
		entrepreneurs = append(entrepreneurs, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return entrepreneurs, total, nil
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func fromNullString(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	return &n.String
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func fromNullUUID(n uuid.NullUUID) *uuid.UUID {
	if !n.Valid {
		return nil
	}
	return &n.UUID
}

func mapCreateRow(e sqlc.CreateEntrepreneurRow, innName, ifutCodeName string) *domain.Entrepreneur {
	return &domain.Entrepreneur{
		ID:                    e.ID,
		InnID:                 e.InnID,
		InnName:               innName,
		LegalName:             e.LegalName,
		RegistrationAuthority: e.RegistrationAuthority,
		RegistrationDate:      e.RegistrationDate,
		RegistrationNumber:    e.RegistrationNumber,
		LegalForm:             e.LegalForm,
		IfutCodeID:            fromNullUUID(e.IfutCodeID),
		IfutCodeName:          ifutCodeName,
		DbibtCode:             e.DbibtCode,
		ActivityStatus:        e.ActivityStatus,
		CharterFund:           e.CharterFund,
		Founders:              e.Founders,
		Email:                 e.Email,
		Phone:                 e.Phone,
		MhobtCode:             e.MhobtCode,
		Address:               e.Address,
		DirectorName:          e.DirectorName,
		SqbApiError:           fromNullString(e.SqbApiError),
		CreatedAt:             e.CreatedAt,
	}
}

func mapGetByIDRow(e sqlc.GetEntrepreneurByIDRow) *domain.Entrepreneur {
	var ifutCodeName string
	if e.IfutCodeName.Valid {
		ifutCodeName = e.IfutCodeName.String
	}
	return &domain.Entrepreneur{
		ID:                    e.ID,
		InnID:                 e.InnID,
		InnName:               e.InnName,
		LegalName:             e.LegalName,
		RegistrationAuthority: e.RegistrationAuthority,
		RegistrationDate:      e.RegistrationDate,
		RegistrationNumber:    e.RegistrationNumber,
		LegalForm:             e.LegalForm,
		IfutCodeID:            fromNullUUID(e.IfutCodeID),
		IfutCodeName:          ifutCodeName,
		DbibtCode:             e.DbibtCode,
		ActivityStatus:        e.ActivityStatus,
		CharterFund:           e.CharterFund,
		Founders:              e.Founders,
		Email:                 e.Email,
		Phone:                 e.Phone,
		MhobtCode:             e.MhobtCode,
		Address:               e.Address,
		DirectorName:          e.DirectorName,
		SqbApiError:           fromNullString(e.SqbApiError),
		CreatedAt:             e.CreatedAt,
	}
}

func mapUpdateRow(e sqlc.UpdateEntrepreneurRow, innName, ifutCodeName string) *domain.Entrepreneur {
	return &domain.Entrepreneur{
		ID:                    e.ID,
		InnID:                 e.InnID,
		InnName:               innName,
		LegalName:             e.LegalName,
		RegistrationAuthority: e.RegistrationAuthority,
		RegistrationDate:      e.RegistrationDate,
		RegistrationNumber:    e.RegistrationNumber,
		LegalForm:             e.LegalForm,
		IfutCodeID:            fromNullUUID(e.IfutCodeID),
		IfutCodeName:          ifutCodeName,
		DbibtCode:             e.DbibtCode,
		ActivityStatus:        e.ActivityStatus,
		CharterFund:           e.CharterFund,
		Founders:              e.Founders,
		Email:                 e.Email,
		Phone:                 e.Phone,
		MhobtCode:             e.MhobtCode,
		Address:               e.Address,
		DirectorName:          e.DirectorName,
		SqbApiError:           fromNullString(e.SqbApiError),
		CreatedAt:             e.CreatedAt,
	}
}
