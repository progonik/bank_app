package dto

import (
	"time"

	"github.com/google/uuid"

	domain "github.com/prodonik/bank_app/internal/domain/ifut_code"
)

type IfutCodeResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type IfutCodeListResponse struct {
	IfutCodes []IfutCodeResponse `json:"ifut_codes"`
	Total     int                `json:"total"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
}

func NewIfutCodeResponse(i *domain.IfutCode) IfutCodeResponse {
	return IfutCodeResponse{
		ID:        i.ID,
		Name:      i.Name,
		CreatedAt: i.CreatedAt,
	}
}

func NewIfutCodeListResponse(codes []*domain.IfutCode, total, limit, offset int) IfutCodeListResponse {
	resp := IfutCodeListResponse{
		IfutCodes: make([]IfutCodeResponse, 0, len(codes)),
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}
	for _, i := range codes {
		resp.IfutCodes = append(resp.IfutCodes, NewIfutCodeResponse(i))
	}
	return resp
}
