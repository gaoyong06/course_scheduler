// 教师节数限制
package constraint

import (
	"course_scheduler/config"
	"course_scheduler/internal/types"
)

var TCLRule1 = &types.Rule{
	Name:     "TCLRule1",
	Type:     "dynamic",
	Fn:       tclRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TCLRule2 = &types.Rule{
	Name:     "TCLRule2",
	Type:     "dynamic",
	Fn:       tclRule2Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TCLRule3 = &types.Rule{
	Name:     "TCLRule3",
	Type:     "dynamic",
	Fn:       tclRule3Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

var TCLRule4 = &types.Rule{
	Name:     "TCLRule4",
	Type:     "dynamic",
	Fn:       tclRule4Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 27. 王老师 上午第4节 最多3次
func tclRule1Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()
	period := timeSlot%config.NumClasses + 1
	preCheckPassed := teacherID == 1 && period == 4

	shouldPenalize := false
	if preCheckPassed {
		count := countTeacherClassInPeriod(teacherID, period, classMatrix)
		shouldPenalize = preCheckPassed && count > 3
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 28. 李老师 上午第4节 最多3次
func tclRule2Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()
	period := timeSlot%config.NumClasses + 1
	preCheckPassed := teacherID == 2 && period == 4

	shouldPenalize := false
	if preCheckPassed {
		count := countTeacherClassInPeriod(teacherID, period, classMatrix)
		shouldPenalize = preCheckPassed && count > 3
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 29. 刘老师 上午第4节 最多3次
func tclRule3Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()
	period := timeSlot%config.NumClasses + 1
	preCheckPassed := teacherID == 3 && period == 4

	shouldPenalize := false
	if preCheckPassed {
		count := countTeacherClassInPeriod(teacherID, period, classMatrix)
		shouldPenalize = preCheckPassed && count > 3
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 30. 张老师 上午第4节 最多3次
func tclRule4Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()
	period := timeSlot%config.NumClasses + 1
	preCheckPassed := teacherID == 4 && period == 4

	shouldPenalize := false
	if preCheckPassed {
		count := countTeacherClassInPeriod(teacherID, period, classMatrix)
		shouldPenalize = preCheckPassed && count > 3
	}

	return preCheckPassed, !shouldPenalize, nil
}

func countTeacherClassInPeriod(teacherID int, period int, classMatrix *types.ClassMatrix) int {
	count := 0
	for _, classMap := range classMatrix.Elements {
		for id, teacherMap := range classMap {
			if teacherID == id {
				for _, timeSlotMap := range teacherMap {
					if timeSlotMap == nil {
						continue
					}
					for timeSlot, element := range timeSlotMap {
						if element.Val.Used == 1 && timeSlot%config.NumClasses+1 == period {
							count++
						}
					}
				}
			}
		}
	}
	return count
}
