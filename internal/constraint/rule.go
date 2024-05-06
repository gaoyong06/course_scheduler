// rule.go
package constraint

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"sort"
)

// 所有固定约束条件
func GetFixedRules(subjects []*models.Subject, teachers []*models.Teacher) []*types.Rule {

	var rules []*types.Rule

	// 班级固排禁排
	classRules := GetClassRules()
	rules = append(rules, classRules...)

	// 科目优先排禁排
	subjectRules := GetSubjectRules(subjects)
	rules = append(rules, subjectRules...)

	// 教师固排禁排
	teacherRules := GetTeacherRules(teachers)
	rules = append(rules, teacherRules...)

	sortRulesByPriority(rules)
	return rules
}

// 所有动态约束条件
func GetDynamicRules(schedule *models.Schedule) []*types.Rule {

	var rules []*types.Rule

	// 连堂课校验(科目课时数大于上课天数时, 禁止同一天排多次课是非连续的, 要排成连堂课)
	rules = append(rules, subjectConnectedRule)

	// 科目互斥限制(科目A与科目B不排在同一天)
	subjectMutexRules := GetSubjectMutexRules()
	rules = append(rules, subjectMutexRules...)

	// 科目顺序限制(体育课不排在数学课前)
	subjectOrderRules := GetSubjectOrderRules()
	rules = append(rules, subjectOrderRules...)

	// 同一个年级,班级,科目相同节次的排课是否超过数量限制
	rules = append(rules, subjectPeriodLimitRule)

	// 科目课时小于天数,禁止同一天排多次相同科目的课
	rules = append(rules, subjectSameDayRule)

	// 教师互斥限制
	teacherMutexRules := GetTeacherMutexRules()
	rules = append(rules, teacherMutexRules...)

	// 教师不跨中午约束
	teacherNoonBreakRules := GetTeacherNoonBreakRules(schedule)
	rules = append(rules, teacherNoonBreakRules...)

	// 教师节数限制
	teacherClassLimitRules := GetTeacherClassLimitRules()
	rules = append(rules, teacherClassLimitRules...)

	// 教师时间段限制
	teacherRangeLimitRules := GetTeacherRangeLimitRules(schedule)
	rules = append(rules, teacherRangeLimitRules...)

	sortRulesByPriority(rules)
	return rules
}

// 获取元素最大得分
func GetMaxElementScore(schedule *models.Schedule, subjects []*models.Subject, teachers []*models.Teacher) int {

	rules := append(GetFixedRules(subjects, teachers), GetDynamicRules(schedule)...)
	maxScore := 0
	for _, rule := range rules {
		maxScore += rule.Score * rule.Weight
	}
	return maxScore
}

// 获取元素最小得分
func GetMinElementScore(schedule *models.Schedule, subjects []*models.Subject, teachers []*models.Teacher) int {

	rules := append(GetFixedRules(subjects, teachers), GetDynamicRules(schedule)...)
	minScore := 0

	for _, rule := range rules {
		minScore -= rule.Penalty * rule.Weight
	}

	return minScore
}

// =================================================

// SortRulesByPriority sorts rules by their priority
func sortRulesByPriority(rules []*types.Rule) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority // Higher priority first
	})
}
