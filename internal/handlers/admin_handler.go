package handlers

import (
	"cooperative-system/internal/config"
	"cooperative-system/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateAdmin(c *gin.Context) {
	user, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to get user from token",
		})
		return
	}

	authUser, ok := user.(models.User)
	if !ok || authUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "only admins can create other admins",
		})
		return
	}

	var reqBody struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	var userToPromote models.User
	if err := config.DB.Where("email = ?", reqBody.Email).First(&userToPromote).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	tx := config.DB.Model(&userToPromote).Update("Role", "admin")
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create admin",
			"details": tx.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "created admin sucessfully",
		"admin":   userToPromote,
	})

}

func DeleteMember(c *gin.Context) {
	user, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to get user from token",
		})
		return
	}

	authUser, ok := user.(models.User)
	if !ok || authUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "only admins can create other admins",
		})
		return
	}

	var reqBody struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	var memberToDelete models.Member
	tx := config.DB.Where("user_id = ?", reqBody.Email).Preload("User").First(&memberToDelete)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	tx = config.DB.Delete(&memberToDelete)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete member",
			"details": tx.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "user deleted",
	})

}
