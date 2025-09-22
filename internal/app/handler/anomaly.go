package handler

import (
	"net/http"
	"strconv"

	"RIP-WEB/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetAllAnomalies(ctx *gin.Context) {
	var anomalies []ds.Anomaly
	var err error

	search := ctx.Query("findanomalies")
	if search == "" {
		anomalies, err = h.Repository.GetAllAnomalies()
	} else {
		anomalies, err = h.Repository.SearchAnomaliesByName(search)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
		"anomalies":  anomalies,
		"cart_count": h.Repository.GetCartCount(),
		"query":      search,
	})
}

func (h *Handler) GetAnomalyById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	anomaly, err := h.Repository.GetAnomalyByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "anomaly.tmpl", anomaly)
}
