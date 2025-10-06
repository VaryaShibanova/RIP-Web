package main

import (
	"fmt"

	"RIP-WEB/internal/app/config"
	"RIP-WEB/internal/app/dsn"
	"RIP-WEB/internal/app/handler"
	"RIP-WEB/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title RIP-WEB API
// @version 1.0
// @description API для системы исследования аномалий деревьев
// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @contact.name API Support
// @contact.url http://localhost:8080
// @contact.email support@rip-web.ru

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println("Database connection string:", postgresString)

	// Инициализация репозитория
	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// Инициализация обработчика с конфигом
	hand := handler.NewHandler(rep, conf)

	// Регистрация обработчиков
	hand.RegisterAPIHandlers(router)
	hand.RegisterStatic(router)

	// Запуск сервера
	serverAddr := fmt.Sprintf("%s:%d", conf.ServiceHost, conf.ServicePort)
	fmt.Printf("Server started on %s\n", serverAddr)
	fmt.Printf("Swagger docs available at http://%s/swagger/index.html\n", serverAddr)

	if err := router.Run(serverAddr); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
