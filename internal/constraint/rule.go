// rule.go
package constraint

import (
	"course_scheduler/internal/types"
	"sort"
)

// 所有固定约束条件
func GetFixedRules() []*types.Rule {

	// var rules []*Rule
	rules := []*types.Rule{
		// CRule1,
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

	sortRulesByPriority(rules)

	return rules
}

// 所有动态约束条件
func GetDynamicRules() []*types.Rule {

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

		// TMERule1,
		// TMERule2,

		TNANRule1,
		TNANRule2,

		TTLRule1,
		TTLRule2,
		TTLRule3,
	}

	sortRulesByPriority(rules)

	return rules
}

// =================================================

// SortRulesByPriority sorts rules by their priority
func sortRulesByPriority(rules []*types.Rule) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority // Higher priority first
	})
}
