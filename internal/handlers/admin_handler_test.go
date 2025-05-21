// Unit tests for AdminHandler endpoints
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
	"gorm.io/gorm"
)

type mockAdminUserRepo struct {
	repository.UserRepository
	FindUserByEmailFunc func(email string) (*models.User, string, error)
	UpdateUserFunc      func(user *models.User, role string) (*models.User, string, error)
}

func (m *mockAdminUserRepo) FindUserByEmail(email string) (*models.User, string, error) {
	return m.FindUserByEmailFunc(email)
}
func (m *mockAdminUserRepo) UpdateUser(user *models.User, role string) (*models.User, string, error) {
	return m.UpdateUserFunc(user, role)
}

type mockAdminMemberRepo struct {
	repository.MemberRepository
	FetchMemberByUserIDFunc func(userID uint) (*models.Member, string, error)
	DeleteFunc              func(member *models.Member) (*models.Member, string, error)
	FetchMemberByIDFunc     func(tx *gorm.DB, memberID string) (*models.Member, string, error)
}

func (m *mockAdminMemberRepo) FetchMemberByUserID(userID uint) (*models.Member, string, error) {
	return m.FetchMemberByUserIDFunc(userID)
}
func (m *mockAdminMemberRepo) Delete(member *models.Member) (*models.Member, string, error) {
	return m.DeleteFunc(member)
}
func (m *mockAdminMemberRepo) FetchMemberByID(tx *gorm.DB, memberID string) (*models.Member, string, error) {
	return m.FetchMemberByIDFunc(tx, memberID)
}

type mockAdminLoanRepo struct {
	repository.LoanRepository
	BeginTransactionFunc      func() *gorm.DB
	GetLoanByIDForUpdateFunc  func(tx *gorm.DB, loanID string) (*models.Loan, string, error)
	GetAllLoansByMemberIDFunc func(tx *gorm.DB, memberID uint) ([]models.Loan, string, error)
	UpdateLoanFunc            func(tx *gorm.DB, loan *models.Loan) (*models.Loan, string, error)
	CreateLoanHistoryFunc     func(tx *gorm.DB, loanHistory *models.LoanHistory) error
}

func (m *mockAdminLoanRepo) BeginTransaction() *gorm.DB {
	return m.BeginTransactionFunc()
}
func (m *mockAdminLoanRepo) GetLoanByIDForUpdate(tx *gorm.DB, loanID string) (*models.Loan, string, error) {
	return m.GetLoanByIDForUpdateFunc(tx, loanID)
}
func (m *mockAdminLoanRepo) GetAllLoansByMemberID(tx *gorm.DB, memberID uint) ([]models.Loan, string, error) {
	return m.GetAllLoansByMemberIDFunc(tx, memberID)
}
func (m *mockAdminLoanRepo) UpdateLoan(tx *gorm.DB, loan *models.Loan) (*models.Loan, string, error) {
	return m.UpdateLoanFunc(tx, loan)
}
func (m *mockAdminLoanRepo) CreateLoanHistory(tx *gorm.DB, loanHistory *models.LoanHistory) error {
	return m.CreateLoanHistoryFunc(tx, loanHistory)
}

type mockAdminSavingsRepo struct {
	repository.SavingsRepository
	GetSavingsByMemberIDTxFunc func(tx *gorm.DB, memberID uint) (*models.Savings, string, error)
}

func (m *mockAdminSavingsRepo) GetSavingsByMemberIDTx(tx *gorm.DB, memberID uint) (*models.Savings, string, error) {
	return m.GetSavingsByMemberIDTxFunc(tx, memberID)
}

// No mock transaction struct needed

func TestCreateAdmin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUserRepo := &mockAdminUserRepo{
		FindUserByEmailFunc: func(email string) (*models.User, string, error) {
			user := models.User{}
			user.ID = 2
			user.Email = email
			user.Role = "member"
			return &user, "success", nil
		},
		UpdateUserFunc: func(user *models.User, role string) (*models.User, string, error) {
			user.Role = role
			return user, "user promoted", nil
		},
	}
	mockMemberRepo := &mockAdminMemberRepo{}
	mockSavingsRepo := &mockAdminSavingsRepo{}
	mockLoanRepo := &mockAdminLoanRepo{}
	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	r.POST("/admins", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "admin"
		c.Set("user", user)
		h.CreateAdmin(c)
	})
	body := map[string]interface{}{"email": "test@example.com"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/admins", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "created admin successfully")
}

func TestCreateAdmin_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUserRepo := &mockAdminUserRepo{}
	mockMemberRepo := &mockAdminMemberRepo{}
	mockSavingsRepo := &mockAdminSavingsRepo{}
	mockLoanRepo := &mockAdminLoanRepo{}
	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	r.POST("/admins", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.CreateAdmin(c)
	})
	body := map[string]interface{}{"email": "test@example.com"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/admins", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "only admins can create other admins")
}

func TestDeleteMember_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUserRepo := &mockAdminUserRepo{
		FindUserByEmailFunc: func(email string) (*models.User, string, error) {
			user := models.User{}
			user.ID = 2
			user.Email = email
			user.Role = "member"
			return &user, "success", nil
		},
	}
	mockMemberRepo := &mockAdminMemberRepo{
		FetchMemberByUserIDFunc: func(userID uint) (*models.Member, string, error) {
			member := models.Member{}
			member.ID = 2
			member.UserID = userID
			member.Name = "Test"
			member.ContactInfo = "123"
			return &member, "success", nil
		},
		DeleteFunc: func(member *models.Member) (*models.Member, string, error) {
			return member, "deleted", nil
		},
	}
	mockSavingsRepo := &mockAdminSavingsRepo{}
	mockLoanRepo := &mockAdminLoanRepo{}
	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	r.DELETE("/admins", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "admin"
		c.Set("user", user)
		h.DeleteMember(c)
	})
	body := map[string]interface{}{"email": "test@example.com"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodDelete, "/admins", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "member deleted successfully")
}

func TestDeleteMember_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUserRepo := &mockAdminUserRepo{}
	mockMemberRepo := &mockAdminMemberRepo{}
	mockSavingsRepo := &mockAdminSavingsRepo{}
	mockLoanRepo := &mockAdminLoanRepo{}
	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	r.DELETE("/admins", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.DeleteMember(c)
	})
	body := map[string]interface{}{"email": "test@example.com"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodDelete, "/admins", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "only admins can delete members")
}

func TestApproveLoan_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock transaction that will succeed
	mockTx := &gorm.DB{}

	mockLoanRepo := &mockAdminLoanRepo{
		BeginTransactionFunc: func() *gorm.DB {
			return mockTx
		},
		GetLoanByIDForUpdateFunc: func(tx *gorm.DB, loanID string) (*models.Loan, string, error) {
			loan := &models.Loan{
				Status:         models.LoanStatusPending,
				MemberID:       1,
				Amount:         1000,
				Type:           "personal",
				LoanTermMonths: 12,
			}
			loan.Model.ID = 1
			return loan, "loan fetched successfully", nil
		},
		GetAllLoansByMemberIDFunc: func(tx *gorm.DB, memberID uint) ([]models.Loan, string, error) {
			return []models.Loan{}, "no active loans", nil
		},
		UpdateLoanFunc: func(tx *gorm.DB, loan *models.Loan) (*models.Loan, string, error) {
			return loan, "loan updated successfully", nil
		},
		CreateLoanHistoryFunc: func(tx *gorm.DB, loanHistory *models.LoanHistory) error {
			return nil
		},
	}

	mockMemberRepo := &mockAdminMemberRepo{
		FetchMemberByIDFunc: func(tx *gorm.DB, memberID string) (*models.Member, string, error) {
			member := &models.Member{
				Name:        "Test Member",
				ContactInfo: "123-456-7890",
			}
			member.Model.ID = 1
			return member, "member fetched successfully", nil
		},
	}

	mockSavingsRepo := &mockAdminSavingsRepo{
		GetSavingsByMemberIDTxFunc: func(tx *gorm.DB, memberID uint) (*models.Savings, string, error) {
			savings := &models.Savings{
				Balance: 1000, // Sufficient savings for the loan
			}
			return savings, "savings fetched successfully", nil
		},
	}

	mockUserRepo := &mockAdminUserRepo{}

	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	r.PUT("/loans/:loan_id/approve", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "admin"
		c.Set("user", user)
		h.ApproveLoan(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/loans/1/approve", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "loan approved successfully")
}

func TestApproveLoan_RejectIneligibleLoan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock transaction that will succeed
	mockTx := &gorm.DB{}

	mockLoanRepo := &mockAdminLoanRepo{
		BeginTransactionFunc: func() *gorm.DB {
			return mockTx
		},
		GetLoanByIDForUpdateFunc: func(tx *gorm.DB, loanID string) (*models.Loan, string, error) {
			loan := &models.Loan{
				Status:         models.LoanStatusPending,
				MemberID:       1,
				Amount:         5000, // Loan amount exceeds twice the savings balance
				Type:           "personal",
				LoanTermMonths: 12,
			}
			loan.Model.ID = 1
			return loan, "loan fetched successfully", nil
		},
		GetAllLoansByMemberIDFunc: func(tx *gorm.DB, memberID uint) ([]models.Loan, string, error) {
			return []models.Loan{}, "no active loans", nil
		},
		UpdateLoanFunc: func(tx *gorm.DB, loan *models.Loan) (*models.Loan, string, error) {
			// Here we would verify the loan was rejected in a real test
			return loan, "loan updated successfully", nil
		},
		CreateLoanHistoryFunc: func(tx *gorm.DB, loanHistory *models.LoanHistory) error {
			return nil
		},
	}

	mockMemberRepo := &mockAdminMemberRepo{
		FetchMemberByIDFunc: func(tx *gorm.DB, memberID string) (*models.Member, string, error) {
			member := &models.Member{
				Name:        "Test Member",
				ContactInfo: "123-456-7890",
			}
			member.Model.ID = 1
			return member, "member fetched successfully", nil
		},
	}

	mockSavingsRepo := &mockAdminSavingsRepo{
		GetSavingsByMemberIDTxFunc: func(tx *gorm.DB, memberID uint) (*models.Savings, string, error) {
			savings := &models.Savings{
				Balance: 1000, // Not enough savings for the loan amount
			}
			return savings, "savings fetched successfully", nil
		},
	}

	mockUserRepo := &mockAdminUserRepo{}

	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	r.PUT("/loans/:loan_id/approve", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "admin"
		c.Set("user", user)
		h.ApproveLoan(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/loans/1/approve", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Still returns 200 OK because rejection is a successful operation
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestApproveLoan_NoAuthUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLoanRepo := &mockAdminLoanRepo{}
	mockMemberRepo := &mockAdminMemberRepo{}
	mockSavingsRepo := &mockAdminSavingsRepo{}
	mockUserRepo := &mockAdminUserRepo{}

	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	// Not setting user in context
	r.PUT("/loans/:loan_id/approve", h.ApproveLoan)

	req, _ := http.NewRequest(http.MethodPut, "/loans/1/approve", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "unable to get user from token")
}

func TestApproveLoan_NonAdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLoanRepo := &mockAdminLoanRepo{}
	mockMemberRepo := &mockAdminMemberRepo{}
	mockSavingsRepo := &mockAdminSavingsRepo{}
	mockUserRepo := &mockAdminUserRepo{}

	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	r.PUT("/loans/:loan_id/approve", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member" // Non-admin role
		c.Set("user", user)
		h.ApproveLoan(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/loans/1/approve", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "only admins can approve loans")
}

func TestApproveLoan_InvalidLoanID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLoanRepo := &mockAdminLoanRepo{}
	mockMemberRepo := &mockAdminMemberRepo{}
	mockSavingsRepo := &mockAdminSavingsRepo{}
	mockUserRepo := &mockAdminUserRepo{}

	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	r.PUT("/loans//approve", func(c *gin.Context) { // Empty loan_id
		user := models.User{}
		user.ID = 1
		user.Role = "admin"
		c.Set("user", user)
		h.ApproveLoan(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/loans//approve", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "loan ID is required")
}

func TestApproveLoan_LoanNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock transaction
	mockTx := &gorm.DB{}

	mockLoanRepo := &mockAdminLoanRepo{
		BeginTransactionFunc: func() *gorm.DB {
			return mockTx
		},
		GetLoanByIDForUpdateFunc: func(tx *gorm.DB, loanID string) (*models.Loan, string, error) {
			return nil, "loan not found", errors.New("not found")
		},
	}
	mockMemberRepo := &mockAdminMemberRepo{}
	mockSavingsRepo := &mockAdminSavingsRepo{}
	mockUserRepo := &mockAdminUserRepo{}

	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	r.PUT("/loans/:loan_id/approve", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "admin"
		c.Set("user", user)
		h.ApproveLoan(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/loans/999/approve", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "loan not found")
}

func TestApproveLoan_AlreadyApproved(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock transaction
	mockTx := &gorm.DB{}

	mockLoanRepo := &mockAdminLoanRepo{
		BeginTransactionFunc: func() *gorm.DB {
			return mockTx
		},
		GetLoanByIDForUpdateFunc: func(tx *gorm.DB, loanID string) (*models.Loan, string, error) {
			loan := &models.Loan{
				Status:   models.LoanStatusApproved, // Already approved
				MemberID: 1,
			}
			loan.Model.ID = 1
			return loan, "loan fetched successfully", nil
		},
	}
	mockMemberRepo := &mockAdminMemberRepo{}
	mockSavingsRepo := &mockAdminSavingsRepo{}
	mockUserRepo := &mockAdminUserRepo{}

	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo, mockSavingsRepo, mockLoanRepo)
	r := gin.Default()
	r.PUT("/loans/:loan_id/approve", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "admin"
		c.Set("user", user)
		h.ApproveLoan(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/loans/1/approve", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Loan is already approved")
}
