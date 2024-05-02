// teacher_time_limit.go
// 教师固排禁排

package constraint

import (
	"course_scheduler/config"
	"course_scheduler/internal/types"
)

var TTLRule1 = &types.Rule{

	Name:     "TTLRule1",
	Type:     "dynamic",
	Fn:       ttlRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TTLRule2 = &types.Rule{

	Name:     "TTLRule2",
	Type:     "dynamic",
	Fn:       ttlRule2Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TTLRule3 = &types.Rule{

	Name:     "TTLRule3",
	Type:     "dynamic",
	Fn:       ttlRule3Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 23. 王老师 上午 最多1节
func ttlRule1Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()
	period := timeSlot%config.NumClasses + 1
	count := countTeacherClassesInRange(1, 1, 4, classMatrix)

	preCheckPassed := teacherID == 1 && period >= 1 && period <= 4
	shouldPenalize := preCheckPassed && count > 1
	return preCheckPassed, !shouldPenalize, nil
}

// 24. 王老师 下午 最多2节
func ttlRule2Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()
	period := timeSlot%config.NumClasses + 1
	count := countTeacherClassesInRange(1, 5, 8, classMatrix)

	preCheckPassed := teacherID == 1 && period >= 5 && period <= 8
	shouldPenalize := preCheckPassed && count > 2
	return preCheckPassed, !shouldPenalize, nil
}

// 25. 王老师 全天(不含晚自习) 最多3节
func ttlRule3Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()
	period := timeSlot%config.NumClasses + 1
	count := countTeacherClassesInRange(1, 1, 8, classMatrix)

	preCheckPassed := teacherID == 1 && period >= 1 && period <= 8
	shouldPenalize := preCheckPassed && count > 3

	return preCheckPassed, !shouldPenalize, nil
}

// 26. 王老师 晚自习 最多1节
func countTeacherClassesInRange(teacherID int, startPeriod, endPeriod int, classMatrix *types.ClassMatrix) int {

	count := 0
	for _, classMap := range classMatrix.Elements {
		for id, teacherMap := range classMap {
			if teacherID == id {
				for _, timeSlotMap := range teacherMap {
					if timeSlotMap == nil {
						continue
					}
					for timeSlot, element := range timeSlotMap {
						if element.Val.Used == 1 && timeSlot >= startPeriod && timeSlot <= endPeriod {
							count++
						}
					}
				}
			}
		}
	}
	return count
}
