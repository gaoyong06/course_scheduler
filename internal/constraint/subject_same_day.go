// 科目课时小于天数,禁止同一天排多次相同科目的课
// 这个约束需要根据当前已排的课程情况来判断是否满足约束。例如，在排课过程中，如果已经为某一天排好了一节语文课，那么在继续为这一天排课时，就需要考虑到这个约束，避免再为这一天排另一节语文课。
// 因此，这个约束需要在排课过程中动态地检查和更新，因此它是一个动态约束条件

package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
)

var SSDRule1 = &Rule{
	Name:     "SSDRule1",
	Type:     "fixed",
	Fn:       ssdRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 科目课时小于天数,禁止同一天排多次相同科目的课
func ssdRule1Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element Element) (bool, bool, error) {

	classSN := element.ClassSN
	SN, _ := types.ParseSN(classSN)
	subjectID := SN.SubjectID

	// 周课时初始化
	classHours := models.GetClassHours()

	preCheckPassed := classHours[subjectID] <= constants.NUM_DAYS

	shouldPenalize := false
	if preCheckPassed {

		// 检查同一天是否安排科目的排课
		ret := isSubjectSameDay(classMatrix, element.ClassSN, element.TimeSlot)
		shouldPenalize = ret
	}
	return preCheckPassed, !shouldPenalize, nil
}

// 检查同一科目是否在同一天已经排课
func isSubjectSameDay(classMatrix map[string]map[int]map[int]map[int]types.Val, sn string, timeSlot int) bool {

	count := 0
	day := timeSlot / constants.NUM_CLASSES

	for _, teacherMap := range classMatrix[sn] {
		for _, venueMap := range teacherMap {
			for timeSlot1, val := range venueMap {

				if val.Used == 1 {
					day1 := timeSlot1 / constants.NUM_CLASSES
					if day == day1 && timeSlot != timeSlot1 {
						count++
					}
				}
			}
		}
	}

	return count > 0
}
