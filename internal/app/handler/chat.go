package handler

import (
	"net/http"
	"strconv"

	"RIP-WEB/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetAllChats(ctx *gin.Context) {
	var chats []ds.Chat
	var err error

	search := ctx.Query("search")
	if search == "" {
		chats, err = h.Repository.GetAllChats()
	} else {
		chats, err = h.Repository.SearchChatsByName(search)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "chats.page.tmpl", gin.H{
		"data":       chats,
		"cart_count": h.Repository.GetCartCount(),
		"search":     search,
	})
}

func (h *Handler) GetChatById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	chat, err := h.Repository.GetChatByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "chat.page.tmpl", chat)
}
