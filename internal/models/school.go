// school.go
package models

type School struct {
	SchoolID int     `json:"school_id"`
	Name     string  `json:"name"`
	Grades   []Grade `json:"grades"`
}
