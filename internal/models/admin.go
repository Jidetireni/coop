package models

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	UserID uint `gorm:"unique"`
	User   User
}
