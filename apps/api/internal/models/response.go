package models

import "github.com/gin-gonic/gin"

func NewResponse(message string, data any) *gin.H {
	return &gin.H{
		"message": message,
		"data":    data,
	}
}
