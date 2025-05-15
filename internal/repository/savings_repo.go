package repository

import (
	"cooperative-system/internal/models"
	"errors" // Added for gorm.ErrRecordNotFound check

	"gorm.io/gorm"
)

type SavingsRepository struct {
	db *gorm.DB
}

func NewSavingsRepository(db *gorm.DB) *SavingsRepository {
	return &SavingsRepository{db: db}
}

// CreateSavingsEntry creates a new savings record in the database within a transaction
func (r *SavingsRepository) CreateSavingsEntry(savings *models.Savings) (*models.Savings, string, error) {
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

// FetchMemberByUserID fetches a member by their user ID.
// This function replaces the previous FetchMemberUserID, which was buggy.
// It's now a method of SavingsRepository, uses r.db, declares the member variable,
// takes userID uint as a parameter, and handles gorm.ErrRecordNotFound.
func (r *SavingsRepository) FetchMemberByUserID(userID uint) (*models.Member, string, error) {
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
func (r *SavingsRepository) GetOrCreateSavings(member *models.Member, userID uint, amount int, description string) (*models.Savings, bool, string, error) {
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
			return createdSavings, true, "savings created", nil
		}
		return nil, false, "failed to find savings record", err
	}
	return &savings, false, "savings found", nil
}

// UpdateSavings updates an existing savings record
func (r *SavingsRepository) UpdateSavings(savings *models.Savings) error {
	return r.db.Save(savings).Error
}

// CreateTransaction creates a new saving transaction record
func (r *SavingsRepository) CreateTransaction(transaction *models.SavingTransaction) error {
	return r.db.Create(transaction).Error
}

// GetSavingsByMemberID fetches a savings record by member ID
func (r *SavingsRepository) GetSavingsByMemberID(memberID uint) (*models.Savings, error) {
	var savings models.Savings
	err := r.db.Preload("Member").Where("member_id = ?", memberID).First(&savings).Error
	if err != nil {
		return nil, err
	}
	return &savings, nil
}

// DeleteSavings deletes a savings record
func (r *SavingsRepository) DeleteSavings(savings *models.Savings) error {
	return r.db.Delete(savings).Error
}

// GetTransactionsByMemberID fetches all transactions for a member
func (r *SavingsRepository) GetTransactionsByMemberID(memberID uint) ([]models.SavingTransaction, error) {
	var transactions []models.SavingTransaction
	err := r.db.Where("member_id = ?", memberID).Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
