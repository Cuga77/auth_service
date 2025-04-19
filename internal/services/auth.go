package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo         repository.Repository
	tokenService *TokenService
	notifier     Notifier
}

func NewAuthService(repo repository.Repository, tokenService *TokenService, notifier Notifier) *AuthService {
	return &AuthService{
		repo:         repo,
		tokenService: tokenService,
		notifier:     notifier,
	}
}

func (s *AuthService) RevokeAllTokens(ctx context.Context, userID string) error {
	return s.repo.RevokeAllTokens(ctx, userID)
}

func (s *AuthService) GenerateTokens(ctx context.Context, userID string, ip net.IP) (*models.TokenPair, error) {
	accessToken, err := s.tokenService.GenerateAccessToken(userID, ip)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash refresh token: %w", err)
	}

	err = s.repo.SaveRefreshToken(ctx, userID, string(hashedToken), ip.String())
	if err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshTokens(ctx context.Context, userID, refreshToken string, clientIP net.IP) (*models.TokenPair, error) {
	if refreshToken == "" {
		return nil, errors.New("empty refresh token")
	}

	tokens, err := s.repo.GetRefreshTokensByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tokens: %w", err)
	}

	var storedToken *models.RefreshToken
	for _, token := range tokens {
		if err := bcrypt.CompareHashAndPassword([]byte(token.TokenHash), []byte(refreshToken)); err == nil {
			storedToken = &token
			break
		}
	}

	if storedToken == nil {
		return nil, errors.New("refresh token not found in DB")
	}

	if storedToken.IP != clientIP.String() {
		msg := fmt.Sprintf("Обнаружена смена IP адреса для пользователя %s. Старый IP: %s, Новый IP: %s",
			userID, storedToken.IP, clientIP.String())
		s.notifier.SendSecurityAlert(userID, msg)

		log.Printf("SECURITY WARNING: %s", msg)
	}

	if time.Now().After(storedToken.ExpiresAt) {
		if err := s.repo.DeleteRefreshToken(ctx, storedToken.ID); err != nil {
			log.Printf(
				"Failed to delete expired refresh token. UserID: %s, TokenID: %s, Error: %v",
				userID,
				storedToken.ID,
				err,
			)
		}
		return nil, errors.New("refresh token expired")
	}

	if storedToken.IP != clientIP.String() {
		log.Printf("SECURITY WARNING: IP changed from %s to %s", storedToken.IP, clientIP.String())
	}

	if err := s.repo.DeleteRefreshToken(ctx, storedToken.ID); err != nil {
		return nil, fmt.Errorf("failed to delete old token: %w", err)
	}

	return s.GenerateTokens(ctx, userID, clientIP)
}
