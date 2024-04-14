package genetic_algorithm

import (
	"course_scheduler/internal/class_adapt"
	"course_scheduler/internal/types"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
)

// 初始化种群
func InitPopulation(classes []class_adapt.Class, classHours map[int]int, populationSize int) ([]*Individual, error) {
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
		var classMatrix map[string]map[int]map[int]map[int]types.Val

		// 如果 assignClassMatrix 返回了错误，就会重新执行打乱课程顺序、初始化课程矩阵、计算匹配结果值和分配课程矩阵这些步骤，直到没有返回错误为止
		for {
			// 打乱课班排课顺序
			for i := len(classes) - 1; i > 0; i-- {
				j := rand.Intn(i + 1)
				classes[i], classes[j] = classes[j], classes[i]
			}
			log.Println("Class order shuffled")

			fmt.Println("=========== classes ===========")
			for _, class := range classes {
				fmt.Println(class.String())
			}

			// 课班适应性矩阵
			classMatrix = class_adapt.InitClassMatrix(classes)
			log.Println("Class matrix initialized")

			// 匹配结果值
			err = class_adapt.MatchScore(classMatrix)
			if err != nil {
				return nil, err
			}
			log.Println("Match scores calculated")

			// 课班适应性矩阵分配
			numAssignedClasses, err = class_adapt.AllocateClassMatrix(classeSNs, classHours, classMatrix)
			fmt.Printf("numAssignedClasses: %d\n", numAssignedClasses)

			if err == nil {
				break
			}

			log.Printf("assignClassMatrix err: %s, retrying...\n", err.Error())
		}

		log.Println("Class matrix assigned")

		// 生成个体
		individual := newIndividual(classMatrix)

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

	// 将新生成的个体添加到种群中
	population = append(population, offspring...)
	// 根据适应度值进行排序
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})

	// Create a new slice to store non-nil individuals
	newPopulation := make([]*Individual, 0, len(population)/2)

	// Copy non-nil individuals to the new slice
	for _, individual := range population[:len(population)/2] {
		if individual != nil {
			newPopulation = append(newPopulation, individual)
		}
	}

	return newPopulation
}
