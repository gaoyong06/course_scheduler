// rule.go
package constraint

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"sort"
)

// 所有固定约束条件
// func GetFixedRules(subjects []*models.Subject, teachers []*models.Teacher, classConstraints []*Class, subjectConstraints []*Subject, teacherConstraints []*Teacher) []*types.Rule {

// 	var rules []*types.Rule

// 	// 班级固排禁排
// 	classRules := GetClassRules(classConstraints)
// 	rules = append(rules, classRules...)

// 	// 科目优先排禁排
// 	subjectRules := GetSubjectRules(subjects, subjectConstraints)
// 	rules = append(rules, subjectRules...)

// 	// 教师固排禁排
// 	teacherRules := GetTeacherRules(teachers, teacherConstraints)
// 	rules = append(rules, teacherRules...)

// 	sortRulesByPriority(rules)
// 	return rules
// }

func GetFixedRules(subjects []*models.Subject, teachers []*models.Teacher, constraints map[string]interface{}) []*types.Rule {

	var rules []*types.Rule

	// 遍历不同类型的约束条件
	for constraintType, constraintValue := range constraints {
		switch constraintType {
		case "Class":
			classConstraints := constraintValue.([]*Class)
			classRules := GetClassRules(classConstraints)
			rules = append(rules, classRules...)
		case "Subject":
			subjectConstraints := constraintValue.([]*Subject)
			subjectRules := GetSubjectRules(subjects, subjectConstraints)
			rules = append(rules, subjectRules...)
		case "Teacher":
			teacherConstraints := constraintValue.([]*Teacher)
			teacherRules := GetTeacherRules(teachers, teacherConstraints)
			rules = append(rules, teacherRules...)
		}
	}

	sortRulesByPriority(rules)
	return rules
}

// 所有动态约束条件
// func GetDynamicRules(schedule *models.Schedule, constraints1 []*SubjectMutex, constraints2 []*SubjectOrder, constraints3 []*TeacherMutex, constraints4 []*TeacherNoonBreak, constraints5 []*TeacherPeriodLimit, constraints6 []*TeacherRangeLimit) []*types.Rule {

// 	var rules []*types.Rule

// 	// 连堂课校验(科目课时数大于上课天数时, 禁止同一天排多次课是非连续的, 要排成连堂课)
// 	rules = append(rules, subjectConnectedRule)

// 	// 科目互斥限制(科目A与科目B不排在同一天)
// 	subjectMutexRules := GetSubjectMutexRules(constraints1)
// 	rules = append(rules, subjectMutexRules...)

// 	// 科目顺序限制(体育课不排在数学课前)
// 	subjectOrderRules := GetSubjectOrderRules(constraints2)
// 	rules = append(rules, subjectOrderRules...)

// 	// 同一个年级,班级,科目相同节次的排课是否超过数量限制
// 	rules = append(rules, subjectPeriodLimitRule)

// 	// 科目课时小于天数,禁止同一天排多次相同科目的课
// 	rules = append(rules, subjectSameDayRule)

// 	// 教师互斥限制
// 	teacherMutexRules := GetTeacherMutexRules(constraints3)
// 	rules = append(rules, teacherMutexRules...)

// 	// 教师不跨中午约束
// 	teacherNoonBreakRules := GetTeacherNoonBreakRules(constraints4)
// 	rules = append(rules, teacherNoonBreakRules...)

// 	// 教师节数限制
// 	teacherClassLimitRules := GetTeacherClassLimitRules(constraints5)
// 	rules = append(rules, teacherClassLimitRules...)

// 	// 教师时间段限制
// 	teacherRangeLimitRules := GetTeacherRangeLimitRules(constraints6)
// 	rules = append(rules, teacherRangeLimitRules...)

// 	sortRulesByPriority(rules)
// 	return rules
// }

// 所有动态约束条件
func GetDynamicRules(schedule *models.Schedule, constraints map[string]interface{}) []*types.Rule {

	var rules []*types.Rule

	// 下面是一些内部规则
	// 连堂课校验(科目课时数大于上课天数时, 禁止同一天排多次课是非连续的, 要排成连堂课)
	rules = append(rules, subjectConnectedRule)

	// 同一个年级,班级,科目相同节次的排课是否超过数量限制
	rules = append(rules, subjectPeriodLimitRule)

	// 科目课时小于天数,禁止同一天排多次相同科目的课
	rules = append(rules, subjectSameDayRule)

	for constraintType, constraintValue := range constraints {
		switch constraintType {
		case "SubjectMutex":

			// 科目互斥限制(科目A与科目B不排在同一天)
			subjectMutexConstraints := constraintValue.([]*SubjectMutex)
			rules = append(rules, GetSubjectMutexRules(subjectMutexConstraints)...)

		case "SubjectOrder":

			// 科目顺序限制(体育课不排在数学课前)
			subjectOrderConstraints := constraintValue.([]*SubjectOrder)
			rules = append(rules, GetSubjectOrderRules(subjectOrderConstraints)...)

		case "TeacherMutex":

			// 教师互斥限制
			teacherMutexConstraints := constraintValue.([]*TeacherMutex)
			rules = append(rules, GetTeacherMutexRules(teacherMutexConstraints)...)

		case "TeacherNoonBreak":

			// 教师不跨中午约束
			teacherNoonBreakConstraints := constraintValue.([]*TeacherNoonBreak)
			rules = append(rules, GetTeacherNoonBreakRules(teacherNoonBreakConstraints)...)

		case "TeacherPeriodLimit":

			// 教师节数限制
			teacherPeriodLimitConstraints := constraintValue.([]*TeacherPeriodLimit)
			rules = append(rules, GetTeacherClassLimitRules(teacherPeriodLimitConstraints)...)

		case "TeacherRangeLimit":

			// 教师时间段限制
			teacherRangeLimitConstraints := constraintValue.([]*TeacherRangeLimit)
			rules = append(rules, GetTeacherRangeLimitRules(teacherRangeLimitConstraints)...)
		}
	}

	sortRulesByPriority(rules)
	return rules
}

// 获取所有元素的最大得分
func GetElementsMaxScore(schedule *models.Schedule, subjects []*models.Subject, teachers []*models.Teacher, constraints map[string]interface{}) int {

	fmt.Println("获取所有元素的最大得分")
	rules := append(GetFixedRules(subjects, teachers, constraints), GetDynamicRules(schedule, constraints)...)
	maxScore := 0
	for _, rule := range rules {
		maxScore += rule.Score * rule.Weight
	}
	return maxScore
}

// 获取所有元素的最小得分
func GetElementsMinScore(schedule *models.Schedule, subjects []*models.Subject, teachers []*models.Teacher, constraints map[string]interface{}) int {

	fmt.Println("获取所有元素的最小得分")
	rules := append(GetFixedRules(subjects, teachers, constraints), GetDynamicRules(schedule, constraints)...)
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
