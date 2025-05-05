package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null"`
}

type Member struct {
	gorm.Model
	UserID      uint   `gorm:"unique"`
	Name        string `gorm:"not null"`
	ContactInfo string `gorm:"not null"`
	User        User
	Savings     []Savings
}

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

type Admin struct {
	gorm.Model
	UserID uint
	User   User
}
