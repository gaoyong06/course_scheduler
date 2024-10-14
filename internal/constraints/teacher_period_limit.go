// 教师节数限制
package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
	"fmt"

	"github.com/samber/lo"
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
	Period          int `json:"period" mapstructure:"period"`                       // 节次(period 从1开始)
	MaxClassesCount int `json:"max_classes_count" mapstructure:"max_classes_count"` // 最多排课次数
}

// 生成字符串
func (t *TeacherPeriodLimit) String() string {
	return fmt.Sprintf("ID: %d, TeacherID: %d, Period: %d, MaxClassesCount: %d", t.ID, t.TeacherID, t.Period, t.MaxClassesCount)
}

// 获取规则
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
		Score:    0, // 遵守规则,没有奖励
		Penalty:  2, // 违反规则,有处罚
		Weight:   1,
		Priority: 1,
	}
}

// 加载规则
func loadTeacherPeriodLimitConstraintsFromDB() []*TeacherPeriodLimit {
	var constraints []*TeacherPeriodLimit
	return constraints
}

// 生成规则校验方法
func (t *TeacherPeriodLimit) genConstraintFn() types.ConstraintFn {
	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachingTask) (bool, bool, error) {

		teacherID := t.TeacherID

		// t.Period 从1开始
		period := t.Period - 1
		maxClassesCount := t.MaxClassesCount

		currTeacherID := element.GetTeacherID()

		// 当前元素排课的节次
		elementPeriods := types.GetElementPeriods(element, schedule)

		preCheckPassed := teacherID == currTeacherID && lo.Contains(elementPeriods, period)

		shouldPenalize := false
		if preCheckPassed {
			count := countTeacherClassInPeriod(teacherID, period, classMatrix, schedule)

			// 如果当前元素已经被排课,要去除掉
			// 否则,则假设在现在的节点排课，count要加1
			if element.Val.Used == 0 {
				count++
			}
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
					for timeSlotStr, element := range timeSlotMap {

						if element.Val.Used == 1 {
							timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
							for _, timeSlot := range timeSlots {
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
	}
	return count
}
