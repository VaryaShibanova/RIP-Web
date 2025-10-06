package handler

import (
	"RIP-WEB/internal/app/ds"
	"RIP-WEB/internal/app/minio"
	"context"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetAnomalies godoc
// @Summary Получение списка аномалий
// @Description Возвращает список аномалий с возможностью фильтрации по названию и году
// @Tags anomalies
// @Produce json
// @Param name query string false "Фильтр по названию"
// @Param year query string false "Фильтр по году"
// @Success 200 {object} AnomaliesListResponse
// @Router /api/anomalies [get]
func (h *Handler) GetAnomalies(ctx *gin.Context) {
	var anomalies []ds.Anomaly

	name := ctx.Query("name")
	year := ctx.Query("year")

	if name != "" || year != "" {
		query := ""
		if name != "" {
			query = name
		}
		anomalies, _ = h.Repository.SearchAnomalies(query)

		if year != "" {
			filtered := []ds.Anomaly{}
			yearInt, _ := strconv.Atoi(year)
			for _, anomaly := range anomalies {
				if anomaly.Year == yearInt {
					filtered = append(filtered, anomaly)
				}
			}
			anomalies = filtered
		}
	} else {
		anomalies, _ = h.Repository.GetAllAnomalies()
	}

	response := make([]gin.H, len(anomalies))
	for i, anomaly := range anomalies {
		response[i] = gin.H{
			"id":        anomaly.ID,
			"name":      anomaly.Name,
			"image_url": anomaly.Image,
			"year":      anomaly.Year,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"anomalies": response,
	})
}

// GetAnomaly godoc
// @Summary Получение информации об аномалии
// @Description Возвращает полную информацию об аномалии по ID
// @Tags anomalies
// @Produce json
// @Param id path int true "ID аномалии"
// @Success 200 {object} AnomalyDetailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/anomalies/{id} [get]
func (h *Handler) GetAnomaly(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	anomaly, err := h.Repository.GetAnomalyByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":          anomaly.ID,
		"name":        anomaly.Name,
		"description": anomaly.Description,
		"image_url":   anomaly.Image,
		"year":        anomaly.Year,
	})
}

// CreateAnomaly godoc
// @Summary Создание новой аномалии
// @Description Создает новую запись об аномалии (требуется аутентификация)
// @Tags anomalies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param anomaly body CreateAnomalyRequest true "Данные аномалии"
// @Success 201 {object} AnomalyDetailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/anomalies [post]
func (h *Handler) CreateAnomaly(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	var request struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description" binding:"required"`
		Year        int    `json:"year" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	anomaly := ds.Anomaly{
		Name:        request.Name,
		Description: request.Description,
		Year:        request.Year,
		IsDelete:    false,
	}

	if err := h.Repository.CreateAnomaly(&anomaly); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":          anomaly.ID,
		"name":        anomaly.Name,
		"description": anomaly.Description,
		"year":        anomaly.Year,
		"created_by":  userID,
	})
}

// UpdateAnomaly godoc
// @Summary Обновление информации об аномалии
// @Description Обновляет данные аномалии (требуется аутентификация)
// @Tags anomalies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID аномалии"
// @Param anomaly body UpdateAnomalyRequest true "Данные для обновления"
// @Success 200 {object} UpdateAnomalyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/anomalies/{id} [put]
func (h *Handler) UpdateAnomaly(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Year        int    `json:"year"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	anomaly, err := h.Repository.GetAnomalyByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	if request.Name != "" {
		anomaly.Name = request.Name
	}
	if request.Description != "" {
		anomaly.Description = request.Description
	}
	if request.Year != 0 {
		anomaly.Year = request.Year
	}

	if err := h.Repository.UpdateAnomaly(anomaly); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Информация об аномалии обновлена",
		"anomaly": gin.H{
			"id":          anomaly.ID,
			"name":        anomaly.Name,
			"description": anomaly.Description,
			"image_url":   anomaly.Image,
			"year":        anomaly.Year,
			"updated_by":  userID,
		},
	})
}

// DeleteAnomaly godoc
// @Summary Удаление аномалии
// @Description Удаляет аномалию и связанное с ней изображение (требуется аутентификация)
// @Tags anomalies
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID аномалии"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/anomalies/{id} [delete]
func (h *Handler) DeleteAnomaly(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	anomaly, err := h.Repository.GetAnomalyByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	if anomaly.Image != "" {
		objectName := minio.ExtractObjectNameFromURL(anomaly.Image)
		minioClient, _ := minio.InitMinio()
		if minioClient != nil {
			minio.DeleteObject(context.Background(), minioClient, "images", objectName)
		}
	}

	if err := h.Repository.GetDB().Delete(&ds.Anomaly{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Аномалия удалена",
		"deleted_by": userID,
	})
}

// UploadAnomalyImage godoc
// @Summary Загрузка изображения для аномалии
// @Description Загружает изображение для аномалии в Minio (требуется аутентификация)
// @Tags anomalies
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID аномалии"
// @Param image formData file true "Изображение"
// @Param filename formData string false "Название файла"
// @Success 200 {object} UploadImageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/anomalies/{id}/image [post]
func (h *Handler) UploadAnomalyImage(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	anomaly, err := h.Repository.GetAnomalyByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден"})
		return
	}

	customName := ctx.PostForm("filename")
	if customName == "" {
		customName = strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	}

	if anomaly.Image != "" {
		objectName := minio.ExtractObjectNameFromURL(anomaly.Image)
		minioClient, _ := minio.InitMinio()
		if minioClient != nil {
			minio.DeleteObject(context.Background(), minioClient, "images", objectName)
		}
	}

	minioClient, err := minio.InitMinio()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка Minio: " + err.Error()})
		return
	}

	objectName, err := minio.UploadImageWithName(context.Background(), minioClient, "images", file, uint(id), customName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки: " + err.Error()})
		return
	}

	imageURL := minio.GetImageURL(objectName)
	if err := h.Repository.UpdateAnomalyImage(uint(id), imageURL); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Изображение загружено",
		"image_url":   imageURL,
		"filename":    customName,
		"uploaded_by": userID,
	})
}
