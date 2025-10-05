package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ДОМЕН М-М (TREE ITEM)

// AddToTree - POST добавление в заявку-черновик
func (h *Handler) AddToTree(ctx *gin.Context) {
	user := h.Repository.GetSystemUser()

	var request struct {
		AnomalyID uint `json:"anomaly_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем существование аномалии
	anomaly, err := h.Repository.GetAnomalyByID(int(request.AnomalyID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	// Создаем или получаем черновую заявку
	tree, err := h.Repository.GetOrCreateDraftTree(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Добавляем аномалию в заявку
	err = h.Repository.AddAnomalyToTree(tree.ID, request.AnomalyID, "", 0)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Аномалия добавлена в заявку",
		"tree_id": tree.ID,
	})
}

// UpdateTreeItem - PUT изменение элемента заявки
func (h *Handler) UpdateTreeItem(ctx *gin.Context) {
	treeIDStr := ctx.Param("id")
	anomalyIDStr := ctx.Param("anomaly_id")

	fmt.Printf("Получены параметры: id=%s, anomaly_id=%s\n", treeIDStr, anomalyIDStr)

	treeID, err := strconv.Atoi(treeIDStr)
	if err != nil {
		fmt.Printf("Ошибка преобразования id: %s\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID заявки"})
		return
	}

	anomalyID, err := strconv.Atoi(anomalyIDStr)
	if err != nil {
		fmt.Printf("Ошибка преобразования anomaly_id: %s\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID аномалии"})
		return
	}

	fmt.Printf("Преобразовано: id=%d, anomaly_id=%d\n", treeID, anomalyID)

	// УПРОЩЕННАЯ СТРУКТУРА - ТОЛЬКО anomalous_rings
	var request struct {
		AnomalousRings string `json:"anomalous_rings"`
		// УДАЛЕНО: CalculatedYear int    `json:"calculated_year"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем заявку для расчета CalculatedYear
	tree, err := h.Repository.GetTreeByID(uint(treeID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Заявка не найдена"})
		return
	}

	// ВЫЧИСЛЯЕМ CalculatedYear автоматически по формуле
	calculatedYear := h.calculateYearForAnomaly(uint(anomalyID), tree.TotalRings, request.AnomalousRings)

	if err := h.Repository.UpdateTreeItem(uint(treeID), uint(anomalyID), request.AnomalousRings, calculatedYear); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Элемент заявки не найден"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":         "Элемент заявки обновлен",
		"anomalous_rings": request.AnomalousRings,
		"calculated_year": calculatedYear, // Возвращаем вычисленное значение
	})
}

// RemoveFromTree - DELETE удаление из заявки
func (h *Handler) RemoveFromTree(ctx *gin.Context) {
	treeID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID заявки"})
		return
	}

	anomalyID, err := strconv.Atoi(ctx.Param("anomaly_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID аномалии"})
		return
	}

	if err := h.Repository.RemoveFromTree(uint(treeID), uint(anomalyID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Элемент удален из заявки"})
}

// treeitem.go - добавим функцию расчета
func (h *Handler) calculateYearForAnomaly(anomalyID uint, totalRings int, anomalousRings string) int {
	// Получаем аномалию для получения Year
	anomaly, err := h.Repository.GetAnomalyByID(int(anomalyID))
	if err != nil || anomaly == nil {
		return 0
	}

	// Парсим anomalous_rings чтобы найти максимальное значение
	maxRing := parseMaxAnomalousRing(anomalousRings)

	// Формула: Year_аномалии + (TotalRings - MaxAnomalousRing)
	calculatedYear := anomaly.Year + (totalRings - maxRing)

	return calculatedYear
}

// Вспомогательная функция для парсинга максимального кольца
func parseMaxAnomalousRing(anomalousRings string) int {
	if anomalousRings == "" {
		return 0
	}

	// Парсим строку вида "45,67,89,112"
	rings := strings.Split(anomalousRings, ",")
	maxRing := 0

	for _, ringStr := range rings {
		ringStr = strings.TrimSpace(ringStr)
		if ring, err := strconv.Atoi(ringStr); err == nil {
			if ring > maxRing {
				maxRing = ring
			}
		}
	}

	return maxRing
}
