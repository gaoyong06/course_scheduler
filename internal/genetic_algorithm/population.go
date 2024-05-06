// population.go
package genetic_algorithm

import (
	"course_scheduler/config"
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"log"
	"math/rand"
	"sort"
)

// 初始化种群
func InitPopulation(classes []types.Class, classHours map[int]int, populationSize int, schedule *models.Schedule, subjects []*models.Subject, teachers []*models.Teacher, subjectVenueMap map[string][]int) ([]*Individual, error) {

	population := make([]*Individual, populationSize)
	errChan := make(chan error, populationSize)

	classSNs := make([]string, len(classes))
	for i, class := range classes {
		classSNs[i] = class.SN.Generate()
	}

	for i := 0; i < populationSize; i++ {
		go func(i int) {
			log.Printf("Initializing individual %d\n", i+1)

			individual, err := createIndividual(classes, classSNs, classHours, schedule, subjects, teachers, subjectVenueMap)
			if err != nil {
				errChan <- err
				return
			}

			population[i] = individual
			log.Printf("Individual %d initialized\n", i)

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

// 更新种群
func UpdatePopulation(population []*Individual, offspring []*Individual) []*Individual {

	size := len(population)

	// 将新生成的个体添加到种群中
	population = append(population, offspring...)

	// 根据适应度值进行排序
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})

	// Create a new slice to store non-nil individuals
	newPopulation := make([]*Individual, 0, size)

	// Copy non-nil individuals to the new slice
	for _, individual := range population[:size] {
		if individual != nil {
			newPopulation = append(newPopulation, individual)
		}
	}

	return newPopulation
}

// 评估种群中每个个体的适应度值，并更新当前找到的最佳个体
// population 种群
// bestIndividual 当前最佳个体
// 返回值: 最佳个体,是否发生替换,错误信息
func UpdateBest(population []*Individual, bestIndividual *Individual) (*Individual, bool, error) {

	replaced := false
	for _, individual := range population {

		// log.Printf("individual(%d) uniqueId: %s, fitness: %d\n", i, individual.UniqueId(), individual.Fitness)
		// 在更新 bestIndividual 时，将当前的 individual 复制一份，然后将 bestIndividual 指向这个复制出来的对象
		// 即使 individual 的值在下一次循环中发生变化，bestIndividual 指向的对象也不会变化
		if individual.Fitness > (*bestIndividual).Fitness {

			log.Printf("UpdateBest individual.Fitness: %d, bestIndividual.Fitness: %d\n", individual.Fitness, bestIndividual.Fitness)
			newBestIndividual := individual.Copy()
			bestIndividual = newBestIndividual
			replaced = true
		}
	}

	// log.Printf("==== UpdateBest DONE! uniqueId: %s, fitness: %d\n", bestIndividual.UniqueId(), bestIndividual.Fitness)
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
	// 在这里实现你的检查逻辑
	// 检查种群中是否有满意的解，根据具体的业务逻辑进行判断
	// 如果找到满意的解则返回 true，否则返回 false
	// 示例逻辑：如果种群中最优个体的适应度已经满足某个阈值，则认为找到了满意的解
	bestIndividual := GetBestIndividual(population)
	return bestIndividual.Fitness <= config.TargetFitness
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
	duplicates := checkDuplicates(population)
	return len(duplicates)
}

// hasDuplicates 种群中是否有相同的个体
func HasDuplicates(population []*Individual) bool {
	duplicates := checkDuplicates(population)
	return len(duplicates) > 0
}

// 检查种群中是否存在时间段冲突的个体
func CheckConflicts(population []*Individual) bool {

	for i, item := range population {
		hasTimeSlotConflicts, conflicts := item.HasTimeSlotConflicts()
		if hasTimeSlotConflicts {
			log.Printf("The %dth individual has time conflicts, conflict info: %v\n", i, conflicts)
			return true
		}
	}
	return false
}

// ============================================

// 创建个体
func createIndividual(classes []types.Class, classeSNs []string, classHours map[int]int, schedule *models.Schedule, subjects []*models.Subject, teachers []*models.Teacher, subjectVenueMap map[string][]int) (*Individual, error) {

	classMatrix := types.NewClassMatrix()

	// 避免所有的 goroutine 对同一个 classes slice 进行操作，导致数据竞争 (data race) 的问题
	classesCopy := make([]types.Class, len(classes))
	copy(classesCopy, classes)
	shuffleClassOrder(classesCopy)

	err := initClassMatrix(classMatrix, classesCopy, teachers, subjectVenueMap)
	if err != nil {
		return nil, err
	}

	calculateFixedScores(classMatrix, subjects, teachers)
	_, err = allocateClassMatrix(classMatrix, classeSNs, classHours, schedule)

	if err != nil {
		return nil, err
	}
	return newIndividual(classMatrix, classHours, schedule, subjects, teachers)
}

// 打乱课程顺序
func shuffleClassOrder(classes []types.Class) {
	for i := len(classes) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		classes[i], classes[j] = classes[j], classes[i]
	}
	log.Println("Class order shuffled")
}

// 初始化课程矩阵
func initClassMatrix(classMatrix *types.ClassMatrix, classes []types.Class, teachers []*models.Teacher, subjectVenueMap map[string][]int) error {

	count := config.NumGrades * config.NumClassesPreGrade * config.NumSubjects
	if len(classes) != count {
		return fmt.Errorf("failed to initialize class matrix: expected %d classes, got %d", count, len(classes))
	}
	classMatrix.Init(classes, teachers, subjectVenueMap)
	log.Printf("Class matrix %p initialized successfully \n", classMatrix)
	return nil
}

// 计算固定得分
func calculateFixedScores(classMatrix *types.ClassMatrix, subjects []*models.Subject, teachers []*models.Teacher) {
	fixedRules := constraint.GetFixedRules(subjects, teachers)
	err := classMatrix.CalcElementFixedScores(fixedRules)
	if err != nil {
		log.Fatalf("Failed to calculate fixed scores: %v", err)
	}
	log.Println("Fixed scores calculated")
}

// 分配课程矩阵
func allocateClassMatrix(classMatrix *types.ClassMatrix, classeSNs []string, classHours map[int]int, schedule *models.Schedule) (int, error) {
	dynamicRules := constraint.GetDynamicRules(schedule)
	return classMatrix.Allocate(classeSNs, classHours, dynamicRules)
}

// checkDuplicates 种群中重复个体的映射，以其唯一ID为键
func checkDuplicates(population []*Individual) map[string][]*Individual {
	duplicates := make(map[string][]*Individual)
	ids := make(map[string]*Individual)

	for _, individual := range population {
		id := individual.UniqueId()
		if existing, ok := ids[id]; ok {
			duplicates[id] = []*Individual{existing, individual}
		} else {
			ids[id] = individual
		}
	}

	return duplicates
}
