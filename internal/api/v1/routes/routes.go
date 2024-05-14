package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"course_scheduler/internal/api/v1/handlers"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, db *gorm.DB) {

	// 创建一个新的路由组
	v1 := r.Group("/api/v1")

	// 注册接收数据路由
	v1.POST("/tasks", handlers.CreateTaskHandler(db))

	// 注册执行排课路由
	v1.GET("/tasks/:task_id/execute", handlers.ExecuteTaskHandler(db))

	// 注册查询排课结果路由
	v1.GET("/tasks/:task_id/result", handlers.GetTaskResultHandler(db))
}
