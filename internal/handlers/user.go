package handlers

import "github.com/gin-gonic/gin"

func (h *AuthHandler) GetUserData(c *gin.Context) {
	userID := c.GetString("user_id")
	c.JSON(200, gin.H{"user_id": userID, "data": "secure_content"})
}
