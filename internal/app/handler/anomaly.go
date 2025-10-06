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

// GetAnomalies - GET список с фильтрацией (JSON API)
func (h *Handler) GetAnomalies(ctx *gin.Context) {
	var anomalies []ds.Anomaly

	// Фильтрация по названию
	name := ctx.Query("name")
	year := ctx.Query("year")

	if name != "" || year != "" {
		// Поиск с фильтрацией
		query := ""
		if name != "" {
			query = name
		}
		anomalies, _ = h.Repository.SearchAnomalies(query)

		// Дополнительная фильтрация по году если нужно
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
		// Все аномалии
		anomalies, _ = h.Repository.GetAllAnomalies()
	}

	// Формируем выходные данные БЕЗ description
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

// GetAnomaly - GET одна запись (JSON API)
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

	// Возвращаем ВСЕ поля для одной аномалии
	ctx.JSON(http.StatusOK, gin.H{
		"id":          anomaly.ID,
		"name":        anomaly.Name,
		"description": anomaly.Description,
		"image_url":   anomaly.Image,
		"year":        anomaly.Year,
	})
}

// CreateAnomaly - POST добавление (без изображения)
func (h *Handler) CreateAnomaly(ctx *gin.Context) {
	// ВХОДНЫЕ ДАННЫЕ - только нужные поля
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

	// ВЫХОДНЫЕ ДАННЫЕ
	ctx.JSON(http.StatusCreated, gin.H{
		"id":          anomaly.ID,
		"name":        anomaly.Name,
		"description": anomaly.Description,
		"year":        anomaly.Year,
	})
}

// UpdateAnomaly - PUT изменение информации об аномалии
func (h *Handler) UpdateAnomaly(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	// ВХОДНЫЕ ДАННЫЕ для обновления
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

	// Обновляем только переданные поля
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

	// ВЫХОДНЫЕ ДАННЫЕ
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Информация об аномалии обновлена",
		"anomaly": gin.H{
			"id":          anomaly.ID,
			"name":        anomaly.Name,
			"description": anomaly.Description,
			"image_url":   anomaly.Image,
			"year":        anomaly.Year,
		},
	})
}

// DeleteAnomaly - DELETE удаление
func (h *Handler) DeleteAnomaly(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	// Получаем аномалию для удаления изображения
	anomaly, err := h.Repository.GetAnomalyByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	// Удаляем изображение из Minio если есть
	if anomaly.Image != "" {
		objectName := minio.ExtractObjectNameFromURL(anomaly.Image)
		minioClient, _ := minio.InitMinio()
		if minioClient != nil {
			minio.DeleteObject(context.Background(), minioClient, "images", objectName)
		}
	}

	// ВАЖНО: Используем HARD DELETE вместо soft delete
	if err := h.Repository.GetDB().Delete(&ds.Anomaly{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Аномалия удалена"})
}

func (h *Handler) UploadAnomalyImage(ctx *gin.Context) {
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

	// ПОЛУЧАЕМ название файла из формы
	customName := ctx.PostForm("filename")
	if customName == "" {
		// Если название не указано, используем оригинальное имя файла
		customName = strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	}

	// Удаляем старое изображение если есть
	if anomaly.Image != "" {
		objectName := minio.ExtractObjectNameFromURL(anomaly.Image)
		minioClient, _ := minio.InitMinio()
		if minioClient != nil {
			minio.DeleteObject(context.Background(), minioClient, "images", objectName)
		}
	}

	// Загружаем новое изображение в Minio
	minioClient, err := minio.InitMinio()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка Minio: " + err.Error()})
		return
	}

	// ПЕРЕДАЕМ кастомное название
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
		"message":   "Изображение загружено",
		"image_url": imageURL,
		"filename":  customName,
	})
}
