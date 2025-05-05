package utils

import (
	"log"

	"github.com/gin-gonic/gin"
)

func RespondWithError(c *gin.Context, code int, message string, details error) {
	if details != nil {
		log.Printf("Error: %v", details)
		c.JSON(code, gin.H{
			"error": message,
		})
	} else {
		c.JSON(code, gin.H{
			"error": message,
		})
	}
}

func SuccessResponse(c *gin.Context, code int, message string, responseName string, data interface{}) {
	c.JSON(code, gin.H{
		"message":    message,
		responseName: data,
	})
}
