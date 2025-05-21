package repository

import (
	"cooperative-system/internal/models"
	"errors" // Added for gorm.ErrRecordNotFound check

	"gorm.io/gorm"
)

type gormSavingsRepository struct {
	db *gorm.DB
}

func NewgormSavingsRepository(db *gorm.DB) *gormSavingsRepository {
	return &gormSavingsRepository{db: db}
}

// CreateSavingsEntry creates a new savings record in the database within a transaction
func (r *gormSavingsRepository) CreateSavingsEntry(savings *models.Savings) (*models.Savings, string, error) {
	// Start a transaction
	tx := r.db.Begin()
	defer func() {
		if rcv := recover(); rcv != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, "failed to start transaction", err
	}

	// Create the savings record
	if err := tx.Create(savings).Error; err != nil {
		tx.Rollback()
		return nil, "failed to create savings entry", err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, "failed to commit transaction", err
	}

	return savings, "savings entry created successfully", nil
}

// takes userID uint as a parameter, and handles gorm.ErrRecordNotFound.
func (r *gormSavingsRepository) FetchMemberByUserID(userID uint) (*models.Member, string, error) {
	var member models.Member
	// Assuming 'user_id' is the foreign key column in the members table referencing the users table.
	if err := r.db.Where("user_id = ?", userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "member not found for the given user ID", err
		}
		return nil, "failed to fetch member by user ID", err
	}
	return &member, "member fetched successfully", nil
}

// GetOrCreateSavings fetches an existing savings record or creates a new one for a member
func (r *gormSavingsRepository) GetOrCreateSavings(member *models.Member, userID uint, amount int, description string) (*models.Savings, bool, string, error) {
	var savings models.Savings
	err := r.db.Preload("Member").Where("member_id = ?", member.ID).First(&savings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No existing savings found, create a new savings record
			savings = models.Savings{
				UserID:       userID,
				MemberID:     member.ID,
				Balance:      amount,
				AmountToSave: amount,
				Description:  description,
			}
			createdSavings, msg, err := r.CreateSavingsEntry(&savings)
			if err != nil {
				return nil, true, msg, err
			}
			return createdSavings, true, "savings created", err
		}
		return nil, false, "failed to find savings record", err
	}
	return &savings, false, "savings found", err
}

// UpdateSavings updates an existing savings record
func (r *gormSavingsRepository) UpdateSavings(savings *models.Savings, updatedData interface{}) (*models.Savings, string, error) {
	if err := r.db.Model(&savings).Updates(updatedData).Error; err != nil {
		return nil, "failed to update savings", err
	}

	return savings, "savings updated successfully", nil

}

// CreateTransaction creates a new saving transaction record
func (r *gormSavingsRepository) CreateTransaction(transaction *models.SavingTransaction) (*models.SavingTransaction, string, error) {
	if err := r.db.Create(transaction).Error; err != nil {
		return nil, "failed to create transaction", err
	}
	return transaction, "transaction created successfully", nil
}

// GetSavingsByMemberID fetches a savings record by member ID
func (r *gormSavingsRepository) GetSavingsByMemberID(memberID uint) (*models.Savings, string, error) {
	var savings models.Savings
	err := r.db.Where("member_id = ?", memberID).First(&savings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "savings not found for the given member ID", err
		}
		return nil, "failed to fetch savings by member ID", err
	}
	return &savings, "savings fetched successfully", nil
}

// DeleteSavings deletes a savings record
func (r *gormSavingsRepository) DeleteSavings(savings *models.Savings) (*models.Savings, string, error) {
	if err := r.db.Delete(savings).Error; err != nil {
		return nil, "failed to delete savings", err
	}
	return savings, "savings deleted successfully", nil
}

// GetTransactionsByMemberID fetches all transactions for a member
func (r *gormSavingsRepository) GetTransactionsByMemberID(memberID uint) ([]models.SavingTransaction, string, error) {
	var transactions []models.SavingTransaction
	err := r.db.Where("member_id = ?", memberID).Find(&transactions).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "no transactions found for the given member ID", err
		}
		return nil, "failed to fetch transactions by member ID", err
	}
	return transactions, "transactions fetched successfully", nil

}

// GetSavingsByMemberIDTx fetches a savings record by member ID within a transaction
func (r *gormSavingsRepository) GetSavingsByMemberIDTx(tx *gorm.DB, memberID uint) (*models.Savings, string, error) {
	var savings models.Savings
	err := tx.Where("member_id = ?", memberID).First(&savings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "savings not found for the given member ID", err
		}
		return nil, "failed to fetch savings by member ID", err
	}
	return &savings, "savings fetched successfully", nil
}
