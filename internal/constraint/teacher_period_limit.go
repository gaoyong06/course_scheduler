// 教师节数限制
package constraint

import (
	"course_scheduler/internal/models"
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
func GetTeacherClassLimitRules(constraints []*TeacherPeriodLimit) []*types.Rule {
	// constraints := loadTeacherPeriodLimitConstraintsFromDB()
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
		Name:     "teacherPeriodLimit",
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
	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

		totalClassesPerDay := schedule.GetTotalClassesPerDay()
		teacherID := t.TeacherID
		period := t.Period
		maxClassesCount := t.MaxClassesCount

		currTeacherID := element.GetTeacherID()
		currTimeSlot := element.GetTimeSlot()
		currPeriod := currTimeSlot % totalClassesPerDay
		preCheckPassed := teacherID == currTeacherID && period == currPeriod

		shouldPenalize := false
		if preCheckPassed {
			count := countTeacherClassInPeriod(teacherID, period, classMatrix, schedule)
			shouldPenalize = preCheckPassed && count > maxClassesCount
		}

		return preCheckPassed, !shouldPenalize, nil
	}
}

// 计算教师某节课的上课次数
func countTeacherClassInPeriod(teacherID int, period int, classMatrix *types.ClassMatrix, schedule *models.Schedule) int {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	count := 0

	// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
	for _, teacherMap := range classMatrix.Elements {
		for id, venueMap := range teacherMap {
			if teacherID == id {
				for _, timeSlotMap := range venueMap {
					for timeSlot, element := range timeSlotMap {
						if element.Val.Used == 1 {
							elementPeriod := timeSlot % totalClassesPerDay
							if elementPeriod == period {
								count++
							}
						}
					}
				}
			}
		}
	}
	return count
}
