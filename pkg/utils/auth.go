package utils

import (
	"cooperative-system/internal/models"

	"github.com/gin-gonic/gin"
)

// GetAuthUser extracts the authenticated user from the context
func GetAuthUser(c *gin.Context) (models.User, bool) {
	user, exist := c.Get("user")
	if !exist {
		return models.User{}, false
	}

	authUser, ok := user.(models.User)
	if !ok {
		return models.User{}, false
	}
	return authUser, true
}
