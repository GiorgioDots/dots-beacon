package respond

import "github.com/gin-gonic/gin"

func Err(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}
