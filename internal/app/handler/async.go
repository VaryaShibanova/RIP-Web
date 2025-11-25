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

// AsyncResultResponse теперь принимает ВСЕ результаты
type AsyncResultResponse struct {
	TreeID      uint             `json:"tree_id" binding:"required"`
	Message     string           `json:"message" binding:"required"`
	Results     []TreeItemResult `json:"results" binding:"required"`
	TotalItems  int              `json:"total_items" binding:"required"`
	FinalStatus string           `json:"final_status" binding:"required"`
}

type TreeItemResult struct {
	TreeItemID     uint   `json:"tree_item_id" binding:"required"`
	CalculatedYear int    `json:"calculated_year" binding:"required"`
	Status         string `json:"status" binding:"required"`
}

// ReceiveAsyncResult обработка ВСЕХ результатов одним запросом
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
		fmt.Printf("❌ JSON bind error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("📦 Received ALL results for tree %d: %d items\n",
		request.TreeID, len(request.Results))

	// 👇 ОБНОВЛЯЕМ ВСЕ TreeItems ОДНИМ МАССИВОМ
	successCount := 0
	failedCount := 0

	for _, result := range request.Results {
		if result.Status == "completed" {
			err := h.Repository.UpdateTreeItemCalculatedYear(
				result.TreeItemID,
				result.CalculatedYear,
			)
			if err != nil {
				fmt.Printf("❌ Failed to update tree_item %d: %v\n", result.TreeItemID, err)
				failedCount++
			} else {
				successCount++
				fmt.Printf("✅ Updated tree_item %d with calculated_year %d\n",
					result.TreeItemID, result.CalculatedYear)
			}
		} else {
			fmt.Printf("⚠️  Skipping tree_item %d with status: %s\n",
				result.TreeItemID, result.Status)
			failedCount++
		}
	}

	// 👇 РАССЧИТЫВАЕМ final_year ДЛЯ ВСЕЙ ЗАЯВКИ
	fmt.Printf("🔢 Calculating final year for tree %d\n", request.TreeID)
	err := h.Repository.CalculateAndUpdateFinalYear(request.TreeID)
	if err != nil {
		fmt.Printf("❌ Failed to calculate final year: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("🎯 All updates completed for tree %d: %d successful, %d failed\n",
		request.TreeID, successCount, failedCount)

	ctx.JSON(http.StatusOK, gin.H{
		"message":            "All tree items processed and final year calculated",
		"tree_id":            request.TreeID,
		"successful_updates": successCount,
		"failed_updates":     failedCount,
		"total_items":        len(request.Results),
		"final_year_updated": true,
	})
}

// StartAsyncCalculations запуск асинхронных расчетов при завершении модератором
func (h *Handler) StartAsyncCalculations(treeID uint) {
	tree, treeItems, err := h.Repository.GetTreeWithItems(treeID)
	if err != nil {
		fmt.Printf("❌ Error getting tree items: %v\n", err)
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
	fmt.Printf("🚀 Sending to Django: tree_id=%d, items_count=%d\n", treeID, len(calcItems))
	for i, item := range calcItems {
		fmt.Printf("   Item %d: tree_item_id=%d, anomaly_id=%d\n",
			i+1, item.TreeItemID, item.AnomalyID)
	}

	// Отправляем в Django сервис
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Printf("❌ Error marshaling request: %v\n", err)
		return
	}

	resp, err := http.Post(
		"http://localhost:8000/api/async/calculate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("❌ Error calling async service: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Читаем ответ от Django
	var response struct {
		Message    string `json:"message"`
		TreeID     uint   `json:"tree_id"`
		TotalItems int    `json:"total_items"`
		Status     string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("❌ Error parsing Django response: %v\n", err)
		return
	}

	fmt.Printf("✅ Async calculations started for tree %d, status: %s\n",
		treeID, response.Status)
	fmt.Printf("   Message: %s, Total items: %d\n",
		response.Message, response.TotalItems)
}
