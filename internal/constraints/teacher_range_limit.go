// teacher_time_limit.go
// 教师时间段限制

package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
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
	ID              int    `json:"id" mapstructure:"id"`                               // 自增ID
	TeacherID       int    `json:"teacher_id" mapstructure:"teacher_id"`               // 教师ID
	Range           string `json:"range" mapstructure:"range"`                         // 时间区间 上午: forenoon, 下午: afternoon, 全天: all_day, 晚自习: night
	MaxClassesCount int    `json:"max_classes_count" mapstructure:"max_classes_count"` // 最多排课次数
}

// 生成字符串
func (t *TeacherRangeLimit) String() string {
	return fmt.Sprintf("ID: %d, TeacherID: %d, Range: %s, MaxClasses: %d", t.ID, t.TeacherID, t.Range, t.MaxClassesCount)
}

// 获取教师时间段限制规则
func GetTeacherRangeLimitRules(constraints []*TeacherRangeLimit) []*types.Rule {

	// constraints := loadTeacherRangeLimitConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule()
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (t *TeacherRangeLimit) genRule() *types.Rule {
	fn := t.genConstraintFn()
	return &types.Rule{
		Name:     "teacherRangeLimit",
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
func (t *TeacherRangeLimit) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

		totalClassesPerDay := schedule.GetTotalClassesPerDay()
		// 规则参数
		teacherID := t.TeacherID
		maxClasses := t.MaxClassesCount

		// 将range转为时间段的起止时间段
		startPeriod, endPeriod := schedule.GetPeriodWithRange(t.Range)
		count := countTeacherClassesInRange(teacherID, startPeriod, endPeriod, classMatrix, schedule)
		currTeacherID := element.GetTeacherID()

		isValidPeriod := false
		currTimeSlots := element.GetTimeSlots()
		for _, currTimeSlot := range currTimeSlots {
			currPeriod := currTimeSlot % totalClassesPerDay
			if currPeriod >= startPeriod && currPeriod <= endPeriod {
				isValidPeriod = true
				break
			}
		}

		preCheckPassed := currTeacherID == teacherID && isValidPeriod
		shouldPenalize := preCheckPassed && count > maxClasses

		return preCheckPassed, !shouldPenalize, nil
	}
}

// 26. 王老师 晚自习 最多1节
func countTeacherClassesInRange(teacherID int, startPeriod, endPeriod int, classMatrix *types.ClassMatrix, schedule *models.Schedule) int {

	count := 0
	totalClassesPerDay := schedule.GetTotalClassesPerDay()

	// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
	for _, classMap := range classMatrix.Elements {
		for _, teacherMap := range classMap {
			for id, venueMap := range teacherMap {
				if teacherID == id {
					for timeSlotStr, element := range venueMap {

						timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
						for _, timeSlot := range timeSlots {
							period := timeSlot % totalClassesPerDay
							if element.Val.Used == 1 && period >= startPeriod && period <= endPeriod {
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
