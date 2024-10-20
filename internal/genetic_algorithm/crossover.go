// crossover.go
package genetic_algorithm

import (
	"course_scheduler/internal/constraints"
	"course_scheduler/internal/models"
	"fmt"
	"log"
	"math/rand"
)

// 交叉操作
// 每个课班是一个染色体
// 交叉在不同个体的，相同课班的染色体之间进行
// 交叉后个体的数量不变
// 参数:
//
//	selected: 选择的个体
//	crossoverRate: 交叉率
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

func Crossover(selected []*Individual, crossoverRate float64, schedule *models.Schedule, teachingTasks []*models.TeachingTask, subjects []*models.Subject, teachers []*models.Teacher, grades []*models.Grade, subjectVenueMap map[string][]int, constraintMap map[string]interface{}) ([]*Individual, int, int, error) {

	offspring := make([]*Individual, 0, len(selected))
	prepared := 0
	executed := 0

	constr1 := constraintMap["Class"].([]*constraints.Class)
	constr2 := constraintMap["Teacher"].([]*constraints.Teacher)

	fmt.Printf("selected count: %d, crossoverRate: %f", len(selected), crossoverRate)

	for i := 0; i < len(selected)-1; i += 2 {
		if rand.Float64() < crossoverRate {

			// 进行交叉和生成子代个体的逻辑
			prepared++
			crossPoint := rand.Intn(len(selected[i].Chromosomes))

			// 复制一份新的个体
			parent1 := selected[i].Copy()
			parent2 := selected[i+1].Copy()

			// 执行交叉操作并进行后续检查
			offspring1, offspring2, err := crossoverAndValidate(parent1, parent2, crossPoint, schedule, grades, teachers, constr1, constr2)

			// 如果交叉操作出现错误, 则撤销当前交叉操作
			if err == nil {

				log.Printf("crossover and validate success. prepared: %d, executed: %d, err: %s", prepared, executed, err)
				// 评估子代个体的适应度并赋值
				offspringClassMatrix1, err1 := offspring1.toClassMatrix(schedule, teachingTasks, subjects, teachers, subjectVenueMap, constraintMap)
				offspringClassMatrix2, err2 := offspring2.toClassMatrix(schedule, teachingTasks, subjects, teachers, subjectVenueMap, constraintMap)
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
				fmt.Printf("crossover parent1.Fitness: %d, parent2.Fitness: %d, offspring1.Fitness: %d, offspring2.Fitness: %d\n", parent1.Fitness, parent2.Fitness, offspring1.Fitness, offspring2.Fitness)

				offspring = append(offspring, offspring1, offspring2)
				executed++

				// 打印交叉明细
				log.Printf("Crossover %s, %s ----> %s, %s\n", parent1.UniqueId, parent2.UniqueId, offspring1.UniqueId, offspring2.UniqueId)

			} else {
				log.Printf("undo the current crossover operation. prepared: %d, executed: %d, err: %s", prepared, executed, err)
				offspring = append(offspring, selected[i], selected[i+1])
			}

		} else {

			// 不进行交叉，直接保留父母个体
			offspring = append(offspring, selected[i], selected[i+1])
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

	// 个体的课时数是否相同
	parent1Count := parent1.GetTimeSlotsCount()
	parent2Count := parent2.GetTimeSlotsCount()
	offspring1Count := offspring1.GetTimeSlotsCount()
	offspring2Count := offspring2.GetTimeSlotsCount()
	if parent1Count != parent2Count || parent1Count != offspring1Count || parent1Count != offspring2Count {
		return nil, nil, fmt.Errorf("invalid time slots count")
	}

	// 判断染色体数量是否相同
	if len(offspring1.Chromosomes) != len(offspring2.Chromosomes) {
		return nil, nil, fmt.Errorf("invalid offspring chromosomes length")
	}

	// 判断基因数量是否相同
	for i, pc1 := range parent1.Chromosomes {
		oc1 := offspring1.Chromosomes[i]
		if len(pc1.Genes) != len(oc1.Genes) {
			return nil, nil, fmt.Errorf("invalid offspring chromosome gene length. pc1 len: %d, oc1 len: %d", len(pc1.Genes), len(oc1.Genes))
		}

		oc2 := offspring2.Chromosomes[i]
		if len(pc1.Genes) != len(oc2.Genes) {
			return nil, nil, fmt.Errorf("invalid offspring chromosome gene length. pc1 len: %d, oc2 len: %d", len(pc1.Genes), len(oc2.Genes))
		}
	}

	for i, pc2 := range parent2.Chromosomes {

		oc1 := offspring1.Chromosomes[i]
		if len(pc2.Genes) != len(oc1.Genes) {
			return nil, nil, fmt.Errorf("invalid offspring chromosome gene length. pc2 len: %d, oc1 len: %d", len(pc2.Genes), len(oc1.Genes))
		}

		oc2 := offspring2.Chromosomes[i]
		if len(pc2.Genes) != len(oc2.Genes) {
			return nil, nil, fmt.Errorf("invalid offspring chromosome gene length. pc2 len: %d, oc2 len: %d", len(pc2.Genes), len(oc2.Genes))
		}
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

	// 提前年级班级信息
	gradeAndClass := parent1.Chromosomes[crossPoint].ExtractGradeAndClass()

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
		gradeAndClass1 := chromosomes1[i].ExtractGradeAndClass()
		if gradeAndClass == gradeAndClass1 {
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
		gradeAndClass2 := chromosomes2[i].ExtractGradeAndClass()
		if gradeAndClass == gradeAndClass2 {
			source = chromosomes2[i]
		} else {
			source = chromosomes1[i]
		}
		target := source.Copy()
		offspring2.Chromosomes[i] = target
	}

	// 修复时间段冲突
	count1, err1 := offspring1.resolveConflicts(schedule, teachers, constr1, constr2)
	if err1 != nil {
		return nil, nil, err1
	}

	count2, err2 := offspring2.resolveConflicts(schedule, teachers, constr1, constr2)
	if err2 != nil {
		return nil, nil, err2
	}

	log.Printf("crossover resolve conflicts success. count1: %d, count2: %d\n", count1, count2)

	// 个体内基因排序
	offspring1.sortChromosomes()
	offspring2.sortChromosomes()

	// 交叉后计算UniqueId
	offspring1.genUniqueId()
	offspring2.genUniqueId()

	fmt.Printf("crossover parent1.UniqueId: %s, parent2.UniqueId: %s, offspring1.UniqueId: %s, offspring2.UniqueId: %s\n", parent1.UniqueId, parent2.UniqueId, offspring1.UniqueId, offspring2.UniqueId)

	// 返回两个子代个体和nil错误
	return offspring1, offspring2, nil
}

// 旧的实现方法备份
// 两个个体之间进行交叉操作，生成两个子代个体
// 返回两个子代个体和错误信息（如果有）
func crossoverIndividualsBAK(parent1, parent2 *Individual, crossPoint int, schedule *models.Schedule, grades []*models.Grade, teachers []*models.Teacher, constr1 []*constraints.Class, constr2 []*constraints.Teacher) (*Individual, *Individual, error) {

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
	if err1 != nil {
		return nil, nil, err1
	}

	count2, err2 := offspring2.resolveConflicts(schedule, teachers, constr1, constr2)
	if err2 != nil {
		return nil, nil, err2
	}

	log.Printf("crossover resolve conflicts success. count1: %d, count2: %d\n", count1, count2)

	// 个体内基因排序
	offspring1.sortChromosomes()
	offspring2.sortChromosomes()

	// 返回两个子代个体和nil错误
	return offspring1, offspring2, nil
}
