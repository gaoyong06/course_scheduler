// subject.go
// 班级固排禁排

package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"math"

	"github.com/samber/lo"
)

// 14. 语数英 周一~周五 第1节 优先排
// 15. 语数英 周一~周五 第2节 优先排
// 16. 语数英 周一~周五 第3节 优先排
var SRule1 = &types.Rule{
	Name:     "SRule1",
	Type:     "fixed",
	Fn:       sRule1Fn,
	Score:    1,
	Penalty:  0,
	Weight:   1,
	Priority: 1,
}

// 副课 安排在第1,2,3节 扣分
var SRule2 = &types.Rule{
	Name:     "SRule2",
	Type:     "fixed",
	Fn:       sRule2Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 17. 主课 周一~周五 第8节 禁排
var SRule3 = &types.Rule{
	Name:     "SRule3",
	Type:     "fixed",
	Fn:       sRule3Fn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

// 18. 主课 周一~周五 第7节 尽量不排
var SRule4 = &types.Rule{
	Name:     "SRule4",
	Type:     "fixed",
	Fn:       sRule4Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 14. 语数英 周一~周五 第1节 优先排
// 15. 语数英 周一~周五 第2节 优先排
// 16. 语数英 周一~周五 第3节 优先排
func sRule1Fn(classMatrix map[string]map[int]map[int]map[int]*types.Element, element types.ClassUnit) (bool, bool, error) {

	classSN := element.GetClassSN()
	timeSlot := element.GetTimeSlot()

	SN, _ := types.ParseSN(classSN)
	subject, err := models.FindSubjectByID(SN.SubjectID)
	if err != nil {
		return false, false, err
	}
	day := timeSlot/constants.NUM_CLASSES + 1
	period := timeSlot%constants.NUM_CLASSES + 1
	preCheckPassed := (period == 1 || period == 2 || period == 3) && (day >= 1 && day <= 5)
	isValid := preCheckPassed && lo.Contains(subject.SubjectGroupIDs, 1)

	return preCheckPassed, isValid, nil
}

// 副课 安排在第1,2,3节 扣分
// 满足该条件扣分, 不满足该该条件, 不增加分数, 也不扣分
func sRule2Fn(classMatrix map[string]map[int]map[int]map[int]*types.Element, element types.ClassUnit) (bool, bool, error) {

	classSN := element.GetClassSN()
	timeSlot := element.GetTimeSlot()

	SN, _ := types.ParseSN(classSN)
	subject, err := models.FindSubjectByID(SN.SubjectID)
	if err != nil {
		return false, false, err
	}
	period := timeSlot%constants.NUM_CLASSES + 1

	preCheckPassed := period == 1 || period == 2 || period == 3

	shouldPenalize := preCheckPassed && lo.Contains(subject.SubjectGroupIDs, 3)
	return preCheckPassed, !shouldPenalize, nil
}

// 17. 主课 周一~周五 第8节 禁排
func sRule3Fn(classMatrix map[string]map[int]map[int]map[int]*types.Element, element types.ClassUnit) (bool, bool, error) {

	classSN := element.GetClassSN()
	timeSlot := element.GetTimeSlot()

	SN, _ := types.ParseSN(classSN)
	subject, err := models.FindSubjectByID(SN.SubjectID)
	if err != nil {
		return false, false, err
	}
	period := timeSlot%constants.NUM_CLASSES + 1
	preCheckPassed := period == 8

	shouldPenalize := preCheckPassed && lo.Contains(subject.SubjectGroupIDs, 2)
	return preCheckPassed, !shouldPenalize, nil
}

// 18. 主课 周一~周五 第7节 尽量不排
func sRule4Fn(classMatrix map[string]map[int]map[int]map[int]*types.Element, element types.ClassUnit) (bool, bool, error) {

	classSN := element.GetClassSN()
	timeSlot := element.GetTimeSlot()

	SN, _ := types.ParseSN(classSN)
	subject, err := models.FindSubjectByID(SN.SubjectID)
	if err != nil {
		return false, false, err
	}
	period := timeSlot%constants.NUM_CLASSES + 1
	preCheckPassed := period == 7

	shouldPenalize := preCheckPassed && lo.Contains(subject.SubjectGroupIDs, 2)
	return preCheckPassed, !shouldPenalize, nil
}
