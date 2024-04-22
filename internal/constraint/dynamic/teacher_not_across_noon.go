// 教师不跨中午(教师排了上午最后一节就不排下午第一节)

package constraint

import (
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/types"
)

var TNANRule1 = &constraint.Rule{
	Name:     "TNANRule1",
	Type:     "fixed",
	Fn:       tnanRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TNANRule2 = &constraint.Rule{
	Name:     "TNANRule2",
	Type:     "fixed",
	Fn:       tnanRule2Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 33. 王老师
func tnanRule1Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element constraint.Element) (bool, bool, error) {

	teacherID := element.TeacherID
	preCheckPassed := teacherID == 1

	shouldPenalize := false
	if preCheckPassed {
		shouldPenalize = isTeacherInBothPeriods(1, 4, 5, classMatrix)
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 34. 李老师
func tnanRule2Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element constraint.Element) (bool, bool, error) {
	teacherID := element.TeacherID
	preCheckPassed := teacherID == 2

	shouldPenalize := false
	if preCheckPassed {
		shouldPenalize = isTeacherInBothPeriods(2, 4, 5, classMatrix)
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 判断教师是否在两个节次都有课
func isTeacherInBothPeriods(teacherID int, period1, period2 int, classMatrix map[string]map[int]map[int]map[int]types.Val) bool {
	for _, classMap := range classMatrix {
		for id, teacherMap := range classMap {
			if id == teacherID {
				for _, periodMap := range teacherMap {
					if val1, ok := periodMap[period1]; ok && val1.Used == 1 {
						if val2, ok := periodMap[period2]; ok && val2.Used == 1 {
							return true
						}
					}
				}
			}
		}
	}
	return false
}
