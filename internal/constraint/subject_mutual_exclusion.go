// 科目互斥限制(科目A与科目B不排在同一天)

package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/types"
	"math"
)

var SMERule1 = &types.Rule{
	Name:     "SMERule1",
	Type:     "dynamic",
	Fn:       smeRule1Fn,
	Score:    0,
	Penalty:  math.MaxInt32,
	Weight:   2,
	Priority: 1,
}

// 37. 活动 体育
func smeRule1Fn(classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, bool, error) {

	classSN := element.GetClassSN()
	SN, _ := types.ParseSN(classSN)

	subjectID := SN.SubjectID
	preCheckPassed := subjectID == 14 || subjectID == 6

	shouldPenalize := false
	if preCheckPassed {
		ret, err := isSubjectsSameDay(14, 6, classMatrix, element)
		if err != nil {
			return false, false, err
		}
		shouldPenalize = ret
	}

	return preCheckPassed, !shouldPenalize, nil
}

// 判断活动课和体育课是否在同一天
func isSubjectsSameDay(subjectAID, subjectBID int, classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, error) {

	timeSlot := element.GetTimeSlot()
	subjectADays := make(map[int]bool)
	subjectBDays := make(map[int]bool)
	for sn, classMap := range classMatrix.Elements {

		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}

		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for ts, element := range venueMap {
					if element.Val.Used == 1 {
						if SN.SubjectID == subjectAID {
							subjectADays[ts/constants.NUM_CLASSES] = true // 将时间段转换为天数
						} else if SN.SubjectID == subjectBID {
							subjectBDays[ts/constants.NUM_CLASSES] = true // 将时间段转换为天数
						}
					}
				}
			}
		}
	}

	elementDay := timeSlot / constants.NUM_CLASSES
	if subjectADays[elementDay] && subjectBDays[elementDay] {
		return true, nil
	}
	return false, nil
}
