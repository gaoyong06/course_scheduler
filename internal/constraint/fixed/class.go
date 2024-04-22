// class.go
// 班级固排禁排

package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/types"
	"math"
)

// 1. 一年级(1)班 语文 王老师 第1节 固排
var CRule1 = &constraint.Rule{
	Name:     "CRule1",
	Type:     "fixed",
	Fn:       cRule1Fn,
	Score:    2,
	Penalty:  0,
	Weight:   1,
	Priority: 1,
}

// 2. 三年级(1)班 第7节 禁排 班会
var CRule2 = &constraint.Rule{
	Name:     "CRule2",
	Type:     "fixed",
	Fn:       cRule2Fn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

// 3. 三年级(2)班 第8节 禁排 班会
var CRule3 = &constraint.Rule{
	Name:     "CRule3",
	Type:     "fixed",
	Fn:       cRule3Fn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

// 4. 四年级 第8节 禁排 班会
var CRule4 = &constraint.Rule{
	Name:     "CRule4",
	Type:     "fixed",
	Fn:       cRule4Fn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

// 5. 四年级(1)班 语文 王老师 第1节 禁排
var CRule5 = &constraint.Rule{
	Name:     "CRule5",
	Type:     "fixed",
	Fn:       cRule5Fn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

// 6. 五年级 数学 李老师 第2节 固排
var CRule6 = &constraint.Rule{
	Name:     "CRule6",
	Type:     "fixed",
	Fn:       cRule6Fn,
	Score:    2,
	Penalty:  0,
	Weight:   1,
	Priority: 1,
}

// 7. 五年级 数学 李老师 第3节 尽量排
var CRule7 = &constraint.Rule{
	Name:     "CRule7",
	Type:     "fixed",
	Fn:       cRule7Fn,
	Score:    1,
	Penalty:  0,
	Weight:   1,
	Priority: 1,
}

// 8. 五年级 数学 李老师 第5节 固排
var CRule8 = &constraint.Rule{
	Name:     "CRule8",
	Type:     "fixed",
	Fn:       cRule8Fn,
	Score:    2,
	Penalty:  0,
	Weight:   1,
	Priority: 1,
}

// 1. 一年级(1)班 语文 王老师 第1节 固排
func cRule1Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element constraint.Element) (bool, bool, error) {

	SN, _ := types.ParseSN(element.ClassSN)
	period := element.TimeSlot%constants.NUM_CLASSES + 1
	preCheckPassed := SN.GradeID == 1 && SN.ClassID == 1 && period == 1
	isValid := preCheckPassed && SN.SubjectID == 1 && element.TeacherID == 1

	return preCheckPassed, isValid, nil
}

// 2. 三年级(1)班 第7节 禁排 班会
func cRule2Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element constraint.Element) (bool, bool, error) {

	SN, _ := types.ParseSN(element.ClassSN)
	period := element.TimeSlot%constants.NUM_CLASSES + 1
	preCheckPassed := SN.GradeID == 3 && SN.ClassID == 1 && period == 7

	isValid := !preCheckPassed
	return preCheckPassed, isValid, nil
}

// 3. 三年级(2)班 第8节 禁排 班会
func cRule3Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element constraint.Element) (bool, bool, error) {

	SN, _ := types.ParseSN(element.ClassSN)
	period := element.TimeSlot%constants.NUM_CLASSES + 1
	preCheckPassed := SN.GradeID == 3 && SN.ClassID == 2 && period == 8

	isValid := !preCheckPassed

	return preCheckPassed, isValid, nil
}

// 4. 四年级 第8节 禁排 班会
func cRule4Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element constraint.Element) (bool, bool, error) {

	SN, _ := types.ParseSN(element.ClassSN)
	period := element.TimeSlot%constants.NUM_CLASSES + 1
	preCheckPassed := SN.GradeID == 4 && period == 8

	isValid := !preCheckPassed

	return preCheckPassed, isValid, nil
}

// 5. 四年级(1)班 语文 王老师 第1节 禁排
func cRule5Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element constraint.Element) (bool, bool, error) {
	SN, _ := types.ParseSN(element.ClassSN)
	period := element.TimeSlot%constants.NUM_CLASSES + 1

	preCheckPassed := SN.GradeID == 4 && SN.SubjectID == 1 && period == 1

	shouldPenalize := preCheckPassed && element.TeacherID == 1
	return preCheckPassed, !shouldPenalize, nil
}

// 6. 五年级 数学 李老师 第2节 固排
func cRule6Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element constraint.Element) (bool, bool, error) {
	SN, _ := types.ParseSN(element.ClassSN)
	period := element.TimeSlot%constants.NUM_CLASSES + 1
	preCheckPassed := SN.GradeID == 5 && period == 2
	isValid := preCheckPassed && SN.SubjectID == 2 && element.TeacherID == 1

	return preCheckPassed, isValid, nil
}

// 7. 五年级 数学 李老师 第3节 尽量排
func cRule7Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element constraint.Element) (bool, bool, error) {
	SN, _ := types.ParseSN(element.ClassSN)
	period := element.TimeSlot%constants.NUM_CLASSES + 1
	preCheckPassed := SN.GradeID == 5 && period == 3
	isValid := preCheckPassed && SN.SubjectID == 2 && element.TeacherID == 2

	return preCheckPassed, isValid, nil
}

// 8. 五年级 数学 李老师 第5节 固排
func cRule8Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element constraint.Element) (bool, bool, error) {
	SN, _ := types.ParseSN(element.ClassSN)
	period := element.TimeSlot%constants.NUM_CLASSES + 1
	preCheckPassed := SN.GradeID == 5 && period == 5
	isValid := preCheckPassed && SN.SubjectID == 2 && element.TeacherID == 2
	return preCheckPassed, isValid, nil
}
