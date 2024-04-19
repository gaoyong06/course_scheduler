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

	// 当前种群内容重复的数量
	dupCount := genetic_algorithm.CountDuplicates(currentPopulation)
	log.Printf("Population size %d: duplicates count %d\n", popSize, dupCount)

	// 定义最佳个体
	bestIndividual := &genetic_algorithm.Individual{}
	uniqueId := ""
	bestGen := -1
	replaced := false

	for gen := 0; gen < maxGen; gen++ {

		// 获取当前最近个体标识符
		uniqueId = bestIndividual.UniqueId()

		// 评估当前种群中每个个体的适应度值，并更新当前找到的最佳个体
		bestIndividual, replaced, err = genetic_algorithm.UpdateBest(currentPopulation, bestIndividual)
		if err != nil {
			log.Panic(err)
		}

		// 如果 bestIndividual 被替换，则记录当前 gen 值
		if replaced {
			bestGen = gen
		}

		// 打印当前代中最好个体的适应度值
		// log.Printf("Generation %d: Best Fitness = %d\n", gen+1, bestIndividual.Fitness)
		log.Printf("Generation %d: Best uniqueId= %s, bestGen=%d, Fitness = %d\n", gen+1, uniqueId, bestGen, bestIndividual.Fitness)

		// 选择
		// 选择的个体是原个体数量的一半
		selectedPopulation, err := genetic_algorithm.Selection(currentPopulation, selectionSize, bestRatio)
		if err != nil {
			log.Panic(err)
		}

		selectedCount := len(selectedPopulation)

		log.Printf("Current population size: %d, duplicates count: %d, selected count: %d\n", popSize, dupCount, selectedCount)

		if selectedCount > 0 {

			// 交叉
			// 交叉前后的个体数量不变
			// offspring, err := genetic_algorithm.Crossover(selectedPopulation, crossoverRate)
			// if err != nil {
			// 	log.Panic(err)
			// }

			crossoverRet := genetic_algorithm.Crossover(selectedPopulation, crossoverRate, classHours)
			if crossoverRet.Err != nil {
				log.Panic(crossoverRet.Err)
			}
			log.Printf("Crossover Gen: %d, selected: %d, offspring: %d, prepared: %d, executed: %d, error: %s\n", gen+1, len(selectedPopulation), len(crossoverRet.Offspring), crossoverRet.Prepared, crossoverRet.Executed, crossoverRet.Err)

			// 变异
			// offspring := crossoverRet.Offspring
			// offspring, err = genetic_algorithm.Mutation(offspring, mutationRate, classHours)
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

		fmt.Printf("\n\n")
	}

	// 评估当前种群中每个个体的适应度值，并更新当前找到的最佳个体
	bestIndividual, replaced, err = genetic_algorithm.UpdateBest(currentPopulation, bestIndividual)
	if err != nil {
		log.Panic(err)
	}
	// 如果 bestIndividual 被替换，则记录当前 gen 值
	if replaced {
		bestGen = maxGen
	}

	// 打印当前代中最好个体的适应度值
	// log.Printf("Generation %d: Best Fitness = %d\n", gen+1, bestIndividual.Fitness)
	log.Printf("Generation %d: Best uniqueId= %s, bestGen=%d, Fitness = %d\n", maxGen, uniqueId, bestGen, bestIndividual.Fitness)

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
			log.Printf("【xy】!!!!!!! %s population中有冲突 ,%v\n", key, b)
		}
	}
}
