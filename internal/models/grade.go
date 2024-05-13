// grade.go
package models

type Grade struct {
	SchoolID int     `json:"school_id"`
	GradeID  int     `json:"grade_id"`
	Name     string  `json:"name"`
	Classes  []Class `json:"classes"`
}
