package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AsyncCalculateRequest запрос на асинхронный расчет
type AsyncCalculateRequest struct {
	TreeID    uint              `json:"tree_id" binding:"required"`
	TreeItems []TreeItemForCalc `json:"tree_items" binding:"required"`
}

type TreeItemForCalc struct {
	TreeItemID     uint   `json:"tree_item_id" binding:"required"`
	AnomalyID      uint   `json:"anomaly_id" binding:"required"`
	TotalRings     int    `json:"total_rings" binding:"required"`
	AnomalousRings string `json:"anomalous_rings" binding:"required"`
	AnomalyYear    int    `json:"anomaly_year" binding:"required"`
}

// AsyncResultResponse ответ для одного TreeItem
type AsyncResultResponse struct {
	TreeID         uint   `json:"tree_id" binding:"required"`
	TreeItemID     uint   `json:"tree_item_id" binding:"required"`
	CalculatedYear int    `json:"calculated_year" binding:"required"`
	Status         string `json:"status" binding:"required"`
}

// FinalResultResponse финальный ответ когда все расчеты завершены
type FinalResultResponse struct {
	TreeID  uint             `json:"tree_id" binding:"required"`
	Message string           `json:"message" binding:"required"`
	Results []TreeItemResult `json:"results" binding:"required"`
}

type TreeItemResult struct {
	TreeItemID     uint   `json:"tree_item_id" binding:"required"`
	CalculatedYear int    `json:"calculated_year" binding:"required"`
	Status         string `json:"status" binding:"required"`
}

// ReceiveAsyncResult обработка результата для одного TreeItem
func (h *Handler) ReceiveAsyncResult(ctx *gin.Context) {
	// ПСЕВДО АВТОРИЗАЦИЯ - простая проверка токена
	authHeader := ctx.GetHeader("Authorization")
	expectedToken := "Bearer abc12345" // Простой токен на 8+ байт

	if authHeader != expectedToken {
		fmt.Printf("❌ Invalid token. Expected: %s, Got: %s\n", expectedToken, authHeader)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
		return
	}

	fmt.Printf("✅ Token validation successful\n")

	var request AsyncResultResponse
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем calculated_year для TreeItem
	err := h.Repository.UpdateTreeItemCalculatedYear(request.TreeItemID, request.CalculatedYear)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("✅ Updated tree_item %d with calculated_year %d\n", request.TreeItemID, request.CalculatedYear)

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "TreeItem result processed",
		"tree_item_id": request.TreeItemID,
	})
}

// ReceiveFinalResult обработка финального результата когда все расчеты завершены
func (h *Handler) ReceiveFinalResult(ctx *gin.Context) {
	// ПСЕВДО АВТОРИЗАЦИЯ
	authHeader := ctx.GetHeader("Authorization")
	expectedToken := "Bearer abc12345"

	if authHeader != expectedToken {
		fmt.Printf("❌ Invalid token in final callback. Expected: %s, Got: %s\n", expectedToken, authHeader)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
		return
	}

	var request FinalResultResponse
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Рассчитываем и обновляем final_year
	err := h.Repository.CalculateAndUpdateFinalYear(request.TreeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("✅ Final year calculated for tree %d\n", request.TreeID)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Final year calculated and tree completed",
		"tree_id": request.TreeID,
	})
}

// StartAsyncCalculations запуск асинхронных расчетов при завершении модератором
func (h *Handler) StartAsyncCalculations(treeID uint) {
	tree, treeItems, err := h.Repository.GetTreeWithItems(treeID)
	if err != nil {
		fmt.Printf("Error getting tree items: %v\n", err)
		return
	}

	// Подготавливаем данные для Django
	calcItems := make([]TreeItemForCalc, len(treeItems))
	for i, item := range treeItems {
		calcItems[i] = TreeItemForCalc{
			TreeItemID:     item.ID,
			AnomalyID:      item.AnomalyID,
			TotalRings:     tree.TotalRings,
			AnomalousRings: item.AnomalousRings,
			AnomalyYear:    item.Anomaly.Year,
		}
	}

	requestData := AsyncCalculateRequest{
		TreeID:    treeID,
		TreeItems: calcItems,
	}

	// ЛОГИРОВАНИЕ
	fmt.Printf("Sending to Django: tree_id=%d, items_count=%d\n", treeID, len(calcItems))

	// Отправляем в Django сервис
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}

	resp, err := http.Post(
		"http://localhost:8000/api/async/calculate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("Error calling async service: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Async calculations started for tree %d, status: %d\n", treeID, resp.StatusCode)
}
