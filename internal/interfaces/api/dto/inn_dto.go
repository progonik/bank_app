package dto

import (
	"time"

	"github.com/google/uuid"

	domain "github.com/prodonik/bank_app/internal/domain/inn"
)

type InnResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type InnListResponse struct {
	Inns   []InnResponse `json:"inns"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

func NewInnResponse(i *domain.Inn) InnResponse {
	return InnResponse{
		ID:        i.ID,
		Name:      i.Name,
		CreatedAt: i.CreatedAt,
	}
}

func NewInnListResponse(inns []*domain.Inn, total, limit, offset int) InnListResponse {
	resp := InnListResponse{
		Inns:   make([]InnResponse, 0, len(inns)),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	for _, i := range inns {
		resp.Inns = append(resp.Inns, NewInnResponse(i))
	}
	return resp
}
