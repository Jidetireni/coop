package handlers

import (
	"cooperative-system/internal/models"
	"cooperative-system/pkg/util"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// Role     string `json:"role"`
}

type UserHandler struct {
	DB *gorm.DB
}

type UserService interface {
	Signup(c *gin.Context)
	Login(c *gin.Context)
}

func (u *UserHandler) Signup(c *gin.Context) {

	// Get the email/pass off the req body
	var reqBody RequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqBody.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to hash password",
			"details": err.Error(),
		})
		return
	}

	// Create the user
	user := models.User{
		Email:    reqBody.Email,
		Password: string(hashedPassword),
		Role:     "member",
	}

	if err := u.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to create user",
			"details": err.Error(),
		})
		return
	}

	// Respond
	c.JSON(http.StatusCreated, gin.H{
		"message": "user created suceesfully",
	})
}

func (u *UserHandler) Login(c *gin.Context) {
	// Get the email and pass off the req body
	var reqBody RequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Look up requested user
	var user models.User
	u.DB.Where("email = ? ", reqBody.Email).First(&user)
	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not found",
		})
		return
	}

	// Compare sent in pass with saved user pass hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqBody.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "incorrect paswword",
			"details": err.Error(),
		})
	}

	// generate a jwt token
	tokenString, err := util.CreateToken(user.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unable to create token",
			"details": err.Error(),
		})
		return
	}

	// send it back
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"userID": user.ID,
		"token":  tokenString,
	})

}

func Validate(c *gin.Context) {

	user, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to get user from token",
		})
		return
	}

	authUser, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user data in context",
		})
		return
	}

	userID := c.Param("id")
	if userID != fmt.Sprintf("%v", authUser.ID) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user ID mismatch",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User validated successfully",
		"user":    user,
	})
}
