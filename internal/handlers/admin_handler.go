package handlers

import (
	"cooperative-system/internal/models"
	"cooperative-system/internal/repository"
	"cooperative-system/pkg/utils"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userRepo    repository.UserRepository
	memberRepo  repository.MemberRepository
	savingsRepo repository.SavingsRepository
	loanRepo    repository.LoanRepository
}

func NewAdminHandler(userRepo repository.UserRepository, memberRepo repository.MemberRepository, savingRepo repository.SavingsRepository, loanRepo repository.LoanRepository) *AdminHandler {
	return &AdminHandler{
		userRepo:    userRepo,
		memberRepo:  memberRepo,
		savingsRepo: savingRepo,
		loanRepo:    loanRepo,
	}
}

type AdminService interface {
	CreateAdmin(c *gin.Context)
	DeleteMember(c *gin.Context)
	ApproveLoan(c *gin.Context)
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

func (h *AdminHandler) ApproveLoan(c *gin.Context) {

	userCtx, exist := c.Get("user")
	if !exist {
		utils.RespondWithError(c, http.StatusInternalServerError, "unable to get user from token", nil)
		return
	}

	authUser, ok := userCtx.(models.User)
	if !ok || authUser.Role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "only admins can approve loans", nil) // Clarified error message
		return
	}

	// Get the loan ID from the URL parameter
	loanID := c.Param("loan_id")
	if loanID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "loan ID is required", nil)
		return
	}

	// --- START TRANSACTIONAL LOGIC ---

	// 1. Start a database transaction
	tx := h.loanRepo.BeginTransaction()
	var errLoop error
	defer func() {
		if rcv := recover(); rcv != nil {
			// h.loanRepo.RollbackTransaction(tx)
			tx.Rollback()
			panic(rcv)
		} else if errLoop != nil {
			tx.Rollback()
		}
	}()

	// fetch the loan within the transaction
	fetchedLoan, message, err := h.loanRepo.GetLoanByIDForUpdate(tx, loanID)
	if err != nil {
		if fetchedLoan == nil {
			utils.RespondWithError(c, http.StatusNotFound, message, err)
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, message, err)
		}
		return
	}

	canProcess, statusMsg, statusErr := models.CheckLoanStatus(fetchedLoan)
	if statusErr != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, statusMsg, statusErr)
		return
	}

	if !canProcess {
		utils.RespondWithError(c, http.StatusBadRequest, statusMsg, nil)
		return
	}

	member, msg, errLoop := h.memberRepo.FetchMemberByID(tx, fmt.Sprint(fetchedLoan.MemberID))
	if errLoop != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to fetch member for eligibility: "+msg, errLoop)
		return
	}

	if member == nil {
		errLoop = errors.New("member not found for eligibility check")
		utils.RespondWithError(c, http.StatusNotFound, errLoop.Error(), nil)
		return
	}

	savings, msg, errLoop := h.savingsRepo.GetSavingsByMemberIDTx(tx, member.ID)
	if errLoop != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to fetch savings for eligibility: "+msg, errLoop)
		return
	}
	if savings == nil {
		errLoop = errors.New("member savings record not found for eligibility check")
		utils.RespondWithError(c, http.StatusNotFound, errLoop.Error(), nil)
		return
	}

	existingLoans, fetchLoansMsg, fetchLoansErr := h.loanRepo.GetAllLoansByMemberID(tx, member.ID) // Renamed msg and err variables
	if fetchLoansErr != nil {
		errLoop = fetchLoansErr // Assign to errLoop for defer rollback
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to fetch existing loans for eligibility: "+fetchLoansMsg, fetchLoansErr)
		return
	}

	isEligible, eligibilityReasons, err := models.CheckLoanEligibility(fetchedLoan, member, savings, existingLoans)
	if err != nil {
		errLoop = err
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to check loan eligibility: "+err.Error(), err)
		return
	}

	now := time.Now()
	if !isEligible {
		rejectionReasonStr := strings.Join(eligibilityReasons, "; ")
		fetchedLoan.Status = models.LoanStatusRejected
		fetchedLoan.RejectionReason = rejectionReasonStr
		fetchedLoan.ReviewedAt = &now
		fetchedLoan.RejectedAt = &now

		updatedLoan, updateMsg, updateErr := h.loanRepo.UpdateLoan(tx, fetchedLoan)
		if updateErr != nil {
			errLoop = updateErr
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to update loan to rejected: "+updateMsg, updateErr)
			return
		}

		rejectionHistory := models.LoanHistory{
			LoanID:    updatedLoan.ID,
			Status:    models.LoanStatusRejected,
			ChangedBy: authUser.ID, // Assuming authUser is the admin user
			Remarks:   "Loan rejected: " + rejectionReasonStr,
		}

		if histErr := h.loanRepo.CreateLoanHistory(tx, &rejectionHistory); histErr != nil {
			errLoop = histErr
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to create rejection history: "+histErr.Error(), histErr)
			return
		}

		if commitErr := tx.Commit().Error; commitErr != nil {
			errLoop = commitErr
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to commit rejection transaction: "+commitErr.Error(), commitErr)
			return
		}

	} else {

		fetchedLoan.Status = models.LoanStatusApproved
		fetchedLoan.ApprovedAt = &now
		fetchedLoan.ReviewedAt = &now
		fetchedLoan.RejectionReason = ""
		fetchedLoan.IsActive = true

		updatedLoan, updateMsg, updateErr := h.loanRepo.UpdateLoan(tx, fetchedLoan)
		if updateErr != nil {
			errLoop = updateErr
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to update loan to approved: "+updateMsg, updateErr)
			return
		}

		approvalHistory := models.LoanHistory{
			LoanID:    updatedLoan.ID,
			Status:    models.LoanStatusApproved,
			ChangedBy: authUser.ID,
			Remarks:   "Loan approved by admin.",
		}
		if histErr := h.loanRepo.CreateLoanHistory(tx, &approvalHistory); histErr != nil {
			errLoop = histErr
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to create approval history: "+histErr.Error(), histErr)
			return
		}

		if commitErr := tx.Commit().Error; commitErr != nil {
			errLoop = commitErr
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to commit approval transaction: "+commitErr.Error(), commitErr)
			return
		}
		utils.SuccessResponse(c, http.StatusOK, "loan approved successfully", "loan", updatedLoan)

	}

	// fetch the loan by id *
	// check if the loan is already approved *
	// check eligibility status, performing eligibility checks *
	// update it's status within a transaction *
	//
}
