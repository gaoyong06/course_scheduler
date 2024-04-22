// score_info.go
package types

// ScoreInfo 存储课班适应性矩阵元素的得分详情
type ScoreInfo struct {
	Score  int     // 最终得分 值越大越好, 默认值: 0
	Passed []*Rule // 满足的约束条件
	Failed []*Rule // 未满足的约束条件
}
