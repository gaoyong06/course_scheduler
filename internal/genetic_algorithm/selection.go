package genetic_algorithm

import "math/rand"

// 选择操作
func Selection(population []*Individual, selectionSize int) []*Individual {
	// 采用轮盘赌选择算法进行选择操作
	// 计算总适应度得分
	totalFitness := 0
	for _, individual := range population {
		totalFitness += individual.Fitness
	}

	// 如果总适应度得分为 0，那么返回一个空的选择结果集合
	if totalFitness == 0 {
		return nil
	}

	// 初始化选择结果集合
	selected := make([]*Individual, 0)

	// 根据适应度概率进行轮盘赌选择
	for i := 0; i < selectionSize; i++ {
		// 生成一个 0 到总适应度得分之间的随机数
		// 注意: 需要检查适应度函数的实现，确保其计算出的得分是非负数
		target := rand.Intn(totalFitness)

		// 轮盘赌选择
		sum := 0
		for _, individual := range population {
			// 适应度值越大，选择概率越大
			sum += individual.Fitness
			if sum >= target {
				selected = append(selected, individual)
				break
			}
		}
	}
	return selected
}
