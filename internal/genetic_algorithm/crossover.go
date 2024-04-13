package genetic_algorithm

import (
	"course_scheduler/internal/evaluation"
	"math/rand"
)

// 交叉
// 每个课班是一个染色体
// 交叉在不同个体的，相同课班的染色体之间进行
func Crossover(selected []*Individual, crossoverRate float64) ([]*Individual, error) {
	offspring := make([]*Individual, 0)
	for i := 0; i < len(selected); i += 2 {
		if rand.Float64() < crossoverRate {
			// Randomly select a crossover point
			crossPoint := rand.Intn(len(selected[i].Chromosomes))
			// Validate crossover
			isValid, err := validateCrossover(selected[i], selected[i+1], crossPoint)
			if err != nil {
				return offspring, err
			}

			if isValid {
				// Perform crossover operation
				offspring1 := &Individual{
					Chromosomes: append(selected[i].Chromosomes[:crossPoint], selected[i+1].Chromosomes[crossPoint:]...),
				}
				offspring2 := &Individual{
					Chromosomes: append(selected[i+1].Chromosomes[:crossPoint], selected[i].Chromosomes[crossPoint:]...),
				}
				offspring = append(offspring, offspring1, offspring2)
			} else {
				offspring = append(offspring, selected[i], selected[i+1])
			}
		} else {
			offspring = append(offspring, selected[i], selected[i+1])
		}
	}
	return offspring, nil
}

// validateCrossover 可换算法验证 验证染色体上的基因在进行基因互换杂交时是否符合基因的约束条件
func validateCrossover(individual1, individual2 *Individual, crossPoint int) (bool, error) {
	// Perform crossover temporarily
	offspring1 := &Individual{
		Chromosomes: append(individual1.Chromosomes[:crossPoint], individual2.Chromosomes[crossPoint:]...),
	}
	offspring2 := &Individual{
		Chromosomes: append(individual2.Chromosomes[:crossPoint], individual1.Chromosomes[crossPoint:]...),
	}

	// Check consistency of gene.Class between offspring1 and offspring2
	if len(offspring1.Chromosomes) != len(offspring2.Chromosomes) {
		return false, nil
	}

	for i, chromosome1 := range offspring1.Chromosomes {
		chromosome2 := offspring2.Chromosomes[i]
		if chromosome1.Genes[0].ClassSN != chromosome2.Genes[0].ClassSN {
			return false, nil
		}
	}

	// Check constraints for each gene in the offspring
	classMatrix1 := offspring1.toClassMatrix()
	for _, chromosome := range offspring1.Chromosomes {
		for _, gene := range chromosome.Genes {

			score, err := evaluation.CalcScore(classMatrix1, gene.ClassSN, gene.TeacherID, gene.VenueID, gene.TimeSlot)
			if err != nil {
				return false, err
			}

			// 之前的判断是 score != -1
			if score < 0 {
				return false, err
			}
		}
	}

	classMatrix2 := offspring1.toClassMatrix()
	for _, chromosome := range offspring2.Chromosomes {
		for _, gene := range chromosome.Genes {

			score, err := evaluation.CalcScore(classMatrix2, gene.ClassSN, gene.TeacherID, gene.VenueID, gene.TimeSlot)
			if err != nil {
				return false, err
			}

			// 之前的判断是 score != -1
			if score < 0 {
				return false, err
			}
		}
	}

	return true, nil
}
