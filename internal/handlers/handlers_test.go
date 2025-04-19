package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auth-service/internal/handlers"
	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := services.NewMockAuthServiceInterface(ctrl)
	handler := handlers.NewAuthHandler(mockAuth, nil)

	t.Run("GenerateTokens", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/tokens?user_id=test", nil)
			c.Request.RemoteAddr = "192.168.1.1:1234"

			mockAuth.EXPECT().
				GenerateTokens(gomock.Any(), "test", gomock.Any()).
				Return(&models.TokenPair{
					AccessToken:  "access",
					RefreshToken: "refresh",
				}, nil)

			handler.GenerateTokens(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	})

	t.Run("RefreshTokens", func(t *testing.T) {
		t.Run("Valid request", func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/refresh", bytes.NewBufferString(
				`{"user_id": "user1", "refresh_token": "token"}`,
			))
			c.Request.RemoteAddr = "192.168.1.1:1234"

			mockAuth.EXPECT().
				RefreshTokens(gomock.Any(), "user1", "token", gomock.Any()).
				Return(&models.TokenPair{
					AccessToken:  "new-access",
					RefreshToken: "new-refresh",
				}, nil)

			handler.RefreshTokens(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	})
}
