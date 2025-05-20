// Unit tests for AdminHandler endpoints
package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cooperative-system/internal/handlers"
	"cooperative-system/internal/models"
	"cooperative-system/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
}

func (m *mockAdminMemberRepo) FetchMemberByUserID(userID uint) (*models.Member, string, error) {
	return m.FetchMemberByUserIDFunc(userID)
}
func (m *mockAdminMemberRepo) Delete(member *models.Member) (*models.Member, string, error) {
	return m.DeleteFunc(member)
}

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
	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo)
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
	h := handlers.NewAdminHandler(&mockAdminUserRepo{}, &mockAdminMemberRepo{})
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
	h := handlers.NewAdminHandler(mockUserRepo, mockMemberRepo)
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
	h := handlers.NewAdminHandler(&mockAdminUserRepo{}, &mockAdminMemberRepo{})
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
