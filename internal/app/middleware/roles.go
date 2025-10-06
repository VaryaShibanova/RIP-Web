package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
