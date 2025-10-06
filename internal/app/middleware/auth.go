package middleware

import (
	"RIP-WEB/internal/app/config"
	"RIP-WEB/internal/app/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config, tokenManager *utils.TokenManager) gin.HandlerFunc {
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
				token := parts[1]

				// Проверяем, не в blacklist ли токен
				if tokenManager.IsTokenBlacklisted(token) {
					ctx.Next()
					return
				}

				claims, err := utils.ValidateJWT(token, cfg.JWTSecret)
				if err == nil {
					ctx.Set("user_id", claims.UserID)
					ctx.Set("login", claims.Login)
					ctx.Set("is_moderator", claims.IsModerator)
					ctx.Set("authenticated", true)
					ctx.Set("token", token)
				}
			}
		}
		ctx.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
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

func RequireModerator() gin.HandlerFunc {
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
