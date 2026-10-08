package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"siakad-mini/helpers"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			helpers.Error(c, 401, "Token tidak valid atau tidak ada", nil)
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := helpers.ParseToken(tokenString)
		if err != nil {
			helpers.Error(c, 401, "Token tidak valid atau kedaluwarsa", nil)
			c.Abort()
			return
		}
		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Set("email", claims["email"].(string))
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}
