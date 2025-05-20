// Unit tests for UserHandler endpoints
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
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	repository.UserRepository
	CreateUserFunc      func(user *models.User) (*models.User, string, error)
	FindUserByEmailFunc func(email string) (*models.User, string, error)
	UpdateUserFunc      func(user *models.User, role string) (*models.User, string, error)
}

func (m *mockUserRepo) CreateUser(user *models.User) (*models.User, string, error) {
	return m.CreateUserFunc(user)
}
func (m *mockUserRepo) FindUserByEmail(email string) (*models.User, string, error) {
	return m.FindUserByEmailFunc(email)
}
func (m *mockUserRepo) UpdateUser(user *models.User, role string) (*models.User, string, error) {
	return m.UpdateUserFunc(user, role)
}

func TestSignup_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &mockUserRepo{
		CreateUserFunc: func(user *models.User) (*models.User, string, error) {
			user.ID = 1
			return user, "user created successfully", nil
		},
	}
	h := handlers.NewUserHandler(mockRepo)
	r := gin.Default()
	r.POST("/signup", h.Signup)
	body := map[string]interface{}{"email": "test@example.com", "password": "password"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "user created successfully")
}

func TestSignup_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewUserHandler(&mockUserRepo{})
	r := gin.Default()
	r.POST("/signup", h.Signup)
	req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer([]byte(`{"email":123}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestSignup_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &mockUserRepo{
		CreateUserFunc: func(user *models.User) (*models.User, string, error) {
			return nil, "repo error", errors.New("db error")
		},
	}
	h := handlers.NewUserHandler(mockRepo)
	r := gin.Default()
	r.POST("/signup", h.Signup)
	body := map[string]interface{}{"email": "test@example.com", "password": "password"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "repo error")
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	mockRepo := &mockUserRepo{
		FindUserByEmailFunc: func(email string) (*models.User, string, error) {
			user := models.User{}
			user.ID = 1
			user.Email = email
			// This is a valid bcrypt hash for "password"
			user.Password = string(hashedPassword)
			user.Role = "member"
			return &user, "success", nil
		},
	}
	h := handlers.NewUserHandler(mockRepo)
	r := gin.Default()
	r.POST("/login", h.Login)
	// Using a valid bcrypt hash that matches the password "password"
	body := map[string]interface{}{"email": "test@example.com", "password": "password"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "login successful")
}

func TestLogin_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewUserHandler(&mockUserRepo{})
	r := gin.Default()
	r.POST("/login", h.Login)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte(`{"email":123}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestLogin_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &mockUserRepo{
		FindUserByEmailFunc: func(email string) (*models.User, string, error) {
			return nil, "not found", errors.New("not found")
		},
	}
	h := handlers.NewUserHandler(mockRepo)
	r := gin.Default()
	r.POST("/login", h.Login)
	body := map[string]interface{}{"email": "test@example.com", "password": "password"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not found")
}
