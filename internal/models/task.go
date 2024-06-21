// task.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	TaskID    uint64    `gorm:"primaryKey;autoIncrement;column:task_id" json:"task_id"`
	TaskData  string    `gorm:"type:text;not null;column:task_data" json:"task_data"`
	Status    string    `gorm:"type:enum('pending','running','success','failed');not null;column:status" json:"status"`
	Progress  int8      `gorm:"type:tinyint unsigned;not null;default:0;column:progress" json:"progress"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`
}

func (Task) TableName() string {
	return "task"
}

// 高并发情况下，创建任务的频率很难达到每纳秒一个。因此，使用时间戳生成 ID 在实际应用中是可行
func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	t.TaskID = uint64(time.Now().UnixNano())
	return
}
