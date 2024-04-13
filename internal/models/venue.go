package models

import "course_scheduler/internal/types"

func GetVenueIDs() []int {

	// 教室id列表
	venueIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	return venueIDs
}

// 教室集合
// 根据课班选取教室
func ClassVenueIDs(SN *types.SN) []int {

	venueIDs := GetVenueIDs()
	classID := SN.ClassID
	venueID := venueIDs[classID]

	var ids []int
	ids = append(ids, venueID)
	return ids
}

// 教室不可排课的时间范围
func venueUnavailableSlots(venue int) []int {
	var timeSlots []int
	return timeSlots
}
