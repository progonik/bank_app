package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	appifut "github.com/prodonik/bank_app/internal/application/ifut_code"
	domain "github.com/prodonik/bank_app/internal/domain/ifut_code"
	"github.com/prodonik/bank_app/internal/interfaces/api/dto"
)

type IfutCodeHandler struct {
	ifutCodeService *appifut.Service
}

func NewIfutCodeHandler(ifutCodeService *appifut.Service) *IfutCodeHandler {
	return &IfutCodeHandler{ifutCodeService: ifutCodeService}
}

// GetAllIfutCodes godoc
// @Summary Get all IFUT codes
// @Description Returns a paginated list of IFUT codes with optional name search
// @Tags ifut-codes
// @Produce json
// @Security BearerAuth
// @Param name query string false "Search by name (partial match)"
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} dto.IfutCodeListResponse "IFUT codes list"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /ifut-codes [get]
func (h *IfutCodeHandler) GetAll(c *gin.Context) {
	filter, err := parseIfutCodeFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	codes, total, err := h.ifutCodeService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, dto.NewIfutCodeListResponse(codes, total, filter.Limit, filter.Offset))
}

func parseIfutCodeFilter(c *gin.Context) (domain.IfutCodeFilter, error) {
	var filter domain.IfutCodeFilter

	if v := c.Query("name"); v != "" {
		filter.Name = &v
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
