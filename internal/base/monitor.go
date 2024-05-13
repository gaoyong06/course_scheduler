package base

import (
	"fmt"
	"log"
	"sort"
	"time"
)

// 遗传算法排课程序的监控结构体
type Monitor struct {
	// 记录每一代最优适应度值
	BestFitnessPerGen map[int]int

	// 记录每一代平均适应度值
	AvgFitnessPerGen map[int]float64

	// 记录每一代最坏适应度值
	WorstFitnessPerGen map[int]int

	// 准备执行交叉操作次数
	NumPreparedCrossover map[int]int
	// 实际执行交叉操作次数
	NumExecutedCrossover map[int]int

	// 准备执行交叉操作次数
	NumPreparedMutation map[int]int
	// 实际执行交叉操作次数
	NumExecutedMutation map[int]int

	// 总计算时间
	TotalTime time.Duration
}

// 构造函数
func NewMonitor() *Monitor {
	return &Monitor{
		BestFitnessPerGen:    make(map[int]int),
		AvgFitnessPerGen:     make(map[int]float64),
		WorstFitnessPerGen:   make(map[int]int),
		NumPreparedCrossover: make(map[int]int),
		NumExecutedCrossover: make(map[int]int),
		NumPreparedMutation:  make(map[int]int),
		NumExecutedMutation:  make(map[int]int),
	}
}

// 打印监控信息
func (m *Monitor) Dump() {

	log.Println("Monitor:")
	// 打印表头
	fmt.Println("| Generation | Best Fitness | Average Fitness | Worst Fitness | Num Prepared Crossover | Num Executed Crossover | Num Prepared Mutation | Num Executed Mutation |")
	fmt.Println("|------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|")

	// 按生成序号排序
	gens := make([]int, 0, len(m.BestFitnessPerGen))
	for gen := range m.BestFitnessPerGen {
		gens = append(gens, gen)
	}
	sort.Ints(gens)

	// 打印每一代的适应度值
	for _, gen := range gens {
		fmt.Printf("| %-11d | %-11d | %-14.2f | %-11d | %-11d | %-11d | %-11d | %-11d |\n",
			gen,
			m.BestFitnessPerGen[gen],
			m.AvgFitnessPerGen[gen],
			m.WorstFitnessPerGen[gen],
			m.NumPreparedCrossover[gen],
			m.NumExecutedCrossover[gen],
			m.NumPreparedMutation[gen],
			m.NumExecutedMutation[gen],
		)
	}
	fmt.Printf("  Total Time: %v\n", m.TotalTime)
}
