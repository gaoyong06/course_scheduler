// 同一个年级,班级,科目相同节次的排课是否超过数量限制
// 系统约束

package constraints

import (
	"course_scheduler/config"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
)

// #### 同一个年级,班级,科目相同节次的排课是否超过数量限制
// 同一个年级,班级,科目相同节次最多排课 {2}次
var subjectPeriodLimitRule = &types.Rule{
	Name:     "subjectPeriodLimit",
	Type:     "dynamic",
	Fn:       splRuleFn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 相同节次的排课是否超过数量限制
func splRuleFn(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	classSN := element.GetClassSN()
	teacherID := element.GetTeacherID()
	venueID := element.GetVenueID()
	timeSlot := element.GetTimeSlot()

	periodCount := countPeriodClasses(classMatrix, classSN, teacherID, venueID, schedule)
	period := timeSlot % totalClassesPerDay

	preCheckPassed := false
	count := 0

	count, preCheckPassed = periodCount[period]

	shouldPenalize := false
	if preCheckPassed {

		// 检查相同节次的排课是否超过数量限制
		shouldPenalize = count > config.SubjectPeriodLimitThreshold
	}
	return preCheckPassed, !shouldPenalize, nil
}

// countPeriodClasses 计算每个时间段的科目数量
func countPeriodClasses(classMatrix *types.ClassMatrix, sn string, teacherID, venueID int, schedule *models.Schedule) map[int]int {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// key: 节次, val: 数量
	periodCount := make(map[int]int)

	for timeSlot, element := range classMatrix.Elements[sn][teacherID][venueID] {

		if element.Val.Used == 1 {
			period := timeSlot % totalClassesPerDay
			periodCount[period]++
		}
	}

	return periodCount
}
