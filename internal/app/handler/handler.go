package handler

import (
	"RIP-WEB/internal/app/config"
	"RIP-WEB/internal/app/repository"
	"RIP-WEB/internal/app/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "RIP-WEB/docs" // Swagger docs
)

type Handler struct {
	Repository *repository.Repository
	Config     *config.Config
}

func NewHandler(r *repository.Repository, cfg *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Config:     cfg,
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

	// Защищенные маршруты
	auth := api.Group("")
	auth.Use(h.AuthMiddleware())
	{
		// Пользовательские маршруты
		auth.GET("/users/me", h.GetCurrentUser)
		auth.POST("/users/logout", h.LogoutUser)
		auth.PUT("/users/profile", h.UpdateUserProfile)

		// Аномалии (только для авторизованных)
		auth.POST("/anomalies", h.RequireAuth(), h.CreateAnomaly)
		auth.PUT("/anomalies/:id", h.RequireAuth(), h.UpdateAnomaly)
		auth.DELETE("/anomalies/:id", h.RequireAuth(), h.DeleteAnomaly)
		auth.POST("/anomalies/:id/image", h.RequireAuth(), h.UploadAnomalyImage)

		// Заявки
		auth.GET("/trees/cart", h.RequireAuth(), h.GetTreeCart)
		auth.GET("/trees", h.RequireAuth(), h.GetTrees)
		auth.POST("/trees/current/items", h.RequireAuth(), h.AddToTree)
		auth.GET("/trees/:id", h.RequireAuth(), h.GetTree)
		auth.PUT("/trees/:id", h.RequireAuth(), h.UpdateTree)
		auth.PUT("/trees/:id/form", h.RequireAuth(), h.FormTree)
		auth.DELETE("/trees/:id", h.RequireAuth(), h.DeleteTree)

		// Tree items
		items := auth.Group("/trees/:id/items")
		items.Use(h.RequireAuth())
		{
			items.PUT("/:anomaly_id", h.UpdateTreeItem)
			items.DELETE("/:anomaly_id", h.RemoveFromTree)
		}

		// Маршруты модератора
		moderator := auth.Group("")
		moderator.Use(h.RequireModerator())
		{
			moderator.PUT("/trees/:id/complete", h.CompleteTree)
		}
	}
}

// AuthMiddleware - middleware для аутентификации через JWT
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			token, err := ctx.Cookie("token")
			if err == nil {
				authHeader = "Bearer " + token
			}
		}

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				claims, err := utils.ValidateJWT(parts[1], h.Config.JWTSecret)
				if err == nil {
					ctx.Set("user_id", claims.UserID)
					ctx.Set("login", claims.Login)
					ctx.Set("is_moderator", claims.IsModerator)
					ctx.Set("authenticated", true)
				}
			}
		}
		ctx.Next()
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
