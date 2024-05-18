// 同一个年级,班级,科目,在同一天的排课是否超过数量限制
// 系统约束

package constraints

import (
	"course_scheduler/config"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"math"
)

// #### 同一个年级,班级,科目,在同一天的排课是否超过数量限制
// 同一个年级,班级,科目,在同一天的排课最多排课 {2}次
var subjectDayLimitRule = &types.Rule{
	Name:     "subjectDayLimit",
	Type:     "dynamic",
	Fn:       sdlRuleFn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

// 相同节次的排课是否超过数量限制
func sdlRuleFn(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	classSN := element.GetClassSN()
	teacherID := element.GetTeacherID()
	venueID := element.GetVenueID()
	timeSlot := element.GetTimeSlot()

	dayCount := countDayClasses(classMatrix, classSN, teacherID, venueID, schedule)
	day := timeSlot / totalClassesPerDay

	preCheckPassed := false
	count := 0

	count, preCheckPassed = dayCount[day]

	shouldPenalize := false
	if preCheckPassed {

		// 检查同一天的排课是否超过数量限制
		shouldPenalize = count > config.SubjectDayLimitThreshold
	}
	return preCheckPassed, !shouldPenalize, nil
}

// countDayClasses 计算每天的科目数量
func countDayClasses(classMatrix *types.ClassMatrix, sn string, teacherID, venueID int, schedule *models.Schedule) map[int]int {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// key: 天, val: 数量
	dayCount := make(map[int]int)

	for timeSlot, element := range classMatrix.Elements[sn][teacherID][venueID] {

		if element.Val.Used == 1 {
			day := timeSlot / totalClassesPerDay
			dayCount[day]++
		}
	}

	return dayCount
}
