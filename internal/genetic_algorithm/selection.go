// selection.go
package genetic_algorithm

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
)

// 选择操作
// 参数:
//
//	population: 种群
//	selectionSize: 选择数量
//	bestRatio: 保留最佳个体概率
//
// 返回值:
//
//	返回 选择的个体、错误信息
func Selection(population []*Individual, selectionSize int, bestRatio float64) ([]*Individual, error) {

	// 个体数量
	popSize := len(population)

	// 相同个体的数量
	dupCount := CountDuplicates(population)

	// 保留最佳个体
	bestCount := 0
	if bestRatio > 0 {
		bestCount = int(math.Max(float64(popSize)*bestRatio, 1))
	}
	log.Printf("Selection current population size: %d, duplicates count: %d, best count: %d\n", popSize, dupCount, bestCount)

	selected := make([]*Individual, 0, selectionSize)
	ids := make(map[string]bool)

	// 按照适应度进行排序
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})

	// 将排名前 bestCount 个个体选中
	for i := 0; i < bestCount; i++ {

		individual := population[i]
		id := individual.UniqueId()
		if !ids[id] {
			selected = append(selected, individual)
			ids[id] = true
		}
	}

	// 使用锦标赛选择选择其余的个体
	count := 0
	for len(selected) < selectionSize {
		index1 := rand.Intn(popSize)
		index2 := rand.Intn(popSize)
		if population[index1].Fitness > population[index2].Fitness {
			individual := population[index1]
			id := individual.UniqueId()
			if !ids[id] {
				selected = append(selected, individual)
				ids[id] = true
			}
		} else {
			individual := population[index2]
			id := individual.UniqueId()
			if !ids[id] {
				selected = append(selected, individual)
				ids[id] = true
			}
		}
		count++
		if count > 10*popSize {
			break
		}
	}
	return selected, nil
}

// validateSelection 检查选择是否有效
func validateSelection(population, selected []*Individual, selectionSize int) error {

	if len(selected) != selectionSize {
		return fmt.Errorf("selection size failed")
	}

	// 检查选择是否包含种群中的最佳个体
	bestIndividual := population[0]
	found := false
	for _, individual := range selected {
		if individual.UniqueId() == bestIndividual.UniqueId() {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("selection does not contain the best individual from the population")
	}

	// 检查选择是否包含重复的个体
	if HasDuplicates(selected) {
		return fmt.Errorf("selection contains duplicate individuals")
	}
	return nil
}
