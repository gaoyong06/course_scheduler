package genetic_algorithm

import (
	"math/rand"
	"sort"
)

// 选择操作
// population 种群
// selectionSize 选择数量
// bestRatio 保留最佳个体概率
func Selection(population []*Individual, selectionSize int, bestRatio float64) []*Individual {

	// 个体数量
	popSize := len(population)

	// 计算总适应度得分
	totalFitness := 0
	for _, individual := range population {
		totalFitness += individual.Fitness
	}

	// fmt.Printf("Selection len(population): %d, totalFitness: %d\n", len(population), totalFitness)

	// 如果总适应度得分为 0，那么返回一个空的选择结果集合
	if totalFitness == 0 {
		return nil
	}

	// 初始化选择结果集合
	selected := make([]*Individual, 0, selectionSize)

	// 保留最佳个体
	bestCount := int(float64(popSize) * bestRatio)

	// 按照适应度进行排序
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})

	// 将排名前 bestCount 个个体添加到 selected 中
	for i := 0; i < bestCount; i++ {
		selected = append(selected, population[i])
	}

	// 竞标赛模式选择
	for i := bestCount; i < selectionSize; i++ {
		individual1 := population[rand.Intn(len(population))]
		individual2 := population[rand.Intn(len(population))]
		if individual1.Fitness > individual2.Fitness {
			selected = append(selected, individual1)
		} else {
			selected = append(selected, individual2)
		}
	}

	return selected
}
