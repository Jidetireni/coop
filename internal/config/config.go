package config

import (
	"cooperative-system/internal/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDb() {

	var err error
	dsn := os.Getenv("DB")
	if dsn == "" {
		log.Fatal("environment variable DB not set")
	}

	// connect a postgtres db using the dns env
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	log.Println("database connected successfully")

}

func LoadEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func SyncDB() {
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Member{})
	DB.AutoMigrate(&models.Savings{})
	DB.AutoMigrate(&models.SavingTransaction{})
	DB.AutoMigrate(&models.Loan{})
}
