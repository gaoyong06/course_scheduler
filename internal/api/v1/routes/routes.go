package routes

import (
	"github.com/gin-gonic/gin"

	"course_scheduler/internal/api/v1/handlers"
	"course_scheduler/internal/api/v1/middlewares"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		// 示例路由
		v1.GET("/handler1", middlewares.Middleware1(), handlers.Handler1)
	}
}
