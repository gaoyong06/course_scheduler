// 教师互斥限制
// teacher_mutual_exclusion.go
package constraint

import (
	"course_scheduler/config"
	"course_scheduler/internal/types"
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
func GetTeacherMutexRules() []*types.Rule {
	constraints := loadTeacherMutexConstraintsFromDB()
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
		Name:     t.String(),
		Type:     "dynamic",
		Fn:       fn,
		Score:    0,
		Penalty:  config.MaxPenaltyScore,
		Weight:   1,
		Priority: 1,
	}
}

// 加载班级固排禁排规则
func loadTeacherMutexConstraintsFromDB() []*SubjectMutex {
	var constraints []*SubjectMutex
	return constraints
}

// 生成规则校验方法
func (t *TeacherMutex) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {

		teacherAID := t.TeacherAID
		teacherBID := t.TeacherBID

		teacherID := element.GetTeacherID()

		preCheckPassed := teacherID == teacherAID || teacherID == teacherBID

		shouldPenalize := false
		if preCheckPassed {
			shouldPenalize = isTeacherSameDay(teacherAID, teacherBID, classMatrix, element)
		}

		return preCheckPassed, !shouldPenalize, nil
	}
}

// 判断教师A,教师B是否同一天都有课
func isTeacherSameDay(teacherAID, teacherBID int, classMatrix *types.ClassMatrix, element types.Element) bool {

	teacher1Days := make(map[int]bool)
	teacher2Days := make(map[int]bool)
	timeSlot := element.GetTimeSlot()

	elementDay := timeSlot / config.NumClasses

	for _, classMap := range classMatrix.Elements {
		for id, teacherMap := range classMap {
			if id == teacherAID {
				for _, timeSlotMap := range teacherMap {
					for timeSlot, element := range timeSlotMap {
						if element.Val.Used == 1 {
							day := timeSlot / config.NumClasses
							teacher1Days[day] = true // 将时间段转换为天数
						}
					}
				}
			} else if id == teacherBID {
				for _, timeSlotMap := range teacherMap {
					for timeSlot, element := range timeSlotMap {
						if element.Val.Used == 1 {
							day := timeSlot / config.NumClasses
							teacher2Days[day] = true // 将时间段转换为天数
						}
					}
				}
			}
		}
	}

	if teacher1Days[elementDay] && teacher2Days[elementDay] {
		return true
	}
	return false
}
