package middleware

import (
	"net/http"

	"github.com/TuanNghia295/Movie-Streaming-App/server/utils"
	"github.com/gin-gonic/gin"
)

// Validate accessToken and cron access
func AuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := utils.GetAccessToken(ctx)

		// ctx.Abort(): là một method trong Gin được sử dụng để dừng quá trình xử lý của middleware và không cho các middleware hoặc handler tiếp theo được gọi

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			ctx.Abort()
			return
		}

		claims, err := utils.ValidateToken(token)

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}
		ctx.Set("userId", claims.UserId)
		ctx.Set("role", claims.Role)

		ctx.Next() // continue execute
	}
}
