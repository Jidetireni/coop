package handlers

import (
	"cooperative-system/internal/models"
	"cooperative-system/internal/repository"
	"cooperative-system/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserHandler struct {
	UserRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		UserRepo: userRepo,
	}
}

type UserService interface {
	Signup(c *gin.Context)
	Login(c *gin.Context)
}

func (u *UserHandler) Signup(c *gin.Context) {
	var reqBody RequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqBody.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to hash password", err)
		return
	}

	user := models.User{
		Email:    reqBody.Email,
		Password: string(hashedPassword),
		Role:     "member",
	}

	createdUser, msg, err := u.UserRepo.CreateUser(&user)
	if err != nil {
		if createdUser == nil {
			utils.RespondWithError(c, http.StatusBadRequest, msg, err)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		}
		return
	}

	data := models.NewUserResponse(createdUser)
	utils.SuccessResponse(c, http.StatusCreated, "user created successfully", "user", data)
}

func (u *UserHandler) Login(c *gin.Context) {
	var reqBody RequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	user, msg, err := u.UserRepo.FindUserByEmail(reqBody.Email)
	if err != nil || user.ID == 0 {
		if user == nil {
			utils.RespondWithError(c, http.StatusNotFound, msg, nil)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqBody.Password)); err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "incorrect password", err)
		return
	}

	tokenString, err := utils.CreateToken(user.Email)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "unable to create token", err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	utils.SuccessResponse(c, http.StatusOK, "login successful", "data", gin.H{
		"user_id": user.ID,
		"token":  tokenString,
	})
}

// func Validate(c *gin.Context) {
// 	user, exist := c.Get("user")
// 	if !exist {
// 		utils.RespondWithError(c, http.StatusInternalServerError, "unable to get user from token", nil)
// 		return
// 	}

// 	authUser, ok := user.(models.User)
// 	if !ok {
// 		utils.RespondWithError(c, http.StatusInternalServerError, "invalid user data in context", nil)
// 		return
// 	}

// 	userID := c.Param("id")
// 	if userID != fmt.Sprintf("%v", authUser.ID) {
// 		utils.RespondWithError(c, http.StatusForbidden, "user ID mismatch", nil)
// 		return
// 	}

// 	utils.SuccessResponse(c, http.StatusOK, "User validated successfully", "user", user)
// }
