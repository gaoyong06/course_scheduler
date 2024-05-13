// 科目互斥限制(科目A与科目B不排在同一天)
package constraint

import (
	"course_scheduler/config"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
)

// | 科目 A | 科目 B |
// | ------ | ------ |
// | 数学   | 英语   |
// | 历史   | 地理   |

// 科目互斥，格式：科目 A+科目 B，A 与 B 不排在同一天
type SubjectMutex struct {
	ID         int `json:"id" mapstructure:"id"`                     // 自增ID
	SubjectAID int `json:"subject_a_id" mapstructure:"subject_a_id"` // 科目A ID
	SubjectBID int `json:"subject_b_id" mapstructure:"subject_b_id"` // 科目B ID
}

// 生成字符串
func (sm *SubjectMutex) String() string {
	return fmt.Sprintf("ID: %d, SubjectAID: %d, SubjectBID: %d", sm.ID, sm.SubjectAID, sm.SubjectBID)
}

// 获取班级固排禁排规则
func GetSubjectMutexRules(constraints []*SubjectMutex) []*types.Rule {
	// constraints := loadSubjectMutexConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule()
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (s *SubjectMutex) genRule() *types.Rule {
	fn := s.genConstraintFn()
	return &types.Rule{
		Name:     "subjectMutex",
		Type:     "dynamic",
		Fn:       fn,
		Score:    0,
		Penalty:  config.MaxPenaltyScore,
		Weight:   2,
		Priority: 1,
	}
}

// 加载班级固排禁排规则
func loadSubjectMutexConstraintsFromDB() []*SubjectMutex {
	var constraints []*SubjectMutex
	return constraints
}

// 生成规则校验方法
func (s *SubjectMutex) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

		subjectAID := s.SubjectAID
		subjectBID := s.SubjectBID

		classSN := element.GetClassSN()
		SN, err := types.ParseSN(classSN)
		if err != nil {
			return false, false, err
		}

		subjectID := SN.SubjectID
		preCheckPassed := subjectID == subjectAID || subjectID == subjectBID

		shouldPenalize := false
		if preCheckPassed {
			ret, err := isSubjectsSameDay(subjectAID, subjectBID, classMatrix, element, schedule)
			if err != nil {
				return false, false, err
			}
			shouldPenalize = ret
		}
		return preCheckPassed, !shouldPenalize, nil
	}
}

func isSubjectsSameDay(subjectAID, subjectBID int, classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule) (bool, error) {

	timeSlot := element.GetTimeSlot()
	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	elementDay := timeSlot / totalClassesPerDay

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
							subjectADays[ts/totalClassesPerDay] = true // 将时间段转换为天数
							subjectATimeSlots[ts] = ts / totalClassesPerDay
						} else if SN.SubjectID == subjectBID {
							subjectBDays[ts/totalClassesPerDay] = true // 将时间段转换为天数
							subjectBTimeSlots[ts] = ts / totalClassesPerDay
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
