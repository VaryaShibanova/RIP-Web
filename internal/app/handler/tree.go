package handler

import (
	"RIP-WEB/internal/app/ds"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/*// GetTreeCart godoc
// @Summary Получение данных корзины
// @Description Возвращает ID пользователя и количество элементов в корзине
// @Tags trees
// @Produce json
// @Success 200 {object} TreeCartPublicResponse
// @Router /api/trees/cart [get]
func (h *Handler) GetTreeCart(ctx *gin.Context) {
	// Временно возвращаем статические данные вместо проверки авторизации
	ctx.JSON(http.StatusOK, gin.H{
		"user_id":    -1,
		"item_count": 0,
	})
}*/

// GetTreeCart godoc
// @Summary Получение данных корзины
// @Description Возвращает ID черновой заявки и количество элементов в ней
// @Tags trees
// @Produce json
// @Security BearerAuth
// @Success 200 {object} TreeCartResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/trees/cart [get]
func (h *Handler) GetTreeCart(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	tree, err := h.Repository.GetDraftTree(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var treeID uint = 0
	var count int64 = 0

	if tree != nil {
		treeID = tree.ID
		count = h.Repository.GetCartCount(userID.(uint))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"tree_id":    treeID,
		"item_count": count,
	})
}

// GetTrees godoc
// @Summary Получение списка заявок
// @Description Возвращает список заявок с фильтрацией по статусу и дате
// @Tags trees
// @Produce json
// @Security BearerAuth
// @Param status query string false "Фильтр по статусу"
// @Param date_from query string false "Фильтр по дате от (формат: YYYY-MM-DD)"
// @Param date_to query string false "Фильтр по дате до (формат: YYYY-MM-DD)"
// @Success 200 {object} TreesListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/trees [get]
func (h *Handler) GetTrees(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	isModerator, _ := ctx.Get("is_moderator")
	status := ctx.Query("status")
	dateFromStr := ctx.Query("date_from")
	dateToStr := ctx.Query("date_to")

	var dateFrom, dateTo time.Time
	var err error

	if dateFromStr != "" {
		dateFrom, err = time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты от"})
			return
		}
	}

	if dateToStr != "" {
		dateTo, err = time.Parse("2006-01-02", dateToStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты до"})
			return
		}
	}

	var trees []ds.Tree
	if isModerator.(bool) {
		// Модератор видит ВСЕ заявки без исключений
		trees, err = h.Repository.GetAllTreesForModerator(status, dateFrom, dateTo)
	} else {
		// Пользователь видит только СВОИ заявки кроме удаленных
		trees, err = h.Repository.GetUserTreesWithFilters(userID.(uint), status, dateFrom, dateTo)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type TreeResponse struct {
		ID                uint   `json:"id"`
		Creator           string `json:"creator"`
		Moderator         string `json:"moderator,omitempty"`
		AmountOfAnomalies int    `json:"amount_of_anomalies"`
		Status            string `json:"status,omitempty"`
		FinalYear         int    `json:"final_year"` // Добавляем final_year
	}

	response := make([]TreeResponse, len(trees))
	for i, tree := range trees {
		var itemCount int64
		h.Repository.GetDB().Model(&ds.TreeItem{}).Where("tree_id = ?", tree.ID).Count(&itemCount)

		response[i] = TreeResponse{
			ID:                tree.ID,
			Creator:           tree.Creator.Login,
			AmountOfAnomalies: int(itemCount),
			FinalYear:         tree.FinalYear, // Добавляем final_year
		}

		if tree.ModeratorID.Valid {
			response[i].Moderator = tree.Moderator.Login
		}

		// Добавляем статус только для модератора
		if isModerator.(bool) {
			response[i].Status = tree.Status
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"trees": response,
	})
}

// GetTree godoc
// @Summary Получение информации о заявке
// @Description Возвращает полную информацию о заявке и ее элементах
// @Tags trees
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} TreeDetailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/trees/{id} [get]
func (h *Handler) GetTree(ctx *gin.Context) {
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

	tree, treeItems, err := h.Repository.GetTreeWithItems(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	isModerator, _ := ctx.Get("is_moderator")

	// Проверяем доступ: модератор ИЛИ создатель заявки
	if !isModerator.(bool) && tree.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к этой заявке"})
		return
	}

	// Для пользователя скрываем удаленные заявки других пользователей
	if !isModerator.(bool) && tree.Status == "удалён" && tree.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
		return
	}

	type SimplifiedTreeResponse struct {
		ID          uint   `json:"id"`
		Description string `json:"description"`
		TotalRings  int    `json:"total_rings"`
		FinalYear   int    `json:"final_year"` // Убедимся что есть
		Status      string `json:"status,omitempty"`
		CreatorID   uint   `json:"creator_id"`
	}

	type SimplifiedTreeItemResponse struct {
		AnomalyID      uint   `json:"anomaly_id"`
		AnomalousRings string `json:"anomalous_rings"`
		CalculatedYear int    `json:"calculated_year"`
		AnomalyName    string `json:"anomaly_name"`
		AnomalyImage   string `json:"anomaly_image"`
	}

	simplifiedTree := SimplifiedTreeResponse{
		ID:          tree.ID,
		Description: tree.Description,
		TotalRings:  tree.TotalRings,
		FinalYear:   tree.FinalYear, // Добавляем final_year
		CreatorID:   tree.CreatorID,
	}

	// Добавляем статус только для модератора
	if isModerator.(bool) {
		simplifiedTree.Status = tree.Status
	}

	simplifiedItems := make([]SimplifiedTreeItemResponse, len(treeItems))
	for i, item := range treeItems {
		simplifiedItems[i] = SimplifiedTreeItemResponse{
			AnomalyID:      item.AnomalyID,
			AnomalousRings: item.AnomalousRings,
			CalculatedYear: item.CalculatedYear,
			AnomalyName:    item.Anomaly.Name,
			AnomalyImage:   item.Anomaly.Image,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"tree":      simplifiedTree,
		"treeItems": simplifiedItems,
	})
}

// UpdateTree godoc
// @Summary Обновление заявки
// @Description Обновляет данные заявки (только для создателя и только черновые заявки)
// @Tags trees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param tree body UpdateTreeRequest true "Данные для обновления"
// @Success 200 {object} TreeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/trees/{id} [put]
func (h *Handler) UpdateTree(ctx *gin.Context) {
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

	tree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if tree.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Можно редактировать только свои заявки"})
		return
	}

	if tree.Status != "черновик" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Можно редактировать только черновые заявки"})
		return
	}

	var updateData struct {
		Description string `json:"description"`
		TotalRings  int    `json:"total_rings"`
		FinalYear   int    `json:"final_year"` // Добавляем final_year
	}

	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tree.Description = updateData.Description
	tree.TotalRings = updateData.TotalRings
	tree.FinalYear = updateData.FinalYear // Обновляем final_year

	if err := h.Repository.UpdateTree(tree); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type SimplifiedTreeResponse struct {
		ID          uint   `json:"id"`
		Description string `json:"description"`
		TotalRings  int    `json:"total_rings"`
		FinalYear   int    `json:"final_year"` // Добавляем final_year
	}

	simplifiedResponse := SimplifiedTreeResponse{
		ID:          tree.ID,
		Description: tree.Description,
		TotalRings:  tree.TotalRings,
		FinalYear:   tree.FinalYear, // Добавляем final_year
	}

	ctx.JSON(http.StatusOK, simplifiedResponse)
}

// FormTree godoc
// @Summary Формирование заявки
// @Description Переводит заявку из статуса "черновик" в "сформирован"
// @Tags trees
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} TreeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/trees/{id}/form [put]
func (h *Handler) FormTree(ctx *gin.Context) {
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

	tree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if tree.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Можно формировать только свои заявки"})
		return
	}

	if tree.Status != "черновик" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Можно формировать только черновые заявки"})
		return
	}

	if err := h.Repository.FormTree(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedTree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type SimplifiedTreeResponse struct {
		ID          uint   `json:"id"`
		Status      string `json:"status"`
		Description string `json:"description"`
		TotalRings  int    `json:"total_rings"`
		FinalYear   int    `json:"final_year"`
	}

	simplifiedResponse := SimplifiedTreeResponse{
		ID:          updatedTree.ID,
		Status:      updatedTree.Status,
		Description: updatedTree.Description,
		TotalRings:  updatedTree.TotalRings,
		FinalYear:   updatedTree.FinalYear,
	}

	ctx.JSON(http.StatusOK, simplifiedResponse)
}

// CompleteTree godoc
// @Summary Завершение заявки модератором
// @Description Завершает или отклоняет заявку (только для модераторов)
// @Tags trees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param action body CompleteTreeRequest true "Действие (complete/reject)"
// @Success 200 {object} CompleteTreeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/trees/{id}/complete [put]
func (h *Handler) CompleteTree(ctx *gin.Context) {
	userID, exists := h.getUserIDFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	isModerator := h.getIsModeratorFromContext(ctx)
	if !isModerator {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Требуются права модератора"})
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var request struct {
		Action string `json:"action" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Action != "complete" && request.Action != "reject" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Действие должно быть 'complete' или 'reject'"})
		return
	}

	// Проверяем существование заявки перед завершением
	existingTree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingTree.Status != "сформирован" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Можно завершать только сформированные заявки"})
		return
	}

	if err := h.Repository.CompleteTree(uint(id), userID, request.Action); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем обновленную заявку с рассчитанными годами
	updatedTree, updatedTreeItems, err := h.Repository.GetTreeWithItems(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type AnomalyCalculatedYear struct {
		AnomalyID      uint   `json:"anomaly_id"`
		AnomalyName    string `json:"anomaly_name"`
		AnomalousRings string `json:"anomalous_rings"`
		CalculatedYear int    `json:"calculated_year"`
	}

	type CompleteTreeResponse struct {
		ID             uint                    `json:"id"`
		Status         string                  `json:"status"`
		FinalYear      int                     `json:"final_year"`
		Anomalies      []AnomalyCalculatedYear `json:"anomalies"`
		TotalAnomalies int                     `json:"total_anomalies"`
	}

	// Формируем список аномалий с calculated_year
	anomalies := make([]AnomalyCalculatedYear, len(updatedTreeItems))
	for i, item := range updatedTreeItems {
		anomalies[i] = AnomalyCalculatedYear{
			AnomalyID:      item.AnomalyID,
			AnomalyName:    item.Anomaly.Name,
			AnomalousRings: item.AnomalousRings,
			CalculatedYear: item.CalculatedYear,
		}
	}

	response := CompleteTreeResponse{
		ID:             updatedTree.ID,
		Status:         updatedTree.Status,
		FinalYear:      updatedTree.FinalYear,
		Anomalies:      anomalies,
		TotalAnomalies: len(anomalies),
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteTree godoc
// @Summary Удаление заявки
// @Description Удаляет заявку (помечает статус как "удалён")
// @Tags trees
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/trees/{id} [delete]
func (h *Handler) DeleteTree(ctx *gin.Context) {
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

	tree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if tree.CreatorID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Можно удалять только свои заявки"})
		return
	}

	if err := h.Repository.DeleteTree(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка удалена"})
}
