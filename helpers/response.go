package helpers

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, code int, message string, data interface{}, meta interface{}) {
	resp := gin.H{"success": true, "message": message, "data": data}
	if meta != nil {
		resp["meta"] = meta
	}
	c.JSON(code, resp)
}

func Error(c *gin.Context, code int, message string, errors interface{}) {
	resp := gin.H{"success": false, "message": message}
	if errors != nil {
		resp["errors"] = errors
	}
	c.JSON(code, resp)
}

func ValidationError(c *gin.Context, errors map[string][]string) {
	c.JSON(422, gin.H{
		"success": false,
		"message": "Validasi gagal",
		"errors":  errors,
	})
}
