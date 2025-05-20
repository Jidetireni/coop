package handlers

import (
	"cooperative-system/internal/models"
	"cooperative-system/internal/repository"
	"cooperative-system/pkg/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userRepo   repository.UserRepository
	memberRepo repository.MemberRepository
}

func NewAdminHandler(userRepo repository.UserRepository, memberRepo repository.MemberRepository) *AdminHandler {
	return &AdminHandler{
		userRepo:   userRepo,
		memberRepo: memberRepo,
	}
}

type AdminService interface {
	CreateAdmin(c *gin.Context)
	DeleteMember(c *gin.Context)
}

func getMemberByIDAndAuthorize(c *gin.Context, repo repository.MemberRepository, authUser *models.User) (*models.Member, error) {
	// Get the member ID from the URL parameter
	memberID := c.Param("id")

	// Check if the member exists in the database
	fetchedMember, message, err := repo.FetchByID(memberID)
	if err != nil {
		if fetchedMember == nil {
			utils.RespondWithError(c, http.StatusNotFound, message, err)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, message, err)
		}
		return nil, fmt.Errorf("failed to fetch member: %w", err)
	}

	if authUser.Role != "admin" && fetchedMember.UserID != authUser.ID {
		utils.RespondWithError(c, http.StatusUnauthorized, message, err)
		return nil, fmt.Errorf("unauthorized access")
	}

	// Return the member if everything is fine
	return fetchedMember, nil
}

func (h *AdminHandler) CreateAdmin(c *gin.Context) {
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

	userToPromote, msg, err := h.userRepo.FindUserByEmail(reqBody.Email)
	if err != nil {
		if userToPromote == nil {
			utils.RespondWithError(c, http.StatusNotFound, msg, err)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "error finding user: "+msg, err)
		}
		return
	}

	promotedUser, msg, err := h.userRepo.UpdateUser(userToPromote, "admin")
	if err != nil {
		if promotedUser == nil {
			utils.RespondWithError(c, http.StatusNotFound, msg, err)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "error promoting user: "+msg, err)
		}
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "created admin successfully", "admin", userToPromote)
}

func (h *AdminHandler) DeleteMember(c *gin.Context) {
	userCtx, exist := c.Get("user")
	if !exist {
		utils.RespondWithError(c, http.StatusInternalServerError, "unable to get user from token", nil)
		return
	}

	authUser, ok := userCtx.(models.User)
	if !ok || authUser.Role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "only admins can delete members", nil) // Clarified error message
		return
	}

	var reqBody struct {
		Email string `json:"email" binding:"required"` // Assuming you want to identify the member by the user's email
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	userToDelete, msg, err := h.userRepo.FindUserByEmail(reqBody.Email)
	if err != nil {
		if userToDelete == nil {
			utils.RespondWithError(c, http.StatusNotFound, msg, nil)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "error finding user: "+msg, err)
		}
		return
	}

	// Use MemberRepository to find and delete the member by UserID

	memberToDelete, msg, err := h.memberRepo.FetchMemberByUserID(userToDelete.ID) // Now this method exists
	if err != nil {
		if memberToDelete == nil { // Assuming FetchMemberByUserID returns (nil, "member not found", nil) for not found
			utils.RespondWithError(c, http.StatusNotFound, msg, nil)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "error fetching member details: "+msg, err)
		}
		return
	}

	deletedMember, msg, err := h.memberRepo.Delete(memberToDelete)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "member deleted successfully", "deleted_member", deletedMember) // Changed to http.StatusOK and provided deleted member
}
