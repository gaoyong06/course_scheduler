package middlewares

import (
	"github.com/gin-gonic/gin"
)

// Middleware1 示例中间件
func Middleware1() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 中间件逻辑
		// ...

		// 继续处理请求
		c.Next()
	}
}
