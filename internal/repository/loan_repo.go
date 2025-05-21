package repository

import (
	"cooperative-system/internal/models"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormLoanRepository struct {
	db *gorm.DB
}

func NewGormLoanRepository(db *gorm.DB) *gormLoanRepository {
	return &gormLoanRepository{db: db}
}

func (r *gormLoanRepository) CreateLoanWithInitialHistory(loan *models.Loan, loanHistory *models.LoanHistory) (*models.Loan, *models.LoanHistory, string, error) {
	// Start a transaction
	tx := r.db.Begin()
	defer func() {
		if rcv := recover(); rcv != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return nil, nil, "failed to start transaction", err
	}

	// Create the loan record
	if err := tx.Create(loan).Error; err != nil {
		tx.Rollback()
		return nil, nil, "failed to create loan", err

	}

	if loan.ID == 0 {
		tx.Rollback()
		return nil, nil, "failed to get loan ID after creation", errors.New("loan ID is zero after creation")
	}
	loanHistory.LoanID = loan.ID

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, nil, "failed to commit transaction", err
	}
	// Create the loan history record
	if err := tx.Create(&loanHistory).Error; err != nil {
		tx.Rollback()
		return nil, nil, "failed to create loan history", err
	}
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, nil, "failed to commit transaction", err
	}

	return loan, loanHistory, "loan created successfully", nil
}

func (r *gormLoanRepository) GetLoanByID(loanID string) (*models.Loan, string, error) {
	var loan models.Loan
	if err := r.db.Where("id = ?", loanID).First(&loan).Error; err != nil {
		return nil, "loan not found", err
	}
	return &loan, "success", nil
}

func (r *gormLoanRepository) GetLoanHistoryByID(loanID string) ([]models.LoanHistory, string, error) {
	var loanHistories []models.LoanHistory

	if err := r.db.Where("loan_id = ?", loanID).Order("created_at ASC").Find(&loanHistories).Error; err != nil {
		return nil, "loan history not found", err
	}

	if len(loanHistories) == 0 {
		return nil, "loan history not found", errors.New("no loan history found")
	}

	return loanHistories, "success", nil

}

func (r *gormLoanRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

// func (r *gormLoanRepository) RollbackTransaction(tx *gorm.DB) {
// 	tx.Rollback()
// }

// func (r *gormLoanRepository) CommitTransaction(tx *gorm.DB) {
// 	tx.Commit()
// }

func (h *gormLoanRepository) GetLoanByIDForUpdate(tx *gorm.DB, loanID string) (*models.Loan, string, error) {
	var loan models.Loan
	// Add Clauses(clause.Locking{Strength: "UPDATE"}) for row-level lock
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", loanID).First(&loan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // More specific error check
			return nil, "loan not found", err
		}
		return nil, "failed to fetch loan for update", err
	}
	return &loan, "loan fetched successfully for update", nil
}

func (h *gormLoanRepository) GetAllLoansByMemberID(tx *gorm.DB, memberID uint) ([]models.Loan, string, error) {
	var loans []models.Loan
	if err := tx.Where("member_id = ?", memberID).Find(&loans).Error; err != nil {
		return nil, "failed to fetch loans", err
	}
	return loans, "loans fetched successfully", nil
}

func (h *gormLoanRepository) UpdateLoan(tx *gorm.DB, loan *models.Loan) (*models.Loan, string, error) {
	if err := tx.Save(loan).Error; err != nil {
		return nil, "failed to update loan", err
	}
	return loan, "loan updated successfully", nil
}

// CreateLoanHistory creates a new loan history record within a transaction
func (h *gormLoanRepository) CreateLoanHistory(tx *gorm.DB, loanHistory *models.LoanHistory) error {
	if err := tx.Create(loanHistory).Error; err != nil {
		return err
	}
	return nil
}
