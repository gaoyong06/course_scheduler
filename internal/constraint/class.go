// class.go
// 班级固排禁排

package constraint

import (
	"course_scheduler/internal/types"
	"fmt"
)

// ##### 班级固排禁排

// | 年级   | 班级 | 科目 | 老师   | 时间    | 限制   | 描述 |
// | ------ | ---- | ---- | ------ | ------- | ------ | ---- |
// | 一年级 | 1 班 | 语文 | 王老师 | 第 1 节 | 固排   |      |
// | 三年级 | 1 班 |      |        | 第 7 节 | 禁排   | 班会 |
// | 三年级 | 2 班 |      |        | 第 8 节 | 禁排   | 班会 |
// | 四年级 |      |      |        | 第 8 节 | 禁排   | 班会 |
// | 四年级 | 1 班 | 语文 | 王老师 | 第 1 节 | 禁排   |      |
// | 五年级 |      | 数学 | 李老师 | 第 2 节 | 固排   |      |
// | 五年级 |      | 数学 | 李老师 | 第 3 节 | 尽量排 |      |
// | 五年级 |      | 数学 | 李老师 | 第 5 节 | 固排   |      |
type Class struct {
	ID        int    `json:"id" mapstructure:"id"`                                     // 自增ID
	GradeID   int    `json:"grade_id" mapstructure:"grade_id"`                         // 年级ID
	ClassID   int    `json:"class_id" mapstructure:"class_id"`                         // 班级ID, 可以为空
	SubjectID int    `json:"subject_id,omitempty" mapstructure:"subject_id,omitempty"` // 科目ID, 可以为空
	TeacherID int    `json:"teacher_id,omitempty" mapstructure:"teacher_id,omitempty"` // 老师ID, 可以为空
	TimeSlot  int    `json:"time_slot" mapstructure:"time_slot"`                       // 时间段与时间点
	Limit     string `json:"limit" mapstructure:"limit"`                               // 限制: 固定排课: fixed, 尽量排: prefer, 禁止排课: not
	Desc      string `json:"desc" mapstructure:"desc"`                                 // 描述
}

// 生成字符串
func (c *Class) String() string {
	return fmt.Sprintf("ID: %d, GradeID: %d, ClassID: %d, SubjectID: %d, TeacherID: %d, TimeSlot: %d, Limit: %s, Desc: %s", c.ID,
		c.GradeID, c.ClassID, c.SubjectID, c.TeacherID, c.TimeSlot, c.Limit, c.Desc)
}

// 获取班级固排禁排规则
func GetClassRules() []*types.Rule {
	constraints := loadClassConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule()
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (c *Class) genRule() *types.Rule {
	fn := c.genConstraintFn()
	return &types.Rule{
		Name:     c.String(),
		Type:     "fixed",
		Fn:       fn,
		Score:    c.getScore(),
		Penalty:  c.getPenalty(),
		Weight:   1,
		Priority: 1,
	}
}

// 加载班级固排禁排规则
func loadClassConstraintsFromDB() []*Class {
	var constraints []*Class
	return constraints
}

// 生成规则校验方法
func (c *Class) genConstraintFn() types.ConstraintFn {
	return func(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {
		preCheckPassed := c.preCheck(element)
		isValid := c.constraintCheck(element)
		return preCheckPassed, isValid, nil
	}
}

// 判断前置条件是否成立
func (c *Class) preCheck(element types.Element) bool {
	SN, err := types.ParseSN(element.ClassSN)
	if err != nil {
		return false
	}
	return c.GradeID == SN.GradeID && (c.ClassID == 0 || c.ClassID == SN.ClassID) && c.TimeSlot == element.TimeSlot
}

// 判断是否符合约束条件
func (c *Class) constraintCheck(element types.Element) bool {
	return (c.SubjectID == 0 || c.SubjectID == element.SubjectID) &&
		(c.TeacherID == 0 || c.TeacherID == element.TeacherID) &&
		(c.Limit == "fixed" || c.Limit == "prefer")
}

// 奖励分
func (c *Class) getScore() int {
	score := 0
	if c.Limit == "fixed" {
		score = 3
	} else if c.Limit == "prefer" {
		score = 2
	}
	return score
}

// 惩罚分
func (c *Class) getPenalty() int {
	penalty := 0
	if c.Limit == "not" {
		penalty = 3
	} else if c.Limit == "avoid" {
		penalty = 2
	}
	return penalty
}
