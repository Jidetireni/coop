package handlers

import (
	"cooperative-system/internal/models"
	"cooperative-system/internal/repository"
	"cooperative-system/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateSavingRequest struct {
	Amount      int    `json:"amount" binding:"required"`
	Description string `json:"description"`
}

type SavingsHandler struct {
	repo       repository.SavingsRepository
	MemberRepo repository.MemberRepository
}

func NewSavingsHandler(savingsRepo repository.SavingsRepository, memberRepo repository.MemberRepository) *SavingsHandler {
	return &SavingsHandler{
		repo:       savingsRepo,
		MemberRepo: memberRepo,
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

	var finalSavingsState *models.Savings = savings

	if !created {
		// Savings record exists, update its balance and other fields using repository
		updateData := map[string]interface{}{
			"Balance":      savings.Balance + reqBody.Amount, // Calculate the new total balance
			"AmountToSave": reqBody.Amount,                   // This seems to track the last amount saved
			"Description":  reqBody.Description,              // Update with the new description
		}

		tempUpdatedSavings, msg, err := s.repo.UpdateSavings(savings, updateData)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
			return
		}
		finalSavingsState = tempUpdatedSavings
	}

	// Create a new saving transaction history using repository
	transaction := &models.SavingTransaction{
		SavingsID:   finalSavingsState.ID,
		MemberID:    member.ID,
		Amount:      reqBody.Amount,
		Description: reqBody.Description,
	}

	createdTransaction, msg, err := s.repo.CreateTransaction(transaction)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		return
	}

	savingsResponse := models.NewSavingsResponse(finalSavingsState)
	transactionResponse := models.NewSavingTransactionResponse(createdTransaction)

	// Respond with the updated savings and new transaction
	utils.SuccessResponse(c, http.StatusCreated, "savings created successfully", "data",
		gin.H{
			"savings":     savingsResponse,
			"transaction": transactionResponse,
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
	savings, msg, err := s.repo.GetSavingsByMemberID(member.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		return
	}

	savingsResponse := models.NewSavingsResponse(savings)

	// Respond with the savings information
	utils.SuccessResponse(c, http.StatusOK, "retrieved savings successfully", "data", gin.H{
		"savings": savingsResponse,
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

	savings, msg, err := s.repo.GetSavingsByMemberID(member.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, msg, err)
		return
	}

	updateData := make(map[string]interface{})
	// Prepare fields to update
	if savingsReq.Amount != 0 {
		updateData["AmountToSave"] = savingsReq.Amount
	}
	if savingsReq.Description != "" {
		updateData["Description"] = savingsReq.Description
	}

	// Update the savings
	updatedSavings, msg, err := s.repo.UpdateSavings(savings, updateData)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		return
	}

	savingsResponse := models.NewSavingsResponse(updatedSavings)
	utils.SuccessResponse(c, http.StatusOK, "updated savings successfully", "data", gin.H{
		"savings": savingsResponse,
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

	savings, msg, err := s.repo.GetSavingsByMemberID(member.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, msg, err)
		return
	}

	// delete the savings record using repository
	deletedSavings, msg, err := s.repo.DeleteSavings(savings)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		return
	}

	savingsResponse := models.NewSavingsResponse(deletedSavings)
	utils.SuccessResponse(c, http.StatusOK, "deleted savings successfully", "data", gin.H{
		"savings": savingsResponse,
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

	transactions, msg, err := s.repo.GetTransactionsByMemberID(member.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		return
	}

	transactionResponses := make([]models.SavingTransactionResponse, len(transactions))
	for i, transaction := range transactions {
		currentTransaction := transaction
		transactionResponses[i] = models.NewSavingTransactionResponse(&currentTransaction)

	}

	utils.SuccessResponse(c, http.StatusOK, "retrieved transactions successfully", "data", gin.H{
		"transactions": transactionResponses,
	})

}
