// main.go
package main

import (
	"log"
	"time"

	"course_scheduler/config"
	"course_scheduler/internal/class_adapt"
	"course_scheduler/internal/genetic_algorithm"
	"course_scheduler/internal/models"
)

func main() {
	startTime := time.Now()

	// 参数定义
	popSize := config.PopSize
	selectionSize := config.SelectionSize
	maxGen := config.MaxGen
	// mutationRate := config.MutationRate
	crossoverRate := config.CrossoverRate
	bestRatio := config.BestRatio

	// 课班初始化
	classes := class_adapt.InitClasses()

	// 周课时初始化
	classHours := models.GetClassHours()

	// 初始化种群
	population, err := genetic_algorithm.InitPopulation(classes, classHours, popSize)
	if err != nil {
		log.Panic(err)
	}

	// 定义最佳个体
	bestIndividual := &genetic_algorithm.Individual{}

	for gen := 0; gen < maxGen; gen++ {

		// 评估种群中每个个体的适应度值，并更新当前找到的最佳个体
		bestIndividual, err = genetic_algorithm.EvaluateAndUpdateBest(population, bestIndividual)
		if err != nil {
			log.Panic(err)
		}

		// 打印当前代中最好个体的适应度值
		log.Printf("Generation %d: Best Fitness = %d\n", gen+1, bestIndividual.Fitness)

		// 选择
		// 选择的个体是原个体数量的一半
		selected := genetic_algorithm.Selection(population, selectionSize, bestRatio)
		if len(selected) > 0 {

			// 交叉
			// 交叉前后的个体数量不变
			crossoverRet := genetic_algorithm.Crossover(selected, crossoverRate)
			if crossoverRet.Error != nil {
				log.Panic(crossoverRet.Error)
			}
			log.Printf("Generation %d: Best Fitness = %d, crossoverRet len(selected):%d, len(offsprings): %d, prepareCrossover: %d, executedCrossover: %d, error: %s\n", gen+1, bestIndividual.Fitness, len(selected), len(crossoverRet.Offsprings), crossoverRet.PrepareCrossover, crossoverRet.ExecutedCrossover, crossoverRet.Error)

			// // 变异
			// offspring, err = genetic_algorithm.Mutation(offspring, mutationRate)
			// if err != nil {
			// 	log.Panic(err)
			// }

			// 更新种群
			// 更新前后的个体数量不变
			population = genetic_algorithm.UpdatePopulation(population, crossoverRet.Offsprings)

		}
	}

	// 打印最好的个体
	log.Printf("最佳个体适应度: %d\n", bestIndividual.Fitness)
	bestIndividual.PrintSchedule()

	// 计算总运行时间
	elapsedTime := time.Since(startTime)
	log.Printf("Total runtime: %v\n", elapsedTime)
}
