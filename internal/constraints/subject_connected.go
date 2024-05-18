// 连堂课校验(科目课时数大于上课天数时, 禁止同一天排多次课是非连续的, 要排成连堂课)
// 系统约束
// 需要根据当前已排的课程情况来判断是否满足约束。例如，在排课过程中，如果已经排好了周一第1节语文课，那么在继续为周一排课时，就需要考虑到这个约束，避免再为周一排一个非连续的语文课。
// 因此，这个约束需要在排课过程中动态地检查和更新，因此它是一个动态约束条件

package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"math"
	"sort"
)

var subjectConnectedRule = &types.Rule{
	Name:     "subjectConnected",
	Type:     "dynamic",
	Fn:       scRuleFn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   1,
	Priority: 1,
}

// 连堂课校验(科目课时数大于上课天数时, 禁止同一天排多次课是非连续的, 要排成连堂课)
func scRuleFn(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

	classSN := element.GetClassSN()

	SN, _ := types.ParseSN(classSN)
	gradeID := SN.GradeID
	classID := SN.ClassID
	subjectID := SN.SubjectID

	// 科目周课时
	numClassesPerWeek := models.GetNumClassesPerWeek(gradeID, classID, subjectID, taskAllocs)

	numWorkdays := schedule.NumWorkdays
	preCheckPassed := numClassesPerWeek > numWorkdays
	isConnected := false
	var err error
	if preCheckPassed {

		// 检查同一科目一天安排多次是否是连堂
		isConnected, err = isSubjectConnected(classMatrix, element, schedule)
		if err != nil {
			return false, false, err
		}
	}
	return preCheckPassed, isConnected, nil
}

// 检查科目是否连续排课
func isSubjectConnected(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule) (bool, error) {

	gradeID := element.GradeID
	classID := element.ClassID
	subjectID := element.SubjectID

	// 每天课节数
	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	subjectTimeSlots := make([]int, 0)

	// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
	for sn, classMap := range classMatrix.Elements {

		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}
		if SN.GradeID != gradeID || SN.ClassID != classID || SN.SubjectID != subjectID {
			continue
		}

		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {

				for timeSlot, element := range venueMap {

					if element.Val.Used == 1 && element.GradeID == gradeID && element.ClassID == classID && element.SubjectID == subjectID {
						subjectTimeSlots = append(subjectTimeSlots, timeSlot)
					}
				}
			}
		}
	}

	sort.Ints(subjectTimeSlots)
	// log.Printf(" gradeID: %d, classID: %d, subjectID: %d, current subject time slots: %#v\n", gradeID, classID, subjectID, subjectTimeSlots)

	dayTimeSlots := make(map[int][]int)
	for i := 0; i < len(subjectTimeSlots); i++ {

		day := subjectTimeSlots[i] / totalClassesPerDay
		dayTimeSlots[day] = append(dayTimeSlots[day], subjectTimeSlots[i])
	}

	// 计算当前时间节点是第几天
	elementDay := element.TimeSlot / totalClassesPerDay
	// 遍历同一天的时间段
	timeSlots := dayTimeSlots[elementDay]
	for i := 0; i < len(timeSlots)-1; i++ {
		if timeSlots[i]+1 != timeSlots[i+1] {
			return false, nil
		}
	}

	// log.Printf("elementDay: %d, timeSlots: %v\n", elementDay, timeSlots)
	return true, nil
}
