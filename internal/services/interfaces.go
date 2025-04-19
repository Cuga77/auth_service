package services

import (
	"context"
	"net"

	"github.com/auth-service/internal/models"
)

type AuthServiceInterface interface {
	GenerateTokens(ctx context.Context, userID string, ip net.IP) (*models.TokenPair, error)
	RefreshTokens(ctx context.Context, userID, refreshToken string, ip net.IP) (*models.TokenPair, error)
	RevokeAllTokens(ctx context.Context, userID string) error
}

type Notifier interface {
	SendSecurityAlert(userID, message string) error
}

//go:generate mockgen -destination=mock_auth_service.go -package=services . AuthServiceInterface
//go:generate mockgen -destination=mock_notifier.go -package=services . Notifier
