// config.go
package config

import "time"

// 遗传算法常量
const (
	PopSize       = 100 // 种群规模 20-100
	SelectionSize = 50  // 选择操作 个体数量, 重要！重要！重要！选择个体数量需要个种群规模的一半
	MaxGen        = 100 // 遗传代数 100-500
	// PopSize       = 50  // 种群规模 20-100
	// SelectionSize = 50  // 选择操作 个体数量 选择的个体是原个体数量的一半
	// MaxGen        = 50  // 最大遗传代数 100-500
	MaxStagnGen   = 50                 // 最大停滞代数（连续n代没有改进, 当达到这个停滞代数时算法会停止运行）
	MutationRate  = 0.05               // 变异率 0.001-0.05
	CrossoverRate = 0.9                // 交叉率 0.4~0.9
	BestRatio     = 0.05               // 选择最佳个体百分比
	TargetFitness = 1000               // 视排课为最大话问题,适应度值值越高质量最优, 需通过实验调整这个值，以找到一个最优的退出条件。但是TargetFitness 的值不能太大，否则可能会导致算法无法收敛，当种群中的某个个体达到或超过这个值时，算法会停止运行并输出结果
	MaxDuration   = 3600 * time.Second // 排课的最长运行时间限制, 60分钟
)

const (
	SubjectPeriodLimitThreshold = 3 // 相同节次排课数量限制
	SubjectDayLimitThreshold    = 2 // 相同节次排课数量限制
)

const (
	MaxPenaltyScore = 3 // 表示ClassMatrix中的元素可以具有的最大可能得分, 这个得分很重要,会直接影响适应度计算的结果, 一般和最高的奖励分是相同的

	MaxRetries = 6 // 创建个体的最大重试次数
)

// 排课优先级
const (
	Fixed  = "fixed"  // 固定排课
	Prefer = "prefer" // 优先排课(尽量排课)
	Not    = "not"    // 禁止排课
	Avoid  = "avoid"  // 尽量不排课
	Min    = "min"    // 最少排课count节
	Max    = "max"    // 最多排课count节
)
