// 科目课时小于天数,禁止同一天排多次相同科目的课
// 系统约束
// 这个约束需要根据当前已排的课程情况来判断是否满足约束。例如，在排课过程中，如果已经为某一天排好了一节语文课，那么在继续为这一天排课时，就需要考虑到这个约束，避免再为这一天排另一节语文课。
// 因此，这个约束需要在排课过程中动态地检查和更新，因此它是一个动态约束条件

package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"

	"github.com/samber/lo"
)

var subjectSameDayRule = &types.Rule{
	Name:     "subjectSameDay",
	Type:     "dynamic",
	Fn:       ssdRuleFn,
	Score:    2,
	Penalty:  6,
	Weight:   1,
	Priority: 1,
}

// 如果上课次数和上课天数相同, 或者小于上课天数 则一天排一次课
// 正常来讲,上课总次数,应该和上课天数相同
func ssdRuleFn(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, teachingTasks []*models.TeachingTask) (bool, bool, error) {

	classSN := element.GetClassSN()
	timeSlots := element.GetTimeSlots()
	numWorkdays := schedule.NumWorkdays

	SN, _ := types.ParseSN(classSN)

	gradeID := SN.GradeID
	classID := SN.ClassID
	subjectID := SN.SubjectID

	// 科目周课时
	total := models.GetNumClassesPerWeek(gradeID, classID, subjectID, teachingTasks)
	connectedCount := models.GetNumConnectedClassesPerWeek(gradeID, classID, subjectID, teachingTasks)
	count := total - connectedCount

	preCheckPassed := count <= numWorkdays

	shouldPenalize := false
	if preCheckPassed {

		// 检查同一天是否安排科目的排课
		shouldPenalize = isSubjectSameDay(classMatrix, classSN, timeSlots, schedule)
	}
	return preCheckPassed, !shouldPenalize, nil
}

// 检查同一科目是否在同一天已经排课
func isSubjectSameDay(classMatrix *types.ClassMatrix, sn string, timeSlots []int, schedule *models.Schedule) bool {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	count := 0
	day := timeSlots[0] / totalClassesPerDay

	for _, teacherMap := range classMatrix.Elements[sn] {
		for _, venueMap := range teacherMap {
			for timeSlotStr, e := range venueMap {

				timeSlots1 := utils.ParseTimeSlotStr(timeSlotStr)
				intersect := lo.Intersect(timeSlots, timeSlots1)

				if e.Val.Used == 1 && len(intersect) == 0 {
					day1 := timeSlots1[0] / totalClassesPerDay
					if day == day1 {
						count++
					}
				}
			}
		}
	}

	return count > 0
}
