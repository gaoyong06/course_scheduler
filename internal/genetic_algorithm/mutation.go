// mutation.go
package genetic_algorithm

import (
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"math/rand"
)

// 变异
// 每个课班是一个染色体
// 替换时间
func Mutation(selected []*Individual, mutationRate float64, classHours map[int]int) ([]*Individual, error) {
	for i := range selected {
		if rand.Float64() < mutationRate {
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

	SN, err := types.ParseSN(gene.ClassSN)
	if err != nil {
		return false, err
	}

	element := &types.Element{
		ClassSN:   gene.ClassSN,
		SubjectID: SN.SubjectID,
		GradeID:   SN.GradeID,
		ClassID:   SN.ClassID,
		TeacherID: newGene.TeacherID,
		VenueID:   newGene.VenueID,
		TimeSlot:  newGene.TimeSlot,
	}

	fixedRules := constraint.GetFixedRules()
	dynamicRules := constraint.GetDynamicRules()

	classMatrix.CalcScore(element, fixedRules, dynamicRules)
	score := classMatrix.Elements[gene.ClassSN][gene.TeacherID][gene.VenueID][gene.TimeSlot].Val.ScoreInfo.Score

	if score < 0 {
		return false, err
	}
	return true, nil
}

// findUnusedTCt 查找基因中未使用的教师,教室,时间段
func findUnusedTCt(chromosome *Chromosome) (int, int, int, error) {

	teachers := models.GetTeachers()
	venueIDs := models.GetVenueIDs()

	unusedTeacherIDs := make([]int, 0)
	unusedVenueIDs := make([]int, 0)
	unusedTimeSlots := make([]int, 0)

	// Find unused teachers, classrooms, and time slots
	for _, teacher := range teachers {
		teacherUsed := false
		for _, gene := range chromosome.Genes {
			if gene.TeacherID == teacher.TeacherID {
				teacherUsed = true
				break
			}
		}
		if !teacherUsed {
			unusedTeacherIDs = append(unusedTeacherIDs, teacher.TeacherID)
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
