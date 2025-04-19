package handlers

import (
	"net"
	"net/http"

	"github.com/auth-service/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthServiceInterface
	notifier    services.Notifier
}

func NewAuthHandler(authService services.AuthServiceInterface, notifier services.Notifier) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		notifier:    notifier,
	}
}

func (h *AuthHandler) GenerateTokens(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	ip := net.ParseIP(c.ClientIP())
	if ip == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid IP address"})
		return
	}

	tokens, err := h.authService.GenerateTokens(c.Request.Context(), userID, ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetString("user_id")
	err := h.authService.RevokeAllTokens(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "logout failed"})
		return
	}
	c.JSON(200, gin.H{"status": "logged out"})
}
