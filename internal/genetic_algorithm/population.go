// population.go
package genetic_algorithm

import (
	"course_scheduler/config"
	constraint "course_scheduler/internal/constraints"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"log"
	"sort"
)

// 初始化种群
func InitPopulation(populationSize int, schedule *models.Schedule, teachingTasks []*models.TeachingTask, subjects []*models.Subject, teachers []*models.Teacher, subjectVenueMap map[string][]int, constraints map[string]interface{}) ([]*Individual, error) {

	population := make([]*Individual, populationSize)
	errChan := make(chan error, populationSize)

	for i := 0; i < populationSize; i++ {
		go func(i int) {
			log.Printf("Initializing individual %d\n", i+1)

			individual, err := createIndividual(schedule, teachingTasks, subjects, teachers, subjectVenueMap, constraints)
			if err != nil {
				errChan <- err
				return
			}

			population[i] = individual
			log.Printf("Individual %d, uniqueId: %s, initialized\n", i, individual.UniqueId)

			// 向 errChan 发送 nil 值，表示该 goroutine 执行成功
			errChan <- nil
		}(i)
	}

	// 检查错误
	for i := 0; i < populationSize; i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	log.Println("Population initialization completed")
	return population, nil
}

func UpdatePopulation(population []*Individual, offspring []*Individual) []*Individual {
	size := len(population)

	// 选择的个体去重
	ids := make(map[string]bool)

	// 将新生成的个体添加到种群中
	population = append(population, offspring...)

	// 根据适应度值进行排序
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})

	newPopulation := make([]*Individual, 0, size)

	for _, individual := range population {
		if len(newPopulation) < size && individual != nil {
			id := individual.UniqueId
			if !ids[id] {
				newPopulation = append(newPopulation, individual)
				ids[id] = true
			}
		}
		// 提前结束循环
		if len(newPopulation) == size {
			break
		}
	}

	return newPopulation
}

// 评估种群中每个个体的适应度值，并更新当前找到的最佳个体
// 参数:
//
//	population: 种群
//	bestIndividual: 当前最佳个体
//
// 返回值:
//
//	返回 最佳个体、是否发生替换、错误信息
func UpdateBest(population []*Individual, bestIndividual *Individual) (*Individual, bool, error) {

	replaced := false
	for i, individual := range population {

		log.Printf("update best individual(%d) uniqueId: %s, fitness: %d\n", i, individual.UniqueId, individual.Fitness)
		// 在更新 bestIndividual 时，将当前的 individual 复制一份，然后将 bestIndividual 指向这个复制出来的对象
		// 即使 individual 的值在下一次循环中发生变化，bestIndividual 指向的对象也不会变化
		if individual.Fitness > (*bestIndividual).Fitness {

			log.Printf("update best individual.Fitness: %d, bestIndividual.Fitness: %d\n", individual.Fitness, bestIndividual.Fitness)
			newBestIndividual := individual.Copy()
			bestIndividual = newBestIndividual
			replaced = true
		}
	}

	return bestIndividual, replaced, nil
}

// 获取质量最优的个体,适应度得分最高的个体
func GetBestIndividual(population []*Individual) *Individual {

	// 获取适应度得分最低的个体作为最优解
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})
	bestIndividual := population[0]

	return bestIndividual
}

// 获取质量最差的个体,适应度得分最低的个体
func GetWorstIndividual(population []*Individual) *Individual {

	// 获取适应度得分最高的个体作为最差解
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness < population[j].Fitness
	})
	worstIndividual := population[0]
	return worstIndividual
}

// IsSatIndividual 检查是否找到满意的解
func IsSatIndividual(population []*Individual) bool {
	// 检查种群中是否有满意的解，根据具体的业务逻辑进行判断
	// 如果找到满意的解则返回 true，否则返回 false
	// 示例逻辑：如果种群中最优个体的适应度已经满足某个阈值，则认为找到了满意的解
	bestIndividual := GetBestIndividual(population)
	return bestIndividual.Fitness >= config.TargetFitness
}

// HasImproved 判断种群是否有改进
// prevBestIndividual 父种群中最优的个体
// population 当前种群
func HasImproved(prevBestIndividual *Individual, population []*Individual) bool {

	prevBestFitness := prevBestIndividual.Fitness
	for _, individual := range population {

		// 如果有更优秀的个体，则种群有改进
		if individual.Fitness > prevBestFitness {
			return true
		}
	}
	return false
}

// 计算平均适应度值
func CalcAvgFitness(generation int, population []*Individual) float64 {
	var totalFitness float64
	for _, individual := range population {
		totalFitness += float64(individual.Fitness)
	}
	averageFitness := totalFitness / float64(len(population))
	return averageFitness
}

// countDuplicates 种群中相同个体的数量
func CountDuplicates(population []*Individual) int {
	duplicates := findDuplicates(population)
	return len(duplicates)
}

// hasDuplicates 种群中是否有相同的个体
func HasDuplicates(population []*Individual) bool {
	duplicates := findDuplicates(population)
	return len(duplicates) > 0
}

// 检查种群中是否存在时间段冲突的个体
func CheckConflicts(population []*Individual) bool {

	for i, item := range population {
		hasTimeSlotConflicts, conflicts := item.HasTimeSlotConflicts()
		if hasTimeSlotConflicts {
			log.Printf("check conflicts failed. The %dth individual has time conflicts, conflict info: %v\n", i, conflicts)
			return true
		}
	}
	return false
}

// ============================================

// 创建个体
func createIndividual(schedule *models.Schedule, teachingTasks []*models.TeachingTask, subjects []*models.Subject, teachers []*models.Teacher, subjectVenueMap map[string][]int, constraints map[string]interface{}) (*Individual, error) {
	allocated := false
	classMatrix, err := types.NewClassMatrix(schedule, teachingTasks, subjects, teachers, subjectVenueMap)
	if err != nil {
		return nil, err
	}

	for retry := 0; retry < config.MaxRetries; retry++ {
		err = classMatrix.Init()
		if err != nil {
			return nil, err
		}

		calcFixedScores(classMatrix, subjects, teachers, schedule, teachingTasks, constraints)
		calcDynamicScores(classMatrix, schedule, teachingTasks, constraints)

		allocateCount, err := allocateClassMatrix(classMatrix, schedule, constraints)
		if err != nil {
			log.Printf("allocate class matrix failed. allocate count: %d, retry: %d, err : %s\n", allocateCount, retry, err)
			continue
		}

		allocated = true
		log.Printf("allocate class matrix success. allocate count  %d\n", allocateCount)
		break
	}

	if !allocated {
		return nil, fmt.Errorf("create individual failed. because allocate class matrix failed")
	}

	return newIndividual(classMatrix, schedule, subjects, teachers, constraints)
}

// 计算固定得分
func calcFixedScores(classMatrix *types.ClassMatrix, subjects []*models.Subject, teachers []*models.Teacher, schedule *models.Schedule, teachingTasks []*models.TeachingTask, constraints map[string]interface{}) {

	rules := constraint.GetFixedRules(subjects, teachers, constraints)
	err := classMatrix.CalcElementFixedScores(schedule, teachingTasks, rules)
	if err != nil {
		log.Fatalf("Failed to calculate fixed scores: %v", err)
	}
}

// 计算动态约束得分
func calcDynamicScores(classMatrix *types.ClassMatrix, schedule *models.Schedule, teachingTasks []*models.TeachingTask, constraints map[string]interface{}) {

	rules := constraint.GetDynamicRules(schedule, constraints)
	err := classMatrix.CalcElementDynamicScores(schedule, teachingTasks, rules)
	if err != nil {
		log.Fatalf("Failed to calculate fixed scores: %v", err)
	}
}

// 分配课程矩阵
func allocateClassMatrix(classMatrix *types.ClassMatrix, schedule *models.Schedule, constraints map[string]interface{}) (int, error) {
	dynamicRules := constraint.GetDynamicRules(schedule, constraints)
	return classMatrix.Allocate(dynamicRules)
}

// findDuplicates 种群中重复个体的映射，以其唯一ID为键
func findDuplicates(population []*Individual) map[string][]*Individual {

	duplicates := make(map[string][]*Individual)
	ids := make(map[string]*Individual)

	for _, individual := range population {
		id := individual.UniqueId
		if existing, ok := ids[id]; ok {
			duplicates[id] = []*Individual{existing, individual}
		} else {
			ids[id] = individual
		}
	}

	return duplicates
}
