package middleware

import (
	"log"
	"strings"

	"github.com/auth-service/internal/services"
	"github.com/gin-gonic/gin"
)

func JWTValidator(tokenService *services.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header required"})
			return
		}

		if len(tokenString) > 7 && strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = tokenString[7:]
		}

		claims, err := tokenService.ParseAccessToken(tokenString)
		if err != nil {
			log.Printf("JWT validation failed: %v", err)
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("ip", claims.IP)
		c.Next()
	}
}
