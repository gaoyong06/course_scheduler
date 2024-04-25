// 相同节次的排课是否超过数量限制

package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/types"
)

var PCRule1 = &types.Rule{
	Name:     "PCRule1",
	Type:     "dynamic",
	Fn:       pcRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 相同节次的排课是否超过数量限制
func pcRule1Fn(classMatrix map[string]map[int]map[int]map[int]*types.Element, element types.ClassUnit) (bool, bool, error) {

	classSN := element.GetClassSN()
	teacherID := element.GetTeacherID()
	venueID := element.GetVenueID()
	timeSlot := element.GetTimeSlot()

	periodCount := countPeriodClasses(classMatrix, classSN, teacherID, venueID)
	period := timeSlot % constants.NUM_CLASSES

	preCheckPassed := false
	count := 0

	count, preCheckPassed = periodCount[period]

	shouldPenalize := false
	if preCheckPassed {

		// 检查相同节次的排课是否超过数量限制
		shouldPenalize = count > constants.PERIOD_THRESHOLD
	}
	return preCheckPassed, !shouldPenalize, nil
}

// countPeriodClasses 计算每个时间段的科目数量
func countPeriodClasses(classMatrix map[string]map[int]map[int]map[int]*types.Element, sn string, teacherID, venueID int) map[int]int {

	periodCount := make(map[int]int)

	for timeSlot, element := range classMatrix[sn][teacherID][venueID] {

		if element.Val.Used == 1 {
			period := timeSlot % constants.NUM_CLASSES
			periodCount[period]++
		}
	}

	return periodCount
}
