// 教师不跨中午(教师排了上午最后一节就不排下午第一节)

package constraint

import (
	"course_scheduler/internal/types"
)

var TNANRule1 = &types.Rule{
	Name:     "TNANRule1",
	Type:     "dynamic",
	Fn:       tnanRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TNANRule2 = &types.Rule{
	Name:     "TNANRule2",
	Type:     "dynamic",
	Fn:       tnanRule2Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 33. 王老师
func tnanRule1Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	preCheckPassed := teacherID == 1

	shouldPenalize := false
	if preCheckPassed {
		shouldPenalize = isTeacherInBothPeriods(1, 4, 5, classMatrix)
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 34. 李老师
func tnanRule2Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {
	teacherID := element.GetTeacherID()
	preCheckPassed := teacherID == 2

	shouldPenalize := false
	if preCheckPassed {
		shouldPenalize = isTeacherInBothPeriods(2, 4, 5, classMatrix)
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 判断教师是否在两个节次都有课
func isTeacherInBothPeriods(teacherID int, period1, period2 int, classMatrix *types.ClassMatrix) bool {
	for _, classMap := range classMatrix.Elements {
		for id, teacherMap := range classMap {
			if id == teacherID {
				for _, periodMap := range teacherMap {
					if element1, ok := periodMap[period1]; ok && element1.Val.Used == 1 {
						if element2, ok := periodMap[period2]; ok && element2.Val.Used == 1 {
							return true
						}
					}
				}
			}
		}
	}
	return false
}
