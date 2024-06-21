package models

import (
	"time"
)

type ScheduleResult struct {
	ResultID  uint64    `gorm:"primaryKey;autoIncrement;column:result_id" json:"result_id"`
	TaskID    uint64    `gorm:"type:bigint unsigned;not null;column:task_id" json:"task_id"`
	SubjectID uint64    `gorm:"type:bigint unsigned;not null;column:subject_id" json:"subject_id"`
	TeacherID uint64    `gorm:"type:bigint unsigned;not null;column:teacher_id" json:"teacher_id"`
	GradeID   uint64    `gorm:"type:bigint unsigned;not null;column:grade_id" json:"grade_id"`
	ClassID   uint64    `gorm:"type:bigint unsigned;not null;column:class_id" json:"class_id"`
	VenueID   uint64    `gorm:"type:bigint unsigned;not null;column:venue_id" json:"venue_id"`
	Weekday   int8      `gorm:"type:tinyint unsigned;not null;column:weekday" json:"weekday"`
	Period    int8      `gorm:"type:tinyint unsigned;not null;column:period" json:"period"`
	StartTime string    `gorm:"type:varchar(255);not null;column:start_time" json:"start_time"`
	EndTime   string    `gorm:"type:varchar(255);not null;column:end_time" json:"end_time"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`
}

func (ScheduleResult) TableName() string {
	return "schedule_result"
}
