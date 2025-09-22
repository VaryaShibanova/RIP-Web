package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetTree(ctx *gin.Context) {
	strId := ctx.Param("id")
	treeID, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Неверный ID дерева",
		})
		logrus.Error(err)
		return
	}

	tree, treeItems, err := h.Repository.GetTreeWithItems(uint(treeID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "tree.tmpl", gin.H{
		"tree":       tree,
		"treeItems":  treeItems,
		"treeID":     treeID,
		"cart_count": h.Repository.GetCartCount(),
	})
}

func (h *Handler) AddToTree(ctx *gin.Context) {
	creatorID := uint(1)
	anomalyID, _ := strconv.Atoi(ctx.PostForm("anomaly_id"))
	anomalousRings := ctx.PostForm("anomalous_rings")
	calculatedYear, _ := strconv.Atoi(ctx.PostForm("calculated_year"))

	tree, err := h.Repository.GetOrCreateDraftTree(creatorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = h.Repository.AddAnomalyToTree(tree.ID, uint(anomalyID), anomalousRings, calculatedYear)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Redirect(http.StatusFound, "/anomalies")
}

func (h *Handler) DeleteTree(ctx *gin.Context) {
	strId := ctx.PostForm("tree_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = h.Repository.DeleteTree(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Redirect(http.StatusFound, "/anomalies")
}
