package services

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/repository"
	"github.com/auth-service/internal/repository/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockNotifier := NewMockNotifier(ctrl)
	tokenSvc := NewTokenService("test-secret")
	authSvc := NewAuthService(mockRepo, tokenSvc, mockNotifier)
	ctx := context.Background()
	userIP := net.ParseIP("192.168.1.1")

	t.Run("GenerateTokens", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo.EXPECT().
				SaveRefreshToken(gomock.Any(), "user1", gomock.Any(), userIP.String()).
				Return(nil)

			pair, err := authSvc.GenerateTokens(ctx, "user1", userIP)
			require.NoError(t, err)
			assert.NotEmpty(t, pair.AccessToken)
			assert.NotEmpty(t, pair.RefreshToken)
		})

		t.Run("Database error", func(t *testing.T) {
			mockRepo.EXPECT().
				SaveRefreshToken(gomock.Any(), "user1", gomock.Any(), userIP.String()).
				Return(repository.ErrDatabase)

			_, err := authSvc.GenerateTokens(ctx, "user1", userIP)
			assert.ErrorIs(t, err, repository.ErrDatabase)
		})
	})

	t.Run("RefreshTokens", func(t *testing.T) {
		refreshToken := "valid-refresh-token"
		hashedToken, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
		storedToken := models.RefreshToken{
			ID:        "token-id",
			UserID:    "user1",
			TokenHash: string(hashedToken),
			IP:        userIP.String(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		t.Run("Valid refresh", func(t *testing.T) {
			mockRepo.EXPECT().
				GetRefreshTokensByUser(ctx, "user1").
				Return([]models.RefreshToken{storedToken}, nil)

			mockRepo.EXPECT().
				DeleteRefreshToken(ctx, "token-id").
				Return(nil)

			mockRepo.EXPECT().
				SaveRefreshToken(gomock.Any(), "user1", gomock.Any(), userIP.String()).
				Return(nil)

			pair, err := authSvc.RefreshTokens(ctx, "user1", refreshToken, userIP)
			require.NoError(t, err)
			assert.NotEmpty(t, pair.AccessToken)
		})

		t.Run("Expired token", func(t *testing.T) {
			expiredToken := storedToken
			expiredToken.ExpiresAt = time.Now().Add(-1 * time.Hour)

			mockRepo.EXPECT().
				GetRefreshTokensByUser(ctx, "user1").
				Return([]models.RefreshToken{expiredToken}, nil)

			mockRepo.EXPECT().
				DeleteRefreshToken(ctx, "token-id").
				Return(nil)

			_, err := authSvc.RefreshTokens(ctx, "user1", refreshToken, userIP)
			assert.ErrorContains(t, err, "expired")
		})
	})
}
