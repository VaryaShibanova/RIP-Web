package handler

import (
	"RIP-WEB/internal/app/config"
	"RIP-WEB/internal/app/repository"
	"RIP-WEB/internal/app/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "RIP-WEB/docs" // Swagger docs
)

type Handler struct {
	Repository   *repository.Repository
	Config       *config.Config
	TokenManager *utils.TokenManager
}

func NewHandler(r *repository.Repository, cfg *config.Config, tokenManager *utils.TokenManager) *Handler {
	return &Handler{
		Repository:   r,
		Config:       cfg,
		TokenManager: tokenManager,
	}
}

func (h *Handler) RegisterAPIHandlers(router *gin.Engine) {
	api := router.Group("/api")

	// Swagger документация
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Публичные маршруты
	api.GET("/anomalies", h.GetAnomalies)
	api.GET("/anomalies/:id", h.GetAnomaly)
	api.POST("/users/register", h.RegisterUser)
	api.POST("/users/login", h.LoginUser)

	// Защищенные маршруты (middleware уже применено глобально)
	{
		// Пользовательские маршруты
		api.GET("/users/me", h.RequireAuth(), h.GetCurrentUser)
		api.POST("/users/logout", h.RequireAuth(), h.LogoutUser)
		api.PUT("/users/profile", h.RequireAuth(), h.UpdateUserProfile)

		// Аномалии (только для авторизованных)
		api.POST("/anomalies", h.RequireAuth(), h.CreateAnomaly)
		api.PUT("/anomalies/:id", h.RequireAuth(), h.UpdateAnomaly)
		api.DELETE("/anomalies/:id", h.RequireAuth(), h.DeleteAnomaly)
		api.POST("/anomalies/:id/image", h.RequireAuth(), h.UploadAnomalyImage)

		// Заявки
		api.GET("/trees/cart", h.RequireAuth(), h.GetTreeCart)
		api.GET("/trees", h.RequireAuth(), h.GetTrees)
		api.POST("/trees/current/items", h.RequireAuth(), h.AddToTree)
		api.GET("/trees/:id", h.RequireAuth(), h.GetTree)
		api.PUT("/trees/:id", h.RequireAuth(), h.UpdateTree)
		api.PUT("/trees/:id/form", h.RequireAuth(), h.FormTree)
		api.DELETE("/trees/:id", h.RequireAuth(), h.DeleteTree)

		// Tree items
		items := api.Group("/trees/:id/items")
		items.Use(h.RequireAuth())
		{
			items.PUT("/:anomaly_id", h.UpdateTreeItem)
			items.DELETE("/:anomaly_id", h.RemoveFromTree)
		}

		// Маршруты модератора
		moderator := api.Group("")
		moderator.Use(h.RequireModerator())
		{
			moderator.PUT("/trees/:id/complete", h.CompleteTree)
		}
	}
}

// RequireAuth - middleware для проверки аутентификации
func (h *Handler) RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authenticated, exists := ctx.Get("authenticated")
		if !exists || !authenticated.(bool) {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется аутентификация",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// RequireModerator - middleware для проверки прав модератора
func (h *Handler) RequireModerator() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isModerator, exists := ctx.Get("is_moderator")
		if !exists || !isModerator.(bool) {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "Требуются права модератора",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.Static("/resources", "./resources")
	router.GET("/favicon.ico", func(ctx *gin.Context) {
		ctx.Status(204)
	})
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
