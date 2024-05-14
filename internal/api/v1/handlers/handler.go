package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"course_scheduler/internal/api/v1/middlewares"
	"course_scheduler/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

// 创建排课任务
func CreateTaskHandler(db *gorm.DB) gin.HandlerFunc {

	return func(c *gin.Context) {
		// 解析请求体中的 JSON 数据
		var taskData map[string]interface{}
		if err := c.ShouldBindJSON(&taskData); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// 将 taskData 转换为 JSON 格式的字符串
		taskDataBytes, err := json.Marshal(taskData)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 在排课任务队列中新增一条排课任务
		task := &models.Task{
			TaskData: string(taskDataBytes),
			Status:   "pending",
		}
		if err := db.Create(task).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 给网站程序返回一个 task_id 作为接收数据的返回值
		c.JSON(201, gin.H{"task_id": strconv.FormatUint(task.TaskID, 10)})
	}
}

// 执行排课任务
func ExecuteTaskHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 URL 中获取 task_id
		taskID := c.Param("task_id")

		// 从排课任务队列中获取到该任务
		var task models.Task
		if err := db.Where("id = ? AND status = ?", taskID, "pending").First(&task).Error; err != nil {
			c.JSON(404, gin.H{"error": "task not found"})
			return
		}

		// 更新任务状态为 running
		if err := db.Model(&task).Update("status", "running").Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 执行排课
		scheduleResults, _, err := middlewares.ExecuteTask(task.TaskID, task.TaskData)
		if err != nil {
			// 执行排课失败，更新任务状态为 failed
			if err := db.Model(&task).Update("status", "failed").Error; err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			// 记录排课错误日志
			errorLog := &models.ScheduleErrorLog{
				TaskID:   task.TaskID,
				ErrorMsg: err.Error(),
			}
			if err := db.Create(errorLog).Error; err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		} else {
			// 执行排课成功，更新任务状态为 success
			if err := db.Model(&task).Update("status", "success").Error; err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			// 将排课结果写入排课结果数据表
			err := db.CreateInBatches(scheduleResults, len(scheduleResults)).Error
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(200, gin.H{"status": task.Status})
	}
}

// 查询排课结果
func GetTaskResultHandler(db *gorm.DB) gin.HandlerFunc {

	return func(c *gin.Context) {
		// 从 URL 中获取 task_id
		taskID := c.Param("task_id")

		// 查询排课任务
		var task models.Task
		if err := db.Where("id = ?", taskID).First(&task).Error; err != nil {
			c.JSON(404, gin.H{"error": "task not found"})
			return
		}

		// 如果任务状态是 success，则返回排课结果
		if task.Status == "success" {
			var result []models.ScheduleResult
			if err := db.Where("task_id = ?", taskID).Find(&result).Error; err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, result)
		} else {
			// 如果任务状态是 pending、running 或 failed，则排课结果为空
			c.JSON(200, []models.ScheduleResult{})
		}
	}
}
