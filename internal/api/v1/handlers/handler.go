package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler1 示例处理程序
func Handler1(c *gin.Context) {
	// 处理请求
	// ...

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
}
