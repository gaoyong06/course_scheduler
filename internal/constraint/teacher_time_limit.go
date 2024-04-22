// teacher_time_limit.go
// 教师固排禁排

package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/types"
)

var TTLRule1 = &Rule{

	Name:     "TTLRule1",
	Type:     "dynamic",
	Fn:       ttlRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TTLRule2 = &Rule{

	Name:     "TTLRule2",
	Type:     "dynamic",
	Fn:       ttlRule2Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TTLRule3 = &Rule{

	Name:     "TTLRule3",
	Type:     "dynamic",
	Fn:       ttlRule3Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 23. 王老师 上午 最多1节
func ttlRule1Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element Element) (bool, bool, error) {

	teacherID := element.TeacherID
	period := element.TimeSlot%constants.NUM_CLASSES + 1
	count := countTeacherClassesInRange(1, 1, 4, classMatrix)

	preCheckPassed := teacherID == 1 && period >= 1 && period <= 4
	shouldPenalize := preCheckPassed && count > 1
	return preCheckPassed, !shouldPenalize, nil
}

// 24. 王老师 下午 最多2节
func ttlRule2Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element Element) (bool, bool, error) {

	teacherID := element.TeacherID
	period := element.TimeSlot%constants.NUM_CLASSES + 1
	count := countTeacherClassesInRange(1, 5, 8, classMatrix)

	preCheckPassed := teacherID == 1 && period >= 5 && period <= 8
	shouldPenalize := preCheckPassed && count > 2
	return preCheckPassed, !shouldPenalize, nil
}

// 25. 王老师 全天(不含晚自习) 最多3节
func ttlRule3Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element Element) (bool, bool, error) {

	teacherID := element.TeacherID
	period := element.TimeSlot%constants.NUM_CLASSES + 1
	count := countTeacherClassesInRange(1, 1, 8, classMatrix)

	preCheckPassed := teacherID == 1 && period >= 1 && period <= 8
	shouldPenalize := preCheckPassed && count > 3

	return preCheckPassed, !shouldPenalize, nil
}

// 26. 王老师 晚自习 最多1节
func countTeacherClassesInRange(teacherID int, startPeriod, endPeriod int, classMatrix map[string]map[int]map[int]map[int]types.Val) int {

	count := 0
	for _, classMap := range classMatrix {
		for id, teacherMap := range classMap {
			if teacherID == id {
				for _, timeSlotMap := range teacherMap {
					if timeSlotMap == nil {
						continue
					}
					for timeSlot, val := range timeSlotMap {
						if val.Used == 1 && timeSlot >= startPeriod && timeSlot <= endPeriod {
							count++
						}
					}
				}
			}
		}
	}
	return count
}
