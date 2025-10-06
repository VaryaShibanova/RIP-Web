package main

import (
	"RIP-WEB/internal/app/ds"
	"RIP-WEB/internal/app/dsn"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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

	// Создаем тестовых пользователей с хэшированными паролями
	users := []ds.Users{
		{
			Login:       "research_user",
			Password:    hashPassword("password123"),
			IsModerator: false,
		},
		{
			Login:       "moderator_user",
			Password:    hashPassword("password123"),
			IsModerator: true,
		},
	}

	for _, user := range users {
		var existingUser ds.Users
		if err := db.Where("login = ?", user.Login).First(&existingUser).Error; err != nil {
			if err := db.Create(&user).Error; err != nil {
				log.Printf("Failed to create user %s: %v", user.Login, err)
			} else {
				log.Printf("Created user: %s", user.Login)
			}
		} else {
			// Обновляем пароль если пользователь уже существует
			existingUser.Password = user.Password
			existingUser.IsModerator = user.IsModerator
			if err := db.Save(&existingUser).Error; err != nil {
				log.Printf("Failed to update user %s: %v", user.Login, err)
			} else {
				log.Printf("Updated user: %s", user.Login)
			}
		}
	}

	fmt.Println("Database migrated and seeded successfully!")
}

func hashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	return string(hashed)
}
