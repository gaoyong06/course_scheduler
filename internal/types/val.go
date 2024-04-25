// val.go
package types

// 课表适应性矩阵元素值
type Val struct {
	ScoreInfo ScoreInfo `json:"score_info"` // 矩阵元素匹配分数明细
	Used      int       `json:"used"`       // 是否占用, 0 未占用, 1 已占用, 默认值: 0
}
