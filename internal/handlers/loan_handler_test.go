// Unit tests for LoanHandler endpoints
package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"cooperative-system/internal/handlers"
	"cooperative-system/internal/models"
	"cooperative-system/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockLoanRepo struct {
	repository.LoanRepository
	CreateLoanRequestObjectFunc func(loan *models.Loan) (*models.Loan, string, error)
	GetLoanByIDFunc             func(loanID string) (*models.Loan, string, error)
}

func (m *mockLoanRepo) CreateLoanRequestObject(loan *models.Loan) (*models.Loan, string, error) {
	return m.CreateLoanRequestObjectFunc(loan)
}

func (m *mockLoanRepo) GetLoanByID(loanID string) (*models.Loan, string, error) {
	return m.GetLoanByIDFunc(loanID)
}

type mockMemberRepoForLoan struct {
	repository.MemberRepository
	FetchMemberByUserIDFunc func(userID uint) (*models.Member, string, error)
}

func (m *mockMemberRepoForLoan) FetchMemberByUserID(userID uint) (*models.Member, string, error) {
	return m.FetchMemberByUserIDFunc(userID)
}

func TestApplyLoan_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLoan := &mockLoanRepo{
		CreateLoanRequestObjectFunc: func(loan *models.Loan) (*models.Loan, string, error) {
			loan.ID = 1
			return loan, "Loan created successfully.", nil
		},
	}
	mockMember := &mockMemberRepoForLoan{
		FetchMemberByUserIDFunc: func(userID uint) (*models.Member, string, error) {
			member := models.Member{}
			member.ID = 1
			member.UserID = userID
			member.Name = "Test"
			member.ContactInfo = "123"
			return &member, "success", nil
		},
	}
	h := handlers.NewLoanHandler(mockLoan, mockMember)
	r := gin.Default()
	r.POST("/loans", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.ApplyLoan(c)
	})
	body := map[string]interface{}{"amount": 1000.00, "description": "desc", "type": "personal", "loan_term_months": 12}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/loans", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "loan application submitted successfully")
}

func TestApplyLoan_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewLoanHandler(&mockLoanRepo{}, &mockMemberRepoForLoan{})
	r := gin.Default()
	r.POST("/loans", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.ApplyLoan(c)
	})
	req, _ := http.NewRequest(http.MethodPost, "/loans", bytes.NewBuffer([]byte(`{"amount": "bad"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestApplyLoan_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewLoanHandler(&mockLoanRepo{}, &mockMemberRepoForLoan{})
	r := gin.Default()
	r.POST("/loans", h.ApplyLoan)
	body := map[string]interface{}{"amount": 1000, "description": "desc", "type": "personal", "loan_term_months": 12}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/loans", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthenticated user")
}

func TestApplyLoan_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLoan := &mockLoanRepo{
		CreateLoanRequestObjectFunc: func(loan *models.Loan) (*models.Loan, string, error) {
			return nil, "failed to apply for loan", errors.New("db error")
		},
	}
	mockMember := &mockMemberRepoForLoan{
		FetchMemberByUserIDFunc: func(userID uint) (*models.Member, string, error) {
			member := models.Member{}
			member.ID = 1
			member.UserID = userID
			member.Name = "Test"
			member.ContactInfo = "123"
			return &member, "success", nil
		},
	}
	h := handlers.NewLoanHandler(mockLoan, mockMember)
	r := gin.Default()
	r.POST("/loans", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.ApplyLoan(c)
	})
	body := map[string]interface{}{"amount": 1000, "description": "desc", "type": "personal", "loan_term_months": 12}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/loans", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to apply for loan")
}

func TestTrackLoanApproval_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLoan := &mockLoanRepo{
		GetLoanByIDFunc: func(loanID string) (*models.Loan, string, error) {
			loan := &models.Loan{}
			loan.Model.ID = 1
			loan.Status = "approved"
			loan.MemberID = 1 // Same as member.ID in mock below
			return loan, "loan fetched successfully", nil
		},
	}
	mockMember := &mockMemberRepoForLoan{
		FetchMemberByUserIDFunc: func(userID uint) (*models.Member, string, error) {
			member := models.Member{}
			member.Model.ID = 1 // Same as loan.MemberID
			member.UserID = userID
			member.Name = "Test"
			member.ContactInfo = "123"
			return &member, "success", nil
		},
	}
	h := handlers.NewLoanHandler(mockLoan, mockMember)
	r := gin.Default()
	r.GET("/loans/:loan_id", func(c *gin.Context) {
		user := models.User{}
		user.Model.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.TrackLoanApproval(c)
	})
	req, _ := http.NewRequest(http.MethodGet, "/loans/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "loan details fetched successfully")
}

func TestTrackLoanApproval_LoanNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLoan := &mockLoanRepo{
		GetLoanByIDFunc: func(loanID string) (*models.Loan, string, error) {
			return nil, "loan not found", errors.New("not found")
		},
	}
	mockMember := &mockMemberRepoForLoan{
		FetchMemberByUserIDFunc: func(userID uint) (*models.Member, string, error) {
			member := models.Member{}
			member.Model.ID = 1
			member.UserID = userID
			return &member, "success", nil
		},
	}
	h := handlers.NewLoanHandler(mockLoan, mockMember)
	r := gin.Default()
	r.GET("/loans/:loan_id", func(c *gin.Context) {
		user := models.User{}
		user.Model.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.TrackLoanApproval(c)
	})
	req, _ := http.NewRequest(http.MethodGet, "/loans/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "loan not found")
}

func TestTrackLoanApproval_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockLoan := &mockLoanRepo{
		GetLoanByIDFunc: func(loanID string) (*models.Loan, string, error) {
			loan := &models.Loan{}
			loan.Model.ID = 1
			loan.MemberID = 2 // Different from the member ID below
			loan.Status = "pending"
			return loan, "loan fetched successfully", nil
		},
	}
	mockMember := &mockMemberRepoForLoan{
		FetchMemberByUserIDFunc: func(userID uint) (*models.Member, string, error) {
			member := models.Member{}
			member.Model.ID = 1 // Different from loan.MemberID
			member.UserID = userID
			member.Name = "Test"
			member.ContactInfo = "123"
			return &member, "success", nil
		},
	}
	h := handlers.NewLoanHandler(mockLoan, mockMember)
	r := gin.Default()
	r.GET("/loans/:loan_id", func(c *gin.Context) {
		user := models.User{}
		user.Model.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.TrackLoanApproval(c)
	})
	req, _ := http.NewRequest(http.MethodGet, "/loans/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "you are not authorized to view this loan")
}

func TestTrackLoanApproval_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewLoanHandler(&mockLoanRepo{}, &mockMemberRepoForLoan{})
	r := gin.Default()
	r.GET("/loans/:loan_id", h.TrackLoanApproval) // No user set in context
	req, _ := http.NewRequest(http.MethodGet, "/loans/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthenticated user")
}
