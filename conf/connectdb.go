package config

import (
	"log"
	"os"

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
