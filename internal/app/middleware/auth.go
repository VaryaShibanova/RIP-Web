package middleware

import (
	"RIP-WEB/internal/app/config"
	"RIP-WEB/internal/app/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

		// Сбрасываем аутентификацию по умолчанию
		ctx.Set("authenticated", false)
		ctx.Set("user_id", uint(0))
		ctx.Set("is_moderator", false)
		ctx.Set("login", "")

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]

				// ПРОВЕРЯЕМ BLACKLIST ПЕРЕД ВАЛИДАЦИЕЙ JWT
				if tokenManager.IsTokenBlacklisted(token) {
					logrus.Info("Token found in blacklist, rejecting request")
					ctx.Set("authenticated", false)
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
					logrus.Infof("User authenticated: %s (ID: %d)", claims.Login, claims.UserID)
				} else {
					logrus.Warnf("JWT validation failed: %v", err)
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
