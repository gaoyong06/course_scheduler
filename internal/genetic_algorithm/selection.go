package genetic_algorithm

import (
	"fmt"
	"math"
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
	bestCount := 0
	if bestRatio > 0 {
		bestCount = int(math.Max(float64(popSize)*bestRatio, 1))
	}

	fmt.Printf("Selection best count: %d\n", bestCount)

	// 按照适应度进行排序
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})

	// 将排名前 bestCount 个个体添加到 selected 中
	for i := 0; i < bestCount; i++ {

		fmt.Printf("Selection best population: %d, fitness: %d\n", i+1, population[i].Fitness)
		selected = append(selected, population[i])
	}

	// 竞标赛模式选择
	// 选择操作做了去重,避免同一个个体被多次选中
	for i := bestCount; i < selectionSize; i++ {
		var individual1, individual2 *Individual
		for {
			candidate := population[rand.Intn(len(population))]
			if !contains(selected, candidate) {
				individual1 = candidate
				break
			}
		}

		for {
			candidate := population[rand.Intn(len(population))]
			if !contains(selected, candidate) {
				individual2 = candidate
				break
			}
		}

		if individual1.Fitness > individual2.Fitness {
			selected = append(selected, individual1)
		} else {
			selected = append(selected, individual2)
		}
	}

	return selected
}

// 是否包含
func contains(slice []*Individual, item *Individual) bool {
	for _, a := range slice {
		if a.UniqueId() == item.UniqueId() {
			return true
		}
	}
	return false
}
