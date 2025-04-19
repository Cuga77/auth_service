package services_test

import (
	"net"
	"testing"

	"github.com/auth-service/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenService(t *testing.T) {
	ts := services.NewTokenService("test-secret")
	userIP := net.ParseIP("192.168.1.1")

	t.Run("GenerateAccessToken", func(t *testing.T) {
		token, err := ts.GenerateAccessToken("user1", userIP)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("ParseAccessToken", func(t *testing.T) {
		token, _ := ts.GenerateAccessToken("user1", userIP)

		claims, err := ts.ParseAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user1", claims.UserID)
		assert.Equal(t, userIP.String(), claims.IP)
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) {
		token1, err := ts.GenerateRefreshToken()
		require.NoError(t, err)

		token2, err := ts.GenerateRefreshToken()
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2)
		assert.NotEmpty(t, token1)
	})
}
