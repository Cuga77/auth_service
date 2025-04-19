package repository

import (
	"context"
	"errors"

	"github.com/auth-service/internal/models"
)

var (
	ErrDatabase = errors.New("database error")
)

type Repository interface {
	SaveRefreshToken(ctx context.Context, userID, tokenHash, ip string) error
	GetRefreshTokensByUser(ctx context.Context, userID string) ([]models.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, id string) error
	RevokeAllTokens(ctx context.Context, userID string) error
	Close() error
}

//go:generate mockgen -destination=repository_mock.go -package=repository github.com/auth-service/internal/repository Repository