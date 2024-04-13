package types

type Val struct {
	Score int // 匹配结果值, 匹配结果值越大越好，匹配结果值为“-1”表示课班不可用当前适应性矩阵的元素下标，默认值: 0
	Used  int // 是否占用, 0 未占用, 1 已占用, 默认值: 0
}
