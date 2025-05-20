package models

import (
	"time"

	"gorm.io/gorm"
)

type Savings struct {
	gorm.Model
	UserID       uint `gorm:"not null"`
	MemberID     uint `gorm:"not null"`
	Balance      int  `gorm:"not null"`
	AmountToSave int  `gorm:"not null"`
	Description  string
}

type SavingTransaction struct {
	gorm.Model
	SavingsID   uint `gorm:"not null"`
	MemberID    uint `gorm:"not null"`
	Amount      int  `gorm:"not null"`
	Description string
	// TransactionDate time.Time
	Savings Savings
}

type SavingsResponse struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Balance      int       `json:"balance"`
	AmountToSave int       `json:"amount_to_save"`
	Description  string    `json:"description"`
	MemberID     uint      `json:"member_id"`
}

func NewSavingsResponse(savings *Savings) SavingsResponse {
	return SavingsResponse{
		ID:           savings.ID,
		CreatedAt:    savings.CreatedAt,
		UpdatedAt:    savings.UpdatedAt,
		Balance:      savings.Balance,
		AmountToSave: savings.AmountToSave,
		Description:  savings.Description,
		MemberID:     savings.MemberID,
	}
}

type SavingTransactionResponse struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Amount      int       `json:"amount"`
	Description string    `json:"description"`
	MemberID    uint      `json:"member_id"`
	SavingsID   uint      `json:"savings_id"`
}

func NewSavingTransactionResponse(transaction *SavingTransaction) SavingTransactionResponse {
	return SavingTransactionResponse{
		ID:          transaction.ID,
		CreatedAt:   transaction.CreatedAt,
		UpdatedAt:   transaction.UpdatedAt,
		Amount:      transaction.Amount,
		Description: transaction.Description,
		MemberID:    transaction.MemberID,
		SavingsID:   transaction.SavingsID,
	}
}
