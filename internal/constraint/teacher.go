// teacher.go
// 教师固排禁排

package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"math"
)

var TRule1 = &types.Rule{
	Name:     "TRule1",
	Type:     "fixed",
	Fn:       tRule1Fn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

var TRule2 = &types.Rule{
	Name:     "TRule2",
	Type:     "fixed",
	Fn:       tRule2Fn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

var TRule3 = &types.Rule{
	Name:     "TRule3",
	Type:     "fixed",
	Fn:       tRule3Fn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

var TRule4 = &types.Rule{
	Name:     "TRule4",
	Type:     "fixed",
	Fn:       tRule4Fn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

var TRule5 = &types.Rule{
	Name:     "TRule5",
	Type:     "fixed",
	Fn:       tRule5Fn,
	Score:    2,
	Penalty:  0,
	Weight:   1,
	Priority: 1,
}

// 9. 数学组 周一 第4节 禁排 教研会
func tRule1Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()

	teacher, err := models.FindTeacherByID(teacherID)
	if err != nil {
		return false, false, err
	}
	day := timeSlot/constants.NUM_CLASSES + 1
	period := timeSlot%constants.NUM_CLASSES + 1

	preCheckPassed := day == 1 && period == 4
	shouldPenalize := preCheckPassed && teacher.TeacherGroupIDs[0] == 2
	return preCheckPassed, !shouldPenalize, nil
}

// 10. 刘老师 周一 第4节 禁排 教研会
func tRule2Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()

	teacher, err := models.FindTeacherByID(teacherID)
	if err != nil {
		return false, false, err
	}
	day := timeSlot/constants.NUM_CLASSES + 1
	period := timeSlot%constants.NUM_CLASSES + 1

	preCheckPassed := day == 1 && period == 4
	shouldPenalize := preCheckPassed && teacher.TeacherID == 3
	return preCheckPassed, !shouldPenalize, nil
}

// 11. 行政领导 周二 第7节 禁排 例会
func tRule3Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()

	teacher, err := models.FindTeacherByID(teacherID)
	if err != nil {
		return false, false, err
	}
	day := timeSlot/constants.NUM_CLASSES + 1
	period := timeSlot%constants.NUM_CLASSES + 1

	preCheckPassed := day == 2 && period == 7
	shouldPenalize := preCheckPassed && teacher.TeacherGroupIDs[0] == 3
	return preCheckPassed, !shouldPenalize, nil
}

// 12. 马老师 周二 第7节 禁排 例会
func tRule4Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()

	teacher, err := models.FindTeacherByID(teacherID)
	if err != nil {
		return false, false, err
	}
	day := timeSlot/constants.NUM_CLASSES + 1
	period := timeSlot%constants.NUM_CLASSES + 1

	preCheckPassed := day == 2 && period == 7
	shouldPenalize := preCheckPassed && teacher.TeacherID == 5
	return preCheckPassed, !shouldPenalize, nil
}

// 13. 王老师 周二 第2节 固排
func tRule5Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	teacherID := element.GetTeacherID()
	timeSlot := element.GetTimeSlot()

	teacher, err := models.FindTeacherByID(teacherID)
	if err != nil {
		return false, false, err
	}
	day := timeSlot/constants.NUM_CLASSES + 1
	period := timeSlot%constants.NUM_CLASSES + 1

	preCheckPassed := day == 2 && period == 2
	isValid := preCheckPassed && teacher.TeacherID == 1

	return preCheckPassed, isValid, nil
}
