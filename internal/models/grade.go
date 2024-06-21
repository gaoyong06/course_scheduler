// grade.go
package models

type Grade struct {
	SchoolID int     `json:"school_id" mapstructure:"school_id"`
	GradeID  int     `json:"grade_id" mapstructure:"grade_id"`
	Name     string  `json:"name" mapstructure:"name"`
	Classes  []Class `json:"classes" mapstructure:"classes"`
}
