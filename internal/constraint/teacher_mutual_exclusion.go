// 教师互斥限制
// teacher_mutual_exclusion.go
package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/types"
)

var TMERule1 = &types.Rule{
	Name:     "TMERule1",
	Type:     "dynamic",
	Fn:       tmeRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TMERule2 = &types.Rule{
	Name:     "TMERule2",
	Type:     "dynamic",
	Fn:       tmeRule2Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 31. 王老师(语文) 马老师(美术)
func tmeRule1Fn(classMatrix map[string]map[int]map[int]map[int]*types.Element, element types.ClassUnit) (bool, bool, error) {
	teacherID := element.GetTeacherID()
	preCheckPassed := teacherID == 1 || teacherID == 5

	shouldPenalize := false
	if preCheckPassed {
		shouldPenalize = isTeacherSameDay(1, 5, classMatrix, element)
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 32. 李老师(数学) 黄老师(体育)
func tmeRule2Fn(classMatrix map[string]map[int]map[int]map[int]*types.Element, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	preCheckPassed := teacherID == 2 || teacherID == 6

	shouldPenalize := false
	if preCheckPassed {
		shouldPenalize = isTeacherSameDay(2, 6, classMatrix, element)
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 判断教师A,教师B是否同一天都有课
func isTeacherSameDay(teacherAID, teacherBID int, classMatrix map[string]map[int]map[int]map[int]*types.Element, element types.ClassUnit) bool {

	teacher1Days := make(map[int]bool)
	teacher2Days := make(map[int]bool)
	timeSlot := element.GetTimeSlot()

	elementDay := timeSlot / constants.NUM_CLASSES

	for _, classMap := range classMatrix {
		for id, teacherMap := range classMap {
			if id == teacherAID {
				for _, timeSlotMap := range teacherMap {
					for timeSlot, element := range timeSlotMap {
						if element.Val.Used == 1 {
							day := timeSlot / constants.NUM_CLASSES
							teacher1Days[day] = true // 将时间段转换为天数
						}
					}
				}
			} else if id == teacherBID {
				for _, timeSlotMap := range teacherMap {
					for timeSlot, element := range timeSlotMap {
						if element.Val.Used == 1 {
							day := timeSlot / constants.NUM_CLASSES
							teacher2Days[day] = true // 将时间段转换为天数
						}
					}
				}
			}
		}
	}

	if teacher1Days[elementDay] && teacher2Days[elementDay] {
		return true
	}
	return false
}
