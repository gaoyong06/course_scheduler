// 教师节数限制
package constraint

import (
	"course_scheduler/config"
	"course_scheduler/internal/types"
	"fmt"
)

// ###### 教师节数限制

// | 教师   | 节次        | 最多排课次数 |
// | ------ | ----------- | ------------ |
// | 王老师 | 上午第 4 节 | 3 次         |
// | 李老师 | 上午第 4 节 | 3 次         |
// | 刘老师 | 上午第 4 节 | 3 次         |
// | 张老师 | 上午第 4 节 | 3 次         |

type TeacherPeriodLimit struct {
	ID              int `json:"id" mapstructure:"id"`                               // 自增ID
	TeacherID       int `json:"teacher_id" mapstructure:"teacher_id"`               // 教师ID
	Period          int `json:"period" mapstructure:"period"`                       // 节次
	MaxClassesCount int `json:"max_classes_count" mapstructure:"max_classes_count"` // 最多排课次数
}

// 生成字符串
func (t *TeacherPeriodLimit) String() string {
	return fmt.Sprintf("ID: %d, TeacherID: %d, Period: %d, MaxClassesCount: %d", t.ID, t.TeacherID, t.Period, t.MaxClassesCount)
}

// 获取班级固排禁排规则
func GetTeacherClassLimitRules() []*types.Rule {
	constraints := loadTeacherPeriodLimitConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule()
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (t *TeacherPeriodLimit) genRule() *types.Rule {
	fn := t.genConstraintFn()
	return &types.Rule{
		Name:     t.String(),
		Type:     "dynamic",
		Fn:       fn,
		Score:    0,
		Penalty:  1,
		Weight:   1,
		Priority: 1,
	}
}

// 加载班级固排禁排规则
func loadTeacherPeriodLimitConstraintsFromDB() []*TeacherPeriodLimit {
	var constraints []*TeacherPeriodLimit
	return constraints
}

// 生成规则校验方法
func (t *TeacherPeriodLimit) genConstraintFn() types.ConstraintFn {
	return func(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {

		teacherID := t.TeacherID
		period := t.Period
		maxClassesCount := t.MaxClassesCount

		currTeacherID := element.GetTeacherID()
		currTimeSlot := element.GetTimeSlot()
		currPeriod := currTimeSlot % config.NumClasses
		preCheckPassed := teacherID == currTeacherID && period == currPeriod

		shouldPenalize := false
		if preCheckPassed {
			count := countTeacherClassInPeriod(teacherID, period, classMatrix)
			shouldPenalize = preCheckPassed && count > maxClassesCount
		}

		return preCheckPassed, !shouldPenalize, nil
	}
}

// 计算教师某节课的上课次数
func countTeacherClassInPeriod(teacherID int, period int, classMatrix *types.ClassMatrix) int {
	count := 0
	for _, classMap := range classMatrix.Elements {
		for id, teacherMap := range classMap {
			if teacherID == id {
				for _, timeSlotMap := range teacherMap {
					if timeSlotMap == nil {
						continue
					}
					for timeSlot, element := range timeSlotMap {
						if element.Val.Used == 1 && timeSlot%config.NumClasses+1 == period {
							count++
						}
					}
				}
			}
		}
	}
	return count
}
