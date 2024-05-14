// 教师不跨中午(教师排了上午最后一节就不排下午第一节)

package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
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
	TeacherID int `json:"teacher_id" mapstructure:"teacher_id"` // 教师ID
}

// 生成字符串
func (t *TeacherNoonBreak) String() string {
	return fmt.Sprintf("TeacherID: %d", t.TeacherID)
}

// 获取班级固排禁排规则
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
		Score:    0,
		Penalty:  1,
		Weight:   1,
		Priority: 1,
	}
}

// 加载班级固排禁排规则
func loadTeacherNoonBreakConstraintsFromDB() []*TeacherNoonBreak {
	var constraints []*TeacherNoonBreak
	return constraints
}

// 生成规则校验方法
func (t *TeacherNoonBreak) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

		totalClassesPerDay := schedule.GetTotalClassesPerDay()

		teacherID := t.TeacherID
		currTeacherID := element.TeacherID
		currTimeSlot := element.TimeSlot
		currPeriod := currTimeSlot % totalClassesPerDay

		// 上午
		_, forenoonEndPeriod := schedule.GetPeriodWithRange("forenoon")
		// 下午
		afternoonStartPeriod, _ := schedule.GetPeriodWithRange("afternoon")
		preCheckPassed := currTeacherID == teacherID && (currPeriod == forenoonEndPeriod || currPeriod == afternoonStartPeriod)
		isReward := false
		if preCheckPassed {
			isReward = !isTeacherInBothPeriods(element, teacherID, forenoonEndPeriod, afternoonStartPeriod, classMatrix, schedule)
		}

		return preCheckPassed, isReward, nil
	}
}

// 判断教师是否在两个节次都有课
func isTeacherInBothPeriods(element types.Element, teacherID int, period1, period2 int, classMatrix *types.ClassMatrix, schedule *models.Schedule) bool {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	elementDay := element.TimeSlot / totalClassesPerDay
	dayPeriodCount := calcTeacherDayClasses(classMatrix, teacherID, schedule)

	if periods, ok := dayPeriodCount[elementDay]; ok {
		if lo.Contains(periods, period1) && lo.Contains(periods, period2) {
			return true
		}
	}
	return false
}

// countTeacherPeriodClasses 计算老师每天的排课节数列表
func calcTeacherDayClasses(classMatrix *types.ClassMatrix, teacherID int, schedule *models.Schedule) map[int][]int {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// key: 天, val: 节次列表
	dayPeriodCount := make(map[int][]int)

	// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
	for _, teacherMap := range classMatrix.Elements {
		for id, venueMap := range teacherMap {
			if teacherID == id {
				for _, timeSlotMap := range venueMap {
					for timeSlot, element := range timeSlotMap {
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
	return dayPeriodCount
}
