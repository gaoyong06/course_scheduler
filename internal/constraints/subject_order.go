// 科目顺序限制(体育课不排在数学课前)

package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
	"fmt"
	"sort"

	"github.com/samber/lo"
)

// ###### 科目顺序限制

// 体育课不排在数学课前

// | 科目 A | 科目 B | 描述 |
// | ------ | ------ | ---- |
// | 体育   | 数学   |      |

type SubjectOrder struct {
	ID         int `json:"id" mapstructure:"id"`                     // 自增ID
	SubjectAID int `json:"subject_a_id" mapstructure:"subject_a_id"` // 科目A ID
	SubjectBID int `json:"subject_b_id" mapstructure:"subject_b_id"` // 科目B ID
}

// 生成字符串
func (s *SubjectOrder) String() string {
	return fmt.Sprintf("ID: %d, SubjectAID: %d, SubjectBID: %d", s.ID, s.SubjectAID, s.SubjectBID)
}

// 获取班级固排禁排规则
func GetSubjectOrderRules(constraints []*SubjectOrder) []*types.Rule {
	// constraints := loadSubjectOrderConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule()
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (s *SubjectOrder) genRule() *types.Rule {
	fn := s.genConstraintFn()
	return &types.Rule{
		Name:     "subjectOrder",
		Type:     "dynamic",
		Fn:       fn,
		Score:    4,
		Penalty:  6,
		Weight:   1,
		Priority: 1,
	}
}

// 加载班级固排禁排规则
func loadSubjectOrderConstraintsFromDB() []*SubjectOrder {
	var constraints []*SubjectOrder
	return constraints
}

// 生成规则校验方法
func (s *SubjectOrder) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, teachingTasks []*models.TeachingTask) (bool, bool, error) {

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
			shouldPenalize, err = isElementSubjectABeforeSubjectB(subjectAID, subjectBID, classMatrix, element, schedule)
			if err != nil {
				return false, false, err
			}
		}
		return preCheckPassed, !shouldPenalize, nil
	}
}

// 判断当前元素的排课课程是否会出现课程A(体育)是在课程B(数学)之前的结果
func isElementSubjectABeforeSubjectB(subjectAID, subjectBID int, classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule) (bool, error) {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// 遍历课程表，同时记录课程A和课程B的上课时间段
	var timeSlotsA, timeSlotsB []int
	timeSlots := element.GetTimeSlots()
	for sn, classMap := range classMatrix.Elements {
		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}
		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for timeSlotStr, e := range venueMap {
					if e.Val.Used == 1 {

						eleTimeSlots := utils.ParseTimeSlotStr(timeSlotStr)
						for _, timeSlot := range eleTimeSlots {

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
	}

	// 如果当前课时是课程A
	if element.SubjectID == subjectAID {
		timeSlotsA = append(timeSlotsA, timeSlots...)
	}

	// 如果当前课时是课程B
	if element.SubjectID == subjectBID {
		timeSlotsB = append(timeSlotsB, timeSlots...)
	}

	// 对上课时间段进行排序
	sort.Ints(timeSlotsA)
	sort.Ints(timeSlotsB)

	// 检查课程A是否在课程B之前

	for _, timeSlotA := range timeSlotsA {
		for _, timeSlotB := range timeSlotsB {

			dayA := timeSlotA / totalClassesPerDay
			dayB := timeSlotB / totalClassesPerDay

			if dayA == dayB && timeSlotB == timeSlotA+1 && (lo.Contains(timeSlots, timeSlotA) || lo.Contains(timeSlots, timeSlotB)) {
				return true, nil
			}
		}
	}
	// 如果没有找到课程B在课程A之后的上课时间，则返回false
	return false, nil
}
