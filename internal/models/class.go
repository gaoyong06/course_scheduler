// class.go
package models

type Class struct {
	SchoolID int    `json:"school_id"`
	ClassID  int    `json:"class_id"`
	Name     string `json:"name"`
}
