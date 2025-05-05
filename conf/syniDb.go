package config

import "cooperative-system/models"

func SyncDB() {
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Member{})
	DB.AutoMigrate(&models.Savings{})
	DB.AutoMigrate(&models.SavingTransaction{})
}
