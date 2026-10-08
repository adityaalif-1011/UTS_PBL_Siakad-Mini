package middleware

import (
	"github.com/gin-gonic/gin"

	"siakad-mini/helpers"
)

func RoleOnly(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		helpers.Error(c, 403, "Akses ditolak", nil)
		c.Abort()
	}
}
