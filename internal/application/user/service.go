package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domain "github.com/prodonik/bank_app/internal/domain/user"
	"github.com/prodonik/bank_app/internal/infrastructure/auth"
)

type Service struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	jwtService  *auth.JWTService
}

func NewService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	jwtService *auth.JWTService,
) *Service {
	return &Service{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtService:  jwtService,
	}
}

type RegisterInput struct {
	FullName string
	Login    string
	Password string
	Role     string
}

type LoginInput struct {
	Login    string
	Password string
	DeviceID string
}

type RefreshInput struct {
	RefreshToken string
	DeviceID     string
}

type UpdateInput struct {
	FullName *string
	Role     *string
	Status   *bool
}

type AuthOutput struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	existing, err := s.userRepo.GetByLogin(ctx, input.Login)
	if err == nil && existing != nil {
		return nil, domain.ErrLoginAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role := input.Role
	if role == "" {
		role = "user"
	}

	u := &domain.User{
		FullName:     input.FullName,
		Role:         role,
		Login:        input.Login,
		PasswordHash: string(hashedPassword),
		Status:       true,
	}

	created, err := s.userRepo.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return created, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthOutput, error) {
	u, err := s.userRepo.GetByLogin(ctx, input.Login)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !u.Status {
		return nil, domain.ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.createOrUpdateSession(ctx, u, input.DeviceID)
}

func (s *Service) Refresh(ctx context.Context, input RefreshInput) (*AuthOutput, error) {
	claims, err := s.jwtService.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokenHash := auth.HashToken(input.RefreshToken)
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	u, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !u.Status {
		return nil, domain.ErrUserInactive
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(u.ID, u.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	newTokenHash := auth.HashToken(tokenPair.RefreshToken)
	if _, err := s.sessionRepo.UpdateToken(ctx, session.ID, newTokenHash); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &AuthOutput{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         u,
	}, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context, filter domain.UserFilter) ([]*domain.User, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.userRepo.GetAll(ctx, filter)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.User, error) {
	existing, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.FullName != nil {
		existing.FullName = *input.FullName
	}
	if input.Role != nil {
		existing.Role = *input.Role
	}
	if input.Status != nil {
		existing.Status = *input.Status
	}

	return s.userRepo.Update(ctx, existing)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.userRepo.GetByID(ctx, id); err != nil {
		return err
	}

	// Delete sessions first
	_ = s.sessionRepo.DeleteByUserID(ctx, id)

	return s.userRepo.Delete(ctx, id)
}

func (s *Service) createOrUpdateSession(ctx context.Context, u *domain.User, deviceID string) (*AuthOutput, error) {
	tokenPair, err := s.jwtService.GenerateTokenPair(u.ID, u.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	refreshTokenHash := auth.HashToken(tokenPair.RefreshToken)

	existing, err := s.sessionRepo.GetByUserAndDevice(ctx, u.ID, deviceID)
	if err == nil && existing != nil {
		if _, err := s.sessionRepo.UpdateToken(ctx, existing.ID, refreshTokenHash); err != nil {
			return nil, fmt.Errorf("failed to update session: %w", err)
		}
	} else {
		session := &domain.Session{
			UserID:           u.ID,
			DeviceID:         deviceID,
			RefreshTokenHash: refreshTokenHash,
		}
		if _, err := s.sessionRepo.Create(ctx, session); err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	return &AuthOutput{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         u,
	}, nil
}
