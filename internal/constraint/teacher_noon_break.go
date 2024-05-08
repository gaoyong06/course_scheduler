// 教师不跨中午(教师排了上午最后一节就不排下午第一节)

package constraint

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
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
func GetTeacherNoonBreakRules() []*types.Rule {
	constraints := loadTeacherNoonBreakConstraintsFromDB()
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
func loadTeacherNoonBreakConstraintsFromDB() []*TeacherNoonBreak {
	var constraints []*TeacherNoonBreak
	return constraints
}

// 生成规则校验方法
func (t *TeacherNoonBreak) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

		teacherID := t.TeacherID
		currTeacherID := element.GetTeacherID()
		preCheckPassed := currTeacherID == teacherID

		shouldPenalize := false
		if preCheckPassed {

			// 上午
			_, forenoonEndPeriod := schedule.GetPeriodWithRange("forenoon")
			// 下午
			afternoonStartPeriod, _ := schedule.GetPeriodWithRange("afternoon")
			shouldPenalize = isTeacherInBothPeriods(teacherID, forenoonEndPeriod, afternoonStartPeriod, classMatrix)
		}

		return preCheckPassed, !shouldPenalize, nil
	}
}

// 判断教师是否在两个节次都有课
func isTeacherInBothPeriods(teacherID int, period1, period2 int, classMatrix *types.ClassMatrix) bool {
	for _, classMap := range classMatrix.Elements {
		for id, teacherMap := range classMap {
			if id == teacherID {
				for _, periodMap := range teacherMap {
					if element1, ok := periodMap[period1]; ok && element1.Val.Used == 1 {
						if element2, ok := periodMap[period2]; ok && element2.Val.Used == 1 {
							return true
						}
					}
				}
			}
		}
	}
	return false
}
