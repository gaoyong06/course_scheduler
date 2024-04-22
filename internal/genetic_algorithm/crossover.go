// crossover.go
package genetic_algorithm

import (
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/types"
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
func Crossover(selected []*Individual, crossoverRate float64, classHours map[int]int) CrossoverResult {

	offspring := make([]*Individual, 0, len(selected))
	prepared := 0
	executed := 0

	// log.Printf("=== Crossover selected %d, crossoverRate: %.02f\n", len(selected), crossoverRate)

	for i := 0; i < len(selected)-1; i += 2 {
		if rand.Float64() < crossoverRate {

			prepared++
			crossPoint := rand.Intn(len(selected[i].Chromosomes)-1) + 1

			// 复制一份新的个体
			parent1 := selected[i].Copy()
			parent2 := selected[i+1].Copy()

			// 交叉操作
			offspring1, offspring2, err := crossoverIndividuals(parent1, parent2, crossPoint, classHours)
			if err != nil {
				return CrossoverResult{
					Offspring: offspring,
					Prepared:  prepared,
					Executed:  executed,
					Err:       err,
				}
			}

			// for k := 0; k < len(selected); k++ {
			// 	log.Printf("=== Crossover selected[%d], %s\n", k, selected[k].UniqueId())
			// }

			// log.Printf("=== Crossover selected[i]: %s, selected[i+1]: %s, parent1: %s, parent2: %s, offspring1: %s,  offspring2: %s\n", selected[i].UniqueId(), selected[i+1].UniqueId(), parent1.UniqueId(), parent2.UniqueId(), offspring1.UniqueId(), offspring2.UniqueId())

			// 修复时间段冲突
			_, _, err1 := offspring1.RepairTimeSlotConflicts()
			_, _, err2 := offspring2.RepairTimeSlotConflicts()

			// 交叉操作后,顺利修复时间段冲突
			if err1 == nil && err2 == nil {
				isValid, err := validateCrossover(offspring1, offspring2, classHours)
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

					// 打印交叉明细
					// log.Printf("Crossover %s, %s ----> %s, %s\n", parent1.UniqueId(), parent2.UniqueId(), offspring1.UniqueId(), offspring2.UniqueId())

				} else {

					offspring = append(offspring, selected[i], selected[i+1])
				}

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

// 两个个体之间进行交叉操作，生成两个子代个体
// 返回两个子代个体和错误信息（如果有）
func crossoverIndividuals(individual1, individual2 *Individual, crossPoint int, classHours map[int]int) (*Individual, *Individual, error) {

	// 检查交叉点是否在有效范围内
	if crossPoint <= 0 || crossPoint >= len(individual1.Chromosomes) {
		return nil, nil, fmt.Errorf("invalid crossPoint %d", crossPoint)
	}

	// 深度复制父代个体的染色体序列
	chromosomes1 := make([]*Chromosome, len(individual1.Chromosomes))
	for i, chromosome := range individual1.Chromosomes {
		chromosomes1[i] = chromosome.Copy()
	}
	chromosomes2 := make([]*Chromosome, len(individual2.Chromosomes))
	for i, chromosome := range individual2.Chromosomes {
		chromosomes2[i] = chromosome.Copy()
	}

	// 初始化子代个体的染色体序列
	offspring1 := &Individual{
		Chromosomes: make([]*Chromosome, len(individual1.Chromosomes)),
	}
	offspring2 := &Individual{
		Chromosomes: make([]*Chromosome, len(individual1.Chromosomes)),
	}

	// 为子代个体1复制基因
	for i := 0; i < len(chromosomes1); i++ {
		var source *Chromosome
		if i < crossPoint {
			source = chromosomes1[i]
		} else {
			source = chromosomes2[i]
		}
		// target := offspring1.Chromosomes[i]
		target := source.Copy()
		offspring1.Chromosomes[i] = target
	}

	// 为子代个体2复制基因
	for i := 0; i < len(chromosomes1); i++ {
		var source *Chromosome
		if i < crossPoint {
			source = chromosomes2[i]
		} else {
			source = chromosomes1[i]
		}
		// target := offspring2.Chromosomes[i]
		target := source.Copy()
		offspring2.Chromosomes[i] = target
	}

	// 个体内基因排序
	offspring1.SortChromosomes()
	offspring2.SortChromosomes()

	// 评估子代个体的适应度并赋值
	fitness1, err1 := offspring1.EvaluateFitness(classHours)
	fitness2, err2 := offspring2.EvaluateFitness(classHours)

	if err1 != nil || err2 != nil {
		return nil, nil, fmt.Errorf("ERROR: offspring evaluate fitness failed. err1: %s, err2: %s", err1.Error(), err2.Error())
	}

	offspring1.Fitness = fitness1
	offspring2.Fitness = fitness2

	// 返回两个子代个体和nil错误
	return offspring1, offspring2, nil
}

// validateCrossover 可换算法验证 验证染色体上的基因在进行基因互换杂交时是否符合基因的约束条件
func validateCrossover(offspring1, offspring2 *Individual, classHours map[int]int) (bool, error) {

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

			SN, err := types.ParseSN(gene.ClassSN)
			if err != nil {
				return false, err
			}

			element := constraint.Element{
				ClassSN:   gene.ClassSN,
				SubjectID: SN.SubjectID,
				GradeID:   SN.GradeID,
				ClassID:   SN.ClassID,
				TeacherID: gene.TeacherID,
				VenueID:   gene.VenueID,
				TimeSlot:  gene.TimeSlot,
			}

			// score, err := evaluation.CalcScore(classMatrix1, classHours, gene.ClassSN, gene.TeacherID, gene.VenueID, gene.TimeSlot)
			score, err := constraint.CalcScore(classMatrix1, element)

			if err != nil {
				return false, err
			}

			// if score.FinalScore < 0 {
			// 	return false, err
			// }
			if score < 0 {
				return false, err
			}
		}
	}

	classMatrix2 := offspring1.toClassMatrix()
	for _, chromosome := range offspring2.Chromosomes {
		for _, gene := range chromosome.Genes {

			SN, err := types.ParseSN(gene.ClassSN)
			if err != nil {
				return false, err
			}

			element := constraint.Element{
				ClassSN:   gene.ClassSN,
				SubjectID: SN.SubjectID,
				GradeID:   SN.GradeID,
				ClassID:   SN.ClassID,
				TeacherID: gene.TeacherID,
				VenueID:   gene.VenueID,
				TimeSlot:  gene.TimeSlot,
			}
			score, err := constraint.CalcScore(classMatrix2, element)

			// score, err := evaluation.CalcScore(classMatrix2, classHours, gene.ClassSN, gene.TeacherID, gene.VenueID, gene.TimeSlot)
			if err != nil {
				return false, err
			}

			if score < 0 {
				return false, err
			}

			// if score.FinalScore < 0 {
			// 	return false, err
			// }
		}
	}
	return true, nil
}
