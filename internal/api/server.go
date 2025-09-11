package api

import (
	"RIP-WEB/internal/app/handler"
	"RIP-WEB/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting dendrochronology server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория дендрохронологических данных")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/resources", "./resources")

	// Основные маршруты
	r.GET("/", handler.GetAnomalies)
	r.GET("/anomaly/:id", handler.GetAnomaly)
	r.GET("/requests", handler.GetRequests)

	r.Run()
	log.Println("Dendrochronology server down")
}
