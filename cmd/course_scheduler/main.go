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

	popSize := config.PopSize
	selectionSize := config.SelectionSize
	maxGen := config.MaxGen
	// mutationRate := config.MutationRate
	crossoverRate := config.CrossoverRate

	classes := class_adapt.InitClasses()
	classHours := models.GetClassHours()

	population, err := genetic_algorithm.InitPopulation(classes, classHours, popSize)
	if err != nil {
		log.Panic(err)
	}

	bestIndividual := &genetic_algorithm.Individual{Fitness: -1}

	for gen := 0; gen < maxGen; gen++ {

		// 评估适应度
		for _, individual := range population {
			individual.Fitness, err = individual.EvaluateFitness()
			if err != nil {
				log.Panic(err)
			}
			if individual.Fitness > bestIndividual.Fitness {
				bestIndividual = individual
			}
		}

		// 打印当前代中最好个体的适应度值
		log.Printf("Generation %d: Best Fitness = %d\n", gen+1, bestIndividual.Fitness)

		// 选择
		selected := genetic_algorithm.Selection(population, selectionSize)

		// 交叉
		crossoverRet := genetic_algorithm.Crossover(selected, crossoverRate)
		// population, err = genetic_algorithm.Crossover(selected, crossoverRate)
		if crossoverRet.Error != nil {
			log.Panic(crossoverRet.Error)
		}
		log.Printf("crossoverRet: %#v\n", crossoverRet)

		// // 变异
		// offspring, err = genetic_algorithm.Mutation(offspring, mutationRate)
		// if err != nil {
		// 	log.Panic(err)
		// }

		// 更新种群
		population = genetic_algorithm.UpdatePopulation(population, crossoverRet.Offsprings)
	}

	// 检查是否有时间段冲突
	conflictExists, conflictDetails := bestIndividual.HasTimeSlotConflicts()
	if conflictExists {
		log.Printf("Individual has time slot conflicts: %v\n", conflictDetails)
	} else {
		log.Println("Individual does not have time slot conflicts")
	}

	// 打印最好的个体
	bestIndividual.PrintSchedule()

	// 计算总运行时间
	elapsedTime := time.Since(startTime)
	log.Printf("Total runtime: %v\n", elapsedTime)
}
