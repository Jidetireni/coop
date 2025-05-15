package handlers

import (
	"cooperative-system/internal/models"
	"cooperative-system/internal/repository"
	"cooperative-system/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateSavingRequest struct {
	Amount      int    `json:"amount" binding:"required"`
	Description string `json:"description"`
}

type SavingsHandler struct {
	repo       *repository.SavingsRepository
	DB         *gorm.DB // Add DB for legacy helpers like getMemberByIDAndAuthorize
	MemberRepo *repository.MemberRepository
}

func NewSavingsHandler(db *gorm.DB) *SavingsHandler {
	return &SavingsHandler{
		repo:       repository.NewSavingsRepository(db),
		DB:         db, // Set DB for legacy helpers
		MemberRepo: repository.NewMemberRepository(db),
	}
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
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	// Use repository to fetch member by user ID
	member, msg, err := s.repo.FetchMemberByUserID(authUser.ID)
	if err != nil {
		status := http.StatusNotFound
		if msg != "member not found for the given user ID" {
			status = http.StatusInternalServerError
		}
		utils.RespondWithError(c, status, msg, err)
		return
	}

	// Bind the incoming JSON request to CreateSavingRequest struct
	var reqBody CreateSavingRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Use repository to get or create savings
	savings, created, msg, err := s.repo.GetOrCreateSavings(member, authUser.ID, reqBody.Amount, reqBody.Description)
	if err != nil {
		status := http.StatusInternalServerError
		if created {
			status = http.StatusInternalServerError
		}
		utils.RespondWithError(c, status, msg, err)
		return
	}

	if !created {
		// Savings record exists, update its balance and other fields using repository
		savings.Balance += reqBody.Amount
		savings.AmountToSave = reqBody.Amount
		savings.Description = reqBody.Description
		if err := s.repo.UpdateSavings(savings); err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to update savings record", err)
			return
		}
	}

	// Create a new saving transaction history using repository
	transaction := &models.SavingTransaction{
		SavingsID:   savings.ID,
		MemberID:    member.ID,
		Amount:      reqBody.Amount,
		Description: reqBody.Description,
	}
	if err := s.repo.CreateTransaction(transaction); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to create transaction record", err)
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
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	// Get the member ID from the URL parameter using MemberRepo
	member, err := getMemberByIDAndAuthorize(c, s.MemberRepo, &authUser)
	if err != nil {
		return
	}

	// Use repository to get savings by member ID
	savings, err := s.repo.GetSavingsByMemberID(member.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to get member savings", err)
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
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, s.MemberRepo, &authUser)
	if err != nil {
		return
	}

	var savingsReq CreateSavingRequest
	if err := c.ShouldBindJSON(&savingsReq); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	savings, err := s.repo.GetSavingsByMemberID(member.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Savings record not found", err)
		return
	}

	// Prepare fields to update
	if savingsReq.Amount != 0 {
		savings.AmountToSave = savingsReq.Amount
	}
	if savingsReq.Description != "" {
		savings.Description = savingsReq.Description
	}

	// Update the savings
	if err := s.repo.UpdateSavings(savings); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update savings record", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"savings": savings,
	})
}

func (s *SavingsHandler) DeleteSavings(c *gin.Context) {
	authUser, ok := getAuthUser(c)
	if !ok {
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, s.MemberRepo, &authUser)
	if err != nil {
		return
	}

	savings, err := s.repo.GetSavingsByMemberID(member.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Savings record not found", err)
		return
	}

	// delete the savings record using repository
	if err := s.repo.DeleteSavings(savings); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to delete savings record", err)
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
		utils.RespondWithError(c, http.StatusForbidden, "you are not authorized to view transactions", nil)
		return
	}

	member, err := getMemberByIDAndAuthorize(c, s.MemberRepo, &authUser)
	if err != nil {
		return
	}

	transactions, err := s.repo.GetTransactionsByMemberID(member.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to fetch transactions", err)
		return
	}

	// send response
	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
	})
}
