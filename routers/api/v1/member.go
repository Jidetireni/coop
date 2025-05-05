package v1

import (
	"cooperative-system/models"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MemberRequestBody struct {
	Name        string `json:"name" binding:"required"`
	ContactInfo string `json:"contact_info" binding:"required"`
}

type MemberHandler struct {
	DB *gorm.DB
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

func respondWithError(c *gin.Context, code int, message string, details error) {
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

func successResponse(c *gin.Context, code int, message string, responseName string, data interface{}) {
	c.JSON(code, gin.H{
		"message":    message,
		responseName: data,
	})
}

func getMemberByIDAndAuthorize(c *gin.Context, db *gorm.DB, authUser *models.User) (*models.Member, error) {
	// Get the member ID from the URL parameter
	memberID := c.Param("id")
	var member models.Member

	// Check if the member exists in the database
	err := db.Where("id = ?", memberID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Respond with Not Found if member does not exist
			respondWithError(c, http.StatusNotFound, "member not found", err)
			return nil, err
		} else {
			// Handle other database errors
			respondWithError(c, http.StatusInternalServerError, "failed to retrieve member", err)
			return nil, err
		}
	}

	if authUser.Role != "admin" && member.UserID != authUser.ID {
		// Check if the authenticated user is authorized to access this member
		respondWithError(c, http.StatusUnauthorized, "You are not authorized to view this member's savings", nil)
		return nil, fmt.Errorf("unauthorized access")
	}

	// Return the member if everything is fine
	return &member, nil
}

func (m *MemberHandler) CreateMember(c *gin.Context) {
	// Parse the request body into struct
	var reqBody MemberRequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		respondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Get authenticated user
	authUser, ok := getAuthUser(c)

	if !ok {
		respondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	// Assign authenticated user details to the member
	member := models.Member{
		UserID:      authUser.ID,
		Name:        reqBody.Name,
		ContactInfo: reqBody.ContactInfo,
	}

	// Create the member in the database
	if err := m.DB.Create(&member).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to create member", err)
		return
	}

	// Create an initial savings record
	savings := models.Savings{
		UserID:       authUser.ID,
		MemberID:     member.ID,
		Balance:      0,
		AmountToSave: 0,
		Description:  "Initial savings record",
	}

	if err := m.DB.Create(&savings).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to create initial savings", err)
		return
	}

	// Return the newly created member and savings
	successResponse(c, http.StatusCreated, "member created successfully", "data", gin.H{
		"member":  member,
		"savings": savings,
	})
}

func (m *MemberHandler) GetAllMembers(c *gin.Context) {
	// get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok || authUser.Role != "admin" {
		respondWithError(c, http.StatusForbidden, "you are not authorized to view all members", nil)
		return
	}

	// fetch all members with their linked users
	var members []models.Member
	if err := m.DB.Preload("User").Find(&members).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to retrieve members", err)
		return
	}

	// send response
	c.JSON(http.StatusOK, gin.H{
		"members": members,
	})
}

func (m *MemberHandler) GetMemberByID(c *gin.Context) {
	// Get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, m.DB, &authUser)
	if err != nil {
		return
	}

	// Return the member
	successResponse(c, http.StatusOK, "retrieved member by id successfully", "data", member)
}

func (m *MemberHandler) UpdateAMember(c *gin.Context) {
	// Parse request body
	var reqBody MemberRequestBody
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		respondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, m.DB, &authUser)
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
	if err := m.DB.Model(&member).Updates(updateFields).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to update member", err)
		return
	}

	// Fetch updated member
	if err := m.DB.Preload("User").First(&member, member.ID).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to fetch updated member", err)
		return
	}

	successResponse(c, http.StatusOK, "updated member successfully", "data", member)
}

func (m *MemberHandler) DeleteAMember(c *gin.Context) {
	// get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, m.DB, &authUser)
	if err != nil {
		return
	}

	// Delete the member
	if err := m.DB.Delete(&member).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to delete member", err)
		return
	}

	successResponse(c, http.StatusOK, "deleted member successfully", "data", member)
}
