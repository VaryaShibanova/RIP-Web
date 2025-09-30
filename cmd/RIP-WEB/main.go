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

	// Инициализация обработчика
	hand := handler.NewHandler(rep)

	// Создание и запуск приложения
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
