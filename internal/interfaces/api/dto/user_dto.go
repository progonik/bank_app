package dto

import (
	"time"

	"github.com/google/uuid"

	domain "github.com/prodonik/bank_app/internal/domain/user"
)

type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Login    string `json:"login" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	DeviceID     string `json:"device_id" binding:"required"`
}

type UpdateUserRequest struct {
	FullName *string `json:"full_name"`
	Role     *string `json:"role"`
	Status   *bool   `json:"status"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	DeviceID     string       `json:"device_id"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	Login     string    `json:"login"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type UserListResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		FullName:  u.FullName,
		Role:      u.Role,
		Login:     u.Login,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

func NewUserListResponse(users []*domain.User, total, limit, offset int) UserListResponse {
	resp := UserListResponse{
		Users:  make([]UserResponse, 0, len(users)),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	for _, u := range users {
		resp.Users = append(resp.Users, NewUserResponse(u))
	}
	return resp
}

func NewAuthResponse(accessToken, refreshToken, deviceID string, u *domain.User) AuthResponse {
	return AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		DeviceID:     deviceID,
		User:         NewUserResponse(u),
	}
}
