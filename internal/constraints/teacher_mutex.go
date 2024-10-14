// 教师互斥限制
// teacher_mutual_exclusion.go
package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
	"fmt"
)

// | 教师 A | 教师 B |
// | ------ | ------ |
// | 张三   | 李四   |
// | 王五   | 赵六   |

// 教师互斥，教师 A, 教师 B不同时上课
type TeacherMutex struct {
	ID         int `json:"id" mapstructure:"id"`                     // 自增ID
	TeacherAID int `json:"teacher_a_id" mapstructure:"teacher_a_id"` // Teacher A's ID
	TeacherBID int `json:"teacher_b_id" mapstructure:"teacher_b_id"` // Teacher B's ID
}

// 生成字符串
func (t *TeacherMutex) String() string {
	return fmt.Sprintf("ID: %d, TeacherAID: %d, TeacherBID: %d", t.ID, t.TeacherAID, t.TeacherBID)
}

// 获取班级固排禁排规则
func GetTeacherMutexRules(constraints []*TeacherMutex) []*types.Rule {
	// constraints := loadTeacherMutexConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule()
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (t *TeacherMutex) genRule() *types.Rule {
	fn := t.genConstraintFn()
	return &types.Rule{
		Name:     "teacherMutex",
		Type:     "dynamic",
		Fn:       fn,
		Score:    4,
		Penalty:  6,
		Weight:   1,
		Priority: 1,
	}
}

// 加载班级固排禁排规则
func loadTeacherMutexConstraintsFromDB() []*TeacherMutex {
	var constraints []*TeacherMutex
	return constraints
}

// 生成规则校验方法
func (t *TeacherMutex) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachingTask) (bool, bool, error) {

		teacherAID := t.TeacherAID
		teacherBID := t.TeacherBID

		teacherID := element.GetTeacherID()

		preCheckPassed := teacherID == teacherAID || teacherID == teacherBID

		shouldPenalize := false
		if preCheckPassed {
			shouldPenalize = isElementTeacherOnSameDay(teacherAID, teacherBID, classMatrix, element, schedule)
		}

		return preCheckPassed, !shouldPenalize, nil
	}
}

// 判断教师A,教师B是否同一天都有课
func isElementTeacherOnSameDay(teacherAID, teacherBID int, classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule) bool {

	teacherADays := make(map[int]bool)
	teacherBDays := make(map[int]bool)

	onSameDay := false

	timeSlots := element.GetTimeSlots()
	totalClassesPerDay := schedule.GetTotalClassesPerDay()

	elementDay := timeSlots[0] / totalClassesPerDay

	for _, classMap := range classMatrix.Elements {
		for id, teacherMap := range classMap {
			if id == teacherAID {
				for _, timeSlotMap := range teacherMap {
					for timeSlotStr, element := range timeSlotMap {
						if element.Val.Used == 1 {

							timeSlots1 := utils.ParseTimeSlotStr(timeSlotStr)
							for _, timeSlot := range timeSlots1 {
								day := timeSlot / totalClassesPerDay
								teacherADays[day] = true // 将时间段转换为天数
							}

						}
					}
				}
			} else if id == teacherBID {
				for _, timeSlotMap := range teacherMap {
					for timeSlotStr, element := range timeSlotMap {
						if element.Val.Used == 1 {

							timeSlots1 := utils.ParseTimeSlotStr(timeSlotStr)
							for _, timeSlot := range timeSlots1 {
								day := timeSlot / totalClassesPerDay
								teacherBDays[day] = true // 将时间段转换为天数
							}
						}
					}
				}
			}
		}
	}

	if element.TeacherID == teacherAID {
		onSameDay = teacherBDays[elementDay]
	}

	if element.TeacherID == teacherBID {
		onSameDay = teacherADays[elementDay]
	}

	// log.Printf("teacher mutex, element.timeSlots: %v, element.TeacherID: %d, teacherAID: %d, teacherBID: %d, elementDay: %d, onSameDay: %v\n", element.TimeSlots, element.TeacherID, teacherAID, teacherBID, elementDay, onSameDay)

	return onSameDay
}
