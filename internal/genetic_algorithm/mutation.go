// mutation.go
package genetic_algorithm

import (
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"log"
	"math/rand"
)

// 变异
// 每个课班是一个染色体
// 替换时间
// Mutation performs mutation on the selected individuals with a given mutation rate
func Mutation(selected []*Individual, mutationRate float64, classHours map[int]int) ([]*Individual, error) {

	prepared := 0
	executed := 0

	for i := range selected {
		if rand.Float64() < mutationRate {
			prepared++

			// Randomly select a chromosome and gene index for mutation
			chromosomeIndex := rand.Intn(len(selected[i].Chromosomes))
			geneIndex := rand.Intn(len(selected[i].Chromosomes[chromosomeIndex].Genes))

			// Get the chromosome and gene to be mutated
			chromosome := selected[i].Chromosomes[chromosomeIndex]
			gene := chromosome.Genes[geneIndex]

			// Find available options for the given class
			unusedTeacherID, unusedVenueID, unusedTimeSlot, err := findUnusedTCt(chromosome)
			if err != nil {
				return nil, err
			}

			// Validate the mutation
			isValid, err := validateMutation(selected[i], gene, unusedTeacherID, unusedVenueID, unusedTimeSlot, classHours)
			if err != nil {
				return nil, err
			}

			if isValid {
				executed++
				// Mutate the gene with the unused values (if available)
				if unusedTeacherID > 0 {
					gene.TeacherID = unusedTeacherID
				}
				if unusedVenueID > 0 {
					gene.VenueID = unusedVenueID
				}
				if unusedTimeSlot > 0 {
					gene.TimeSlot = unusedTimeSlot
				}
			}
		}
	}

	log.Printf("Prepared mutations: %d, Executed mutations: %d\n", prepared, executed)
	return selected, nil
}

// validateMutation 可行性验证 用于验证染色体上的基因在进行基因变异更换时是否符合基因的约束条件
func validateMutation(individual *Individual, gene Gene, unusedTeacherID, unusedVenueID, unusedTimeSlot int, classHours map[int]int) (bool, error) {
	// Check if the mutation will result in a valid gene
	newGene := gene
	if unusedTeacherID != 0 {
		newGene.TeacherID = unusedTeacherID
	}
	if unusedVenueID != 0 {
		newGene.VenueID = unusedVenueID
	}
	if unusedTimeSlot != 0 {
		newGene.TimeSlot = unusedTimeSlot
	}

	// Calculate the score for the new gene
	classMatrix := individual.toClassMatrix()

	// SN, err := types.ParseSN(gene.ClassSN)
	// if err != nil {
	// 	return false, err
	// }

	oldElement := classMatrix.Elements[gene.ClassSN][gene.TeacherID][gene.VenueID][gene.TimeSlot]
	oldElement.Val.Used = 0

	newElement := classMatrix.Elements[gene.ClassSN][newGene.TeacherID][newGene.VenueID][newGene.TimeSlot]
	newElement.Val.Used = 1

	// &types.Element{
	// 	ClassSN:   gene.ClassSN,
	// 	SubjectID: SN.SubjectID,
	// 	GradeID:   SN.GradeID,
	// 	ClassID:   SN.ClassID,
	// 	TeacherID: newGene.TeacherID,
	// 	VenueID:   newGene.VenueID,
	// 	TimeSlot:  newGene.TimeSlot,
	// }

	// 更新元素的得分和矩阵的总分数
	fixedRules := constraint.GetFixedRules()
	dynamicRules := constraint.GetDynamicRules()

	oldScore := classMatrix.Score
	classMatrix.UpdateElementScore(newElement, fixedRules, dynamicRules)
	newScore := classMatrix.SumUsedElementsScore()
	classMatrix.Score = newScore

	// elementScore := classMatrix.Elements[gene.ClassSN][gene.TeacherID][gene.VenueID][gene.TimeSlot].Val.ScoreInfo.Score

	oldElementScore := oldElement.Val.ScoreInfo.Score
	newElementScore := newElement.Val.ScoreInfo.Score

	if newElementScore <= 0 {
		return false, nil
	}

	fitness1 := individual.Fitness
	individual.EvaluateFitness(classMatrix, classHours)
	fitness2 := individual.Fitness

	fmt.Println("FFFFFFFFFFFFUCK")

	log.Printf("Mutations fitness1:  %d, fitness2: %d, oldScore: %d, newScore: %d, oldElementScore: %d, newElementScore: %d\n", fitness1, fitness2, oldScore, newScore, oldElementScore, newElementScore)

	return true, nil
}

// findUnusedTCt 查找基因中未使用的教师,教室,时间段
func findUnusedTCt(chromosome *Chromosome) (int, int, int, error) {

	// teachers := models.GetTeachers()
	// venueIDs := models.GetVenueIDs()

	SN, err := types.ParseSN(chromosome.ClassSN)
	if err != nil {
		return -1, -1, -1, err
	}

	subjectID := SN.SubjectID
	classID := SN.ClassID

	teacherIDs := models.ClassTeacherIDs(subjectID)
	venueIDs := models.ClassVenueIDs(classID)
	// timeSlots := types.ClassTimeSlots(teacherIDs, venueIDs)

	unusedTeacherIDs := make([]int, 0)
	unusedVenueIDs := make([]int, 0)
	unusedTimeSlots := make([]int, 0)

	// Find unused teachers, classrooms, and time slots
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
	timeSlots := types.ClassTimeSlots(unusedTeacherIDs, unusedVenueIDs)
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

	// Return a random unused teacher, classroom, and time slot (if available)
	return getRandomUnused(unusedTeacherIDs), getRandomUnused(unusedVenueIDs), getRandomUnused(unusedTimeSlots), nil
}

// getRandomUnused returns a random unused value from the given slice, or an empty string if the slice is empty
func getRandomUnused(values []int) int {
	if len(values) == 0 {
		return 0
	}
	return values[rand.Intn(len(values))]
}
