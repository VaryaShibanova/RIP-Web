package handler

import (
	"RIP-WEB/internal/app/repository"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(302, "/anomalies")
	})

	router.GET("/anomalies", h.GetAllAnomalies)  // GET - получение услуг (ORM)
	router.GET("/anomaly/:id", h.GetAnomalyById) // GET - просмотр услуги (ORM)
	router.GET("/tree/:id", h.GetTree)           // GET - просмотр заявки (ORM)
	router.POST("/tree/add", h.AddToTree)        // POST - добавление в заявку (ORM)
	router.POST("/tree/delete", h.DeleteTree)    // POST - удаление заявки (SQL UPDATE)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/resources", "./resources")

	router.GET("/favicon.ico", func(ctx *gin.Context) {
		ctx.Status(204) // No Content
	})
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

func (h *Handler) DeleteTree(ctx *gin.Context) {
	strId := ctx.PostForm("tree_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Неверный ID дерева"})
		return
	}

	err = h.Repository.DeleteTree(uint(id))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(302, "/anomalies")
}

func (h *Handler) AddToTree(ctx *gin.Context) {
	creatorID := uint(1)
	anomalyID, err := strconv.Atoi(ctx.PostForm("anomaly_id"))
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Неверный ID аномалии"})
		return
	}

	// Получаем данные аномалии из БД
	anomaly, err := h.Repository.GetAnomalyByID(anomalyID) // ORM вызов
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Ошибка получения данных аномалии"})
		return
	}
	if anomaly == nil {
		ctx.JSON(404, gin.H{"error": "Аномалия не найдена"})
		return
	}

	calculatedYear := 0
	// Пустая строка для колец - заполнится позже
	anomalousRings := ""

	// получение или создание черновика
	tree, err := h.Repository.GetOrCreateDraftTree(creatorID) // ORM вызов
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// добавление аномалии в заявку
	err = h.Repository.AddAnomalyToTree(tree.ID, uint(anomalyID), anomalousRings, calculatedYear) //ORM вызов
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(302, "/anomalies")
}
