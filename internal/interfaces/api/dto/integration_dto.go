package dto

import "time"

type IntegrationUserResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Role         string    `json:"role"`
	UserFullName string    `json:"user_full_name"`
	UserLogin    string    `json:"user_login"`
	UserRole     string    `json:"user_role"`
	UserStatus   bool      `json:"user_status"`
	CreatedAt    time.Time `json:"created_at"`
}

type IntegrationResponse struct {
	ID          string                    `json:"id"`
	Code        string                    `json:"code"`
	Name        string                    `json:"name"`
	Active      bool                      `json:"active"`
	ActiveUntil *time.Time                `json:"active_until"`
	IsUsable    bool                      `json:"is_usable"`
	Users       []IntegrationUserResponse `json:"users"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

type UpdateIntegrationRequest struct {
	Active      bool       `json:"active"`
	ActiveUntil *time.Time `json:"active_until"`
}
