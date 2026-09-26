package services

import (
	"context"
	"errors"
	"time"

	"github.com/Aayx2hOG/automata/internal/auth"
	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/repositories"
)

type AuthService struct {
	users         repositories.UserRepository
	refreshTokens repositories.RefreshTokenRepository
	jwtManager    *auth.JWTManager
	refreshTTL    time.Duration
}

func NewAuthService(
	users repositories.UserRepository,
	refreshTokens repositories.RefreshTokenRepository,
	jwtManager *auth.JWTManager,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:         users,
		refreshTokens: refreshTokens,
		jwtManager:    jwtManager,
		refreshTTL:    refreshTTL,
	}
}

type AuthResult struct {
	User         *models.User
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*AuthResult, error) {
	existingUser, err := s.users.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, models.ErrorUserAlreadyExists
	}
	if err != nil && !errors.Is(err, models.ErrorUserNotFound) {
		return nil, err
	}
	hash, err := auth.HashPassword(password, auth.DefaultArgon2Params())
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		PasswordHash: hash,
		Role:         models.RoleUser,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, models.ErrorUserNotFound) {
			return nil, models.ErrorInvalidCredentials
		}
		return nil, err
	}
	valid, err := auth.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, models.ErrorInvalidCredentials
	}
	if !user.IsActive {
		return nil, models.ErrorInvalidCredentials
	}
	return s.issueTokens(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (*AuthResult, error) {
	hash := auth.HashRefreshToken(rawRefreshToken)

	stored, err := s.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	if stored.RevokedAt != nil {
		_ = s.refreshTokens.RevokeAllForUser(ctx, stored.UserID)
		return nil, models.ErrorInvalidToken
	}
	if time.Now().After(stored.ExpiresAt) {
		return nil, models.ErrorInvalidToken
	}

	user, err := s.users.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, models.ErrorInvalidToken
	}
	access, err := s.jwtManager.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	raw, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if err = s.refreshTokens.Rotate(ctx, hash, auth.HashRefreshToken(raw), user.ID, time.Now().Add(s.refreshTTL)); err != nil {
		return nil, err
	}
	return &AuthResult{User: user, AccessToken: access, RefreshToken: raw}, nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := auth.HashRefreshToken(rawRefreshToken)
	return s.refreshTokens.Revoke(ctx, hash)
}

func (s *AuthService) issueTokens(ctx context.Context, user *models.User) (*AuthResult, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	rawRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.refreshTTL)
	hash := auth.HashRefreshToken(rawRefreshToken)
	if err := s.refreshTokens.Create(ctx, user.ID, hash, expiresAt); err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
	}, nil
}
