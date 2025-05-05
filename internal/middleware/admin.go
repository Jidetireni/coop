package middleware

import (
	"cooperative-system/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAdmin(c *gin.Context) {
	user, exist := c.Get("user")
	if !exist {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "unable to get user from token",
		})
		return
	}

	authUser, ok := user.(models.User)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user data in context",
		})
		return
	}

	if authUser.Role != "admin" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "you are not authorized to perform this action",
		})
		return
	}

	c.Next()
}
