// mutation.go
package genetic_algorithm

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"log"
	"math/rand"
)

// 变异
// 变异即是染色体基因位更改为其他结果，如替换老师或者时间或者教室，替换的老师或者时间或者教室从未出现在对应课班上，但是是符合老师或者教室的约束性条件，理论上可以匹配该课班
// 每个课班是一个染色体
// Mutation performs mutation on the selected individuals with a given mutation rate
func Mutation(selected []*Individual, mutationRate float64, schedule *models.Schedule, teachAllocs []*models.TeachTaskAllocation, subjects []*models.Subject, teachers []*models.Teacher, venueMap map[string][]int) ([]*Individual, int, int, error) {

	prepared := 0
	executed := 0

	for i := range selected {
		if rand.Float64() < mutationRate {
			prepared++

			// 变异的个体
			// individual := selected[i]

			// 随机选择染色体和基因索引进行突变
			chromosomeIndex := rand.Intn(len(selected[i].Chromosomes))
			geneIndex := rand.Intn(len(selected[i].Chromosomes[chromosomeIndex].Genes))

			// 获取要突变的染色体和基因
			chromosome := selected[i].Chromosomes[chromosomeIndex]
			gene := chromosome.Genes[geneIndex]

			// 找到给定课班的可用参数信息
			unusedTeacherID, unusedVenueID, unusedTimeSlot, err := findUnusedTCt(chromosome, schedule, teachers, venueMap)
			// fmt.Printf("Mutation unusedTeacherID: %d, unusedVenueID: %d, unusedTimeSlot: %d\n", unusedTeacherID, unusedVenueID, unusedTimeSlot)
			if err != nil {
				return nil, prepared, executed, err
			}

			// 变异校验
			isValid, err := validateMutation(selected[i], chromosomeIndex, geneIndex, unusedTeacherID, unusedVenueID, unusedTimeSlot)
			if err != nil {
				return nil, prepared, executed, err
			}

			// 校验通过
			if isValid {

				executed++

				// 用未使用的值(如果有的话)改变基因
				if unusedTeacherID > 0 {
					gene.TeacherID = unusedTeacherID
				}
				if unusedVenueID > 0 {
					gene.VenueID = unusedVenueID
				}
				if unusedTimeSlot > 0 {
					gene.TimeSlot = unusedTimeSlot
				}
				chromosome.Genes[geneIndex] = gene

				selected[i].Chromosomes[chromosomeIndex] = chromosome

				// 修复时间段冲突
				_, _, err := selected[i].RepairTimeSlotConflicts(schedule)
				if err != nil {
					return nil, prepared, executed, err
				}

				// 个体内基因排序
				selected[i].SortChromosomes()

				// 更新个体适应度
				classMatrix, err := selected[i].toClassMatrix(schedule, teachAllocs, subjects, teachers, venueMap)
				if err != nil {
					return nil, prepared, executed, err
				}

				newFitness, err := selected[i].EvaluateFitness(classMatrix, schedule, subjects, teachers)
				if err != nil {
					return nil, prepared, executed, err
				}
				selected[i].Fitness = newFitness
			}
		}
	}

	log.Printf("Prepared mutations: %d, Executed mutations: %d\n", prepared, executed)
	return selected, prepared, executed, nil
}

// validateMutation 可行性验证 用于验证染色体上的基因在进行基因变异更换时是否符合基因的约束条件
func validateMutation(individual *Individual, chromosomeIndex, geneIndex, unusedTeacherID, unusedVenueID, unusedTimeSlot int) (bool, error) {

	// 检查突变是否会产生有效的基因，找到未使用的教师、教室和时间段
	newGene := individual.Chromosomes[chromosomeIndex].Genes[geneIndex]

	if models.IsTeacherIDValid(unusedTeacherID) {
		newGene.TeacherID = unusedTeacherID
	}

	if models.IsVenueIDValid(unusedVenueID) {
		newGene.VenueID = unusedVenueID
	}

	if unusedTimeSlot >= 0 {
		newGene.TimeSlot = unusedTimeSlot
	}

	return true, nil
}

// findUnusedTCt 查找基因中未使用的教师,教室,时间段
func findUnusedTCt(chromosome *Chromosome, schedule *models.Schedule, teachers []*models.Teacher, venueMap map[string][]int) (int, int, int, error) {

	SN, err := types.ParseSN(chromosome.ClassSN)
	if err != nil {
		return -1, -1, -1, err
	}

	subjectID := SN.SubjectID
	gradeID := SN.GradeID
	classID := SN.ClassID

	teacherIDs := models.ClassTeacherIDs(gradeID, classID, subjectID, teachers)
	venueIDs := models.ClassVenueIDs(gradeID, classID, subjectID, venueMap)
	// timeSlots := types.ClassTimeSlots(teacherIDs, venueIDs)

	unusedTeacherIDs := make([]int, 0)
	unusedVenueIDs := make([]int, 0)
	unusedTimeSlots := make([]int, 0)

	// 找到闲置的老师、教室和时间段
	for _, teacherID := range teacherIDs {
		teacherUsed := false
		for _, gene := range chromosome.Genes {
			if gene.TeacherID == teacherID {
				teacherUsed = true
				break
			}
		}
		if !teacherUsed {
			unusedTeacherIDs = append(unusedTeacherIDs, teacherID)
		}
	}

	for _, venueID := range venueIDs {
		venueUsed := false
		for _, gene := range chromosome.Genes {
			if gene.VenueID == venueID {
				venueUsed = true
				break
			}
		}
		if !venueUsed {
			unusedVenueIDs = append(unusedVenueIDs, venueID)
		}
	}

	// 时间集合
	timeSlots := types.ClassTimeSlots(schedule, unusedTeacherIDs, unusedVenueIDs)
	for _, timeSlot := range timeSlots {
		timeSlotUsed := false
		for _, gene := range chromosome.Genes {
			if gene.TimeSlot == timeSlot {
				timeSlotUsed = true
				break
			}
		}
		if !timeSlotUsed {
			unusedTimeSlots = append(unusedTimeSlots, timeSlot)
		}
	}

	// fmt.Printf("unusedTeacherIDs: %#v, unusedVenueIDs: %#v, unusedTimeSlots: %#v\n", unusedTeacherIDs, unusedVenueIDs, unusedTimeSlots)

	// 返回一个随机的未使用的老师、教室和时间段(如果有的话)
	return getRandomUnused(unusedTeacherIDs), getRandomUnused(unusedVenueIDs), getRandomUnused(unusedTimeSlots), nil
}

// getRandomUnused returns a random unused value from the given slice, or an empty string if the slice is empty
func getRandomUnused(values []int) int {
	if len(values) == 0 {
		return 0
	}
	return values[rand.Intn(len(values))]
}
