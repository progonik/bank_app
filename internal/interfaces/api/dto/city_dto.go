package dto

import (
	"time"

	"github.com/google/uuid"

	domain "github.com/prodonik/bank_app/internal/domain/city"
)

type CreateCityRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateCityRequest struct {
	Name *string `json:"name"`
}

type CityResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CityListResponse struct {
	Cities []CityResponse `json:"cities"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func NewCityResponse(c *domain.City) CityResponse {
	return CityResponse{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
	}
}

func NewCityListResponse(cities []*domain.City, total, limit, offset int) CityListResponse {
	resp := CityListResponse{
		Cities: make([]CityResponse, 0, len(cities)),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	for _, c := range cities {
		resp.Cities = append(resp.Cities, NewCityResponse(c))
	}
	return resp
}
