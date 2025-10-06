package middleware

import (
	"RIP-WEB/internal/app/config"
	"RIP-WEB/internal/app/session"
	"RIP-WEB/internal/app/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config, sessionManager *session.Manager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Проверяем JWT токен
		authHeader := ctx.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				claims, err := utils.ValidateJWT(parts[1], cfg.JWTSecret)
				if err == nil {
					ctx.Set("user_id", claims.UserID)
					ctx.Set("login", claims.Login)
					ctx.Set("is_moderator", claims.IsModerator)
					ctx.Set("authenticated", true)
					ctx.Next()
					return
				}
			}
		}

		// Проверяем сессию в Redis через куки
		sessionID, err := ctx.Cookie("session_id")
		if err == nil && sessionID != "" {
			sess, err := sessionManager.GetSession(ctx.Request.Context(), sessionID)
			if err == nil && sess != nil {
				ctx.Set("user_id", sess.UserID)
				ctx.Set("login", sess.Login)
				ctx.Set("is_moderator", sess.IsModerator)
				ctx.Set("authenticated", true)

				// Обновляем TTL сессии
				sessionManager.CreateSession(ctx.Request.Context(), sessionID, sess)
				ctx.Next()
				return
			}
		}

		ctx.Next()
	}
}
