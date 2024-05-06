// 科目课时小于天数,禁止同一天排多次相同科目的课
// 系统约束
// 这个约束需要根据当前已排的课程情况来判断是否满足约束。例如，在排课过程中，如果已经为某一天排好了一节语文课，那么在继续为这一天排课时，就需要考虑到这个约束，避免再为这一天排另一节语文课。
// 因此，这个约束需要在排课过程中动态地检查和更新，因此它是一个动态约束条件

package constraint

import (
	"course_scheduler/config"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
)

var subjectSameDayRule = &types.Rule{
	Name:     "subjectSameDayRule",
	Type:     "dynamic",
	Fn:       ssdRuleFn,
	Score:    0,
	Penalty:  config.MaxPenaltyScore,
	Weight:   1,
	Priority: 1,
}

// 科目课时小于天数,禁止同一天排多次相同科目的课
func ssdRuleFn(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {

	classSN := element.GetClassSN()
	timeSlot := element.GetTimeSlot()

	SN, _ := types.ParseSN(classSN)
	subjectID := SN.SubjectID

	// 周课时初始化
	classHours := models.GetClassHours()

	preCheckPassed := classHours[subjectID] <= config.NumDays

	shouldPenalize := false
	if preCheckPassed {

		// 检查同一天是否安排科目的排课
		ret := isSubjectSameDay(classMatrix, classSN, timeSlot)
		shouldPenalize = ret
	}
	return preCheckPassed, !shouldPenalize, nil
}

// 检查同一科目是否在同一天已经排课
func isSubjectSameDay(classMatrix *types.ClassMatrix, sn string, timeSlot int) bool {

	count := 0
	day := timeSlot / config.NumClasses

	for _, teacherMap := range classMatrix.Elements[sn] {
		for _, venueMap := range teacherMap {
			for timeSlot1, element := range venueMap {

				if element.Val.Used == 1 && timeSlot != timeSlot1 {
					day1 := timeSlot1 / config.NumClasses
					if day == day1 {
						count++
					}
				}
			}
		}
	}

	return count > 0
}
