// teacher_time_limit.go
// 教师时间段限制

package constraint

import (
	"course_scheduler/config"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
)

// ###### 教师时间段限制

// | 教师   | 时间段           | 最多排课节数 |
// | ------ | ---------------- | ------------ |
// | 王老师 | 上午             | 1 节         |
// | 王老师 | 下午             | 2 节         |
// | 王老师 | 全天(不含晚自习) | 3 节         |
// | 王老师 | 晚自习           | 1 节         |

type TeacherRangeLimit struct {
	ID         int    `json:"id" mapstructure:"id"`                   // 自增ID
	TeacherID  int    `json:"teacher_id" mapstructure:"teacher_id"`   // 教师ID
	Range      string `json:"range" mapstructure:"range"`             // 时间区间 上午: forenoon, 下午: afternoon, 全天: all_day, 晚自习: night
	MaxClasses int    `json:"max_classes" mapstructure:"max_classes"` // 最多排课节数
}

// 生成字符串
func (t *TeacherRangeLimit) String() string {
	return fmt.Sprintf("ID: %d, TeacherID: %d, Range: %s, MaxClasses: %d", t.ID, t.TeacherID, t.Range, t.MaxClasses)
}

// 获取教师时间段限制规则
func GetTeacherRangeLimitRules(schedule *models.Schedule) []*types.Rule {

	constraints := loadTeacherRangeLimitConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule(schedule)
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (t *TeacherRangeLimit) genRule(schedule *models.Schedule) *types.Rule {
	fn := t.genConstraintFn(schedule)
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
func loadTeacherRangeLimitConstraintsFromDB() []*TeacherRangeLimit {
	var constraints []*TeacherRangeLimit
	return constraints
}

// 生成规则校验方法
func (t *TeacherRangeLimit) genConstraintFn(schedule *models.Schedule) types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {

		// 规则参数
		teacherID := t.TeacherID
		maxClasses := t.MaxClasses

		// 将range转为时间段的起止时间段
		startPeriod, endPeriod := schedule.GetPeriodWithRange(t.Range)

		currTeacherID := element.GetTeacherID()
		currTimeSlot := element.GetTimeSlot()
		currPeriod := currTimeSlot % config.NumClasses
		count := countTeacherClassesInRange(teacherID, startPeriod, endPeriod, classMatrix)

		preCheckPassed := currTeacherID == teacherID && currPeriod >= startPeriod && currPeriod <= endPeriod
		shouldPenalize := preCheckPassed && count > maxClasses
		return preCheckPassed, !shouldPenalize, nil
	}
}

// 26. 王老师 晚自习 最多1节
func countTeacherClassesInRange(teacherID int, startPeriod, endPeriod int, classMatrix *types.ClassMatrix) int {

	count := 0
	for _, classMap := range classMatrix.Elements {
		for id, teacherMap := range classMap {
			if teacherID == id {
				for _, timeSlotMap := range teacherMap {
					if timeSlotMap == nil {
						continue
					}
					for timeSlot, element := range timeSlotMap {
						if element.Val.Used == 1 && timeSlot >= startPeriod && timeSlot <= endPeriod {
							count++
						}
					}
				}
			}
		}
	}
	return count
}
