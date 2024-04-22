// rule.go
package constraint

import (
	"course_scheduler/internal/types"
	"sort"
)

// 课班适应性矩阵中的一个元素
type Element struct {
	ClassSN   string // 科目_年级_班级
	SubjectID int    // 科目
	GradeID   int    // 年级
	ClassID   int    // 班级
	TeacherID int    // 教室
	VenueID   int    // 教室
	TimeSlot  int    // 时间段
}

// 约束处理函数
// 参数
//
//	ClassMatrix 课班适应性矩阵
//	element 课班适应性矩阵元素
//
// 返回值
//
//	bool true: 满足前置条件, false: 不满足前置条件
//	bool true: 满足约束,增加score, false: 不满足约束,增加penalty
//	error 错误信息
type ConstraintFn func(classMatrix map[string]map[int]map[int]map[int]types.Val, element Element) (bool, bool, error)

// Rule 表示评分规则
type Rule struct {
	Name     string       // 规则名称
	Type     string       // 固定约束条件: fixed, 动态约束条件: dynamic
	Fn       ConstraintFn // 约束条件处理方法 约束条件前置检查方法，返回值 true: 满足前置条件, false: 不满足前置条件, 返回值 true: 满足约束,增加score, false: 不满足约束,增加penalty
	Score    int          // 得分
	Penalty  int          // 惩罚分
	Weight   int          // 权重
	Priority int          // 优先级
}

// CalcFixed 计算固定约束条件得分
func CalcFixed(classMatrix map[string]map[int]map[int]map[int]types.Val, element Element) (int, error) {

	rules := getFixedRules()
	sortRulesByPriority(rules)
	score := 0
	penalty := 0
	for _, rule := range rules {
		if rule.Type == "fixed" {
			if preCheckPassed, result, err := rule.Fn(classMatrix, element); preCheckPassed && err == nil {

				if result {
					score += rule.Score * rule.Weight
				} else {
					penalty += rule.Penalty * rule.Weight
				}
			}
		}
	}
	finalScore := score - penalty
	return finalScore, nil
}

// CalcDynamic 计算动态约束条件得分
func CalcDynamic(classMatrix map[string]map[int]map[int]map[int]types.Val, element Element) (int, error) {

	rules := getDynamicRules()
	sortRulesByPriority(rules)
	score := 0
	penalty := 0
	for _, rule := range rules {
		if rule.Type == "dynamic" {
			if preCheckPassed, result, err := rule.Fn(classMatrix, element); preCheckPassed && err == nil {
				if result {
					score += rule.Score
				} else {
					penalty += rule.Penalty
				}
			}
		}
	}
	finalScore := score - penalty
	return finalScore, nil
}

// 计算固定约束得分和动态约束得分
func CalcScore(classMatrix map[string]map[int]map[int]map[int]types.Val, element Element) (int, error) {

	fixedScore, err1 := CalcFixed(classMatrix, element)
	if err1 != nil {
		return 0, nil
	}

	dynamicScore, err2 := CalcDynamic(classMatrix, element)
	if err2 != nil {
		return 0, nil
	}

	finalScore := fixedScore + dynamicScore
	return finalScore, nil
}

// // 实时更新动态约束条件得分
// func UpdateDynamic(rules []*Rule, params Element) {

// 	sortRulesByPriority(rules)
// 	// 实际更新逻辑
// }

// SortRulesByPriority sorts rules by their priority
func sortRulesByPriority(rules []*Rule) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority // Higher priority first
	})
}

// 所有固定约束条件
func getFixedRules() []*Rule {

	// var rules []*Rule
	rules := []*Rule{
		// CRule1,
		// CRule2,
		// ...
	}

	return rules
}

// 所有动态约束条件
func getDynamicRules() []*Rule {
	var rules []*Rule
	return rules
}
