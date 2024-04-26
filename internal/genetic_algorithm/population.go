// population.go
package genetic_algorithm

import (
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/types"
	"fmt"
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

		var err error
		var numAssignedClasses int
		classMatrix := types.NewClassMatrix()

		// var classMatrix map[string]map[int]map[int]map[int]types.Val

		// 设置最大重试次数
		const maxRetries = 1000

		// 如果 assignClassMatrix 返回了错误，就会重新执行打乱课程顺序、初始化课程矩阵、计算匹配结果值和分配课程矩阵这些步骤，直到没有返回错误为止
		for retryCount := 0; ; retryCount++ {
			// 打乱课班排课顺序
			for i := len(classes) - 1; i > 0; i-- {
				j := rand.Intn(i + 1)
				classes[i], classes[j] = classes[j], classes[i]
			}
			log.Println("Class order shuffled")

			// fmt.Println("=========== classes ===========")
			// for _, class := range classes {
			// 	fmt.Println(class.String())
			// }

			// 课班适应性矩阵
			// classMatrix = class_adapt.InitClassMatrix(classes)
			classMatrix.Init(classes)
			log.Println("Class matrix initialized")

			// 计算课班适应性矩阵各个元素固定约束条件下的得分

			fixedRules := constraint.GetFixedRules()
			err = classMatrix.CalcElementFixedScores(fixedRules)
			if err != nil {
				return nil, err
			}
			log.Println("Fixed scores calculated")

			// utils.PrintClassMatrix(classMatrix.Elements)

			// 课班适应性矩阵分配
			dynamicRules := constraint.GetDynamicRules()
			numAssignedClasses, err = classMatrix.Allocate(classeSNs, classHours, dynamicRules)
			log.Printf("numAssignedClasses: %d\n", numAssignedClasses)

			if err == nil {
				break
			}

			log.Printf("assignClassMatrix err: %s, retrying...\n", err.Error())

			// 检查重试次数是否达到最大值
			if retryCount >= maxRetries {
				return nil, fmt.Errorf("max retries (%d) reached while trying to allocate class matrix", maxRetries)
			}
		}

		log.Println("Class matrix assigned")

		// 生成个体
		individual, err := newIndividual(classMatrix, classHours)
		if err != nil {
			return nil, err
		}

		fmt.Println("打印矩阵中有冲突的元素")
		classMatrix.PrintConstraintElement()

		// fmt.Println("================================")
		// individual.PrintSchedule()

		population = append(population, individual)
		log.Println("Individual initialized")
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
