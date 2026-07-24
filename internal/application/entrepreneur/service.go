package entrepreneur

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"

	appintegration "github.com/prodonik/bank_app/internal/application/integration"
	domain "github.com/prodonik/bank_app/internal/domain/entrepreneur"
	ifutdomain "github.com/prodonik/bank_app/internal/domain/ifut_code"
	domainn "github.com/prodonik/bank_app/internal/domain/inn"
	"github.com/prodonik/bank_app/internal/infrastructure/bitrix"
	"github.com/prodonik/bank_app/internal/infrastructure/sqb"
)

type Service struct {
	repo         domain.Repository
	innRepo      domainn.Repository
	ifutCodeRepo ifutdomain.Repository
	sqbClient    *sqb.Client
	bitrixClient *bitrix.Client
	integrations *appintegration.Service
	db           *sql.DB
}

func NewService(repo domain.Repository, innRepo domainn.Repository, ifutCodeRepo ifutdomain.Repository, sqbClient *sqb.Client, bitrixClient *bitrix.Client, integrations *appintegration.Service, db *sql.DB) *Service {
	return &Service{repo: repo, innRepo: innRepo, ifutCodeRepo: ifutCodeRepo, sqbClient: sqbClient, bitrixClient: bitrixClient, integrations: integrations, db: db}
}

type CreateInput struct {
	InnName               string
	LegalName             string
	RegistrationAuthority string
	RegistrationDate      string
	RegistrationNumber    string
	LegalForm             string
	IfutCode              string
	DbibtCode             int32
	ActivityStatus        bool
	CharterFund           int32
	Founders              string
	Email                 string
	Phone                 string
	MhobtCode             string
	Address               string
	ActivitySubRegion     string
	DirectorName          string
}

type UpdateInput struct {
	LegalName             *string
	RegistrationAuthority *string
	RegistrationDate      *string
	RegistrationNumber    *string
	LegalForm             *string
	IfutCode              *string
	DbibtCode             *int32
	ActivityStatus        *bool
	CharterFund           *int32
	Founders              *string
	Email                 *string
	Phone                 *string
	MhobtCode             *string
	Address               *string
	ActivitySubRegion     *string
	DirectorName          *string
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Entrepreneur, error) {
	// Strip spaces from INN
	input.InnName = strings.ReplaceAll(input.InnName, " ", "")

	// Resolve INN
	innRecord, err := s.innRepo.GetByName(ctx, input.InnName)
	if err != nil {
		if !errors.Is(err, domainn.ErrInnNotFound) {
			return nil, err
		}
		innRecord, err = s.innRepo.Create(ctx, &domainn.Inn{Name: input.InnName})
		if err != nil {
			return nil, err
		}
	}

	// Resolve IFUT code
	var ifutCodeID *uuid.UUID
	var ifutCodeName string
	if input.IfutCode != "" {
		ifutRecord, err := s.ifutCodeRepo.GetByName(ctx, input.IfutCode)
		if err != nil {
			if !errors.Is(err, ifutdomain.ErrIfutCodeNotFound) {
				return nil, err
			}
			ifutRecord, err = s.ifutCodeRepo.Create(ctx, &ifutdomain.IfutCode{Name: input.IfutCode})
			if err != nil {
				return nil, err
			}
		}
		ifutCodeID = &ifutRecord.ID
		ifutCodeName = ifutRecord.Name
	}

	e := &domain.Entrepreneur{
		InnID:                 innRecord.ID,
		InnName:               innRecord.Name,
		LegalName:             input.LegalName,
		RegistrationAuthority: input.RegistrationAuthority,
		RegistrationDate:      input.RegistrationDate,
		RegistrationNumber:    input.RegistrationNumber,
		LegalForm:             input.LegalForm,
		IfutCodeID:            ifutCodeID,
		IfutCodeName:          ifutCodeName,
		DbibtCode:             input.DbibtCode,
		ActivityStatus:        input.ActivityStatus,
		CharterFund:           input.CharterFund,
		Founders:              input.Founders,
		Email:                 input.Email,
		Phone:                 input.Phone,
		MhobtCode:             input.MhobtCode,
		Address:               input.Address,
		ActivitySubRegion:     input.ActivitySubRegion,
		DirectorName:          input.DirectorName,
	}

	created, err := s.repo.Create(ctx, e)
	if err != nil {
		return nil, err
	}

	if s.sqbClient != nil && s.integrationUsable(ctx, "sqb") {
		resp, err := s.sqbClient.SendLead(ctx, created)
		if err != nil {
			errMsg := err.Error()
			log.Printf("sqb: failed to send lead for entrepreneur %s: %v", created.ID, err)
			_ = s.repo.SetSqbApiError(ctx, created.ID, &errMsg)
			created.SqbApiError = &errMsg
		} else if resp.ErrorCode != "" {
			errMsg := resp.ErrorCode + ": " + resp.ErrorDescription
			log.Printf("sqb: lead error for entrepreneur %s — %s", created.ID, errMsg)
			_ = s.repo.SetSqbApiError(ctx, created.ID, &errMsg)
			created.SqbApiError = &errMsg
		} else {
			log.Printf("sqb: lead sent for entrepreneur %s — status: %s", created.ID, resp.Status)
		}
	}

	if s.bitrixClient != nil && s.bitrixClient.Enabled() && strings.HasPrefix(created.RegistrationAuthority, "birdarcha") && s.integrationUsable(ctx, "bitrix") {
		leadID, err := s.bitrixClient.CreateLead(ctx, created)
		if err != nil {
			log.Printf("bitrix: failed to create lead for entrepreneur %s: %v", created.ID, err)
		} else {
			log.Printf("bitrix: created lead=%d for entrepreneur %s", leadID, created.ID)
		}
	}

	return created, nil
}

func (s *Service) integrationUsable(ctx context.Context, code string) bool {
	if s.integrations == nil {
		return true
	}
	usable, err := s.integrations.IsUsable(ctx, code)
	if err != nil {
		log.Printf("integration: failed to check %s state: %v", code, err)
		return false
	}
	if !usable {
		log.Printf("integration: %s is inactive or expired, skipping outbound call", code)
	}
	return usable
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Entrepreneur, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context, filter domain.EntrepreneurFilter) ([]*domain.Entrepreneur, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.repo.GetAll(ctx, filter)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Entrepreneur, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.LegalName != nil {
		existing.LegalName = *input.LegalName
	}
	if input.RegistrationAuthority != nil {
		existing.RegistrationAuthority = *input.RegistrationAuthority
	}
	if input.RegistrationDate != nil {
		existing.RegistrationDate = *input.RegistrationDate
	}
	if input.RegistrationNumber != nil {
		existing.RegistrationNumber = *input.RegistrationNumber
	}
	if input.LegalForm != nil {
		existing.LegalForm = *input.LegalForm
	}
	if input.IfutCode != nil {
		if *input.IfutCode == "" {
			existing.IfutCodeID = nil
			existing.IfutCodeName = ""
		} else {
			ifutRecord, err := s.ifutCodeRepo.GetByName(ctx, *input.IfutCode)
			if err != nil {
				if !errors.Is(err, ifutdomain.ErrIfutCodeNotFound) {
					return nil, err
				}
				ifutRecord, err = s.ifutCodeRepo.Create(ctx, &ifutdomain.IfutCode{Name: *input.IfutCode})
				if err != nil {
					return nil, err
				}
			}
			existing.IfutCodeID = &ifutRecord.ID
			existing.IfutCodeName = ifutRecord.Name
		}
	}
	if input.DbibtCode != nil {
		existing.DbibtCode = *input.DbibtCode
	}
	if input.ActivityStatus != nil {
		existing.ActivityStatus = *input.ActivityStatus
	}
	if input.CharterFund != nil {
		existing.CharterFund = *input.CharterFund
	}
	if input.Founders != nil {
		existing.Founders = *input.Founders
	}
	if input.Email != nil {
		existing.Email = *input.Email
	}
	if input.Phone != nil {
		existing.Phone = *input.Phone
	}
	if input.MhobtCode != nil {
		existing.MhobtCode = *input.MhobtCode
	}
	if input.Address != nil {
		existing.Address = *input.Address
	}
	if input.ActivitySubRegion != nil {
		existing.ActivitySubRegion = *input.ActivitySubRegion
	}
	if input.DirectorName != nil {
		existing.DirectorName = *input.DirectorName
	}

	return s.repo.Update(ctx, existing)
}

func (s *Service) GetAllWithSqbError(ctx context.Context, limit, offset int) ([]*domain.Entrepreneur, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetAllWithSqbError(ctx, limit, offset)
}

func (s *Service) RetrySqbFailed(ctx context.Context) (sent, failed int) {
	entrepreneurs, _, err := s.repo.GetAllWithSqbError(ctx, 1000, 0)
	if err != nil {
		log.Printf("sqb-retry: failed to fetch failed records: %v", err)
		return 0, 0
	}

	if len(entrepreneurs) == 0 {
		log.Println("sqb-retry: no failed records to retry")
		return 0, 0
	}

	log.Printf("sqb-retry: retrying %d records...", len(entrepreneurs))

	if !s.integrationUsable(ctx, "sqb") {
		log.Println("sqb-retry: sqb integration is inactive or expired")
		return 0, 0
	}

	for _, e := range entrepreneurs {
		if s.sqbClient == nil {
			break
		}
		resp, err := s.sqbClient.SendLead(ctx, e)
		if err != nil {
			errMsg := err.Error()
			_ = s.repo.SetSqbApiError(ctx, e.ID, &errMsg)
			log.Printf("sqb-retry: failed %s: %v", e.ID, err)
			failed++
		} else if resp.ErrorCode != "" {
			errMsg := resp.ErrorCode + ": " + resp.ErrorDescription
			_ = s.repo.SetSqbApiError(ctx, e.ID, &errMsg)
			log.Printf("sqb-retry: error %s: %s", e.ID, errMsg)
			failed++
		} else {
			_ = s.repo.SetSqbApiError(ctx, e.ID, nil)
			log.Printf("sqb-retry: sent %s — success", e.ID)
			sent++
		}
	}

	log.Printf("sqb-retry: complete — sent=%d, failed=%d", sent, failed)
	return sent, failed
}

func (s *Service) UpdateBirdarchaToken(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO syncer_state (key, value, updated_at)
		VALUES ('birdarcha_token', $1, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = NOW()
	`, token)
	if err != nil {
		return err
	}
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO syncer_state (key, value, updated_at)
		VALUES ('birdarcha_token_refresh_needed', 'false', NOW())
		ON CONFLICT (key) DO UPDATE SET value = 'false', updated_at = NOW()
	`)
	return nil
}

func (s *Service) NeedsTokenRefresh(ctx context.Context) (bool, error) {
	var value string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM syncer_state WHERE key = 'birdarcha_token_refresh_needed'`,
	).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return value == "true", nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
