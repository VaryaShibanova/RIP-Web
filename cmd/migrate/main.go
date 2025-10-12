package main

import (
	"RIP-WEB/internal/app/ds"
	"RIP-WEB/internal/app/dsn"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dsnString := dsn.FromEnv()
	if dsnString == "" {
		log.Fatal("Database connection string is empty")
	}

	db, err := gorm.Open(postgres.Open(dsnString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Автомиграция
	err = db.AutoMigrate(
		&ds.Anomaly{},
		&ds.Tree{},
		&ds.TreeItem{},
		&ds.Users{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("Database migrated successfully!")
}
