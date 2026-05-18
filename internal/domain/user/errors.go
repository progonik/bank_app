package user

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrSessionNotFound    = errors.New("session not found")
)
