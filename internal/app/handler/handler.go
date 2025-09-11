package handler

import (
	"RIP-WEB/internal/app/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetAnomalies(ctx *gin.Context) {
	var anomalies []repository.Anomaly
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		anomalies, err = h.Repository.GetAnomalies()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		anomalies, err = h.Repository.GetAnomaliesByPattern(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":      time.Now().Format("15:04:05"),
		"anomalies": anomalies,
		"query":     searchQuery,
	})
}

func (h *Handler) GetAnomaly(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	anomaly, err := h.Repository.GetAnomaly(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "anomaly.html", gin.H{
		"anomaly": anomaly,
	})
}

func (h *Handler) GetRequests(ctx *gin.Context) {
	requests, err := h.Repository.GetRequests()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "requests.html", gin.H{
		"time":     time.Now().Format("15:04:05"),
		"requests": requests,
	})
}
