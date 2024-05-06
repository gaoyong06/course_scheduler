// grade.go
package models

type Grade struct {
	GradeID int     `json:"grade_id"`
	Name    string  `json:"name"`
	Classes []Class `json:"classes"`
}
