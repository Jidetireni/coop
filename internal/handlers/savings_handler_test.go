// Unit tests for SavingsHandler endpoints
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

type mockSavingsRepo struct {
	repository.SavingsRepository
	FetchMemberByUserIDFunc       func(userID uint) (*models.Member, string, error)
	GetOrCreateSavingsFunc        func(member *models.Member, userID uint, amount int, description string) (*models.Savings, bool, string, error)
	UpdateSavingsFunc             func(savings *models.Savings, updateFields interface{}) (*models.Savings, string, error)
	CreateTransactionFunc         func(transaction *models.SavingTransaction) (*models.SavingTransaction, string, error)
	GetSavingsByMemberIDFunc      func(memberID uint) (*models.Savings, string, error)
	DeleteSavingsFunc             func(savings *models.Savings) (*models.Savings, string, error)
	GetTransactionsByMemberIDFunc func(memberID uint) ([]models.SavingTransaction, string, error)
}

func (m *mockSavingsRepo) FetchMemberByUserID(userID uint) (*models.Member, string, error) {
	return m.FetchMemberByUserIDFunc(userID)
}
func (m *mockSavingsRepo) GetOrCreateSavings(member *models.Member, userID uint, amount int, description string) (*models.Savings, bool, string, error) {
	return m.GetOrCreateSavingsFunc(member, userID, amount, description)
}
func (m *mockSavingsRepo) UpdateSavings(savings *models.Savings, updateFields interface{}) (*models.Savings, string, error) {
	return m.UpdateSavingsFunc(savings, updateFields)
}
func (m *mockSavingsRepo) CreateTransaction(transaction *models.SavingTransaction) (*models.SavingTransaction, string, error) {
	return m.CreateTransactionFunc(transaction)
}
func (m *mockSavingsRepo) GetSavingsByMemberID(memberID uint) (*models.Savings, string, error) {
	return m.GetSavingsByMemberIDFunc(memberID)
}
func (m *mockSavingsRepo) DeleteSavings(savings *models.Savings) (*models.Savings, string, error) {
	return m.DeleteSavingsFunc(savings)
}
func (m *mockSavingsRepo) GetTransactionsByMemberID(memberID uint) ([]models.SavingTransaction, string, error) {
	return m.GetTransactionsByMemberIDFunc(memberID)
}

type mockMemberRepoForSavings struct {
	repository.MemberRepository
	FetchByIDFunc func(memberID string) (*models.Member, string, error)
}

func (m *mockMemberRepoForSavings) FetchByID(memberID string) (*models.Member, string, error) {
	return m.FetchByIDFunc(memberID)
}

func TestCreateSavings_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSavings := &mockSavingsRepo{
		FetchMemberByUserIDFunc: func(userID uint) (*models.Member, string, error) {
			member := models.Member{}
			member.ID = 1
			member.UserID = userID
			member.Name = "Test"
			member.ContactInfo = "123"
			return &member, "success", nil
		},
		GetOrCreateSavingsFunc: func(member *models.Member, userID uint, amount int, description string) (*models.Savings, bool, string, error) {
			savings := models.Savings{}
			savings.ID = 1
			savings.UserID = userID
			savings.MemberID = member.ID
			savings.Balance = amount
			savings.AmountToSave = amount
			savings.Description = description
			return &savings, true, "savings created", nil
		},
		CreateTransactionFunc: func(transaction *models.SavingTransaction) (*models.SavingTransaction, string, error) {
			t := models.SavingTransaction{}
			t.ID = 1
			t.MemberID = transaction.MemberID
			t.Amount = transaction.Amount
			t.Description = transaction.Description
			t.SavingsID = transaction.SavingsID
			return &t, "transaction created", nil
		},
	}
	mockMember := &mockMemberRepoForSavings{}
	h := handlers.NewSavingsHandler(mockSavings, mockMember)
	r := gin.Default()
	r.POST("/savings", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.CreateSavings(c)
	})
	body := map[string]interface{}{"amount": 100, "description": "desc"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/savings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "savings created successfully")
}

func TestCreateSavings_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSavings := &mockSavingsRepo{
		FetchMemberByUserIDFunc: func(userID uint) (*models.Member, string, error) {
			member := models.Member{}
			member.ID = 1
			member.UserID = userID
			member.Name = "Test"
			member.ContactInfo = "123"
			return &member, "success", nil
		},
	}
	h := handlers.NewSavingsHandler(mockSavings, &mockMemberRepoForSavings{})
	r := gin.Default()
	r.POST("/savings", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.CreateSavings(c)
	})
	req, _ := http.NewRequest(http.MethodPost, "/savings", bytes.NewBuffer([]byte(`{"amount": "bad"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestCreateSavings_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSavings := &mockSavingsRepo{
		FetchMemberByUserIDFunc: func(userID uint) (*models.Member, string, error) {
			member := models.Member{}
			member.ID = 1
			member.UserID = userID
			member.Name = "Test"
			member.ContactInfo = "123"
			return &member, "success", nil
		},
	}
	h := handlers.NewSavingsHandler(mockSavings, &mockMemberRepoForSavings{})
	r := gin.Default()
	r.POST("/savings", h.CreateSavings)
	body := map[string]interface{}{"amount": 100, "description": "desc"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/savings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthenticated user")
}

func TestCreateSavings_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSavings := &mockSavingsRepo{
		FetchMemberByUserIDFunc: func(userID uint) (*models.Member, string, error) {
			member := models.Member{}
			member.ID = 1
			member.UserID = userID
			member.Name = "Test"
			member.ContactInfo = "123"
			return &member, "success", nil
		},
		GetOrCreateSavingsFunc: func(member *models.Member, userID uint, amount int, description string) (*models.Savings, bool, string, error) {
			return nil, true, "repo error", errors.New("db error")
		},
	}
	mockMember := &mockMemberRepoForSavings{}
	h := handlers.NewSavingsHandler(mockSavings, mockMember)
	r := gin.Default()
	r.POST("/savings", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.CreateSavings(c)
	})
	body := map[string]interface{}{"amount": 100, "description": "desc"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/savings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "repo error")
}
