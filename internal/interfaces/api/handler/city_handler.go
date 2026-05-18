package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appcity "github.com/prodonik/bank_app/internal/application/city"
	domain "github.com/prodonik/bank_app/internal/domain/city"
	"github.com/prodonik/bank_app/internal/interfaces/api/dto"
)

type CityHandler struct {
	cityService *appcity.Service
}

func NewCityHandler(cityService *appcity.Service) *CityHandler {
	return &CityHandler{cityService: cityService}
}

// CreateCity godoc
// @Summary Create a new city
// @Description Creates a new city with the given name
// @Tags cities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCityRequest true "City details"
// @Success 201 {object} dto.CityResponse "City created"
// @Failure 400 {object} dto.ErrorResponse "Validation error"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /cities [post]
func (h *CityHandler) Create(c *gin.Context) {
	var req dto.CreateCityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	city, err := h.cityService.Create(c.Request.Context(), appcity.CreateInput{
		Name: req.Name,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, dto.NewCityResponse(city))
}

// GetCityByID godoc
// @Summary Get city by ID
// @Description Returns a city by its UUID
// @Tags cities
// @Produce json
// @Security BearerAuth
// @Param id path string true "City ID (UUID)"
// @Success 200 {object} dto.CityResponse "City found"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "City not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /cities/{id} [get]
func (h *CityHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid city id"})
		return
	}

	city, err := h.cityService.GetByID(c.Request.Context(), id)
	if err != nil {
		handleCityError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewCityResponse(city))
}

// GetAllCities godoc
// @Summary Get all cities
// @Description Returns a paginated list of cities with optional name search
// @Tags cities
// @Produce json
// @Security BearerAuth
// @Param name query string false "Search by name (partial match)"
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} dto.CityListResponse "Cities list"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /cities [get]
func (h *CityHandler) GetAll(c *gin.Context) {
	filter, err := parseCityFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	cities, total, err := h.cityService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, dto.NewCityListResponse(cities, total, filter.Limit, filter.Offset))
}

// UpdateCity godoc
// @Summary Update city
// @Description Updates a city's name
// @Tags cities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "City ID (UUID)"
// @Param request body dto.UpdateCityRequest true "Fields to update"
// @Success 200 {object} dto.CityResponse "City updated"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "City not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /cities/{id} [put]
func (h *CityHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid city id"})
		return
	}

	var req dto.UpdateCityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	city, err := h.cityService.Update(c.Request.Context(), id, appcity.UpdateInput{
		Name: req.Name,
	})
	if err != nil {
		handleCityError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewCityResponse(city))
}

// DeleteCity godoc
// @Summary Delete city
// @Description Deletes a city by its UUID
// @Tags cities
// @Produce json
// @Security BearerAuth
// @Param id path string true "City ID (UUID)"
// @Success 200 {object} map[string]string "City deleted"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "City not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /cities/{id} [delete]
func (h *CityHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid city id"})
		return
	}

	if err := h.cityService.Delete(c.Request.Context(), id); err != nil {
		handleCityError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "city deleted successfully"})
}

func parseCityFilter(c *gin.Context) (domain.CityFilter, error) {
	var filter domain.CityFilter

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

func handleCityError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrCityNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}
}
