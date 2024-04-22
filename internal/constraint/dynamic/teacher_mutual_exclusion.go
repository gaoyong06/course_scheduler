// 教师互斥限制

package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/types"
)

var TMERule1 = &constraint.Rule{
	Name:     "TMERule1",
	Type:     "fixed",
	Fn:       tmeRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TMERule2 = &constraint.Rule{
	Name:     "TMERule2",
	Type:     "fixed",
	Fn:       tmeRule2Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 31. 王老师 马老师
func tmeRule1Fn(element constraint.Element) (bool, bool, error) {
	teacherID := element.TeacherID
	preCheckPassed := teacherID == 1 || teacherID == 5

	shouldPenalize := false
	if preCheckPassed {
		shouldPenalize = isTeacherSameDay(1, 5, element.ClassMatrix)
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 32. 李老师 黄老师
func tmeRule2Fn(element constraint.Element) (bool, bool, error) {
	teacherID := element.TeacherID
	preCheckPassed := teacherID == 2 || teacherID == 6

	shouldPenalize := false
	if preCheckPassed {
		shouldPenalize = isTeacherSameDay(2, 6, element.ClassMatrix)
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 判断教师A,教师B是否同一天都有课
func isTeacherSameDay(teacherAID, teacherBID int, classMatrix map[string]map[int]map[int]map[int]types.Val) bool {
	teacher1Days := make(map[int]bool)
	teacher2Days := make(map[int]bool)
	for _, classMap := range classMatrix {
		for id, teacherMap := range classMap {
			if id == teacherAID {
				for _, timeSlotMap := range teacherMap {
					for timeSlot, val := range timeSlotMap {
						if val.Used == 1 {
							teacher1Days[timeSlot/constants.NUM_CLASSES] = true // 将时间段转换为天数
						}
					}
				}
			} else if id == teacherBID {
				for _, timeSlotMap := range teacherMap {
					for timeSlot, val := range timeSlotMap {
						if val.Used == 1 {
							teacher2Days[timeSlot/constants.NUM_CLASSES] = true // 将时间段转换为天数
						}
					}
				}
			}
		}
	}
	for day := 0; day < constants.NUM_DAYS; day++ {
		if teacher1Days[day] && teacher2Days[day] {
			return true
		}
	}
	return false
}
