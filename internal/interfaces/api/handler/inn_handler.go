package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	appinn "github.com/prodonik/bank_app/internal/application/inn"
	domain "github.com/prodonik/bank_app/internal/domain/inn"
	"github.com/prodonik/bank_app/internal/interfaces/api/dto"
)

type InnHandler struct {
	innService *appinn.Service
}

func NewInnHandler(innService *appinn.Service) *InnHandler {
	return &InnHandler{innService: innService}
}

// GetAllINNs godoc
// @Summary Get all INNs
// @Description Returns a paginated list of INNs with optional name search
// @Tags inns
// @Produce json
// @Security BearerAuth
// @Param name query string false "Search by name (partial match)"
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} dto.InnListResponse "INNs list"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /inns [get]
func (h *InnHandler) GetAll(c *gin.Context) {
	filter, err := parseInnFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	inns, total, err := h.innService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, dto.NewInnListResponse(inns, total, filter.Limit, filter.Offset))
}

func parseInnFilter(c *gin.Context) (domain.InnFilter, error) {
	var filter domain.InnFilter

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
