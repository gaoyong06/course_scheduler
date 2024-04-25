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
func pcRule1Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element types.ClassUnit) (bool, bool, error) {

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
func countPeriodClasses(classMatrix map[string]map[int]map[int]map[int]types.Val, sn string, teacherID, venueID int) map[int]int {

	periodCount := make(map[int]int)

	for timeSlot, val := range classMatrix[sn][teacherID][venueID] {

		// 这是不能使用val.Used==1来做判断
		// 因为val.Used的赋值是在AllocateClassMatrix阶段执行的
		// 此时还没有执行到AllocateClassMatrix
		// 所以,此时val.Used都是0
		if val.Used == 1 {
			period := timeSlot % constants.NUM_CLASSES
			periodCount[period]++
		}
	}

	return periodCount
}
