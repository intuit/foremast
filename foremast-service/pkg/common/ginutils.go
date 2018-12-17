package common

import "github.com/gin-gonic/gin"


// ErrorResponse .... use ErrorResponse to handle error
func ErrorResponse(c *gin.Context, code int, err string) {
	c.JSON(code, gin.H{
		"error": err,
	})
}
