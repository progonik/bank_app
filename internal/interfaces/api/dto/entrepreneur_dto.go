package dto

import (
	"time"

	"github.com/google/uuid"

	domain "github.com/prodonik/bank_app/internal/domain/entrepreneur"
)

type CreateEntrepreneurRequest struct {
	Inn                   string `json:"inn" binding:"required,min=9,max=14"`
	LegalName             string `json:"legal_name"`
	RegistrationAuthority string `json:"registration_authority"`
	RegistrationDate      string `json:"registration_date" binding:"required"`
	RegistrationNumber    string `json:"registration_number"`
	LegalForm             string `json:"legal_form"`
	IfutCode              string `json:"ifut_code"`
	DbibtCode             int32  `json:"dbibt_code"`
	ActivityStatus        *bool  `json:"activity_status"`
	CharterFund           int32  `json:"charter_fund"`
	Founders              string `json:"founders"`
	Email                 string `json:"email" binding:"required,email"`
	Phone                 string `json:"phone" binding:"required,min=9,max=32"`
	MhobtCode             string `json:"mhobt_code"`
	Address               string `json:"address"`
	DirectorName          string `json:"director_name"`
}

type UpdateEntrepreneurRequest struct {
	LegalName             *string `json:"legal_name"`
	RegistrationAuthority *string `json:"registration_authority"`
	RegistrationDate      *string `json:"registration_date"`
	RegistrationNumber    *string `json:"registration_number"`
	LegalForm             *string `json:"legal_form"`
	IfutCode              *string `json:"ifut_code"`
	DbibtCode             *int32  `json:"dbibt_code"`
	ActivityStatus        *bool   `json:"activity_status"`
	CharterFund           *int32  `json:"charter_fund"`
	Founders              *string `json:"founders"`
	Email                 *string `json:"email"`
	Phone                 *string `json:"phone"`
	MhobtCode             *string `json:"mhobt_code"`
	Address               *string `json:"address"`
	DirectorName          *string `json:"director_name"`
}

type EntrepreneurResponse struct {
	ID                    uuid.UUID  `json:"id"`
	InnID                 uuid.UUID  `json:"inn_id"`
	InnName               string     `json:"inn_name"`
	LegalName             string     `json:"legal_name"`
	RegistrationAuthority string     `json:"registration_authority"`
	RegistrationDate      string     `json:"registration_date"`
	RegistrationNumber    string     `json:"registration_number"`
	LegalForm             string     `json:"legal_form"`
	IfutCodeID            *uuid.UUID `json:"ifut_code_id"`
	IfutCodeName          string     `json:"ifut_code_name"`
	DbibtCode             int32      `json:"dbibt_code"`
	ActivityStatus        bool       `json:"activity_status"`
	CharterFund           int32      `json:"charter_fund"`
	Founders              string     `json:"founders"`
	Email                 string     `json:"email"`
	Phone                 string     `json:"phone"`
	MhobtCode             string     `json:"mhobt_code"`
	Address               string     `json:"address"`
	DirectorName          string     `json:"director_name"`
	SqbApiError           *string    `json:"sqb_api_error"`
	CreatedAt             time.Time  `json:"created_at"`
}

type UpdateBirdarchaTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type EntrepreneurListResponse struct {
	Entrepreneurs []EntrepreneurResponse `json:"entrepreneurs"`
	Total         int                    `json:"total"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

func NewEntrepreneurResponse(e *domain.Entrepreneur) EntrepreneurResponse {
	return EntrepreneurResponse{
		ID:                    e.ID,
		InnID:                 e.InnID,
		InnName:               e.InnName,
		LegalName:             e.LegalName,
		RegistrationAuthority: e.RegistrationAuthority,
		RegistrationDate:      e.RegistrationDate,
		RegistrationNumber:    e.RegistrationNumber,
		LegalForm:             e.LegalForm,
		IfutCodeID:            e.IfutCodeID,
		IfutCodeName:          e.IfutCodeName,
		DbibtCode:             e.DbibtCode,
		ActivityStatus:        e.ActivityStatus,
		CharterFund:           e.CharterFund,
		Founders:              e.Founders,
		Email:                 e.Email,
		Phone:                 e.Phone,
		MhobtCode:             e.MhobtCode,
		Address:               e.Address,
		DirectorName:          e.DirectorName,
		SqbApiError:           e.SqbApiError,
		CreatedAt:             e.CreatedAt,
	}
}

func NewEntrepreneurListResponse(entrepreneurs []*domain.Entrepreneur, total, limit, offset int) EntrepreneurListResponse {
	resp := EntrepreneurListResponse{
		Entrepreneurs: make([]EntrepreneurResponse, 0, len(entrepreneurs)),
		Total:         total,
		Limit:         limit,
		Offset:        offset,
	}
	for _, e := range entrepreneurs {
		resp.Entrepreneurs = append(resp.Entrepreneurs, NewEntrepreneurResponse(e))
	}
	return resp
}
