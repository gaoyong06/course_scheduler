// crossover.go
package genetic_algorithm

import (
	"course_scheduler/internal/evaluation"
	"fmt"
	"math/rand"
)

// 交叉操作返回值
type CrossoverResult struct {
	Offsprings        []*Individual // 交叉操作后返回个体
	PrepareCrossover  int           // 准备执行交叉操作的次数
	ExecutedCrossover int           // 实际执行交叉操作的次数
	Error             error         // 错误信息
}

// 交叉
// 每个课班是一个染色体
// 交叉在不同个体的，相同课班的染色体之间进行
// 交叉后个体的数量不变
// 交叉
// 每个课班是一个染色体
// 交叉在不同个体的，相同课班的染色体之间进行
// 交叉后个体的数量不变
func Crossover(selected []*Individual, crossoverRate float64) CrossoverResult {

	offspring := make([]*Individual, 0, len(selected))
	prepareCrossover := 0
	executedCrossover := 0

	for i := 0; i < len(selected); i += 2 {

		if rand.Float64() < crossoverRate {
			prepareCrossover++
			crossPoint := rand.Intn(len(selected[i].Chromosomes))
			offspring1, offspring2 := crossoverIndividuals(selected[i], selected[i+1], crossPoint)

			// Repair time slot conflicts
			// offspring1.RepairTimeSlotConflicts1()
			// offspring2.RepairTimeSlotConflicts1()
			fmt.Println("offspring1 开始修复冲突")
			conflictCount1, repairs1, err1 := offspring1.RepairTimeSlotConflicts()
			fmt.Printf("offspring1 结束修复冲突 conflictCount1: %d, repairs1: %#v, isRepaired1: %v\n", conflictCount1, repairs1, err1)

			fmt.Println("\n")

			fmt.Println("offspring2 开始修复冲突")
			conflictCount2, repairs2, err2 := offspring2.RepairTimeSlotConflicts()

			fmt.Printf("offspring2 结束修复冲突 conflictCount2: %d, repairs2: %#v, isRepaired2: %v\n", conflictCount2, repairs2, err2)
			fmt.Println("\n\n\n")

			isValid, err := validateCrossover(offspring1, offspring2)
			if err != nil {
				return CrossoverResult{
					Offsprings:        offspring,
					PrepareCrossover:  prepareCrossover,
					ExecutedCrossover: executedCrossover,
					Error:             err,
				}
			}

			if isValid {
				offspring = append(offspring, offspring1, offspring2)
				executedCrossover++
			} else {
				offspring = append(offspring, selected[i], selected[i+1])
			}
		} else {
			offspring = append(offspring, selected[i], selected[i+1])
		}
	}

	// fmt.Printf("Prepare crossover: %d, Executed crossover: %d\n", prepareCrossover, executedCrossover)

	return CrossoverResult{
		Offsprings:        offspring,
		PrepareCrossover:  prepareCrossover,
		ExecutedCrossover: executedCrossover,
		Error:             nil,
	}
}

// 两个个体之间的交叉操作
func crossoverIndividuals(individual1, individual2 *Individual, crossPoint int) (*Individual, *Individual) {

	offspring1 := &Individual{
		Chromosomes: append(individual1.Chromosomes[:crossPoint], individual2.Chromosomes[crossPoint:]...),
	}
	offspring2 := &Individual{
		Chromosomes: append(individual2.Chromosomes[:crossPoint], individual1.Chromosomes[crossPoint:]...),
	}

	return offspring1, offspring2
}

// validateCrossover 可换算法验证 验证染色体上的基因在进行基因互换杂交时是否符合基因的约束条件
func validateCrossover(offspring1, offspring2 *Individual) (bool, error) {

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

			if score < 0 {
				return false, err
			}
		}
	}
	return true, nil
}
