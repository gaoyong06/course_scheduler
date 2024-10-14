package types

import "course_scheduler/internal/models"

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
type ConstraintFn func(classMatrix *ClassMatrix, element Element, schedule *models.Schedule, teachingTasks []*models.TeachingTask) (bool, bool, error)

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
