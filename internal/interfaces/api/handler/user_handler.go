package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appuser "github.com/prodonik/bank_app/internal/application/user"
	domain "github.com/prodonik/bank_app/internal/domain/user"
	"github.com/prodonik/bank_app/internal/interfaces/api/dto"
)

type UserHandler struct {
	userService *appuser.Service
}

func NewUserHandler(userService *appuser.Service) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with the provided details
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} dto.UserResponse "User created successfully"
// @Failure 400 {object} dto.ErrorResponse "Validation error"
// @Failure 409 {object} dto.ErrorResponse "Login already exists"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), appuser.RegisterInput{
		FullName: req.FullName,
		Login:    req.Login,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.NewUserResponse(user))
}

// Login godoc
// @Summary Login user
// @Description Authenticates user and returns access and refresh tokens
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse "Login successful"
// @Failure 400 {object} dto.ErrorResponse "Validation error"
// @Failure 401 {object} dto.ErrorResponse "Invalid credentials"
// @Failure 403 {object} dto.ErrorResponse "User account is inactive"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	deviceID := uuid.New().String()

	output, err := h.userService.Login(c.Request.Context(), appuser.LoginInput{
		Login:    req.Login,
		Password: req.Password,
		DeviceID: deviceID,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewAuthResponse(output.AccessToken, output.RefreshToken, deviceID, output.User))
}

// Refresh godoc
// @Summary Refresh access token
// @Description Rotates the refresh token and returns a new token pair
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh token and device ID"
// @Success 200 {object} dto.AuthResponse "Tokens refreshed successfully"
// @Failure 400 {object} dto.ErrorResponse "Validation error"
// @Failure 401 {object} dto.ErrorResponse "Invalid or expired refresh token"
// @Failure 403 {object} dto.ErrorResponse "User account is inactive"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/refresh [post]
func (h *UserHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	output, err := h.userService.Refresh(c.Request.Context(), appuser.RefreshInput{
		RefreshToken: req.RefreshToken,
		DeviceID:     req.DeviceID,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewAuthResponse(output.AccessToken, output.RefreshToken, req.DeviceID, output.User))
}

// Logout godoc
// @Summary Logout user
// @Description Invalidates all active sessions for the authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Logged out successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	if err := h.userService.Logout(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// GetMe godoc
// @Summary Get current user
// @Description Returns the authenticated user's profile
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse "User profile"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewUserResponse(user))
}

// GetByID godoc
// @Summary Get user by ID
// @Description Returns a user by their UUID
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} dto.UserResponse "User found"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid user id"})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewUserResponse(user))
}

// GetAll godoc
// @Summary Get all users
// @Description Returns a paginated list of users with optional filters
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param full_name query string false "Filter by full name (partial match)"
// @Param role query string false "Filter by role"
// @Param login query string false "Filter by login (partial match)"
// @Param status query boolean false "Filter by status"
// @Param created_from query string false "Filter by created_at >= (RFC3339)"
// @Param created_to query string false "Filter by created_at <= (RFC3339)"
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} dto.UserListResponse "Users list"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users [get]
func (h *UserHandler) GetAll(c *gin.Context) {
	filter, err := parseUserFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	users, total, err := h.userService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, dto.NewUserListResponse(users, total, filter.Limit, filter.Offset))
}

// Update godoc
// @Summary Update user
// @Description Updates a user's full_name, role, or status
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body dto.UpdateUserRequest true "Fields to update"
// @Success 200 {object} dto.UserResponse "User updated"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid user id"})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, appuser.UpdateInput{
		FullName: req.FullName,
		Role:     req.Role,
		Status:   req.Status,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.NewUserResponse(user))
}

// Delete godoc
// @Summary Delete user
// @Description Deletes a user and all their sessions
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]string "User deleted"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid user id"})
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not found in context")
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user_id type")
	}
	return userID, nil
}

func parseUserFilter(c *gin.Context) (domain.UserFilter, error) {
	var filter domain.UserFilter

	if v := c.Query("full_name"); v != "" {
		filter.FullName = &v
	}
	if v := c.Query("role"); v != "" {
		filter.Role = &v
	}
	if v := c.Query("login"); v != "" {
		filter.Login = &v
	}
	if v := c.Query("status"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return filter, errors.New("invalid status parameter")
		}
		filter.Status = &b
	}
	if v := c.Query("created_from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return filter, errors.New("invalid created_from parameter (use RFC3339)")
		}
		filter.CreatedFrom = &t
	}
	if v := c.Query("created_to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return filter, errors.New("invalid created_to parameter (use RFC3339)")
		}
		filter.CreatedTo = &t
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

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrLoginAlreadyExists):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrUserInactive):
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrSessionNotFound):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrUserNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}
}
