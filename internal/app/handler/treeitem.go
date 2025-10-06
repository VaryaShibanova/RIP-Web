package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AddToTree godoc
// @Summary Добавление аномалии в заявку
// @Description Добавляет аномалию в черновую заявку пользователя
// @Tags tree-items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item body AddToTreeRequest true "Данные для добавления"
// @Success 200 {object} AddToTreeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/trees/current/items [post]
func (h *Handler) AddToTree(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	var request struct {
		AnomalyID uint `json:"anomaly_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	anomaly, err := h.Repository.GetAnomalyByID(int(request.AnomalyID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	tree, err := h.Repository.GetOrCreateDraftTree(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

// UpdateTreeItem godoc
// @Summary Обновление элемента заявки
// @Description Обновляет данные элемента заявки (аномальные кольца)
// @Tags tree-items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param anomaly_id path int true "ID аномалии"
// @Param item body UpdateTreeItemRequest true "Данные для обновления"
// @Success 200 {object} UpdateTreeItemResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/trees/{id}/items/{anomaly_id} [put]
func (h *Handler) UpdateTreeItem(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	treeIDStr := ctx.Param("id")
	anomalyIDStr := ctx.Param("anomaly_id")

	treeID, err := strconv.Atoi(treeIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID заявки"})
		return
	}

	anomalyID, err := strconv.Atoi(anomalyIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID аномалии"})
		return
	}

	tree, err := h.Repository.GetTreeByID(uint(treeID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if tree.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к этой заявке"})
		return
	}

	var request struct {
		AnomalousRings string `json:"anomalous_rings"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repository.UpdateTreeItem(uint(treeID), uint(anomalyID), request.AnomalousRings, 0); err != nil {
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
		"calculated_year": 0,
	})
}

// RemoveFromTree godoc
// @Summary Удаление элемента из заявки
// @Description Удаляет аномалию из заявки
// @Tags tree-items
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param anomaly_id path int true "ID аномалии"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/trees/{id}/items/{anomaly_id} [delete]
func (h *Handler) RemoveFromTree(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

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

	tree, err := h.Repository.GetTreeByID(uint(treeID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if tree.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к этой заявке"})
		return
	}

	if err := h.Repository.RemoveFromTree(uint(treeID), uint(anomalyID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Элемент удален из заявки"})
}

// Вспомогательная функция для парсинга максимального кольца
func parseMaxAnomalousRing(anomalousRings string) int {
	if anomalousRings == "" {
		return 0
	}

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
