// config.go
package config

const (
	PopSize       = 1    // 种群规模 20-100
	SelectionSize = 1    // 选择操作 个体数量
	MaxGen        = 1    // 遗传代数 100-500
	MutationRate  = 0.1  // 变异率 0.001-0.05
	CrossoverRate = 0.9  // 交叉率 0.4~0.9
	BestRatio     = 0.02 // 选择最佳个体百分比
)
