package models

import "gorm.io/gorm"

type Member struct {
	gorm.Model
	UserID      uint   `gorm:"unique"`
	Name        string `gorm:"not null"`
	ContactInfo string `gorm:"not null"`
	User        User
	Savings     []Savings
}
