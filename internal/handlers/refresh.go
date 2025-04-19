package handlers

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type refreshRequest struct {
	UserID       string `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) RefreshTokens(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	clientIP := net.ParseIP(c.ClientIP())
	if clientIP == nil {
		clientIP = net.IPv4(0, 0, 0, 0)
	}

	tokens, err := h.authService.RefreshTokens(
		c.Request.Context(),
		req.UserID,
		req.RefreshToken,
		clientIP,
	)

	if err != nil {
		errorMsg := "failed to refresh tokens"
		if err.Error() == "refresh token not found in DB" {
			c.JSON(http.StatusNotFound, gin.H{"error": errorMsg + ": token not found"})
		} else if err.Error() == "refresh token expired" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": errorMsg + ": token expired"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		}
		return
	}

	c.JSON(http.StatusOK, tokens)
}
