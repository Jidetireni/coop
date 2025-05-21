package handlers

import (
	"cooperative-system/internal/models"
	"cooperative-system/internal/repository"
	"cooperative-system/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoanRequest struct {
	Amount      float64 `json:"amount" binding:"required"`
	Description string  `json:"description"`
	Type        string  `json:"type" binding:"required"`
	// InterestRate   float64 `json:"interest_rate" binding:"required"`
	LoanTermMonths uint `json:"loan_term_months" binding:"required"`
}

type LoanHandler struct {
	repo       repository.LoanRepository
	memberRepo repository.MemberRepository
}

func NewLoanHandler(loanRepo repository.LoanRepository, memberRepo repository.MemberRepository) *LoanHandler {
	return &LoanHandler{
		repo:       loanRepo,
		memberRepo: memberRepo,
	}
}

type LoanService interface {
	ApplyLoan(c *gin.Context)
	TrackLoanApproval(c *gin.Context)
}

func (l *LoanHandler) ApplyLoan(c *gin.Context) {
	// Bind the incoming JSON request to LoanRequest struct
	var reqBody LoanRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if reqBody.Amount <= 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "loan amount must be greater than zero", nil)
		return
	}

	if reqBody.LoanTermMonths <= 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "loan term must be greater than zero months", nil)
		return
	}

	if _, ok := models.AllowedLoanTypes[reqBody.Type]; !ok {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid loan type", nil)
		return
	}

	authUser, ok := getAuthUser(c)
	if !ok {
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	member, msg, err := l.memberRepo.FetchMemberByUserID(authUser.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		return
	}

	calculatedInterestRate := models.GetInterestRate(reqBody.Type, reqBody.LoanTermMonths)
	if calculatedInterestRate == 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid loan type or term", nil)
		return
	}

	totalRepayableAmount, err := models.CalculateTotalRepayableAmount(reqBody.Amount, calculatedInterestRate, reqBody.LoanTermMonths)
	if totalRepayableAmount < reqBody.Amount {
		utils.RespondWithError(c, http.StatusBadRequest, " ", err)
		return
	}

	installmentAmount, err := models.CalculateInstallmentAmount(totalRepayableAmount, reqBody.LoanTermMonths)
	if installmentAmount <= 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "installment amount must be greater than zero", err)
		return
	}

	loan := models.Loan{
		Amount:               reqBody.Amount,
		Description:          reqBody.Description,
		MemberID:             member.ID,
		InterestRate:         calculatedInterestRate,
		Status:               "pending",
		Type:                 reqBody.Type,
		LoanTermMonths:       reqBody.LoanTermMonths,
		TotalRepayableAmount: totalRepayableAmount,
		InstallmentAmount:    installmentAmount,
	}

	// Call the repository method to create a new loan
	createdLoan, msg, err := l.repo.CreateLoanRequestObject(&loan)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		return
	}

	loanResponse := models.NewLoanResponse(createdLoan)

	utils.SuccessResponse(c, http.StatusCreated, "loan application submitted successfully", "data", gin.H{
		"loan": loanResponse,
	})

}

func (l *LoanHandler) TrackLoanApproval(c *gin.Context) {
	loanID := c.Param("loan_id")
	if loanID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "loan ID is required", nil)
		return
	}

	authUser, ok := getAuthUser(c)
	if !ok {
		utils.RespondWithError(c, http.StatusUnauthorized, "unauthenticated user", nil)
		return
	}

	loan, msg, err := l.repo.GetLoanByID(loanID)
	if err != nil {
		if loan == nil {
			utils.RespondWithError(c, http.StatusNotFound, msg, err)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "error fetching loan: "+msg, err)
		}
		return
	}

	member, msg, err := l.memberRepo.FetchMemberByUserID(authUser.ID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, msg, err)
		return
	}

	if loan.MemberID != member.ID {
		utils.RespondWithError(c, http.StatusForbidden, "you are not authorized to view this loan", nil)
		return
	}

	loanResponse := models.NewLoanResponse(loan)
	utils.SuccessResponse(c, http.StatusOK, "loan details fetched successfully", "data", gin.H{
		"loan": loanResponse,
	})

}
