package config

const (
	PopSize       = 100 // 种群规模 20-100
	SelectionSize = 50
	MaxGen        = 100 // 遗传代数 100-500
	MutationRate  = 0.1 // 变异率 0.001-0.05
	CrossoverRate = 1.0 // 交叉率 0.4~0.9
	BestRatio     = 0.1 // 选择最近个体百分比
)
