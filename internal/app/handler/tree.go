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
			"status":      "error",
			"description": "Неверный ID дерева",
		})
		logrus.Error(err)
		return
	}

	tree, treeItems, err := h.Repository.GetTreeWithItems(uint(treeID)) //ORM вызов
	if err != nil {
		// Если дерево не найдено, возвращаем JSON ошибку
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"description": "Заявка не найдена",
		})
		return
	}

	// Проверяем, что дерево не удалено
	if tree.Status == "удалён" {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"description": "you can't watch deleted tree",
		})
		return
	}

	// Если заявка активна - показываем HTML страницу
	ctx.HTML(http.StatusOK, "tree.tmpl", gin.H{
		"tree":       tree,
		"treeItems":  treeItems,
		"treeID":     treeID,
		"cart_count": h.Repository.GetCartCount(),
	})
}
