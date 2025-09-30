package handler

import (
	"RIP-WEB/internal/app/ds"
	"RIP-WEB/internal/app/minio"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ДОМЕН УСЛУГИ (ANOMALY)

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

	ctx.JSON(http.StatusOK, gin.H{
		"anomalies": anomalies,
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

	ctx.JSON(http.StatusOK, anomaly)
}

// CreateAnomaly - POST добавление (без изображения)
func (h *Handler) CreateAnomaly(ctx *gin.Context) {
	var anomaly ds.Anomaly
	if err := ctx.ShouldBindJSON(&anomaly); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repository.CreateAnomaly(&anomaly); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, anomaly)
}

// UpdateAnomaly - PUT изменение
func (h *Handler) UpdateAnomaly(ctx *gin.Context) {
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

	var updateData ds.Anomaly
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем только разрешенные поля
	anomaly.Name = updateData.Name
	anomaly.Description = updateData.Description
	anomaly.Year = updateData.Year

	if err := h.Repository.UpdateAnomaly(anomaly); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, anomaly)
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

	if err := h.Repository.DeleteAnomaly(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Аномалия удалена"})
}

// UploadAnomalyImage - POST добавление изображения
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

	objectName, err := minio.UploadImage(context.Background(), minioClient, "images", file, uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки: " + err.Error()})
		return
	}

	// Сохраняем URL изображения в БД
	imageURL := minio.GetImageURL(objectName)
	if err := h.Repository.UpdateAnomalyImage(uint(id), imageURL); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Изображение загружено",
		"image_url": imageURL,
	})
}
