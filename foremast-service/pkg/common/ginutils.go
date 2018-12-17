package common

import "github.com/gin-gonic/gin"

func ErrorResponse(c *gin.Context, code int, err string) {
	c.JSON(code, gin.H{
		"error": err,
	})
}
