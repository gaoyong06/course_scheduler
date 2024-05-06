// 连堂课校验(科目课时数大于上课天数时, 禁止同一天排多次课是非连续的, 要排成连堂课)
// 系统约束
// 需要根据当前已排的课程情况来判断是否满足约束。例如，在排课过程中，如果已经排好了周一第1节语文课，那么在继续为周一排课时，就需要考虑到这个约束，避免再为周一排一个非连续的语文课。
// 因此，这个约束需要在排课过程中动态地检查和更新，因此它是一个动态约束条件

package constraint

import (
	"course_scheduler/config"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"sort"
)

var subjectConnectedRule = &types.Rule{
	Name:     "subjectConnected",
	Type:     "dynamic",
	Fn:       scRuleFn,
	Score:    0,
	Penalty:  3,
	Weight:   1,
	Priority: 1,
}

// 科目课时小于天数,禁止同一天排多次相同科目的课
func scRuleFn(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {

	classSN := element.GetClassSN()

	SN, _ := types.ParseSN(classSN)
	subjectID := SN.SubjectID

	// 周课时初始化
	classHours := models.GetClassHours()

	preCheckPassed := classHours[subjectID] > config.NumDays

	shouldPenalize := false
	if preCheckPassed {

		// 检查同一科目一天安排多次是否是连堂
		ret, err := isSubjectConnected(subjectID, classMatrix)
		if err != nil {
			return false, false, err
		}
		shouldPenalize = ret
	}
	return preCheckPassed, !shouldPenalize, nil
}

// 检查科目是否连续排课
func isSubjectConnected(subjectID int, classMatrix *types.ClassMatrix) (bool, error) {
	subjectTimeSlots := make([]int, 0)
	for sn, classMap := range classMatrix.Elements {
		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}
		if SN.SubjectID != subjectID {
			continue
		}
		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for timeSlot, element := range venueMap {
					if element.Val.Used == 1 {
						subjectTimeSlots = append(subjectTimeSlots, timeSlot)
					}
				}
			}
		}
	}
	sort.Ints(subjectTimeSlots)
	for i := 0; i < len(subjectTimeSlots)-1; i++ {
		if subjectTimeSlots[i]+1 == subjectTimeSlots[i+1] {
			return true, nil
		}
	}
	return false, nil
}
