package handler

import (
	"errors"
	"net/http"
	"strconv"

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

// UpdateTreeItem - PUT изменение значений в м-м
func (h *Handler) UpdateTreeItem(ctx *gin.Context) {
	treeID, err := strconv.Atoi(ctx.Param("tree_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID заявки"})
		return
	}

	anomalyID, err := strconv.Atoi(ctx.Param("anomaly_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID аномалии"})
		return
	}

	var request struct {
		AnomalousRings string `json:"anomalous_rings"`
		CalculatedYear int    `json:"calculated_year"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repository.UpdateTreeItem(uint(treeID), uint(anomalyID), request.AnomalousRings, request.CalculatedYear); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Элемент заявки не найден"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Элемент заявки обновлен"})
}

// RemoveFromTree - DELETE удаление из заявки
func (h *Handler) RemoveFromTree(ctx *gin.Context) {
	treeID, err := strconv.Atoi(ctx.Param("tree_id"))
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
