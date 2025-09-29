package handler

import (
	"net/http"
	"strconv"

	"RIP-WEB/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetAllAnomalies(ctx *gin.Context) { //все аномалии
	var anomalies []ds.Anomaly
	var err error

	search := ctx.Query("findanomalies")
	if search == "" {
		anomalies, err = h.Repository.GetAllAnomalies() // ORM вызов
	} else {
		anomalies, err = h.Repository.SearchAnomalies(search) // ORM вызов
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	// Получаем количество услуг в корзине и ID текущего дерева
	cartCount := h.Repository.GetCartCount()
	currentTree, _ := h.Repository.GetDraftTree(uint(1)) // creatorID = 1
	currentTreeID := 0
	if currentTree != nil {
		currentTreeID = int(currentTree.ID)
	}

	ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
		"anomalies":       anomalies,
		"cart_count":      cartCount,
		"current_tree_id": currentTreeID,
		"query":           search,
	})
}

func (h *Handler) GetAnomalyById(ctx *gin.Context) { //конкретная аномалия
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	anomaly, err := h.Repository.GetAnomalyByID(id) // ORM вызов
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "anomaly.tmpl", anomaly)
}
