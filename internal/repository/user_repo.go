package repository

import (
	"cooperative-system/internal/models"
	"errors"
	"log"

	"gorm.io/gorm"
)

type gormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *gormUserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) CreateUser(user *models.User) (*models.User, string, error) {

	if err := r.db.Create(&user).Error; err != nil {
		return nil, "failed to create user", err
	}
	return user, "user created successfully", nil

}

func (r *gormUserRepository) FindUserByEmail(email string) (*models.User, string, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "user not found", err
		}
		log.Printf("Error finding user by email %s: %v", email, err)
		return nil, "failed to find user", err
	}
	return &user, "success", nil
}

func (r *gormUserRepository) UpdateUser(user *models.User, role string) (*models.User, string, error) {

	if err := r.db.Model(&user).Update("role", role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "user not found", err
		}
		log.Printf("Error updating user %v: %v", user, err)
		return nil, "failed to update user", err

	}

	return user, "user updated successfully", nil
}
