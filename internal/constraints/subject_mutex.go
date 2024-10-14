// 科目互斥限制(科目A与科目B不排在同一天)
package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
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
		Score:    4,
		Penalty:  6,
		Weight:   1,
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

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachingTask) (bool, bool, error) {

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
			shouldPenalize, err = isElementSubjectOnSameDay(subjectAID, subjectBID, classMatrix, element, schedule)
			if err != nil {
				return false, false, err
			}
		}
		return preCheckPassed, !shouldPenalize, nil
	}
}

// 判断当前元素排课科目,是否和subjectAID或者subjectBID,在同一天
func isElementSubjectOnSameDay(subjectAID, subjectBID int, classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule) (bool, error) {

	timeSlots := element.GetTimeSlots()
	totalClassesPerDay := schedule.GetTotalClassesPerDay()

	// 这里使用第一个时间段
	elementDay := timeSlots[0] / totalClassesPerDay

	// key: day, val:bool
	subjectADays := make(map[int]bool)
	subjectBDays := make(map[int]bool)

	onSameDay := false

	for sn, classMap := range classMatrix.Elements {

		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}

		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for timeSlotStr, e := range venueMap {
					if e.Val.Used == 1 {
						ts := utils.ParseTimeSlotStr(timeSlotStr)
						for _, t := range ts {
							if SN.SubjectID == subjectAID {
								subjectADays[t/totalClassesPerDay] = true // 将时间段转换为天数
							} else if SN.SubjectID == subjectBID {
								subjectBDays[t/totalClassesPerDay] = true // 将时间段转换为天数
							}
						}
					}
				}
			}
		}
	}

	if element.SubjectID == subjectAID {
		onSameDay = subjectBDays[elementDay]
	}

	if element.SubjectID == subjectBID {
		onSameDay = subjectADays[elementDay]
	}

	// log.Printf("subject mutex, element.timeSlots: %v, element.subjectID: %d, subjectAID: %d, subjectBID: %d, elementDay: %d, onSameDay: %v\n", element.TimeSlots, element.SubjectID, subjectAID, subjectBID, elementDay, onSameDay)

	return onSameDay, nil
}
