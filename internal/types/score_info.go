// score_info.go
package types

// ScoreInfo 存储课班适应性矩阵元素的得分详情
type ScoreInfo struct {
	Score         int      // 最终得分 值越大越好, 默认值: 0
	FixedScore    int      // 固定约束条件得分
	DynamicScore  int      // 动态约束条件得分
	FixedPassed   []string // 满足的固定约束条件
	FixedFailed   []string // 未满足的固定约束条件
	DynamicPassed []string // 满足的动态约束条件
	DynamicFailed []string // 未满足的动态约束条件
}
