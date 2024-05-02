// time_table.go
package types

import (
	"course_scheduler/config"
)

// 时间表
type TimeTable struct {
	TimeSlots []int        // 所有可用的时间段
	Used      map[int]bool // 已经被占用的时间段
}

// 初始化时间表
func initTimeTable() *TimeTable {

	var timeSlots []int
	used := make(map[int]bool)
	for i := 0; i < config.NumDays; i++ {
		for j := 0; j < config.NumClasses; j++ {
			timeSlot := i*config.NumClasses + j
			used[timeSlot] = false
			timeSlots = append(timeSlots, timeSlot)
		}
	}

	timeTable := &TimeTable{
		TimeSlots: timeSlots,
		Used:      used,
	}

	return timeTable
}
