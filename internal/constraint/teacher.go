// teacher.go
// 教师固排禁排

package constraint

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"

	"github.com/samber/lo"
)

// ##### 教师固排禁排

// - 教师默认分组(按学科)
// - 教师自定义分组(行政领导)

// | 教师分组 | 教师   | 时间         | 限制 | 描述   |
// | -------- | ------ | ------------ | ---- | ------ |
// | 数学组   |        | 周一 第 4 节 | 禁排 | 教研会 |
// |          | 刘老师 | 周一 第 4 节 | 禁排 | 教研会 |
// | 行政组 |        | 周二 第 7 节 | 禁排 | 例会   |
// |          | 马老师 | 周二 第 7 节 | 禁排 | 例会   |
// |          | 王老师 | 周二 第 2 节 | 尽量排 |        |
// 教师教师固排禁排约束
type Teacher struct {
	ID             int    `json:"id" mapstructure:"id"`                             // 自增ID
	TeacherGroupID int    `json:"teacher_group_id" mapstructure:"teacher_group_id"` // 教师分组ID
	TeacherID      int    `json:"teacher_id" mapstructure:"teacher_id"`             // 老师ID
	TimeSlot       int    `json:"time_slot" mapstructure:"time_slot"`               // 时间段
	Limit          string `json:"limit" mapstructure:"limit"`                       // 限制: 固定排: fixed, 优先排: prefer, 禁排: not, 尽量不排: avoid
	Desc           string `json:"desc" mapstructure:"desc"`                         // 描述
}

// 生成字符串
func (t *Teacher) String() string {
	return fmt.Sprintf("ID: %d, TeacherGroupID: %d, TeacherID: %d, TimeSlot: %d, Limit: %s, Desc: %s", t.ID,
		t.TeacherGroupID, t.TeacherID, t.TimeSlot, t.Limit, t.Desc)
}

// 获取班级固排禁排规则
func GetTeacherRules(teachers []*models.Teacher, constraints []*Teacher) []*types.Rule {
	// constraints := loadTeacherConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule(teachers)
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (c *Teacher) genRule(teachers []*models.Teacher) *types.Rule {
	fn := c.genConstraintFn(teachers)
	return &types.Rule{
		Name:     "teacher",
		Type:     "fixed",
		Fn:       fn,
		Score:    c.getScore(),
		Penalty:  c.getPenalty(),
		Weight:   1,
		Priority: 1,
	}
}

// 加载班级固排禁排规则
func loadTeacherConstraintsFromDB() []*Teacher {
	var constraints []*Teacher
	return constraints
}

// 生成规则校验方法
func (t *Teacher) genConstraintFn(teachers []*models.Teacher) types.ConstraintFn {
	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {
		teacherGroupID := t.TeacherGroupID
		teacherID := t.TeacherID
		timeSlot := t.TimeSlot

		currTeacherID := element.GetTeacherID()
		currTimeSlot := element.GetTimeSlot()

		currTeacher, err := models.FindTeacherByID(currTeacherID, teachers)
		if err != nil {
			return false, false, err
		}

		preCheckPassed := false
		isReward := false

		// 固排,优先排是: 排了有奖励,不排有处罚
		if t.Limit == "fixed" || t.Limit == "prefer" {
			preCheckPassed := currTimeSlot == timeSlot
			isReward = preCheckPassed && (teacherGroupID == 0 || lo.Contains(currTeacher.TeacherGroupIDs, teacherGroupID)) && (teacherID == 0 || teacherID == currTeacherID)
		}

		// 禁排,尽量不排是: 不排没关系, 排了就处罚
		if t.Limit == "not" || t.Limit == "avoid" {
			preCheckPassed = currTimeSlot == timeSlot && (teacherGroupID == 0 || lo.Contains(currTeacher.TeacherGroupIDs, teacherGroupID)) && (teacherID == 0 || teacherID == currTeacherID)
			isReward = false
		}
		return preCheckPassed, isReward, nil
	}
}

// 奖励分
func (t *Teacher) getScore() int {
	score := 0
	if t.Limit == "fixed" {
		score = 3
	} else if t.Limit == "prefer" {
		score = 2
	}
	return score
}

// 惩罚分
func (t *Teacher) getPenalty() int {
	penalty := 0
	if t.Limit == "not" {
		penalty = 3
	} else if t.Limit == "avoid" {
		penalty = 2
	}
	return penalty
}
