// mutation.go
package genetic_algorithm

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
	"log"
	"math/rand"

	"github.com/samber/lo"
)

// 变异
// 变异即是染色体基因位更改为其他结果，如替换老师或者时间或者教室，替换的老师或者时间或者教室从未出现在对应课班上，但是是符合老师或者教室的约束性条件，理论上可以匹配该课班
// 每个课班是一个染色体
func Mutation(selected []*Individual, mutationRate float64, schedule *models.Schedule, teachAllocs []*models.TeachTaskAllocation, subjects []*models.Subject, teachers []*models.Teacher, grades []*models.Grade, venueMap map[string][]int, constraints map[string]interface{}) ([]*Individual, int, int, error) {

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
			unusedTeacherID, unusedVenueID, unusedTimeSlotStr, err := findUnusedTCt(selected[i], chromosome, gene, schedule, teachAllocs, teachers, venueMap)
			// fmt.Printf("Mutation unusedTeacherID: %d, unusedVenueID: %d, unusedTimeSlot: %d\n", unusedTeacherID, unusedVenueID, unusedTimeSlot)
			if err != nil {
				return nil, prepared, executed, err
			}

			// 变异校验
			isValid, err := validateMutation(selected[i], chromosomeIndex, geneIndex, unusedTeacherID, unusedVenueID, unusedTimeSlotStr)
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

				if unusedTimeSlotStr != "" {
					timeSlots := utils.ParseTimeSlotStr(unusedTimeSlotStr)
					gene.TimeSlots = timeSlots
				}
				chromosome.Genes[geneIndex] = gene

				selected[i].Chromosomes[chromosomeIndex] = chromosome

				// 修复时间段冲突
				_, _, err := selected[i].RepairTimeSlotConflicts(schedule, grades)
				if err != nil {
					return nil, prepared, executed, err
				}

				// 个体内基因排序
				selected[i].SortChromosomes()

				// 更新个体适应度
				classMatrix, err := selected[i].toClassMatrix(schedule, teachAllocs, subjects, teachers, venueMap, constraints)
				if err != nil {
					return nil, prepared, executed, err
				}

				newFitness, err := selected[i].EvaluateFitness(classMatrix, schedule, subjects, teachers, constraints)
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
func validateMutation(individual *Individual, chromosomeIndex, geneIndex, unusedTeacherID, unusedVenueID int, unusedTimeSlotStr string) (bool, error) {

	// 检查突变是否会产生有效的基因，找到未使用的教师、教室和时间段
	newGene := individual.Chromosomes[chromosomeIndex].Genes[geneIndex]

	if models.IsTeacherIDValid(unusedTeacherID) {
		newGene.TeacherID = unusedTeacherID
	}

	if models.IsVenueIDValid(unusedVenueID) {
		newGene.VenueID = unusedVenueID
	}

	if unusedTimeSlotStr != "" {

		timeSlots := utils.ParseTimeSlotStr(unusedTimeSlotStr)

		if newGene.IsConnected && len(timeSlots) == 2 {
			newGene.TimeSlots = timeSlots
		}

		if !newGene.IsConnected && len(timeSlots) == 1 {
			newGene.TimeSlots = timeSlots
		}
	}

	return true, nil
}

// findUnusedTCt 查找基因中未使用的教师,教室,时间段
// TODO: 这个方法逻辑太长了，需要优化
func findUnusedTCt(individual *Individual, chromosome *Chromosome, gene *Gene, schedule *models.Schedule, teachAllocs []*models.TeachTaskAllocation, teachers []*models.Teacher, venueMap map[string][]int) (int, int, string, error) {

	SN, err := types.ParseSN(chromosome.ClassSN)
	if err != nil {
		return -1, -1, "", err
	}

	subjectID := SN.SubjectID
	gradeID := SN.GradeID
	classID := SN.ClassID
	isConnected := gene.IsConnected

	teacherIDs := models.ClassTeacherIDs(gradeID, classID, subjectID, teachers)
	venueIDs := models.ClassVenueIDs(gradeID, classID, subjectID, venueMap)
	// timeSlots := types.ClassTimeSlots(teacherIDs, venueIDs)

	unusedTeacherIDs := make([]int, 0)
	unusedVenueIDs := make([]int, 0)
	unusedGeneTimeSlotStrs := make([]string, 0)
	unusedTeacherTimeSlotStrs := make([]string, 0)
	// unusedVenueTimeSlots := make([]int, 0)

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

	// 基因未使用的时间集合
	timeSlotStrs, err := types.ClassTimeSlots(schedule, teachAllocs, gradeID, classID, subjectID, unusedTeacherIDs, unusedVenueIDs)
	if err != nil {
		return -1, -1, "", err
	}

	for _, timeSlotStr := range timeSlotStrs {

		timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
		timeSlotUsed := false
		for _, gene := range chromosome.Genes {

			intersect := lo.Intersect(timeSlots, gene.TimeSlots)
			if len(intersect) > 0 {
				timeSlotUsed = true
				break
			}
		}
		if !timeSlotUsed {
			unusedGeneTimeSlotStrs = append(unusedGeneTimeSlotStrs, timeSlotStr)
		}
	}

	// 空闲时间段,是依赖教师和教室的
	// fmt.Printf("unusedTeacherIDs: %#v, unusedVenueIDs: %#v, unusedTimeSlots: %#v\n", unusedTeacherIDs, unusedVenueIDs, unusedTimeSlots)
	unusedTeacherID := -1
	unusedVenueID := -1
	unusedTimeSlotStr := ""

	if len(unusedTeacherIDs) > 0 {
		unusedTeacherID = getRandomUnused(unusedTeacherIDs)
		for _, timeSlotStr := range timeSlotStrs {
			timeSlotUsed := false
			// 确定该教师的空闲时间段
			for _, chromosome := range individual.Chromosomes {
				for _, gene := range chromosome.Genes {

					timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
					intersect := lo.Intersect(timeSlots, gene.TimeSlots)

					if len(intersect) > 0 && gene.TeacherID == unusedTeacherID {
						timeSlotUsed = true
						break
					}
				}
			}
			if !timeSlotUsed {
				unusedTeacherTimeSlotStrs = append(unusedTeacherTimeSlotStrs, timeSlotStr)
			}
		}
	}

	// TODO: 如果这里的教室,是专用教学场所, 则这里还需要将专用教学场地的时间纳入参与计算
	unusedVenueID = getRandomUnused(unusedVenueIDs)

	unusedTimeSlotStrs := make([]string, 0)
	if len(unusedGeneTimeSlotStrs) > 0 && len(unusedTeacherTimeSlotStrs) > 0 {
		unusedTimeSlotStrs = lo.Intersect(unusedGeneTimeSlotStrs, unusedTeacherTimeSlotStrs)
	} else {
		unusedTimeSlotStrs = unusedGeneTimeSlotStrs
	}

	unusedTimeSlotStr = getRandomUnusedTimeSlotStr(unusedTimeSlotStrs, isConnected)

	// 返回一个随机的未使用的老师、教室和时间段(如果有的话)
	return unusedTeacherID, unusedVenueID, unusedTimeSlotStr, nil
}

// 根据是否是连堂课,从values中随机取一个未使用的时间段字符串
func getRandomUnusedTimeSlotStr(values []string, isConnected bool) string {

	if len(values) == 0 {
		return ""
	}

	var connectedGroup []string
	var normalGroup []string

	for _, value := range values {

		timeSlots := utils.ParseTimeSlotStr(value)
		if len(timeSlots) == 2 {
			connectedGroup = append(connectedGroup, value)
		} else {
			normalGroup = append(normalGroup, value)
		}
	}

	if isConnected {
		return connectedGroup[rand.Intn(len(connectedGroup))]
	} else {
		return normalGroup[rand.Intn(len(normalGroup))]
	}
}

// 从values中随机取一个数字
func getRandomUnused(values []int) int {
	if len(values) == 0 {
		return 0
	}
	return values[rand.Intn(len(values))]
}
