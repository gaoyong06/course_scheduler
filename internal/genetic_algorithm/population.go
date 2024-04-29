// population.go
package genetic_algorithm

import (
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/types"
	"log"
	"math/rand"
	"sort"
	"time"
)

// 初始化种群
func InitPopulation(classes []types.Class, classHours map[int]int, populationSize int) ([]*Individual, error) {

	var classeSNs []string
	for _, class := range classes {
		sn := class.SN.Generate()
		classeSNs = append(classeSNs, sn)
	}

	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	population := make([]*Individual, 0)
	for i := 0; i < populationSize; i++ {
		log.Printf("Initializing individual %d\n", i+1)

		individual, err := createIndividual(classes, classeSNs, classHours)
		if err != nil {
			return nil, err
		}

		population = append(population, individual)
		log.Println("Individual initialized")
		log.Println("")
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
func createIndividual(classes []types.Class, classeSNs []string, classHours map[int]int) (*Individual, error) {

	classMatrix := types.NewClassMatrix()
	shuffleClassOrder(classes)
	initClassMatrix(classMatrix, classes)
	calculateFixedScores(classMatrix)
	_, err := allocateClassMatrix(classMatrix, classeSNs, classHours)

	if err != nil {
		return nil, err
	}
	return newIndividual(classMatrix, classHours)
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
func initClassMatrix(classMatrix *types.ClassMatrix, classes []types.Class) {
	classMatrix.Init(classes)
	log.Println("Class matrix initialized")
}

// 计算固定得分
func calculateFixedScores(classMatrix *types.ClassMatrix) {
	fixedRules := constraint.GetFixedRules()
	err := classMatrix.CalcElementFixedScores(fixedRules)
	if err != nil {
		log.Fatalf("Failed to calculate fixed scores: %v", err)
	}
	log.Println("Fixed scores calculated")
}

// 分配课程矩阵
func allocateClassMatrix(classMatrix *types.ClassMatrix, classeSNs []string, classHours map[int]int) (int, error) {
	dynamicRules := constraint.GetDynamicRules()
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
