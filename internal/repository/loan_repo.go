package repository

import (
	"cooperative-system/internal/models"

	"gorm.io/gorm"
)

type gormLoanRepository struct {
	db *gorm.DB
}

func NewGormLoanRepository(db *gorm.DB) *gormLoanRepository {
	return &gormLoanRepository{db: db}
}

func (r *gormLoanRepository) CreateLoanRequestObject(loan *models.Loan) (*models.Loan, error) {
	if err := r.db.Create(loan).Error; err != nil {
		return nil, err
	}
	return loan, nil
}
