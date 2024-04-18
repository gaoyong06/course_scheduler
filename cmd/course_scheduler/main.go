// main.go
package main

import (
	"fmt"
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

	// 初始化当前种群
	currentPopulation, err := genetic_algorithm.InitPopulation(classes, classHours, popSize)
	if err != nil {
		log.Panic(err)
	}

	// 定义最佳个体
	bestIndividual := &genetic_algorithm.Individual{}
	uniqueId := ""

	for gen := 0; gen < maxGen; gen++ {

		// 获取当前最近个体标识符
		uniqueId = bestIndividual.UniqueId()

		// 评估当前种群中每个个体的适应度值，并更新当前找到的最佳个体
		bestIndividual, err = genetic_algorithm.UpdateBest(currentPopulation, bestIndividual)
		if err != nil {
			log.Panic(err)
		}

		// 打印当前代中最好个体的适应度值
		// log.Printf("Generation %d: Best Fitness = %d\n", gen+1, bestIndividual.Fitness)
		log.Printf("Generation %d: Best uniqueId= %s, Fitness = %d\n", gen+1, uniqueId, bestIndividual.Fitness)

		// 选择
		// 选择的个体是原个体数量的一半
		selectedPopulation := genetic_algorithm.Selection(currentPopulation, selectionSize, bestRatio)
		if len(selectedPopulation) > 0 {

			// 交叉
			// 交叉前后的个体数量不变
			// offspring, err := genetic_algorithm.Crossover(selectedPopulation, crossoverRate)
			// if err != nil {
			// 	log.Panic(err)
			// }

			crossoverRet := genetic_algorithm.Crossover(selectedPopulation, crossoverRate)
			if crossoverRet.Err != nil {
				log.Panic(crossoverRet.Err)
			}
			log.Printf("Generation %d: crossoverRet len(selected):%d, len(offspring): %d, prepared: %d, executed: %d, error: %s\n", gen+1, len(selectedPopulation), len(crossoverRet.Offspring), crossoverRet.Prepared, crossoverRet.Executed, crossoverRet.Err)

			// // 变异
			// offspring, err = genetic_algorithm.Mutation(offspring, mutationRate)
			// if err != nil {
			// 	log.Panic(err)
			// }

			// 更新种群
			// 更新前后的个体数量不变
			// TODO: 这里会引发currentPopulation内边个体有时间段冲突
			xy(currentPopulation, "#1")
			currentPopulation = genetic_algorithm.UpdatePopulation(currentPopulation, crossoverRet.Offspring)
			// currentPopulation = genetic_algorithm.UpdatePopulation(currentPopulation, offspring)
		}
	}

	// 评估当前种群中每个个体的适应度值，并更新当前找到的最佳个体
	bestIndividual, err = genetic_algorithm.UpdateBest(currentPopulation, bestIndividual)
	if err != nil {
		log.Panic(err)
	}

	// 打印最好的个体
	log.Printf("最佳个体适应度: %d, uniqueId: %s\n", bestIndividual.Fitness, uniqueId)
	bestIndividual.PrintSchedule()

	// 计算总运行时间
	elapsedTime := time.Since(startTime)
	log.Printf("Total runtime: %v\n", elapsedTime)
}

func xy(population []*genetic_algorithm.Individual, key string) {

	for _, item := range population {
		a, b := item.HasTimeSlotConflicts()
		if a {
			fmt.Printf("【xy】!!!!!!! %s population中有冲突 ,%v\n", key, b)
		}
	}
}
