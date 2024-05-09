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
// monitor 业务监控
// startTime 当前时间
func Execute(input *base.ScheduleInput, monitor *base.Monitor, startTime time.Time) (*Individual, int, error) {

	// 种群大小
	popSize := config.PopSize
	// 选择操作,选择个体的数量
	selectionSize := config.SelectionSize
	// 变异率
	mutationRate := config.MutationRate
	// 交叉率
	crossoverRate := config.CrossoverRate
	// 选择最佳个体百分比
	bestRatio := config.BestRatio
	// 是否找到满意的解
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
	classes := types.InitClasses(input.TeachTaskAllocations)

	// 初始化当前种群
	constraints := input.ConvertConstraints()
	currentPopulation, err := InitPopulation(classes, popSize, input.Schedule, input.TeachTaskAllocations, input.Subjects, input.Teachers, input.SubjectVenueMap, constraints)
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
		// 下面的bestIndividual会发生更新,所以在这里复制一份
		prevBestIndividual := bestIndividual.Copy()
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
		if HasImproved(prevBestIndividual, currentPopulation) {
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
				offspring, prepared, executed, err := Crossover(selectedPopulation, crossoverRate, input.Schedule, input.TeachTaskAllocations, input.Subjects, input.Teachers, input.SubjectVenueMap, constraints)
				if err != nil {
					log.Panic(err)
				}
				monitor.NumPreparedCrossover[gen] = prepared
				monitor.NumExecutedCrossover[gen] = executed

				// 变异
				offspring, prepared, executed, err = Mutation(offspring, mutationRate, input.Schedule, input.TeachTaskAllocations, input.Subjects, input.Teachers, input.SubjectVenueMap, constraints)
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
				currentPopulation = UpdatePopulation(currentPopulation, offspring)
			}
		}

		// 在每次循环迭代时更新 gen 的值
		gen++
		stop = TerminationCondition(gen, foundSatIndividual, genWithoutImprovement, startTime)
	}

	// 打印当前代中最好个体的适应度值
	log.Printf("Generation %d: Best uniqueId= %s, bestGen=%d, Fitness = %d\n", gen, uniqueId, bestGen, bestIndividual.Fitness)
	return bestIndividual, bestGen, nil
}
