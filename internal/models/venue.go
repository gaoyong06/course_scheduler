package models

import (
	"fmt"
)

// 教学场地
type Venue struct {
	Name     string `json:"name" mapstructure:"name"`         // 场地名称
	Type     string `json:"type" mapstructure:"type"`         // 场地类型 exclusive: 专用教学场所, shared: 共享教学场所
	Capacity int    `json:"capacity" mapstructure:"capacity"` // 教学场所能容纳的, 最多上课班级, 专用教学场所: 为固定值1, 共享教学场所: 默认值为 0,表示不限制
}

// 教室集合
// 根据课班选取教室
// subjectVenueMap key: subjectID_gradeID_classID value: venueIDs
func ClassVenueIDs(gradeID, classID, subjectID int, subjectVenueMap map[string][]int) []int {

	sn := fmt.Sprintf("%d_%d_%d", subjectID, gradeID, classID)
	venueIDs, ok := subjectVenueMap[sn]

	// 如果为设置教学场地,则每个班级默认指定一个
	if !ok || len(venueIDs) == 0 {
		defaultVenueID := gradeID*100 + classID
		venueIDs = []int{defaultVenueID}
	}
	return venueIDs
}

// 教室不可排课的时间范围
func venueUnavailableSlots(venue int) []int {
	var timeSlots []int
	return timeSlots
}

// 判断venueID是否合法
func IsVenueIDValid(venueID int) bool {
	return venueID > 0
}
