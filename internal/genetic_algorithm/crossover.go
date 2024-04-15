// crossover.go
package genetic_algorithm

import (
	"course_scheduler/internal/evaluation"
	"fmt"
	"math/rand"
)

// 交叉操作返回值
// 交叉操作返回值
type CrossoverResult struct {
	Offspring []*Individual // 交叉操作后生成的新个体
	Prepared  int           // 准备执行交叉操作的次数
	Executed  int           // 实际执行交叉操作的次数
	Err       error         // 错误信息
}

// 交叉
// 每个课班是一个染色体
// 交叉在不同个体的，相同课班的染色体之间进行
// 交叉后个体的数量不变
func Crossover(selected []*Individual, crossoverRate float64) CrossoverResult {

	offspring := make([]*Individual, 0, len(selected))
	prepared := 0
	executed := 0

	for i := 0; i < len(selected); i += 2 {
		if rand.Float64() < crossoverRate {
			prepared++
			crossPoint := rand.Intn(len(selected[i].Chromosomes))
			// 复制一份新的个体
			parent1 := selected[i].Copy()
			parent2 := selected[i+1].Copy()
			// 修复时间段冲突
			parent1.RepairTimeSlotConflicts()
			parent2.RepairTimeSlotConflicts()
			// 交叉操作
			offspring1, offspring2 := crossoverIndividuals(parent1, parent2, crossPoint)
			// 修复时间段冲突
			offspring1.RepairTimeSlotConflicts()
			offspring2.RepairTimeSlotConflicts()

			isValid, err := validateCrossover(offspring1, offspring2)
			if err != nil {
				return CrossoverResult{
					Offspring: offspring,
					Prepared:  prepared,
					Executed:  executed,
					Err:       err,
				}
			}
			if isValid {
				offspring = append(offspring, offspring1, offspring2)
				executed++
			} else {
				offspring = append(offspring, selected[i], selected[i+1])
			}

		} else {
			// 复制一份新的个体
			offspring = append(offspring, selected[i].Copy(), selected[i+1].Copy())
		}
	}

	return CrossoverResult{
		Offspring: offspring,
		Prepared:  prepared,
		Executed:  executed,
		Err:       nil,
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

	// 检查交叉后的个体是否有时间段冲突
	hasConflicts1, conflictDetails1 := offspring1.HasTimeSlotConflicts()
	if hasConflicts1 {
		return false, fmt.Errorf("offspring1 has time slot conflicts: %v", conflictDetails1)
	}

	hasConflicts2, conflictDetails2 := offspring2.HasTimeSlotConflicts()
	if hasConflicts2 {
		return false, fmt.Errorf("offspring2 has time slot conflicts: %v", conflictDetails2)
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
