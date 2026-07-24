package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	appent "github.com/prodonik/bank_app/internal/application/entrepreneur"
	domain "github.com/prodonik/bank_app/internal/domain/entrepreneur"
	"github.com/prodonik/bank_app/internal/interfaces/api/dto"
)

type EntrepreneurHandler struct {
	entrepreneurService *appent.Service
}

func NewEntrepreneurHandler(entrepreneurService *appent.Service) *EntrepreneurHandler {
	return &EntrepreneurHandler{entrepreneurService: entrepreneurService}
}

// CreateEntrepreneur godoc
// @Summary Create a new entrepreneur
// @Description Creates a new entrepreneur. If the INN name exists, it binds the existing INN; otherwise creates a new one.
// @Tags entrepreneurs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateEntrepreneurRequest true "Entrepreneur details"
// @Success 201 {object} dto.EntrepreneurResponse "Entrepreneur created"
// @Failure 400 {object} dto.ErrorResponse "Validation error"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /entrepreneurs [post]
func (h *EntrepreneurHandler) Create(c *gin.Context) {
	var req dto.CreateEntrepreneurRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: formatValidationErrors(err)})
		return
	}

	reqJSON, _ := json.Marshal(req)
	log.Printf("entrepreneur: incoming create request: %s", string(reqJSON))

	activityStatus := true
	if req.ActivityStatus != nil {
		activityStatus = *req.ActivityStatus
	}

	e, err := h.entrepreneurService.Create(c.Request.Context(), appent.CreateInput{
		InnName:               req.Inn,
		LegalName:             req.LegalName,
		RegistrationAuthority: req.RegistrationAuthority,
		RegistrationDate:      req.RegistrationDate,
		RegistrationNumber:    req.RegistrationNumber,
		LegalForm:             req.LegalForm,
		IfutCode:              req.IfutCode,
		ActivityType:          req.ActivityType,
		DbibtCode:             req.DbibtCode,
		ActivityStatus:        activityStatus,
		CharterFund:           req.CharterFund,
		Founders:              req.Founders,
		Email:                 req.Email,
		Phone:                 req.Phone,
		MhobtCode:             req.MhobtCode,
		Address:               req.Address,
		ActivitySubRegion:     req.ActivitySubRegion,
		DirectorName:          req.DirectorName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, dto.NewEntrepreneurResponse(e))
}

// GetEntrepreneurByID godoc
// @Summary Get entrepreneur by ID
// @Description Returns an entrepreneur by its UUID
// @Tags entrepreneurs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Entrepreneur ID (UUID)"
// @Success 200 {object} dto.EntrepreneurResponse "Entrepreneur found"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Entrepreneur not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /entrepreneurs/{id} [get]
func (h *EntrepreneurHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid entrepreneur id"})
		return
	}

	e, err := h.entrepreneurService.GetByID(c.Request.Context(), id)
	if err != nil {
		handleEntrepreneurError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewEntrepreneurResponse(e))
}

// GetAllEntrepreneurs godoc
// @Summary Get all entrepreneurs
// @Description Returns a paginated list of entrepreneurs with optional filters
// @Tags entrepreneurs
// @Produce json
// @Security BearerAuth
// @Param legal_name query string false "Search by legal name (partial match)"
// @Param inn_name query string false "Search by INN name (partial match)"
// @Param activity_status query bool false "Filter by activity status"
// @Param director_name query string false "Search by director name (partial match)"
// @Param date_from query string false "Filter from date (YYYY-MM-DD, e.g. 2024-01-01)"
// @Param date_to query string false "Filter to date inclusive (YYYY-MM-DD, e.g. 2024-12-31)"
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} dto.EntrepreneurListResponse "Entrepreneurs list"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /entrepreneurs [get]
func (h *EntrepreneurHandler) GetAll(c *gin.Context) {
	filter, err := parseEntrepreneurFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	entrepreneurs, total, err := h.entrepreneurService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, dto.NewEntrepreneurListResponse(entrepreneurs, total, filter.Limit, filter.Offset))
}

// UpdateEntrepreneur godoc
// @Summary Update entrepreneur
// @Description Updates an entrepreneur's fields
// @Tags entrepreneurs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Entrepreneur ID (UUID)"
// @Param request body dto.UpdateEntrepreneurRequest true "Fields to update"
// @Success 200 {object} dto.EntrepreneurResponse "Entrepreneur updated"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Entrepreneur not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /entrepreneurs/{id} [put]
func (h *EntrepreneurHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid entrepreneur id"})
		return
	}

	var req dto.UpdateEntrepreneurRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	e, err := h.entrepreneurService.Update(c.Request.Context(), id, appent.UpdateInput{
		LegalName:             req.LegalName,
		RegistrationAuthority: req.RegistrationAuthority,
		RegistrationDate:      req.RegistrationDate,
		RegistrationNumber:    req.RegistrationNumber,
		LegalForm:             req.LegalForm,
		IfutCode:              req.IfutCode,
		ActivityType:          req.ActivityType,
		DbibtCode:             req.DbibtCode,
		ActivityStatus:        req.ActivityStatus,
		CharterFund:           req.CharterFund,
		Founders:              req.Founders,
		Email:                 req.Email,
		Phone:                 req.Phone,
		MhobtCode:             req.MhobtCode,
		Address:               req.Address,
		ActivitySubRegion:     req.ActivitySubRegion,
		DirectorName:          req.DirectorName,
	})
	if err != nil {
		handleEntrepreneurError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewEntrepreneurResponse(e))
}

// DeleteEntrepreneur godoc
// @Summary Delete entrepreneur
// @Description Deletes an entrepreneur by its UUID
// @Tags entrepreneurs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Entrepreneur ID (UUID)"
// @Success 200 {object} map[string]string "Entrepreneur deleted"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Entrepreneur not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /entrepreneurs/{id} [delete]
func (h *EntrepreneurHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid entrepreneur id"})
		return
	}

	if err := h.entrepreneurService.Delete(c.Request.Context(), id); err != nil {
		handleEntrepreneurError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "entrepreneur deleted successfully"})
}

// GetSqbFailed godoc
// @Summary Get entrepreneurs with SQB errors
// @Description Returns a paginated list of entrepreneurs that failed to send to SQB
// @Tags entrepreneurs
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} dto.EntrepreneurListResponse "Failed entrepreneurs list"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /entrepreneurs/sqb-failed [get]
func (h *EntrepreneurHandler) GetSqbFailed(c *gin.Context) {
	limit := 20
	offset := 0
	if v := c.Query("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil {
			limit = l
		}
	}
	if v := c.Query("offset"); v != "" {
		if o, err := strconv.Atoi(v); err == nil {
			offset = o
		}
	}

	entrepreneurs, total, err := h.entrepreneurService.GetAllWithSqbError(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, dto.NewEntrepreneurListResponse(entrepreneurs, total, limit, offset))
}

// RetrySqbFailed godoc
// @Summary Retry sending failed entrepreneurs to SQB
// @Description Fetches all entrepreneurs with sqb_api_error and resends them to SQB. Clears the error on success.
// @Tags entrepreneurs
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Retry results"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /entrepreneurs/sqb-retry [post]
func (h *EntrepreneurHandler) RetrySqbFailed(c *gin.Context) {
	sent, failed := h.entrepreneurService.RetrySqbFailed(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"message": "retry complete",
		"sent":    sent,
		"failed":  failed,
	})
}

// UpdateBirdarchaToken godoc
// @Summary Update birdarcha access token
// @Description Updates the birdarcha API access token used by the background syncer
// @Tags entrepreneurs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateBirdarchaTokenRequest true "New token"
// @Success 200 {object} map[string]string "Token updated"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /entrepreneurs/birdarcha-token [put]
func (h *EntrepreneurHandler) UpdateBirdarchaToken(c *gin.Context) {
	var req dto.UpdateBirdarchaTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "token is required"})
		return
	}

	if err := h.entrepreneurService.UpdateBirdarchaToken(c.Request.Context(), req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "birdarcha token updated"})
}

// CheckTokenRefreshNeeded godoc
// @Summary Check if birdarcha token refresh is needed
// @Description Returns whether the extension should reload birdarcha.uz and push a fresh token
// @Tags entrepreneurs
// @Produce json
// @Success 200 {object} map[string]bool "needs_refresh flag"
// @Router /entrepreneurs/birdarcha-token/needs-refresh [get]
func (h *EntrepreneurHandler) CheckTokenRefreshNeeded(c *gin.Context) {
	needed, err := h.entrepreneurService.NeedsTokenRefresh(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"needs_refresh": needed})
}

func parseEntrepreneurFilter(c *gin.Context) (domain.EntrepreneurFilter, error) {
	var filter domain.EntrepreneurFilter

	if v := c.Query("legal_name"); v != "" {
		filter.LegalName = &v
	}
	if v := c.Query("inn_name"); v != "" {
		filter.InnName = &v
	}
	if v := c.Query("activity_status"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return filter, errors.New("invalid activity_status parameter")
		}
		filter.ActivityStatus = &b
	}
	if v := c.Query("director_name"); v != "" {
		filter.DirectorName = &v
	}
	if v := c.Query("date_from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filter, errors.New("invalid date_from parameter, use YYYY-MM-DD format")
		}
		filter.DateFrom = &t
	}
	if v := c.Query("date_to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filter, errors.New("invalid date_to parameter, use YYYY-MM-DD format")
		}
		// Set to end of day to make it inclusive
		endOfDay := t.Add(24*time.Hour - time.Nanosecond)
		filter.DateTo = &endOfDay
	}

	filter.Limit = 20
	if v := c.Query("limit"); v != "" {
		l, err := strconv.Atoi(v)
		if err != nil {
			return filter, errors.New("invalid limit parameter")
		}
		filter.Limit = l
	}

	if v := c.Query("offset"); v != "" {
		o, err := strconv.Atoi(v)
		if err != nil {
			return filter, errors.New("invalid offset parameter")
		}
		filter.Offset = o
	}

	return filter, nil
}

func handleEntrepreneurError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEntrepreneurNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}
}

func formatValidationErrors(err error) string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return err.Error()
	}

	msgs := make([]string, 0, len(ve))
	for _, fe := range ve {
		field := fe.Field()
		switch fe.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", field))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", field, fe.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s characters", field, fe.Param()))
		case "email":
			msgs = append(msgs, fmt.Sprintf("%s must be a valid email address", field))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", field))
		}
	}

	result := msgs[0]
	for i := 1; i < len(msgs); i++ {
		result += "; " + msgs[i]
	}
	return result
}
