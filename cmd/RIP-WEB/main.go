package main

import (
	"fmt"

	"RIP-WEB/internal/app/config"
	"RIP-WEB/internal/app/dsn"
	"RIP-WEB/internal/app/handler"
	"RIP-WEB/internal/app/repository"
	"RIP-WEB/internal/pkg"

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

	// Инициализация менеджера сессий
	sessionManager := session.NewManager(
		conf.RedisHost,
		conf.RedisPort,
		conf.RedisPassword,
		conf.RedisDB,
		conf.JWTExpiration,
	)

	postgresString := dsn.FromEnv()
	fmt.Println("Database connection string:", postgresString)

	// Инициализация репозитория
	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	// Инициализация обработчика с конфигом и менеджером сессий
	hand := handler.NewHandler(rep, conf, sessionManager)

	// Создание и запуск приложения
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
