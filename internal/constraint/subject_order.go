// 科目顺序限制(体育课不排在数学课前)

package constraint

import (
	"course_scheduler/internal/types"
	"sort"
)

var SORule1 = &types.Rule{
	Name:     "SORule1",
	Type:     "dynamic",
	Fn:       soRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 38. 体育 数学
func soRule1Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element types.ClassUnit) (bool, bool, error) {

	classSN := element.GetClassSN()
	SN, _ := types.ParseSN(classSN)
	subjectID := SN.SubjectID
	preCheckPassed := subjectID == 6 || subjectID == 2
	shouldPenalize := false
	if preCheckPassed {
		ret, err := isSubjectABeforeSubjectB(6, 2, classMatrix)
		if err != nil {
			return false, false, err
		}
		shouldPenalize = ret
	}
	return preCheckPassed, !shouldPenalize, nil
}

// 判断体育课后是否就是数学课
// 判断课程A(体育)是在课程B(数学)之前
func isSubjectABeforeSubjectB(subjectAID, subjectBID int, classMatrix map[string]map[int]map[int]map[int]types.Val) (bool, error) {

	// 遍历课程表，同时记录课程A和课程B的上课时间段
	var timeSlotsA, timeSlotsB []int
	for sn, classMap := range classMatrix {
		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}
		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for timeSlot, val := range venueMap {
					if val.Used == 1 {
						if SN.SubjectID == subjectAID {
							timeSlotsA = append(timeSlotsA, timeSlot)
						} else if SN.SubjectID == subjectBID {
							timeSlotsB = append(timeSlotsB, timeSlot)
						}
					}
				}
			}
		}
	}

	// 如果课程A或课程B没有上课时间段，则返回false
	if len(timeSlotsA) == 0 || len(timeSlotsB) == 0 {
		return false, nil
	}

	// 对上课时间段进行排序
	sort.Ints(timeSlotsA)
	sort.Ints(timeSlotsB)
	// 检查课程B是否在课程A之后
	for _, timeSlotA := range timeSlotsA {
		for _, timeSlotB := range timeSlotsB {
			if timeSlotB == timeSlotA+1 {
				return true, nil
			}
		}
	}
	// 如果没有找到课程B在课程A之后的上课时间，则返回false
	return false, nil
}
