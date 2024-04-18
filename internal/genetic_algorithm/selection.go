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
	selected := population[:bestCount]
	restPopulation := population[bestCount:]

	// 竞标赛模式选择
	// 选择操作做了去重,避免同一个个体被多次选中
	for i := bestCount; i < selectionSize; i++ {
		var individual1, individual2 *Individual

		index1 := rand.Intn(len(restPopulation))
		individual1 = restPopulation[index1]
		restPopulation = append(restPopulation[:index1], restPopulation[index1+1:]...)

		index2 := rand.Intn(len(restPopulation))
		individual2 = restPopulation[index2]
		restPopulation = append(restPopulation[:index2], restPopulation[index2+1:]...)

		if individual1.Fitness > individual2.Fitness {
			selected = append(selected, individual1)
		} else {
			selected = append(selected, individual2)
		}
	}

	return selected
}
