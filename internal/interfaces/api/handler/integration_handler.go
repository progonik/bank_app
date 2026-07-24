package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	app "github.com/prodonik/bank_app/internal/application/integration"
	domain "github.com/prodonik/bank_app/internal/domain/integration"
	"github.com/prodonik/bank_app/internal/interfaces/api/dto"
)

type IntegrationHandler struct {
	service *app.Service
}

func NewIntegrationHandler(service *app.Service) *IntegrationHandler {
	return &IntegrationHandler{service: service}
}

func (h *IntegrationHandler) GetAll(c *gin.Context) {
	integrations, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]dto.IntegrationResponse, 0, len(integrations))
	now := time.Now()
	for _, integration := range integrations {
		items = append(items, mapIntegrationResponse(integration, now))
	}
	c.JSON(http.StatusOK, gin.H{"integrations": items})
}

func (h *IntegrationHandler) Update(c *gin.Context) {
	var req dto.UpdateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	integration, err := h.service.UpdateState(c.Request.Context(), c.Param("code"), app.UpdateStateInput{
		Active:      req.Active,
		ActiveUntil: req.ActiveUntil,
	})
	if err != nil {
		if errors.Is(err, domain.ErrIntegrationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapIntegrationResponse(integration, time.Now()))
}

func mapIntegrationResponse(integration *domain.Integration, now time.Time) dto.IntegrationResponse {
	users := make([]dto.IntegrationUserResponse, 0, len(integration.Users))
	for _, user := range integration.Users {
		users = append(users, dto.IntegrationUserResponse{
			ID:           user.ID.String(),
			UserID:       user.UserID.String(),
			Role:         user.Role,
			UserFullName: user.UserFullName,
			UserLogin:    user.UserLogin,
			UserRole:     user.UserRole,
			UserStatus:   user.UserStatus,
			CreatedAt:    user.CreatedAt,
		})
	}

	isUsable := integration.Active && (integration.ActiveUntil == nil || !integration.ActiveUntil.Before(now))
	return dto.IntegrationResponse{
		ID:          integration.ID.String(),
		Code:        integration.Code,
		Name:        integration.Name,
		Active:      integration.Active,
		ActiveUntil: integration.ActiveUntil,
		IsUsable:    isUsable,
		Users:       users,
		CreatedAt:   integration.CreatedAt,
		UpdatedAt:   integration.UpdatedAt,
	}
}
