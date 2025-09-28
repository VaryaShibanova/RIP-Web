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
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверный ID дерева",
		})
		logrus.Error(err)
		return
	}

	tree, treeItems, err := h.Repository.GetTreeWithItems(uint(treeID))
	if err != nil {
		// Если дерево не найдено, возвращаем 404
		ctx.HTML(http.StatusNotFound, "tree.tmpl", gin.H{
			"tree":      nil,
			"treeItems": nil,
			"treeID":    treeID,
		})
		return
	}

	// Проверяем, что дерево не удалено
	if tree.Status == "удалён" {
		ctx.HTML(http.StatusNotFound, "tree.tmpl", gin.H{
			"tree":      nil,
			"treeItems": nil,
			"treeID":    treeID,
		})
		return
	}

	ctx.HTML(http.StatusOK, "tree.tmpl", gin.H{
		"tree":       tree,
		"treeItems":  treeItems,
		"treeID":     treeID,
		"cart_count": h.Repository.GetCartCount(),
	})
}
