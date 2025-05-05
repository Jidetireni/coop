package v1

import (
	"cooperative-system/models"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateSavingRequest struct {
	Amount      int    `json:"amount" binding:"required"`
	Description string `json:"description"`
}

type SavingsHandler struct {
	DB *gorm.DB
}

type SavingsService interface {
	CreateSavings(c *gin.Context)
	GetSavingByID(c *gin.Context)
	UpdateSavings(c *gin.Context)
	DeleteSavings(c *gin.Context)
	GetTransactionsForMember(c *gin.Context)
}

func (s *SavingsHandler) CreateSavings(c *gin.Context) {
	// Get authenticated user from context
	authUser, ok := getAuthUser(c)
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	// Find the member associated with the authenticated user
	var member models.Member
	if err := s.DB.Preload("User").Where("user_id = ?", authUser.ID).First(&member).Error; err != nil {
		respondWithError(c, http.StatusNotFound, "failed to find member", err)
		return
	}

	// Bind the incoming JSON request to CreateSavingRequest struct
	var reqBody CreateSavingRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		respondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Try to find an existing savings record for the member
	var savings models.Savings
	err := s.DB.Preload("Member").Where("member_id = ?", member.ID).First(&savings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No existing savings found, create a new savings record
			savings = models.Savings{
				UserID:       authUser.ID,
				MemberID:     member.ID,
				Balance:      reqBody.Amount,
				AmountToSave: reqBody.Amount,
				Description:  reqBody.Description,
			}
			if err := s.DB.Create(&savings).Error; err != nil {
				respondWithError(c, http.StatusInternalServerError, "failed to create savings record", err)
				return
			}
		} else {
			// Some other database error occurred
			respondWithError(c, http.StatusInternalServerError, "failed to find savings record", err)
			return
		}
	} else {
		// Savings record exists, update its balance and other fields
		savings.Balance += reqBody.Amount
		savings.AmountToSave = reqBody.Amount
		savings.Description = reqBody.Description

		if err := s.DB.Save(&savings).Error; err != nil {
			respondWithError(c, http.StatusInternalServerError, "failed to update savings record", err)
			return
		}
	}

	// Create a new saving transaction history
	transaction := models.SavingTransaction{
		SavingsID:   savings.ID,
		MemberID:    member.ID,
		Amount:      reqBody.Amount,
		Description: reqBody.Description,
	}

	if err := s.DB.Create(&transaction).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to create transaction record", err)
		return
	}

	// Respond with the updated savings and new transaction
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Savings updated successfully",
		"savings":     savings,
		"transaction": transaction,
	})
}

func (s *SavingsHandler) GetSavingByID(c *gin.Context) {
	// Get authenticated user from context
	authUser, ok := getAuthUser(c)
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	// Get the member ID from the URL parameter
	member, err := getMemberByIDAndAuthorize(c, s.DB, &authUser)
	if err != nil {
		return
	}

	// Find the savings associated with the member
	var savings models.Savings
	if err := s.DB.Preload("Member").Where("member_id = ?", member.ID).First(&savings).Error; err != nil {
		// If unable to find savings record, respond with an error
		respondWithError(c, http.StatusInternalServerError, "failed to get member savings", err)
		return
	}

	// Respond with the savings information
	c.JSON(http.StatusOK, gin.H{
		"savings": savings,
	})
}

func (s *SavingsHandler) UpdateSavings(c *gin.Context) {
	// get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, s.DB, &authUser)
	if err != nil {
		return
	}

	var savingsReq CreateSavingRequest
	if err := c.ShouldBindJSON(&savingsReq); err != nil {
		respondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	var savings models.Savings
	if err := s.DB.Where("member_id = ?", member.ID).First(&savings).Error; err != nil {
		respondWithError(c, http.StatusNotFound, "Savings record not found", err)
		return
	}

	// Prepare fields to update
	updateFields := make(map[string]interface{})
	if savingsReq.Amount != 0 {
		updateFields["amount"] = savingsReq.Amount

	}
	if savingsReq.Description != "" {
		updateFields["description"] = savingsReq.Description
	}

	// Update the savings
	if err := s.DB.Model(&savings).Updates(updateFields).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to update savings record", err)
		return
	}

	s.DB.Preload("User").Preload("Member").Where("member_id = ?", member.ID).First(&savings)
	c.JSON(http.StatusOK, gin.H{
		"savings": savings,
	})

}

func (s *SavingsHandler) DeleteSavings(c *gin.Context) {

	authUser, ok := getAuthUser(c)
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	//
	member, err := getMemberByIDAndAuthorize(c, s.DB, &authUser)
	if err != nil {
		return
	}

	// find the savings record associated with the user
	var savings models.Savings
	if err := s.DB.Where("member_id = ?", member.ID).First(&savings).Error; err != nil {
		respondWithError(c, http.StatusNotFound, "Savings record not found", err)
		return

	}

	// delete the savings record
	if err := s.DB.Delete(&savings).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to delete savings", err)
		return
	}

	// response
	c.JSON(http.StatusOK, gin.H{
		"message": "Savings record deleted successfully",
		"deleted": savings,
	})

}

func (s *SavingsHandler) GetTransactionsForMember(c *gin.Context) {

	// get authenticated user
	authUser, ok := getAuthUser(c)
	if !ok || authUser.Role != "admin" {
		respondWithError(c, http.StatusForbidden, "you are not authorized to view all members", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, s.DB, &authUser)
	if err != nil {
		return
	}

	var transactions []models.SavingTransaction
	if err := s.DB.Where("member_id = ?", member.ID).Find(&transactions).Error; err != nil {
		respondWithError(c, http.StatusInternalServerError, "failed to retrieve transactions", err)
		return
	}

	// send response
	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
	})
}
