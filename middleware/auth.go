package middleware

import (
	config "cooperative-system/conf"
	"cooperative-system/models"
	"cooperative-system/pkg/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAuth(c *gin.Context) {
	// Get the token off the in the request cookie
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "operation not allowed, could not get token from cookie",
			"details": err.Error(),
		})
		return
	}

	// verify the token extracted from the cookie
	email, err := util.VerifyToken(tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "operation not allowed, could not verify token",
			"details": err.Error(),
		})
		return
	}

	var user models.User
	config.DB.Where("email = ?", email).First(&user)
	if user.ID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "operation not allowed, user not found",
			"details": err.Error(),
		})

		return
	}

	c.Set("user", user)
	c.Next()
}
