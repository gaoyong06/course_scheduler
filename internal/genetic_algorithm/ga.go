package genetic_algorithm

import (
	"course_scheduler/config"
	"course_scheduler/internal/base"
	"errors"
	"log"
	"math"
	"time"
)

// 遗传算法的实现
// 参数:
//
//	scheduleInput: 排课输入数据
//	monitor: 业务监控
//	startTime: 当前时间
//
// 返回值:
//
//	返回 最佳个体、最佳个体所在的遗传代数、错误信息
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
	// 最大停滞代数
	maxStagnGen := config.MaxStagnGen
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
	bestIndividual := &Individual{
		Chromosomes: nil,
		Fitness:     math.MinInt32,
	}
	// 最差的个体
	var worstIndividual *Individual

	// 初始化种群
	log.Println("Population initialized")

	// 约束条件
	constraints := input.Constraints()

	// 初始化当前种群
	currentPopulation, err := InitPopulation(popSize, input.Schedule, input.TeachingTasks, input.Subjects, input.Teachers, input.SubjectVenueMap, constraints)
	if err != nil {
		return bestIndividual, bestGen, err
	}
	initPopulationTime := time.Since(startTime)
	log.Printf("Init population runtime: %v\n", initPopulationTime)

	// 相同的个体数量过多可能意味着种群缺乏多样性，从而导致算法提前收敛，无法探索到更优解
	// 当前种群内容重复的数量
	dupCount := CountDuplicates(currentPopulation)
	log.Printf("Population size %d: duplicates count %d\n", popSize, dupCount)

	for !stop {
		log.Println("Current Generation:", gen)
		// 获取当前最近个体标识符
		uniqueId = bestIndividual.UniqueId

		// 评估当前种群中每个个体的适应度值，并更新当前找到的最佳个体
		// 下面的bestIndividual会发生更新,所以在这里复制一份
		prevBestIndividual := bestIndividual.Copy()
		bestIndividual, replaced, err = UpdateBest(currentPopulation, bestIndividual)
		if err != nil {
			return bestIndividual, bestGen, err
		}

		log.Printf("ga loop gen: %d, uniqueId: %s, replaced: %v\n", gen, uniqueId, replaced)

		// 如果 bestIndividual 被替换，则记录当前 gen 值
		if replaced {
			bestGen = gen
		}

		// 计算最优,最差,平均适应度
		worstIndividual = GetWorstIndividual(currentPopulation)
		monitor.BestFitnessPerGen[gen] = bestIndividual.Fitness
		monitor.WorstFitnessPerGen[gen] = worstIndividual.Fitness
		monitor.AvgFitnessPerGen[gen] = CalcAvgFitness(gen, currentPopulation)

		// 检查是否连续 n 代没有改进
		if HasImproved(prevBestIndividual, currentPopulation) {
			genWithoutImprovement = 0
		} else {
			genWithoutImprovement++
			if genWithoutImprovement >= maxStagnGen {
				log.Println("Termination condition met: No improvement for", genWithoutImprovement, "generations.")
				break
			}
		}

		// 检查是否找到满意的解
		foundSatIndividual = IsSatIndividual(currentPopulation)
		// if !foundSatIndividual {

		// 选择操作（锦标赛）
		// 选择的个体是原个体数量的一半
		selectedPopulation, err := Selection(currentPopulation, selectionSize, bestRatio)
		if err != nil {
			return bestIndividual, bestGen, err
		}

		selectedCount := len(selectedPopulation)
		log.Printf("Current population size: %d, selected count: %d, duplicates count: %d\n", popSize, selectedCount, dupCount)

		if selectedCount > 0 {

			// 交叉
			// 交叉前后的个体数量不变
			offspring, prepared, executed, err := Crossover(selectedPopulation, crossoverRate, input.Schedule, input.TeachingTasks, input.Subjects, input.Teachers, input.Grades, input.SubjectVenueMap, constraints)
			if err != nil {
				return bestIndividual, bestGen, err
			}
			monitor.NumPreparedCrossover[gen] = prepared
			monitor.NumExecutedCrossover[gen] = executed

			// 变异
			offspring, prepared, executed, err = Mutation(offspring, mutationRate, input.Schedule, input.TeachingTasks, input.Subjects, input.Teachers, input.Grades, input.SubjectVenueMap, constraints)
			if err != nil {
				return bestIndividual, bestGen, err
			}
			monitor.NumPreparedMutation[gen] = prepared
			monitor.NumExecutedMutation[gen] = executed

			// 更新种群
			// 更新前后的个体数量不变
			hasConflicts := CheckConflicts(currentPopulation)
			if hasConflicts {
				err = errors.New("population time slot conflicts")
				return bestIndividual, bestGen, err
			}
			currentPopulation = UpdatePopulation(currentPopulation, offspring)
		}
		// }

		// 在每次循环迭代时更新 gen 的值
		gen++
		stop = TerminationCondition(gen, foundSatIndividual, genWithoutImprovement, startTime)
	}

	// 打印当前代中最好个体的适应度值
	log.Printf("Generation %d: Best uniqueId= %s, bestGen=%d, Fitness = %d\n", gen, uniqueId, bestGen, bestIndividual.Fitness)
	return bestIndividual, bestGen, nil
}
