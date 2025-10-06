package main

import (
	"RIP-WEB/internal/app/ds"
	"RIP-WEB/internal/app/dsn"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Подключаемся к БД
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

	// Читаем SQL файл из папки resources
	sql, err := os.ReadFile("resources/database.sql")
	if err != nil {
		log.Fatalf("Failed to read database.sql: %v", err)
	}

	// Выполняем SQL
	if err := db.Exec(string(sql)).Error; err != nil {
		log.Fatalf("Failed to execute seed SQL: %v", err)
	}

	fmt.Println("✅ Database seeded successfully!")
	fmt.Println("📊 Data overview:")

	// Проверяем данные
	var userCount int64
	db.Model(&ds.Users{}).Count(&userCount)
	fmt.Printf("   Users: %d\n", userCount)

	var anomalyCount int64
	db.Model(&ds.Anomaly{}).Where("is_delete = false").Count(&anomalyCount)
	fmt.Printf("   Anomalies: %d\n", anomalyCount)

	var treeCount int64
	db.Model(&ds.Tree{}).Where("status != 'удалён'").Count(&treeCount)
	fmt.Printf("   Trees: %d\n", treeCount)

	var treeItemCount int64
	db.Model(&ds.TreeItem{}).Count(&treeItemCount)
	fmt.Printf("   Tree Items: %d\n", treeItemCount)
}
