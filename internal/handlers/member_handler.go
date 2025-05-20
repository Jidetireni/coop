package handlers

import (
	"cooperative-system/internal/models"
	"cooperative-system/internal/repository"
	"cooperative-system/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MemberRequestBody struct {
	Name        string `json:"name" binding:"required"`
	ContactInfo string `json:"contact_info" binding:"required"`
}

type MemberHandler struct {
	// DB *gorm.DB
	MemberRepo repository.MemberRepository
}

func NewMemberHandler(memberRepo repository.MemberRepository) *MemberHandler {
	return &MemberHandler{
		MemberRepo: memberRepo,
	}
}

type MemberService interface {
	CreateMember(c *gin.Context)
	GetAllMembers(c *gin.Context)
	GetMemberByID(c *gin.Context)
	UpdateAMember(c *gin.Context)
	DeleteAMember(c *gin.Context)
}

// Helper function to extract authenticated user
func getAuthUser(c *gin.Context) (models.User, bool) {
	// get the user from the auth middleware/token
	user, exist := c.Get("user")
	if !exist {
		return models.User{}, false
	}

	// make sure that the authUser is of type model.User
	authUser, ok := user.(models.User)
	if !ok {
		return models.User{}, false
	}
	return authUser, true

}

func (m *MemberHandler) CreateMember(c *gin.Context) {
	// Parse the request body into struct
	var reqBody MemberRequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok {
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	// Assign authenticated user details to the member
	member := models.Member{
		UserID:      authUser.ID,
		Name:        reqBody.Name,
		ContactInfo: reqBody.ContactInfo,
	}

	// Create an initial savings record
	savings := models.Savings{
		UserID:       authUser.ID,
		MemberID:     member.ID,
		Balance:      0,
		AmountToSave: 0,
		Description:  "Initial savings record",
	}

	createdMember, createdSavings, message, err := m.MemberRepo.CreateMemberWithSavings(&member, &savings)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, message, err)
		return
	}

	memberReponse := models.NewMemberResponse(createdMember)
	savingsResponse := models.NewSavingsResponse(createdSavings)

	// Return the newly created member and savings
	utils.SuccessResponse(c, http.StatusCreated, "member created successfully", "data", gin.H{
		"member":  memberReponse,
		"savings": savingsResponse,
	})
}

func (m *MemberHandler) GetAllMembers(c *gin.Context) {
	// get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok || authUser.Role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "you are not authorized to view all members", nil)
		return
	}

	// fetch all members with their linked users
	var members []models.Member
	fetchedMembersList, message, err := m.MemberRepo.FetchAll(members)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, message, err)
		return
	}

	memberReponses := make([]models.MemberResponse, len(fetchedMembersList))
	for i, member := range fetchedMembersList {
		currentMember := member
		memberReponses[i] = models.NewMemberResponse(&currentMember)
	}

	// send response
	utils.SuccessResponse(c, http.StatusOK, message, "data", gin.H{
		"members": memberReponses,
	})
}

func (m *MemberHandler) GetMemberByID(c *gin.Context) {
	// Get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok {
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, m.MemberRepo, &authUser)
	if err != nil {
		return
	}

	memberResponse := models.NewMemberResponse(member)
	// Return the member
	utils.SuccessResponse(c, http.StatusOK, "retrieved member by id successfully", "data", gin.H{
		"member": memberResponse,
	})
}

func (m *MemberHandler) UpdateAMember(c *gin.Context) {
	// Parse request body
	var reqBody MemberRequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok {
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, m.MemberRepo, &authUser)
	if err != nil {
		return
	}

	// Prepare fields to update
	updateFields := make(map[string]interface{})
	if reqBody.Name != "" {
		updateFields["name"] = reqBody.Name
	}
	if reqBody.ContactInfo != "" {
		updateFields["contact_info"] = reqBody.ContactInfo
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	// Update member
	updateMember, message, err := m.MemberRepo.Update(member, updateFields)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, message, err)
		return
	}

	memberResponse := models.NewMemberResponse(updateMember)
	utils.SuccessResponse(c, http.StatusOK, "updated member successfully", "data", gin.H{
		"member": memberResponse,
	})
}

func (m *MemberHandler) DeleteAMember(c *gin.Context) {
	// get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok {
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, m.MemberRepo, &authUser)
	if err != nil {
		return
	}

	// Delete the member
	deleteMember, message, err := m.MemberRepo.Delete(member)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, message, err)
		return
	}

	memberResponse := models.NewMemberResponse(deleteMember)
	utils.SuccessResponse(c, http.StatusOK, "deleted member successfully", "data", memberResponse)
}
