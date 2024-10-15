// class_subject.go
package models

// 老师教授科目信息
type ClassSubject struct {
	GradeID    int   `json:"grade_id" mapstructure:"grade_id"`     // 年级
	ClassID    int   `json:"class_id" mapstructure:"class_id"`     // 班级
	SubjectIDs []int `json:"subject_id" mapstructure:"subject_id"` // 科目id数组
}
