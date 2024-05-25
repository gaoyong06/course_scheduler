// teacher_time_limit.go
// 教师时间段限制

package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
	"fmt"
	"log"
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

// 获取规则
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
		Score:    0, // 遵守规则, 没有奖励
		Penalty:  2, // 违反规则, 有处罚
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
		currPeriod := -1
		for _, currTimeSlot := range currTimeSlots {
			currPeriod = currTimeSlot % totalClassesPerDay
			if currPeriod >= startPeriod && currPeriod <= endPeriod {
				isValidPeriod = true
				break
			}
		}

		preCheckPassed := currTeacherID == teacherID && isValidPeriod

		if preCheckPassed {
			log.Printf("element.TimeSlots: %v, currTeacherID: %d, currPeriod: %d, count: %d, isValidPeriod: %v, preCheckPassed: %v\n", element.TimeSlots, currTeacherID, currPeriod, count, isValidPeriod, preCheckPassed)
		}
		// 这里count++是指,假设给当前节点排课count会+1
		count++

		shouldPenalize := preCheckPassed && count > maxClasses

		return preCheckPassed, !shouldPenalize, nil
	}
}

// 统计特定教师在某个时间区间的的排课节数
func countTeacherClassesInRange(teacherID int, startPeriod, endPeriod int, classMatrix *types.ClassMatrix, schedule *models.Schedule) int {

	count := 0
	totalClassesPerDay := schedule.GetTotalClassesPerDay()

	// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
	// 统计矩阵内的其他元素
	for _, classMap := range classMatrix.Elements {

		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for timeSlotStr, e := range venueMap {
					if e.Val.Used == 1 && e.TeacherID == teacherID {

						timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
						for _, timeSlot := range timeSlots {
							period := timeSlot % totalClassesPerDay
							// log.Printf("countTeacherClassesInRange, teacherID: %d, startPeriod: %d, endPeriod: %d, period: %d, count: %d\n", teacherID, startPeriod, endPeriod, period, count)
							if period >= startPeriod && period <= endPeriod {
								count++
							}
						}
					}
				}
			}
		}
	}

	// log.Printf("countTeacherClassesInRange, teacherID: %d, startPeriod: %d, endPeriod: %d, count: %d\n", teacherID, startPeriod, endPeriod, count)
	return count
}
