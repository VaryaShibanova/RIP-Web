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

	// Защищенные маршруты (требуют аутентификации)
	{
		// Пользовательские маршруты
		api.GET("/users/me", h.RequireAuth(), h.GetCurrentUser)
		api.POST("/users/logout", h.RequireAuth(), h.LogoutUser)
		api.PUT("/users/profile", h.RequireAuth(), h.UpdateUserProfile)

		// Заявки (доступны всем авторизованным пользователям)
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
	}

	// Маршруты модератора - управление аномалиями
	moderatorAnomalies := api.Group("/anomalies")
	moderatorAnomalies.Use(h.RequireAuth(), h.RequireModerator())
	{
		moderatorAnomalies.POST("", h.CreateAnomaly)
		moderatorAnomalies.PUT("/:id", h.UpdateAnomaly)
		moderatorAnomalies.DELETE("/:id", h.DeleteAnomaly)
		moderatorAnomalies.POST("/:id/image", h.UploadAnomalyImage)
	}

	// Маршруты модератора - завершение заявок
	moderatorTrees := api.Group("/trees")
	moderatorTrees.Use(h.RequireAuth(), h.RequireModerator())
	{
		moderatorTrees.PUT("/:id/complete", h.CompleteTree)
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

// getUserIDFromContext безопасно извлекает user_id из контекста
func (h *Handler) getUserIDFromContext(ctx *gin.Context) (uint, bool) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return 0, false
	}

	switch v := userID.(type) {
	case uint:
		return v, true
	case float64:
		return uint(v), true
	case int:
		return uint(v), true
	case int64:
		return uint(v), true
	default:
		return 0, false
	}
}

// getIsModeratorFromContext безопасно извлекает is_moderator из контекста
func (h *Handler) getIsModeratorFromContext(ctx *gin.Context) bool {
	isModerator, exists := ctx.Get("is_moderator")
	if !exists {
		return false
	}

	switch v := isModerator.(type) {
	case bool:
		return v
	default:
		return false
	}
}

// getLoginFromContext безопасно извлекает login из контекста
func (h *Handler) getLoginFromContext(ctx *gin.Context) string {
	login, exists := ctx.Get("login")
	if !exists {
		return ""
	}

	switch v := login.(type) {
	case string:
		return v
	default:
		return ""
	}
}
