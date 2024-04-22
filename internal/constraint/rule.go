// rule.go
package constraint

import (
	"course_scheduler/internal/types"
	"sort"
)

// CalcFixed 计算固定约束条件得分
func CalcFixed(classMatrix map[string]map[int]map[int]map[int]types.Val, element *types.Element) (int, error) {

	rules := getFixedRules()
	sortRulesByPriority(rules)
	score := 0
	penalty := 0
	for _, rule := range rules {
		if rule.Type == "fixed" {
			if preCheckPassed, result, err := rule.Fn(classMatrix, element); preCheckPassed && err == nil {

				tempVal := classMatrix[element.ClassSN][element.TeacherID][element.VenueID][element.TimeSlot]

				if result {
					score += rule.Score * rule.Weight
					tempVal.ScoreInfo.Passed = append(tempVal.ScoreInfo.Passed, rule)
				} else {
					penalty += rule.Penalty * rule.Weight
					tempVal.ScoreInfo.Failed = append(tempVal.ScoreInfo.Failed, rule)
				}
				classMatrix[element.ClassSN][element.TeacherID][element.VenueID][element.TimeSlot] = tempVal
			}
		}
	}

	finalScore := score - penalty
	return finalScore, nil
}

// CalcDynamic 计算动态约束条件得分
func CalcDynamic(classMatrix map[string]map[int]map[int]map[int]types.Val, element *types.Element) (int, error) {

	rules := getDynamicRules()
	sortRulesByPriority(rules)
	score := 0
	penalty := 0
	for _, rule := range rules {
		if rule.Type == "dynamic" {
			if preCheckPassed, result, err := rule.Fn(classMatrix, element); preCheckPassed && err == nil {

				tempVal := classMatrix[element.ClassSN][element.TeacherID][element.VenueID][element.TimeSlot]
				if result {
					score += rule.Score * rule.Weight
					tempVal.ScoreInfo.Passed = append(tempVal.ScoreInfo.Passed, rule)
				} else {
					penalty += rule.Penalty * rule.Weight
					tempVal.ScoreInfo.Failed = append(tempVal.ScoreInfo.Failed, rule)
				}
				classMatrix[element.ClassSN][element.TeacherID][element.VenueID][element.TimeSlot] = tempVal
			}
		}
	}
	finalScore := score - penalty
	return finalScore, nil
}

// 计算固定约束得分和动态约束得分
func CalcScore(classMatrix map[string]map[int]map[int]map[int]types.Val, element *types.Element) (int, error) {

	score1, err1 := CalcFixed(classMatrix, element)
	if err1 != nil {
		return 0, err1
	}

	score2, err2 := CalcDynamic(classMatrix, element)
	if err2 != nil {
		return 0, err1
	}

	score := score1 + score2

	// 更新val
	tempVal := classMatrix[element.ClassSN][element.TeacherID][element.VenueID][element.TimeSlot]
	tempVal.ScoreInfo.Score = score
	classMatrix[element.ClassSN][element.TeacherID][element.VenueID][element.TimeSlot] = tempVal

	return score, nil
}

// // 实时更新动态约束条件得分
// func UpdateDynamic(rules []*Rule, params Element) {

// 	sortRulesByPriority(rules)
// 	// 实际更新逻辑
// }

// =================================================

// SortRulesByPriority sorts rules by their priority
func sortRulesByPriority(rules []*types.Rule) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority // Higher priority first
	})
}

// 所有固定约束条件
func getFixedRules() []*types.Rule {

	// var rules []*Rule
	rules := []*types.Rule{
		CRule1,
		CRule2,
		CRule3,
		CRule4,
		CRule5,
		CRule6,
		CRule7,
		CRule8,

		SRule1,
		SRule2,
		SRule3,
		SRule4,

		TRule1,
		TRule2,
		TRule3,
		TRule4,
		TRule5,
	}

	return rules
}

// 所有动态约束条件
func getDynamicRules() []*types.Rule {

	// var rules []*Rule

	rules := []*types.Rule{
		PCRule1,

		SCRule1,

		SMERule1,

		SORule1,

		SSDRule1,

		TCLRule1,
		TCLRule2,
		TCLRule3,
		TCLRule4,

		TMERule1,
		TMERule2,

		TNANRule1,
		TNANRule2,

		TTLRule1,
		TTLRule2,
		TTLRule3,
	}

	return rules
}
