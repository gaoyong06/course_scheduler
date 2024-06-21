package models

import (
	"time"
)

type ScheduleErrorLog struct {
	ErrorID   uint64    `gorm:"primaryKey;autoIncrement;column:error_id" json:"error_id"`
	TaskID    uint64    `gorm:"type:bigint unsigned;not null;column:task_id" json:"task_id"`
	ErrorMsg  string    `gorm:"type:text;not null;column:error_message" json:"error_message"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`
}

func (ScheduleErrorLog) TableName() string {
	return "schedule_error_log"
}
