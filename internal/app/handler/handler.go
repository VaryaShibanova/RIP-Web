package handler

import (
	"RIP-WEB/internal/app/repository"

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

func (h *Handler) RegisterAPIHandlers(router *gin.Engine) {
	api := router.Group("/api")

	// Домен услуги (Anomaly)
	anomalies := api.Group("/anomalies")
	{
		anomalies.GET("", h.GetAnomalies)
		anomalies.POST("", h.CreateAnomaly)
		anomalies.GET("/:id", h.GetAnomaly)
		anomalies.PUT("/:id", h.UpdateAnomaly)
		anomalies.DELETE("/:id", h.DeleteAnomaly)
		anomalies.POST("/:id/image", h.UploadAnomalyImage)
	}

	// Домен заявки (Tree)
	trees := api.Group("/trees")
	{
		trees.GET("/cart", h.GetTreeCart)
		trees.GET("", h.GetTrees)
		trees.GET("/:id", h.GetTree)
		trees.PUT("/:id", h.UpdateTree)
		trees.PUT("/:id/form", h.FormTree)
		trees.PUT("/:id/complete", h.CompleteTree)
		trees.DELETE("/:id", h.DeleteTree)
		trees.POST("/current/items", h.AddToTree)
	}

	// Домен м-м (TreeItem)
	treeItems := api.Group("/trees/:tree_id/items")
	{
		treeItems.PUT("/:anomaly_id", h.UpdateTreeItem)
		treeItems.DELETE("/:anomaly_id", h.RemoveFromTree)
	}

	// Домен пользователя (Users)
	users := api.Group("/users")
	{
		users.POST("/register", h.RegisterUser)
		users.GET("/profile", h.GetUserProfile)
		users.PUT("/profile", h.UpdateUserProfile)
		users.POST("/login", h.LoginUser)
		users.POST("/logout", h.LogoutUser)
	}
}

// Старые HTML handlers (оставляем для обратной совместимости)
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// Регистрируем API handlers
	h.RegisterAPIHandlers(router)

	// Старые HTML routes
	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(302, "/anomalies")
	})

	// Используем старые методы для HTML
	router.GET("/anomalies", h.GetAllAnomaliesHTML)
	router.GET("/anomaly/:id", h.GetAnomalyByIdHTML)
	router.GET("/tree/:id", h.GetTreeHTML)
	router.POST("/tree/add", h.AddToTreeHTML)
	router.POST("/tree/delete", h.DeleteTreeHTML)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/resources", "./resources")

	router.GET("/favicon.ico", func(ctx *gin.Context) {
		ctx.Status(204) // No Content
	})
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
