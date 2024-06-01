// crossover.go
package genetic_algorithm

import (
	"course_scheduler/internal/constraints"
	"course_scheduler/internal/models"
	"fmt"
	"log"
	"math/rand"
)

// 交叉操作返回值

// 交叉
// 每个课班是一个染色体
// 交叉在不同个体的，相同课班的染色体之间进行
// 交叉后个体的数量不变
func Crossover(selected []*Individual, crossoverRate float64, schedule *models.Schedule, teachTaskAllocations []*models.TeachTaskAllocation, subjects []*models.Subject, teachers []*models.Teacher, grades []*models.Grade, subjectVenueMap map[string][]int, constraintMap map[string]interface{}) ([]*Individual, int, int, error) {

	offspring := make([]*Individual, 0, len(selected))
	prepared := 0
	executed := 0

	constr1 := constraintMap["Class"].([]*constraints.Class)
	constr2 := constraintMap["Teacher"].([]*constraints.Teacher)

	for i := 0; i < len(selected)-1; i += 2 {
		if rand.Float64() < crossoverRate {
			prepared++
			crossPoint := rand.Intn(len(selected[i].Chromosomes))

			// 复制一份新的个体
			parent1 := selected[i].Copy()
			parent2 := selected[i+1].Copy()

			// 交叉操作并进行后续检查
			offspring1, offspring2, err := crossoverAndValidate(parent1, parent2, crossPoint, schedule, grades, teachers, constr1, constr2)

			// 如果交叉操作出现错误, 则撤销当前交叉操作
			if err != nil {

				log.Printf("undo the current crossover operation. prepared: %d, executed: %d, err: %s", prepared, executed, err)
				offspring = append(offspring, selected[i], selected[i+1])

			} else {

				// 评估子代个体的适应度并赋值
				offspringClassMatrix1, err1 := offspring1.toClassMatrix(schedule, teachTaskAllocations, subjects, teachers, subjectVenueMap, constraintMap)
				offspringClassMatrix2, err2 := offspring2.toClassMatrix(schedule, teachTaskAllocations, subjects, teachers, subjectVenueMap, constraintMap)
				if err1 != nil || err2 != nil {
					return offspring, prepared, executed, fmt.Errorf("ERROR: offspring evaluate fitness failed. err1: %s, err2: %s", err1.Error(), err2.Error())
				}

				fitness1, err1 := offspring1.evaluateFitness(offspringClassMatrix1, schedule, subjects, teachers, constraintMap)
				fitness2, err2 := offspring2.evaluateFitness(offspringClassMatrix2, schedule, subjects, teachers, constraintMap)

				if err1 != nil || err2 != nil {
					return offspring, prepared, executed, fmt.Errorf("ERROR: offspring evaluate fitness failed. err1: %s, err2: %s", err1.Error(), err2.Error())
				}

				offspring1.Fitness = fitness1
				offspring2.Fitness = fitness2

				// 交叉后父代和子代的适应度
				// fmt.Printf("individual1.Fitness: %d, individual2.Fitness: %d, offspring1.Fitness: %d, offspring2.Fitness: %d\n", individual1.Fitness, individual2.Fitness, offspring1.Fitness, offspring2.Fitness)

				offspring = append(offspring, offspring1, offspring2)
				executed++

				// 打印交叉明细
				// log.Printf("Crossover %s, %s ----> %s, %s\n", parent1.UniqueId(), parent2.UniqueId(), offspring1.UniqueId(), offspring2.UniqueId())
			}
		}
	}

	return offspring, prepared, executed, nil
}

// 可换算法验证 用于验证染色体上的基因在进行基因互换杂交时是否符合基因的约束条件
func crossoverAndValidate(parent1, parent2 *Individual, crossPoint int, schedule *models.Schedule, grades []*models.Grade, teachers []*models.Teacher, constr1 []*constraints.Class, constr2 []*constraints.Teacher) (*Individual, *Individual, error) {

	// 交叉操作
	offspring1, offspring2, err := crossoverIndividuals(parent1, parent2, crossPoint, schedule, grades, teachers, constr1, constr2)
	if err != nil {
		return nil, nil, err
	}

	// 判断染色体数量是否相同
	if len(offspring1.Chromosomes) != len(offspring2.Chromosomes) {
		return nil, nil, fmt.Errorf("invalid offspring chromosomes length")
	}

	// 判断各个染色体中课班信息是否相同
	for i, chromosome1 := range offspring1.Chromosomes {
		chromosome2 := offspring2.Chromosomes[i]
		if chromosome1.Genes[0].ClassSN != chromosome2.Genes[0].ClassSN {
			return nil, nil, fmt.Errorf("invalid offspring chromosomes length")
		}
	}
	return offspring1, offspring2, nil
}

// 两个个体之间进行交叉操作，生成两个子代个体
// 返回两个子代个体和错误信息（如果有）
func crossoverIndividuals(parent1, parent2 *Individual, crossPoint int, schedule *models.Schedule, grades []*models.Grade, teachers []*models.Teacher, constr1 []*constraints.Class, constr2 []*constraints.Teacher) (*Individual, *Individual, error) {

	// 检查交叉点是否在有效范围内
	if crossPoint <= 0 || crossPoint >= len(parent1.Chromosomes) {
		return nil, nil, fmt.Errorf("invalid crossPoint %d", crossPoint)
	}

	// 深度复制父代个体的染色体序列
	chromosomes1 := make([]*Chromosome, len(parent1.Chromosomes))
	for i, chromosome := range parent1.Chromosomes {
		chromosomes1[i] = chromosome.Copy()
	}
	chromosomes2 := make([]*Chromosome, len(parent2.Chromosomes))
	for i, chromosome := range parent2.Chromosomes {
		chromosomes2[i] = chromosome.Copy()
	}

	// 初始化子代个体的染色体序列
	offspring1 := &Individual{
		Chromosomes: make([]*Chromosome, len(parent1.Chromosomes)),
	}
	offspring2 := &Individual{
		Chromosomes: make([]*Chromosome, len(parent1.Chromosomes)),
	}

	// 为子代个体1复制基因
	for i := 0; i < len(chromosomes1); i++ {
		var source *Chromosome
		if i < crossPoint {
			source = chromosomes1[i]
		} else {
			source = chromosomes2[i]
		}
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
		target := source.Copy()
		offspring2.Chromosomes[i] = target
	}

	// 修复时间段冲突
	count1, err1 := offspring1.resolveConflicts(schedule, teachers, constr1, constr2)
	count2, err2 := offspring2.resolveConflicts(schedule, teachers, constr1, constr2)

	if err1 == nil && err2 == nil {

		// 个体内基因排序
		offspring1.sortChromosomes()
		offspring2.sortChromosomes()

		log.Printf("crossover resolve conflicts success. count1: %d, count2: %d\n", count1, count2)

		// 返回两个子代个体和nil错误
		return offspring1, offspring2, nil
	}

	return nil, nil, fmt.Errorf("ERROR: offspring repair timeSlot conflicts failed. err1: %v, err2: %v", err1, err2)
}
