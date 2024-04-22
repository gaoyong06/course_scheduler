// 科目互斥限制(科目A与科目B不排在同一天)

package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/types"
)

var SMERule1 = &types.Rule{
	Name:     "SMERule1",
	Type:     "dynamic",
	Fn:       smeRule1Fn,
	Score:    0,
	Penalty:  1,
	Weight:   1,
	Priority: 1,
}

// 37. 活动 体育
func smeRule1Fn(classMatrix map[string]map[int]map[int]map[int]types.Val, element *types.Element) (bool, bool, error) {

	classSN := element.ClassSN
	SN, _ := types.ParseSN(classSN)

	subjectID := SN.SubjectID
	preCheckPassed := subjectID == 14 || subjectID == 6

	shouldPenalize := false
	if preCheckPassed {
		ret, err := isSubjectsSameDay(14, 6, classMatrix)
		if err != nil {
			return false, false, err
		}
		shouldPenalize = ret
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 判断活动课和体育课是否在同一天
func isSubjectsSameDay(subjectAID, subjectBID int, classMatrix map[string]map[int]map[int]map[int]types.Val) (bool, error) {

	subjectADays := make(map[int]bool)
	subjectBDays := make(map[int]bool)
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
							subjectADays[timeSlot/constants.NUM_CLASSES] = true // 将时间段转换为天数
						} else if SN.SubjectID == subjectBID {
							subjectBDays[timeSlot/constants.NUM_CLASSES] = true // 将时间段转换为天数
						}
					}
				}
			}
		}
	}

	for day := 0; day < constants.NUM_DAYS; day++ {
		if subjectADays[day] && subjectBDays[day] {
			return true, nil
		}
	}
	return false, nil
}
