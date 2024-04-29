// 科目互斥限制(科目A与科目B不排在同一天)
package constraint

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/types"
	"fmt"
)

var SMERule1 = &types.Rule{
	Name:     "SMERule1",
	Type:     "dynamic",
	Fn:       smeRule1Fn,
	Score:    0,
	Penalty:  constants.MAX_PENALTY_SCORE,
	Weight:   1,
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

func isSubjectsSameDay(subjectAID, subjectBID int, classMatrix *types.ClassMatrix, element types.ClassUnit) (bool, error) {

	timeSlot := element.GetTimeSlot()
	elementDay := timeSlot / constants.NUM_CLASSES

	// key: day, val:bool
	subjectADays := make(map[int]bool)
	subjectBDays := make(map[int]bool)

	// key: timeSlot val: day
	subjectATimeSlots := make(map[int]int)
	subjectBTimeSlots := make(map[int]int)

	for sn, classMap := range classMatrix.Elements {

		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}

		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for ts, e := range venueMap {
					if e.Val.Used == 1 {
						if SN.SubjectID == subjectAID {
							subjectADays[ts/constants.NUM_CLASSES] = true // 将时间段转换为天数
							subjectATimeSlots[ts] = ts / constants.NUM_CLASSES
						} else if SN.SubjectID == subjectBID {
							subjectBDays[ts/constants.NUM_CLASSES] = true // 将时间段转换为天数
							subjectBTimeSlots[ts] = ts / constants.NUM_CLASSES
						}
					}
				}
			}
		}
	}

	if subjectADays[elementDay] && subjectBDays[elementDay] {

		// fmt.Printf("subjectAID: %d, subjectADays: %v, subjectBID: %d, subjectBDays: %v\n", subjectAID, subjectADays, subjectBID, subjectBDays)

		// Print time slots of both subjects on the same day (elementDay) in a single line
		var subjectATimeSlotsStr, subjectBTimeSlotsStr string
		for ts, day := range subjectATimeSlots {
			if day == elementDay {
				subjectATimeSlotsStr += fmt.Sprintf(" %d", ts)
			}
		}
		for ts, day := range subjectBTimeSlots {
			if day == elementDay {
				subjectBTimeSlotsStr += fmt.Sprintf(" %d", ts)
			}
		}

		// fmt.Printf("Current timeSlot: %d Subject A Time Slots on elementDay: %s, Subject B Time Slots on elementDay: %s\n", timeSlot, subjectATimeSlotsStr, subjectBTimeSlotsStr)

		return true, nil
	}
	return false, nil
}
