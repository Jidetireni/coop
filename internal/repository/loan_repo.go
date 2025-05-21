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

func (r *gormLoanRepository) CreateLoanRequestObject(loan *models.Loan) (*models.Loan, string, error) {
	if err := r.db.Create(loan).Error; err != nil {
		return nil, "failed to create loan", err
	}
	return loan, "loan created successfully", nil
}

func (r *gormLoanRepository) GetLoanByID(loanID string) (*models.Loan, string, error) {
	var loan models.Loan
	if err := r.db.Where("id = ?", loanID).First(&loan).Error; err != nil {
		return nil, "loan not found", err
	}
	return &loan, "success", nil
}
