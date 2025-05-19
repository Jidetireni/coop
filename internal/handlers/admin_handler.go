package handlers

import (
	"cooperative-system/internal/config"
	"cooperative-system/internal/models"
	"cooperative-system/internal/repository"
	"cooperative-system/pkg/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getMemberByIDAndAuthorize(c *gin.Context, repo *repository.MemberRepository, authUser *models.User) (*models.Member, error) {
	// Get the member ID from the URL parameter
	memberID := c.Param("id")
	var member models.Member

	// Check if the member exists in the database
	fetchedMember, message, err := repo.FetchByID(memberID, &member)

	if authUser.Role != "admin" && fetchedMember.UserID != authUser.ID {
		utils.RespondWithError(c, http.StatusUnauthorized, message, err)
		return nil, fmt.Errorf("unauthorized access")
	}

	// Return the member if everything is fine
	return &member, nil
}

func CreateAdmin(c *gin.Context) {
	user, exist := c.Get("user")
	if !exist {
		utils.RespondWithError(c, http.StatusInternalServerError, "unable to get user from token", nil)
		return
	}

	authUser, ok := user.(models.User)
	if !ok || authUser.Role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "only admins can create other admins", nil)
		return
	}

	var reqBody struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Use UserRepository to find the user by email
	userRepo := repository.NewUserRepository(config.DB) // Assuming config.DB is your gorm.DB instance
	userToPromote, err := userRepo.FindUserByEmail(reqBody.Email)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "user not found", err)
		return
	}

	// Update user role to admin
	// This assumes you have an UpdateUser method in your UserRepository
	// If not, you might need to add it or handle the update differently
	userToPromote.Role = "admin"
	if err := config.DB.Save(userToPromote).Error; err != nil { // Or userRepo.UpdateUser(userToPromote)
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to promote user to admin", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "created admin successfully", "admin", userToPromote)
}

func DeleteMember(c *gin.Context) {
	user, exist := c.Get("user")
	if !exist {
		utils.RespondWithError(c, http.StatusInternalServerError, "unable to get user from token", nil)
		return
	}

	authUser, ok := user.(models.User)
	if !ok || authUser.Role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "only admins can delete members", nil) // Clarified error message
		return
	}

	var reqBody struct {
		Email string `json:"email"` // Assuming you want to identify the member by the user's email
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Use UserRepository to find the user by email
	userRepo := repository.NewUserRepository(config.DB)
	userToDelete, err := userRepo.FindUserByEmail(reqBody.Email)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "user not found for the given email", err)
		return
	}

	// Use MemberRepository to find and delete the member by UserID
	memberRepo := repository.NewMemberRepository(config.DB)
	memberToDelete, msg, err := memberRepo.FetchMemberByUserID(userToDelete.ID) // Now this method exists
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, msg, err)
		return
	}

	deletedMember, msg, err := memberRepo.Delete(memberToDelete)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "member deleted successfully", "deleted_member", deletedMember) // Changed to http.StatusOK and provided deleted member
}
