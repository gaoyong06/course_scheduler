// class.go
// 班级固排禁排

package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"

	"github.com/samber/lo"
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
	TimeSlots []int  `json:"time_slots" mapstructure:"time_slots"`                     // 时间段与时间点
	Limit     string `json:"limit" mapstructure:"limit"`                               // 限制: 固定排课: fixed, 尽量排: prefer, 尽量不排课: avoid, 禁止排课: not
	Desc      string `json:"desc" mapstructure:"desc"`                                 // 描述
}

// 生成字符串
func (c *Class) String() string {
	return fmt.Sprintf("ID: %d, GradeID: %d, ClassID: %d, SubjectID: %d, TeacherID: %d, TimeSlots: %v, Limit: %s, Desc: %s", c.ID,
		c.GradeID, c.ClassID, c.SubjectID, c.TeacherID, c.TimeSlots, c.Limit, c.Desc)
}

// 获取班级固排禁排规则
func GetClassRules(constraints []*Class) []*types.Rule {
	// constraints := loadClassConstraintsFromDB()
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
		Name:     "class",
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
	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

		SN, err := types.ParseSN(element.ClassSN)
		if err != nil {
			return false, false, err
		}

		preCheckPassed := false
		isReward := false

		// 当前时间段,是否包含在约束时间段内
		intersect := lo.Intersect(c.TimeSlots, element.TimeSlots)
		isContain := len(intersect) > 0

		// 固排,优先排是: 排了有奖励,不排有处罚
		if c.Limit == "fixed" || c.Limit == "prefer" {

			preCheckPassed = c.GradeID == SN.GradeID && (c.ClassID == 0 || c.ClassID == SN.ClassID) && isContain
			isReward = preCheckPassed && (c.SubjectID == 0 || c.SubjectID == element.SubjectID) &&
				(c.TeacherID == 0 || c.TeacherID == element.TeacherID)
		}

		// 禁排,尽量不排是: 不排没关系, 排了就处罚
		if c.Limit == "not" || c.Limit == "avoid" {

			preCheckPassed = c.GradeID == SN.GradeID && (c.ClassID == 0 || c.ClassID == SN.ClassID) && isContain && (c.SubjectID == 0 || c.SubjectID == element.SubjectID) &&
				(c.TeacherID == 0 || c.TeacherID == element.TeacherID)
			isReward = false
		}

		// if element.ClassSN == "1_9_1" && c.GradeID == 9 && c.ClassID == 1 && c.SubjectID == 1 {
		// 	fmt.Printf("class constraint, sn: %s, timeSlots: %v, limit: %s,  preCheckPassed: %v, isReward: %v\n", element.ClassSN, element.TimeSlots, c.Limit, preCheckPassed, isReward)
		// }

		return preCheckPassed, isReward, nil
	}
}

// 奖励分
func (c *Class) getScore() int {
	score := 0
	if c.Limit == "fixed" {
		score = 6
	} else if c.Limit == "prefer" {
		score = 4
	}
	return score
}

// 惩罚分
func (c *Class) getPenalty() int {
	penalty := 0
	if c.Limit == "not" {
		penalty = 6
	} else if c.Limit == "avoid" {
		penalty = 4
	}
	return penalty
}
