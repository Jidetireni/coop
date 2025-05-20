// Unit tests for MemberHandler endpoints
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

type mockMemberRepo struct {
	repository.MemberRepository
	CreateMemberWithSavingsFunc func(member *models.Member, savings *models.Savings) (*models.Member, *models.Savings, string, error)
	FetchAllFunc                func(members []models.Member) ([]models.Member, string, error)
	FetchByIDFunc               func(memberID string) (*models.Member, string, error)
	UpdateFunc                  func(member *models.Member, updateFields interface{}) (*models.Member, string, error)
	DeleteFunc                  func(member *models.Member) (*models.Member, string, error)
	FetchMemberByUserIDFunc     func(userID uint) (*models.Member, string, error)
}

func (m *mockMemberRepo) CreateMemberWithSavings(member *models.Member, savings *models.Savings) (*models.Member, *models.Savings, string, error) {
	return m.CreateMemberWithSavingsFunc(member, savings)
}
func (m *mockMemberRepo) FetchAll(members []models.Member) ([]models.Member, string, error) {
	return m.FetchAllFunc(members)
}
func (m *mockMemberRepo) FetchByID(memberID string) (*models.Member, string, error) {
	return m.FetchByIDFunc(memberID)
}
func (m *mockMemberRepo) Update(member *models.Member, updateFields interface{}) (*models.Member, string, error) {
	return m.UpdateFunc(member, updateFields)
}
func (m *mockMemberRepo) Delete(member *models.Member) (*models.Member, string, error) {
	return m.DeleteFunc(member)
}
func (m *mockMemberRepo) FetchMemberByUserID(userID uint) (*models.Member, string, error) {
	return m.FetchMemberByUserIDFunc(userID)
}

func TestCreateMember_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewMemberHandler(&mockMemberRepo{})
	r := gin.Default()
	r.POST("/members", func(c *gin.Context) {
		user := models.User{}
		user.ID = 2
		user.Role = "member"
		c.Set("user", user)
		h.CreateMember(c)
	})

	// Missing name
	body := map[string]interface{}{"contact_info": "123456"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/members", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")

	// Missing contact_info
	body = map[string]interface{}{"name": "Jane Doe"}
	jsonBody, _ = json.Marshal(body)
	req, _ = http.NewRequest(http.MethodPost, "/members", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestCreateMember_UserTypeMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewMemberHandler(&mockMemberRepo{})
	r := gin.Default()
	r.POST("/members", func(c *gin.Context) {
		// Set user as wrong type (string instead of models.User)
		c.Set("user", "not-a-user")
		h.CreateMember(c)
	})
	body := map[string]interface{}{"name": "John Doe", "contact_info": "123456"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/members", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthenticated user")
}

func TestCreateMember_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewMemberHandler(&mockMemberRepo{})
	r := gin.Default()
	r.POST("/members", func(c *gin.Context) {
		user := models.User{}
		user.ID = 3
		user.Role = "member"
		c.Set("user", user)
		h.CreateMember(c)
	})
	req, _ := http.NewRequest(http.MethodPost, "/members", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestCreateMember_NonJSONBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewMemberHandler(&mockMemberRepo{})
	r := gin.Default()
	r.POST("/members", func(c *gin.Context) {
		user := models.User{}
		user.ID = 4
		user.Role = "member"
		c.Set("user", user)
		h.CreateMember(c)
	})
	req, _ := http.NewRequest(http.MethodPost, "/members", bytes.NewBuffer([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestCreateMember_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewMemberHandler(&mockMemberRepo{})
	r := gin.Default()
	r.POST("/members", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.CreateMember(c)
	})
	req, _ := http.NewRequest(http.MethodPost, "/members", bytes.NewBuffer([]byte(`{"name":123}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestCreateMember_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlers.NewMemberHandler(&mockMemberRepo{})
	r := gin.Default()
	r.POST("/members", h.CreateMember)
	body := map[string]interface{}{"name": "John Doe", "contact_info": "123456"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/members", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthenticated user")
}

func TestCreateMember_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &mockMemberRepo{
		CreateMemberWithSavingsFunc: func(member *models.Member, savings *models.Savings) (*models.Member, *models.Savings, string, error) {
			return nil, nil, "repo error", errors.New("db error")
		},
	}
	h := handlers.NewMemberHandler(mockRepo)
	r := gin.Default()
	r.POST("/members", func(c *gin.Context) {
		user := models.User{}
		user.ID = 1
		user.Role = "member"
		c.Set("user", user)
		h.CreateMember(c)
	})
	body := map[string]interface{}{"name": "John Doe", "contact_info": "123456"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/members", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "repo error")
}
