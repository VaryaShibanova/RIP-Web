package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ДОМЕН ЗАЯВКИ (TREE)

// GetTreeCart - GET иконки корзины
func (h *Handler) GetTreeCart(ctx *gin.Context) {
	user := h.Repository.GetSystemUser()

	tree, err := h.Repository.GetDraftTree(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var treeID uint = 0
	var count int64 = 0

	if tree != nil {
		treeID = tree.ID
		count = h.Repository.GetCartCount(user.ID)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"tree_id":    treeID,
		"item_count": count,
	})
}

// GetTrees - GET список заявок с фильтрацией
func (h *Handler) GetTrees(ctx *gin.Context) {
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

	trees, err := h.Repository.GetTreesWithFilters(status, dateFrom, dateTo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем для клиента (только логины)
	type TreeResponse struct {
		ID          uint      `json:"id"`
		Status      string    `json:"status"`
		Description string    `json:"description"`
		TotalRings  int       `json:"total_rings"`
		FinalYear   int       `json:"final_year"`
		DateCreate  time.Time `json:"date_create"`
		DateUpdate  time.Time `json:"date_update"`
		DateFinish  time.Time `json:"date_finish,omitempty"`
		Creator     string    `json:"creator"`
		Moderator   string    `json:"moderator,omitempty"`
	}

	response := make([]TreeResponse, len(trees))
	for i, tree := range trees {
		response[i] = TreeResponse{
			ID:          tree.ID,
			Status:      tree.Status,
			Description: tree.Description,
			TotalRings:  tree.TotalRings,
			FinalYear:   tree.FinalYear,
			DateCreate:  tree.DateCreate,
			DateUpdate:  tree.DateUpdate,
			Creator:     tree.Creator.Login,
		}

		if tree.ModeratorID.Valid {
			response[i].Moderator = tree.Moderator.Login
		}
		if tree.DateFinish.Valid {
			response[i].DateFinish = tree.DateFinish.Time
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"trees": response,
	})
}

// GetTree - GET одна запись заявки (JSON API)
func (h *Handler) GetTree(ctx *gin.Context) {
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

	// Проверяем что заявка не удалена
	if tree.Status == "удалён" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка удалена"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"tree":      tree,
		"treeItems": treeItems,
	})
}

// UpdateTree - PUT изменения полей заявки
func (h *Handler) UpdateTree(ctx *gin.Context) {
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

	var updateData struct {
		Description string `json:"description"`
		TotalRings  int    `json:"total_rings"`
	}

	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем только разрешенные поля
	tree.Description = updateData.Description
	tree.TotalRings = updateData.TotalRings

	if err := h.Repository.UpdateTree(tree); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tree)
}

// FormTree - PUT сформировать заявку
func (h *Handler) FormTree(ctx *gin.Context) {
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

	// Проверяем что заявка в статусе черновика
	if tree.Status != "черновик" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Можно формировать только черновые заявки"})
		return
	}

	if err := h.Repository.FormTree(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем обновленную заявку
	updatedTree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedTree)
}

// CompleteTree - PUT завершить/отклонить заявку
func (h *Handler) CompleteTree(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var request struct {
		Action string `json:"action" binding:"required"` // "complete" или "reject"
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Action != "complete" && request.Action != "reject" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Действие должно быть 'complete' или 'reject'"})
		return
	}

	moderator := h.Repository.GetModerator()

	if err := h.Repository.CompleteTree(uint(id), moderator.ID, request.Action); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем обновленную заявку
	updatedTree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedTree)
}

// DeleteTree - DELETE удаление заявки
func (h *Handler) DeleteTree(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	if err := h.Repository.DeleteTree(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка удалена"})
}
