// 教师不跨中午(教师排了上午最后一节就不排下午第一节)

package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
	"fmt"

	"github.com/samber/lo"
)

// ###### 教师不跨中午约束

// 教师排了上午最后一节就不排下午第一节

// | 教师   |
// | ------ |
// | 王老师 |
// | 李老师 |
type TeacherNoonBreak struct {
	ID        int `json:"id" mapstructure:"id"`                 // 自增ID
	TeacherID int `json:"teacher_id" mapstructure:"teacher_id"` // 教师ID
}

// 生成字符串
func (t *TeacherNoonBreak) String() string {
	return fmt.Sprintf("ID: %d, TeacherID: %d", t.ID, t.TeacherID)
}

// 获取规则
func GetTeacherNoonBreakRules(constraints []*TeacherNoonBreak) []*types.Rule {
	// constraints := loadTeacherNoonBreakConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule()
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (t *TeacherNoonBreak) genRule() *types.Rule {
	fn := t.genConstraintFn()
	return &types.Rule{
		Name:     "teacherNoonBreak",
		Type:     "dynamic",
		Fn:       fn,
		Score:    0, // 排课时不跨中午,下午不会有奖励
		Penalty:  4, // 排课是跨中午,下午会有处罚
		Weight:   1,
		Priority: 1,
	}
}

// 加载规则
func loadTeacherNoonBreakConstraintsFromDB() []*TeacherNoonBreak {
	var constraints []*TeacherNoonBreak
	return constraints
}

// 生成规则校验方法
func (t *TeacherNoonBreak) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

		teacherID := t.TeacherID
		currTeacherID := element.TeacherID
		elementPeriods := types.GetElementPeriods(element, schedule)

		// 上午
		_, forenoonEndPeriod := schedule.GetPeriodWithRange("forenoon")
		// 下午
		afternoonStartPeriod, _ := schedule.GetPeriodWithRange("afternoon")

		periods := []int{forenoonEndPeriod, afternoonStartPeriod}
		intersect := lo.Intersect(elementPeriods, periods)
		isContain := len(intersect) > 0

		preCheckPassed := currTeacherID == teacherID && isContain
		isReward := false
		if preCheckPassed {
			isReward = !isTeacherInBothPeriods(element, teacherID, forenoonEndPeriod, afternoonStartPeriod, classMatrix, schedule)
		}

		return preCheckPassed, isReward, nil
	}
}

// 判断如果给当前元素的排课,是否会出现特定教师跨上午,下午排课
func isTeacherInBothPeriods(element types.Element, teacherID int, forenoonEndPeriod, afternoonStartPeriod int, classMatrix *types.ClassMatrix, schedule *models.Schedule) bool {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	elementDay := element.TimeSlots[0] / totalClassesPerDay
	dayPeriodCount := calcTeacherDayClasses(classMatrix, teacherID, schedule)

	// 判断元素所在的天,是否已经有排课
	if periods, ok := dayPeriodCount[elementDay]; ok {

		// 当前元素排课的节次
		elementPeriods := types.GetElementPeriods(element, schedule)

		// 判断当前元素的节次是上午最后一节,还是下午第一节
		// 如果是上午最后一节,则判断下午第一节,是否已经有排课
		if lo.Contains(elementPeriods, forenoonEndPeriod) && lo.Contains(periods, afternoonStartPeriod) {
			return true
		}

		// 如果是下午第1节,则判断上午最后一节,是否已经有排课
		if lo.Contains(elementPeriods, afternoonStartPeriod) && lo.Contains(periods, forenoonEndPeriod) {
			return true
		}
	}
	return false
}

// countTeacherPeriodClasses 计算老师目前每天的排课节数列表
func calcTeacherDayClasses(classMatrix *types.ClassMatrix, teacherID int, schedule *models.Schedule) map[int][]int {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// key: 天, val: 当前排课的节次列表
	dayPeriodCount := make(map[int][]int)

	// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
	for _, teacherMap := range classMatrix.Elements {
		for id, venueMap := range teacherMap {
			if teacherID == id {
				for _, timeSlotMap := range venueMap {
					for timeSlotStr, element := range timeSlotMap {

						timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
						for _, timeSlot := range timeSlots {
							if element.Val.Used == 1 {
								period := timeSlot % totalClassesPerDay
								day := timeSlot / totalClassesPerDay
								dayPeriodCount[day] = append(dayPeriodCount[day], period)
							}
						}
					}
				}
			}
		}
	}
	return dayPeriodCount
}
