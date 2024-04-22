// rule.go
package constraint

import (
	"course_scheduler/internal/types"
	"sort"
)

// 课班适应性矩阵中的一个元素
type Element struct {
	ClassMatrix map[string]map[int]map[int]map[int]types.Val
	ClassSN     string
	TeacherID   int
	VenueID     int
	TimeSlot    int
}

// Rule 表示评分规则
type Rule struct {
	Name     string                                    // 规则名称
	Type     string                                    // 固定约束条件: fixed, 动态约束条件: dynamic
	Fn       func(element Element) (bool, bool, error) // 约束条件处理方法 约束条件前置检查方法，返回值 true: 满足前置条件, false: 不满足前置条件, 返回值 true: 满足约束,增加score, false: 不满足约束,增加penalty
	Score    int                                       // 得分
	Penalty  int                                       // 惩罚分
	Weight   int                                       // 权重
	Priority int                                       // 优先级
}

// CalcFixed 计算固定约束条件得分
func CalcFixed(rules []*Rule, params Element) (int, int, error) {

	sortRulesByPriority(rules)
	score := 0
	penalty := 0
	for _, rule := range rules {
		if rule.Type == "fixed" {
			if preCheckPassed, result, err := rule.Fn(params); preCheckPassed && err == nil {

				if result {
					score += rule.Score
				} else {
					penalty += rule.Penalty
				}
			}
		}
	}
	return score, penalty, nil
}

// CalcDynamic 计算动态约束条件得分
func CalcDynamic(rules []*Rule, params Element) (int, int, error) {

	sortRulesByPriority(rules)
	score := 0
	penalty := 0
	for _, rule := range rules {
		if rule.Type == "dynamic" {
			if preCheckPassed, result, err := rule.Fn(params); preCheckPassed && err == nil {
				if result {
					score += rule.Score
				} else {
					penalty += rule.Penalty
				}
			}
		}
	}
	return score, penalty, nil
}

// 实时更新动态约束条件得分
func UpdateDynamic(rules []*Rule, params Element) {

	sortRulesByPriority(rules)
	// 实际更新逻辑
}

// 分数计算
func CalcScore(rules []*Rule, params Element) int {
	sortRulesByPriority(rules)
	score, penalty := 0, 0
	for _, rule := range rules {

		// 检查是否满足约束规则前置条件
		preCheckPassed, result, err := rule.Fn(params)
		if !preCheckPassed {
			continue
		}

		// 计算得分和惩罚
		if result && err == nil {
			score += rule.Weight * rule.Score
		} else {
			penalty += rule.Weight * rule.Penalty
		}
	}
	finalScore := score - penalty
	return finalScore
}

// SortRulesByPriority sorts rules by their priority
func sortRulesByPriority(rules []*Rule) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority // Higher priority first
	})
}
