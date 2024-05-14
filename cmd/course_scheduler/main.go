// main.go
package main

import (
	"course_scheduler/internal/api/v1/routes"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {

	// 初始化数据库
	dsn := "root:root@tcp(127.0.0.1:3306)/course_scheduler?charset=utf8mb4&parseTime=true&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 创建一个新的 Gin 引擎
	r := gin.Default()

	// 加载 v1 版本的路由
	// 注册路由
	routes.SetupRoutes(r, db)

	// 启动 HTTP 服务器
	port := ":8081"
	fmt.Printf("Server is running on port %s\n", port)
	r.Run(port)
}
