// mutation.go
package genetic_algorithm

import (
	"course_scheduler/internal/constraints"

	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
	"errors"
	"fmt"
	"log"
	"math/rand"

	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// 变异操作
// 变异即是染色体基因位更改为其他结果，如替换老师或者时间或者教室，替换的老师或者时间或者教室从未出现在对应课班上，但是是符合老师或者教室的约束性条件，理论上可以匹配该课班
// 每个课班是一个染色体
// 参数:
//
//	selected: 选择的个体
//	mutationRate: 变异率
//	schedule: 课表方案
//	teachingTasks: 教学计划
//	teachers: 教师信息
//	grades: 年级信息
//	subjectVenueMap: 科目与教学场地
//	constraintMap: 约束条件
//
// 返回值:
//
//	返回 交叉后的个体、准备交叉次数、实际交叉次数、错误信息

func Mutation(selected []*Individual, mutationRate float64, schedule *models.Schedule, teachingTasks []*models.TeachingTask, subjects []*models.Subject, teachers []*models.Teacher, grades []*models.Grade, venueMap map[string][]int, constraintMap map[string]interface{}) ([]*Individual, int, int, error) {

	prepared := 0
	executed := 0

	for i := range selected {
		if rand.Float64() < mutationRate {
			prepared++

			// 变异的个体
			// 随机选择染色体和基因索引进行突变
			chromosomeIndex := rand.Intn(len(selected[i].Chromosomes))
			geneIndex := rand.Intn(len(selected[i].Chromosomes[chromosomeIndex].Genes))

			// 获取要突变的染色体和基因
			chromosome := selected[i].Chromosomes[chromosomeIndex]
			gene := chromosome.Genes[geneIndex]

			// 基因变异和校验
			err := mutationAndValidate(selected[i], chromosome, gene, schedule, teachingTasks, subjects, teachers, venueMap, constraintMap)
			if err != nil {
				log.Printf("mutation failed. err: %v\n", err)
			} else {
				executed++
			}
		}
	}

	log.Printf("Prepared mutations: %d, Executed mutations: %d\n", prepared, executed)
	return selected, prepared, executed, nil
}

// mutationAndValidate 可行性验证 用于验证染色体上的基因在进行基因变异更换时是否符合基因的约束条件
func mutationAndValidate(individual *Individual, chromosome *Chromosome, gene *Gene, schedule *models.Schedule, teachingTasks []*models.TeachingTask, subjects []*models.Subject, teachers []*models.Teacher, venueMap map[string][]int, constraintMap map[string]interface{}) error {

	err := mutationGene(individual, chromosome, gene, schedule, teachingTasks, subjects, teachers, venueMap, constraintMap)

	// 校验的过程...
	return err
}

// 基因变异
func mutationGene(individual *Individual, chromosome *Chromosome, gene *Gene, schedule *models.Schedule, teachingTasks []*models.TeachingTask, subjects []*models.Subject, teachers []*models.Teacher, venueMap map[string][]int, constraintMap map[string]interface{}) error {

	constr1 := constraintMap["Class"].([]*constraints.Class)
	constr2 := constraintMap["Teacher"].([]*constraints.Teacher)

	// 查找基因中未使用的教师或教室或时间段
	teacherID, venueID, timeSlotStr, err := findRandomScheduleForGene(individual, chromosome, gene, schedule, teachers, venueMap, constr1, constr2)
	if err != nil {
		return err
	}
	fmt.Printf("find random schedule for gene teacherID: %d, venueID: %d, timeSlotStr: %s\n", teacherID, venueID, timeSlotStr)

	// 用未使用的值(如果有的话)改变基因
	if teacherID > 0 {
		gene.TeacherID = teacherID
	}

	if venueID > 0 {
		gene.VenueID = venueID
	}

	if timeSlotStr != "" {
		timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
		gene.TimeSlots = timeSlots
	}

	// 修复个体时间段冲突
	_, err = individual.resolveConflicts(schedule, teachers, constr1, constr2)
	if err != nil {
		return err
	}

	// 个体内基因排序
	individual.sortChromosomes()

	// 更新个体适应度
	classMatrix, err := individual.toClassMatrix(schedule, teachingTasks, subjects, teachers, venueMap, constraintMap)
	if err != nil {
		return err
	}

	newFitness, err := individual.evaluateFitness(classMatrix, schedule, subjects, teachers, constraintMap)
	if err != nil {
		return err
	}
	individual.Fitness = newFitness

	// 更新UniqueId
	uniqueId := individual.genUniqueId()
	individual.UniqueId = uniqueId

	return nil

}

// findRandomScheduleForGene 查找基因中未使用的教师或教室或时间段
func findRandomScheduleForGene(individual *Individual, chromosome *Chromosome, gene *Gene, schedule *models.Schedule, teachers []*models.Teacher, venueMap map[string][]int, constr1 []*constraints.Class, constr2 []*constraints.Teacher) (int, int, string, error) {

	SN, err := types.ParseSN(gene.ClassSN)
	if err != nil {
		return 0, 0, "", err
	}

	gradeID := SN.GradeID
	classID := SN.ClassID
	teacherID := gene.TeacherID
	venueID := gene.VenueID
	isConnected := gene.IsConnected

	// 随机获取一个闲置的教师
	idleTeacherID, err := randomIdleTeacherID(chromosome, gene, teachers)
	if err != nil {
		return 0, 0, "", err
	}

	if idleTeacherID > 0 {
		teacherID = idleTeacherID
	}

	// 随机获取一个闲置的教室
	idleVenueID, err := randomIdleVenueID(chromosome, gene, venueMap)
	if err != nil {
		return 0, 0, "", err
	}

	if idleVenueID > 0 {
		venueID = idleVenueID
	}

	// 班级可用时间段
	classConnected, classNormal := individual.getClassValidTimeSlots(schedule, constr1)
	classKey := fmt.Sprintf("%d_%d", gradeID, classID)
	teacherIDStr := cast.ToString(teacherID)
	var timeSlotStrs []string

	// 教师可用时间段
	teacherConnected, teacherNormal, err := individual.getTeacherValidTimeSlots(schedule, teachers, constr2)
	if err != nil {
		return 0, 0, "", err
	}

	// 即是班级可用的时间段,又是教师可用的时间段
	if isConnected {
		timeSlotStrs = lo.Intersect(classConnected[classKey], teacherConnected[teacherIDStr])
	} else {
		timeSlotStrs = lo.Intersect(classNormal[classKey], teacherNormal[teacherIDStr])
	}

	// 随机从可用时间段中取一个
	timeSlotStrVal, err := randomSample(timeSlotStrs)
	if err != nil {
		return 0, 0, "", err
	}
	timeSlotStr := timeSlotStrVal.(string)

	return teacherID, venueID, timeSlotStr, nil
}

// 随机获取基因中未使用的教师ID
func randomIdleTeacherID(chromosome *Chromosome, gene *Gene, teachers []*models.Teacher) (int, error) {

	SN, err := types.ParseSN(gene.ClassSN)
	if err != nil {
		return 0, err
	}

	subjectID := SN.SubjectID
	gradeID := SN.GradeID
	classID := SN.ClassID

	teacherID := 0
	teacherIDs := models.ClassTeacherIDs(gradeID, classID, subjectID, teachers)
	unusedTeacherIDs := make([]int, 0)

	// 找到闲置的老师
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

	if len(unusedTeacherIDs) > 0 {
		teacherIDVal, err := randomSample(unusedTeacherIDs)
		if err != nil {
			return 0, err
		}
		teacherID = teacherIDVal.(int)
	}

	return teacherID, nil
}

// 随机获取基因中未使用的教学场地ID
func randomIdleVenueID(chromosome *Chromosome, gene *Gene, venueMap map[string][]int) (int, error) {

	SN, err := types.ParseSN(gene.ClassSN)
	if err != nil {
		return 0, err
	}

	subjectID := SN.SubjectID
	gradeID := SN.GradeID
	classID := SN.ClassID

	venueID := 0
	venueIDs := models.ClassVenueIDs(gradeID, classID, subjectID, venueMap)
	unusedVenueIDs := make([]int, 0)

	// 找到闲置的教室
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

	// TODO: 如果这里的教室,是专用教学场所, 则这里还需要将专用教学场地的时间纳入参与计算
	if len(unusedVenueIDs) > 0 {
		venueIDVal, err := randomSample(unusedVenueIDs)
		if err != nil {
			return 0, err
		}
		venueID = venueIDVal.(int)
	}

	return venueID, nil
}

// 从values中随机取一个
func randomSample(values interface{}) (interface{}, error) {
	switch v := values.(type) {
	case []string:
		if len(v) == 0 {
			return "", nil
		}
		return v[rand.Intn(len(v))], nil
	case []int:
		if len(v) == 0 {
			return 0, nil
		}
		return v[rand.Intn(len(v))], nil
	default:
		return nil, errors.New("invalid type for values")
	}
}
