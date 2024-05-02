// subject.go
// 班级固排禁排

package constraint

import (
	"course_scheduler/config"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"

	"github.com/samber/lo"
)

// 14. 语数英 周一~周五 第1节 优先排
// 15. 语数英 周一~周五 第2节 优先排
// 16. 语数英 周一~周五 第3节 优先排
var SRule1 = &types.Rule{
	Name:     "SRule1",
	Type:     "fixed",
	Fn:       sRule1Fn,
	Score:    2,
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
	Penalty:  2,
	Weight:   1,
	Priority: 2,
}

// 17. 主课 周一~周五 第8节 禁排
var SRule3 = &types.Rule{
	Name:     "SRule3",
	Type:     "fixed",
	Fn:       sRule3Fn,
	Score:    0,
	Penalty:  config.MaxPenaltyScore,
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
func sRule1Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	subjectGroupID := 1
	classSN := element.GetClassSN()
	timeSlot := element.GetTimeSlot()

	SN, _ := types.ParseSN(classSN)
	subject, err := models.FindSubjectByID(SN.SubjectID)
	if err != nil {
		return false, false, err
	}
	day := timeSlot/config.NumClasses + 1
	period := timeSlot%config.NumClasses + 1

	// 判断subjectGroupID是否已经排课完成
	isSubjectGroupScheduled, err := isSubjectGroupScheduled(classMatrix, subjectGroupID)
	if err != nil {
		return false, false, err
	}
	preCheckPassed := (period == 1 || period == 2 || period == 3) && (day >= 1 && day <= 5)

	// FindAvailableSubjectsByGroupID
	shouldPenalize := preCheckPassed && !lo.Contains(subject.SubjectGroupIDs, subjectGroupID) && !isSubjectGroupScheduled

	// fmt.Printf("sRule1Fn sn: %s, timeSlot: %d, subjectGroupIDs: %d\n", classSN, timeSlot, subject.SubjectGroupIDs)
	return preCheckPassed, !shouldPenalize, nil
}

// 副课 安排在第1,2,3节 扣分
// 满足该条件扣分, 不满足该该条件, 不增加分数, 也不扣分
func sRule2Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	classSN := element.GetClassSN()
	timeSlot := element.GetTimeSlot()

	SN, _ := types.ParseSN(classSN)
	subject, err := models.FindSubjectByID(SN.SubjectID)
	if err != nil {
		return false, false, err
	}
	period := timeSlot%config.NumClasses + 1

	preCheckPassed := period == 1 || period == 2 || period == 3

	shouldPenalize := preCheckPassed && lo.Contains(subject.SubjectGroupIDs, 3)
	return preCheckPassed, !shouldPenalize, nil
}

// 17. 主课 周一~周五 第8节 禁排
func sRule3Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	classSN := element.GetClassSN()
	timeSlot := element.GetTimeSlot()

	SN, _ := types.ParseSN(classSN)
	subject, err := models.FindSubjectByID(SN.SubjectID)
	if err != nil {
		return false, false, err
	}
	period := timeSlot%config.NumClasses + 1
	preCheckPassed := period == 8

	shouldPenalize := preCheckPassed && lo.Contains(subject.SubjectGroupIDs, 2)
	// fmt.Printf("sRule3Fn sn: %s, timeSlot: %d, shouldPenalize: %v\n", classSN, timeSlot, shouldPenalize)
	return preCheckPassed, !shouldPenalize, nil
}

// 18. 主课 周一~周五 第7节 尽量不排
func sRule4Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	classSN := element.GetClassSN()
	timeSlot := element.GetTimeSlot()

	SN, _ := types.ParseSN(classSN)
	subject, err := models.FindSubjectByID(SN.SubjectID)
	if err != nil {
		return false, false, err
	}
	period := timeSlot%config.NumClasses + 1
	preCheckPassed := period == 7

	shouldPenalize := preCheckPassed && lo.Contains(subject.SubjectGroupIDs, 2)
	return preCheckPassed, !shouldPenalize, nil
}

// 判断subjectGroupID的课程是否已经排完
// Check if all courses in the subject group are scheduled
func isSubjectGroupScheduled(classMatrix *types.ClassMatrix, subjectGroupID int) (bool, error) {

	// 根据科目分组得到所有的科目
	subjects, err := models.FindSubjectsByGroupID(subjectGroupID)
	if err != nil {
		return false, err
	}

	// 根据科目得到该科目的一周课时
	classHours := models.GetClassHours()

	for _, subject := range subjects {
		subjectID := subject.SubjectID
		subjectClassHours := classHours[subjectID]
		totalScheduledHours := 0

		for sn, classMap := range classMatrix.Elements {
			SN, err := types.ParseSN(sn)
			if err != nil {
				return false, err
			}

			if SN.SubjectID == subjectID {
				for _, teacherMap := range classMap {
					for _, venueMap := range teacherMap {
						for _, element := range venueMap {
							if element.Val.Used == 1 {
								totalScheduledHours++
							}
						}
					}
				}
			}
		}

		if subjectClassHours != totalScheduledHours {

			// fmt.Printf("subjectID: %d, subjectClassHours: %d, totalScheduledHours: %d\n", subjectID, subjectClassHours, totalScheduledHours)
			return false, nil
		}
	}

	return true, nil
}
