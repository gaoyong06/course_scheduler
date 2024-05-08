package genetic_algorithm

import (
	"course_scheduler/config"
	"course_scheduler/internal/base"
	"course_scheduler/internal/types"
	"log"
	"time"
)

// 遗传算法的实现
// scheduleInput 排课输入数据
// startTime 当前时间
func Execute(scheduleInput *base.ScheduleInput, monitor *base.Monitor, startTime time.Time) (*Individual, int, error) {

	// 种群大小
	popSize := config.PopSize
	// 选择操作,选择个体的数量
	selectionSize := config.SelectionSize
	// 最大迭代次数
	maxGen := config.MaxGen
	// 变异率
	mutationRate := config.MutationRate
	// 交叉率
	crossoverRate := config.CrossoverRate
	// 选择最佳个体百分比
	bestRatio := config.BestRatio
	// 找到满意的解
	foundSatIndividual := false
	// 连续 n 代没有改进
	genWithoutImprovement := 0
	// 是否进入搜索循环
	stop := false
	// 当前代数
	gen := 0
	// 最佳个体所在的代数
	bestGen := -1
	// 最佳个体唯一id
	uniqueId := ""
	// 最佳个体是否发生替换
	replaced := false
	// 最优的个体
	bestIndividual := &Individual{}
	// 最差的个体
	var worstIndividual *Individual

	// 初始化种群
	log.Println("Population initialized")
	// population, err := InitPopulation(scheduleInput, config.PopSize)
	// if err != nil {
	// 	log.Println("Population initialized failed. ", err)
	// 	return nil, err
	// }

	// 课班初始化
	classes := types.InitClasses(scheduleInput.TeachTaskAllocations)

	// 初始化当前种群
	currentPopulation, err := InitPopulation(classes, popSize, scheduleInput.Schedule, scheduleInput.TeachTaskAllocations, scheduleInput.Subjects, scheduleInput.Teachers, scheduleInput.SubjectVenueMap)
	if err != nil {
		log.Panic(err)
	}
	initPopulationTime := time.Since(startTime)
	log.Printf("Init population runtime: %v\n", initPopulationTime)

	// 当前种群内容重复的数量
	dupCount := CountDuplicates(currentPopulation)
	log.Printf("Population size %d: duplicates count %d\n", popSize, dupCount)

	for !stop {
		log.Println("Current Generation:", gen)
		// 获取当前最近个体标识符
		uniqueId = bestIndividual.UniqueId()

		// 评估当前种群中每个个体的适应度值，并更新当前找到的最佳个体
		bestIndividual, replaced, err = UpdateBest(currentPopulation, bestIndividual)
		if err != nil {
			log.Panic(err)
		}

		// 如果 bestIndividual 被替换，则记录当前 gen 值
		if replaced {
			bestGen = gen
		}

		// 评估适应度
		// for _, individual := range population {
		// 	individual.Fitness = EvaluateFitness(scheduleInput, individual.LessonListMap)
		// 	individual.Generation = currentIteration
		// }
		// log.Println("Fitness evaluation completed")

		// 计算最优,最差,平均适应度
		// bestIndividual = GetBestIndividual(currentPopulation)
		worstIndividual = GetWorstIndividual(currentPopulation)
		monitor.BestFitnessPerGen[gen] = bestIndividual.Fitness
		monitor.WorstFitnessPerGen[gen] = worstIndividual.Fitness
		monitor.AvgFitnessPerGen[gen] = CalcAvgFitness(gen, currentPopulation)

		// 检查是否连续 n 代没有改进
		if HasImproved(bestIndividual, currentPopulation) {
			genWithoutImprovement = 0
		} else {
			genWithoutImprovement++
			if genWithoutImprovement >= config.MaxStagnGen {
				log.Println("Termination condition met: No improvement for", genWithoutImprovement, "generations.")
				break
			}
		}

		// 检查是否找到满意的解
		foundSatIndividual = IsSatIndividual(currentPopulation)
		if !foundSatIndividual {

			// 选择操作（锦标赛）
			// 选择的个体是原个体数量的一半
			selectedPopulation, err := Selection(currentPopulation, selectionSize, bestRatio)
			if err != nil {
				log.Panic(err)
			}

			selectedCount := len(selectedPopulation)
			log.Printf("Current population size: %d, duplicates count: %d, selected count: %d\n", popSize, dupCount, selectedCount)

			if selectedCount > 0 {

				// 交叉
				// 交叉前后的个体数量不变
				offspring, prepared, executed, err := Crossover(selectedPopulation, crossoverRate, scheduleInput.Schedule, scheduleInput.TeachTaskAllocations, scheduleInput.Subjects, scheduleInput.Teachers, scheduleInput.SubjectVenueMap)
				if err != nil {
					log.Panic(err)
				}
				monitor.NumPreparedCrossover[gen] = prepared
				monitor.NumExecutedCrossover[gen] = executed

				// log.Printf("Crossover Gen: %d, selected: %d, offspring: %d, prepared: %d, executed: %d, error: %s\n", gen, len(selectedPopulation), len(crossoverRet.Offspring), crossoverRet.Prepared, crossoverRet.Executed, crossoverRet.Err)

				// 变异
				offspring, prepared, executed, err = Mutation(offspring, mutationRate, scheduleInput.Schedule, scheduleInput.TeachTaskAllocations, scheduleInput.Subjects, scheduleInput.Teachers, scheduleInput.SubjectVenueMap)
				if err != nil {
					log.Panic(err)
				}
				monitor.NumPreparedMutation[gen] = prepared
				monitor.NumExecutedMutation[gen] = executed

				// 更新种群
				// 更新前后的个体数量不变
				hasConflicts := CheckConflicts(currentPopulation)
				if hasConflicts {
					log.Panic("Population time slot conflicts")
				}
				// currentPopulation = genetic_algorithm.UpdatePopulation(currentPopulation, crossoverRet.Offspring)
				currentPopulation = UpdatePopulation(currentPopulation, offspring)
			}
		}

		// 在每次循环迭代时更新 currentIteration 的值
		gen++
		stop = TerminationCondition(gen, foundSatIndividual, genWithoutImprovement, startTime)
	}

	// 评估当前种群中每个个体的适应度值，并更新当前找到的最佳个体
	bestIndividual, replaced, err = UpdateBest(currentPopulation, bestIndividual)
	if err != nil {
		log.Panic(err)
	}
	// 如果 bestIndividual 被替换，则记录当前 gen 值
	if replaced {
		bestGen = maxGen - 1
	}

	// 打印当前代中最好个体的适应度值
	// log.Printf("Generation %d: Best Fitness = %d\n", gen, bestIndividual.Fitness)
	log.Printf("Generation %d: Best uniqueId= %s, bestGen=%d, Fitness = %d\n", maxGen-1, uniqueId, bestGen, bestIndividual.Fitness)

	// 输出最终排课结果
	// bestIndividual = GetBestIndividual(population)
	return bestIndividual, bestGen, nil
}
