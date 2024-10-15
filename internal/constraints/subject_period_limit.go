// 同一个年级,班级,科目相同节次的排课是否超过数量限制
// 系统约束

package constraints

import (
	"course_scheduler/config"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
)

// #### 同一个年级,班级,科目相同节次的排课是否超过数量限制
// 同一个年级,班级,科目相同节次最多排课 {2}次
var subjectPeriodLimitRule = &types.Rule{
	Name:     "subjectPeriodLimit",
	Type:     "dynamic",
	Fn:       splRuleFn,
	Score:    0,
	Penalty:  2,
	Weight:   1,
	Priority: 1,
}

// 相同节次的排课是否超过数量限制
func splRuleFn(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, teachingTasks []*models.TeachingTask) (bool, bool, error) {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	classSN := element.GetClassSN()
	teacherID := element.GetTeacherID()
	venueID := element.GetVenueID()
	timeSlots := element.GetTimeSlots()

	periodCount := countPeriodClasses(classMatrix, classSN, teacherID, venueID, schedule)

	shouldPenalize := false
	preCheckPassed := false
	count := 0

	for _, timeSlot := range timeSlots {

		period := timeSlot % totalClassesPerDay
		count, preCheckPassed = periodCount[period]

		if preCheckPassed {
			// 检查相同节次的排课是否超过数量限制
			shouldPenalize = count > config.SubjectPeriodLimitThreshold
			if shouldPenalize {
				return true, false, nil
			}
		}
	}

	return preCheckPassed, true, nil
}

// countPeriodClasses 计算每个时间段的科目数量
func countPeriodClasses(classMatrix *types.ClassMatrix, sn string, teacherID, venueID int, schedule *models.Schedule) map[int]int {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// key: 节次, val: 数量
	periodCount := make(map[int]int)

	for timeSlotStr, element := range classMatrix.Elements[sn][teacherID][venueID] {

		timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
		for _, timeSlot := range timeSlots {
			if element.Val.Used == 1 {
				period := timeSlot % totalClassesPerDay
				periodCount[period]++
			}
		}
	}
	return periodCount
}
