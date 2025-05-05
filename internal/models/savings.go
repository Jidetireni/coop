package models

import "gorm.io/gorm"

type Savings struct {
	gorm.Model
	UserID       uint `gorm:"not null"`
	MemberID     uint `gorm:"not null"`
	Balance      int  `gorm:"not null"`
	AmountToSave int  `gorm:"not null"`
	Description  string
	Member       Member
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
