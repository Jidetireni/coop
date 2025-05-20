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
	CreateLoanRequestObjectFunc func(loan *models.Loan) (*models.Loan, error)
}

func (m *mockLoanRepo) CreateLoanRequestObject(loan *models.Loan) (*models.Loan, error) {
	return m.CreateLoanRequestObjectFunc(loan)
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
		CreateLoanRequestObjectFunc: func(loan *models.Loan) (*models.Loan, error) {
			loan.ID = 1
			return loan, nil
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
		CreateLoanRequestObjectFunc: func(loan *models.Loan) (*models.Loan, error) {
			return nil, errors.New("db error")
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
