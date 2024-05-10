// 连堂课校验(科目课时数大于上课天数时, 禁止同一天排多次课是非连续的, 要排成连堂课)
// 系统约束
// 需要根据当前已排的课程情况来判断是否满足约束。例如，在排课过程中，如果已经排好了周一第1节语文课，那么在继续为周一排课时，就需要考虑到这个约束，避免再为周一排一个非连续的语文课。
// 因此，这个约束需要在排课过程中动态地检查和更新，因此它是一个动态约束条件

package constraint

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
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

// 连堂课校验(科目课时数大于上课天数时, 禁止同一天排多次课是非连续的, 要排成连堂课)
func scRuleFn(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

	classSN := element.GetClassSN()

	SN, _ := types.ParseSN(classSN)
	gradeID := SN.GradeID
	classID := SN.ClassID
	subjectID := SN.SubjectID

	// 科目周课时
	classHours := models.GetNumClassesPerWeek(gradeID, classID, subjectID, taskAllocs)

	numWorkdays := schedule.NumWorkdays
	preCheckPassed := classHours > numWorkdays

	// if subjectID == 1 {
	// 	fmt.Printf("classHours: %d, numWorkdays: %d, preCheckPassed: %v\n", classHours, numWorkdays, preCheckPassed)
	// 	panic("aaaaa")
	// }

	isValid := false
	var err error
	if preCheckPassed {

		// 检查同一科目一天安排多次是否是连堂
		isValid, err = isSubjectConnected(classMatrix, element, schedule)
		if err != nil {
			return false, false, err
		}

		if subjectID == 1 {
			fmt.Printf("gradeID: %d, classID: %d, subjectID: %d, timeSlot: %d, isValid: %v\n", element.GradeID, element.ClassID, element.SubjectID, element.TimeSlot, isValid)
		}

	}

	return preCheckPassed, isValid, nil
}

// // 检查科目是否连续排课
// func isSubjectConnected(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule) (bool, error) {

// 	gradeID := element.GradeID
// 	classID := element.ClassID
// 	subjectID := element.SubjectID

// 	// // 每天课节数
// 	// totalClassesPerDay := schedule.GetTotalClassesPerDay()

// 	// // 每周的课节数
// 	// totalClassesPerWeek := schedule.TotalClassesPerWeek()

// 	// 一周课程时间段
// 	// weekTimeSlots := schedule.GenWeekTimeSlots()

// 	subjectTimeSlots := make([]int, 0)
// 	for sn, classMap := range classMatrix.Elements {
// 		SN, err := types.ParseSN(sn)
// 		if err != nil {
// 			return false, err
// 		}
// 		if SN.GradeID != gradeID && SN.ClassID != classID && SN.SubjectID != subjectID {
// 			for _, teacherMap := range classMap {
// 				for _, venueMap := range teacherMap {
// 					for timeSlot, element := range venueMap {
// 						if element.Val.Used == 1 {
// 							subjectTimeSlots = append(subjectTimeSlots, timeSlot)
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	sort.Ints(subjectTimeSlots)
// 	for i := 0; i < len(subjectTimeSlots)-1; i++ {
// 		if subjectTimeSlots[i]+1 == subjectTimeSlots[i+1] {
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }

// 检查科目是否连续排课
func isSubjectConnected(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule) (bool, error) {

	gradeID := element.GradeID
	classID := element.ClassID
	subjectID := element.SubjectID

	// 每天课节数
	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	subjectTimeSlots := make([]int, 0)
	for sn, classMap := range classMatrix.Elements {
		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}
		if SN.GradeID != gradeID && SN.ClassID != classID && SN.SubjectID != subjectID {
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

	// 计算当前时间节点是第几天
	day := element.TimeSlot / totalClassesPerDay

	// 遍历同一天的时间段
	for i := 0; i < len(subjectTimeSlots)-1; i++ {

		// 获取时间段对应的天数
		timeSlotDay := subjectTimeSlots[i] / totalClassesPerDay
		if timeSlotDay == day && subjectTimeSlots[i]+1 == subjectTimeSlots[i+1] {
			return true, nil
		}
	}
	return false, nil
}
