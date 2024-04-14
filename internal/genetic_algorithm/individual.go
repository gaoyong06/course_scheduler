package genetic_algorithm

import (
	"course_scheduler/internal/class_adapt"
	"course_scheduler/internal/constants"
	"course_scheduler/internal/evaluation"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"sort"
	"strings"
)

// Individual 个体结构体，代表一个完整的课表排课方案
type Individual struct {
	Chromosomes []Chromosome // 染色体序列
	Fitness     int          // 适应度
}

// 生成个体
// classMatrix 课班适应性矩阵
// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Val
// key: [9][13][9][40],
func newIndividual(classMatrix map[string]map[int]map[int]map[int]types.Val) *Individual {

	// fmt.Println("================ classMatrix =====================")
	// printClassMatrix(classMatrix)

	// 所有课班选择点位完毕后即可得到一个随机课表，作为种群一个个体
	individual := &Individual{
		// 种群的个体中每个课班选择作为一个染色体
		Chromosomes: []Chromosome{},
	}

	totalGenes := 0
	for sn, classMap := range classMatrix {

		// 种群的个体中每个课班选择作为一个染色体
		chromosome := Chromosome{
			ClassSN: sn,
			// 将每个课班的时间、教室、老师作为染色体上的基因
			Genes: []Gene{},
		}

		numGenesInChromosome := 0

		for teacherID, teacherMap := range classMap {
			for venueID, venueMap := range teacherMap {
				for timeSlot, val := range venueMap {

					if val.Used == 1 {

						// 将每个课班的时间、教室、老师作为染色体上的基因
						gene := Gene{
							ClassSN:   sn,
							TeacherID: teacherID,
							VenueID:   venueID,
							TimeSlot:  timeSlot,
						}
						chromosome.Genes = append(chromosome.Genes, gene)
						numGenesInChromosome++
						totalGenes++
					}
				}
			}
		}

		fmt.Printf("Chromosome for class %s has %d genes\n", sn, numGenesInChromosome)

		// 种群的个体中每个课班选择作为一个染色体
		individual.Chromosomes = append(individual.Chromosomes, chromosome)
	}

	fmt.Printf("Total number of chromosomes: %d\n", len(individual.Chromosomes))
	fmt.Printf("Total number of genes: %d\n", totalGenes)

	individual.SortChromosomes()
	return individual
}

func (i *Individual) toClassMatrix() map[string]map[int]map[int]map[int]types.Val {
	// 汇总课班集合
	classes1 := class_adapt.InitClasses()

	// 初始化课班适应性矩阵
	classMatrix := class_adapt.InitClassMatrix(classes1)
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			// 检查中间键是否存在，如果不存在，则创建它们
			if _, ok := classMatrix[gene.ClassSN]; !ok {
				classMatrix[gene.ClassSN] = make(map[int]map[int]map[int]types.Val)
			}
			if _, ok := classMatrix[gene.ClassSN][gene.TeacherID]; !ok {
				classMatrix[gene.ClassSN][gene.TeacherID] = make(map[int]map[int]types.Val)
			}
			if _, ok := classMatrix[gene.ClassSN][gene.TeacherID][gene.VenueID]; !ok {
				classMatrix[gene.ClassSN][gene.TeacherID][gene.VenueID] = make(map[int]types.Val)
			}

			if _, ok := classMatrix[gene.ClassSN][gene.TeacherID][gene.VenueID][gene.TimeSlot]; !ok {
				classMatrix[gene.ClassSN][gene.TeacherID][gene.VenueID][gene.TimeSlot] = types.Val{Score: 0, Used: 0}
			}
			if val, ok := classMatrix[gene.ClassSN][gene.TeacherID][gene.VenueID][gene.TimeSlot]; ok {
				// 键存在，更新值
				val.Used = 1
				classMatrix[gene.ClassSN][gene.TeacherID][gene.VenueID][gene.TimeSlot] = val
			} else {
				// 键不存在，创建新的值并赋值为 1
				classMatrix[gene.ClassSN][gene.TeacherID][gene.VenueID][gene.TimeSlot] = types.Val{Score: 0, Used: 1}
			}
		}
	}
	return classMatrix
}

// SortChromosomes 对个体中的染色体进行排序
func (i *Individual) SortChromosomes() {
	sort.Slice(i.Chromosomes, func(a, b int) bool {
		return i.Chromosomes[a].ClassSN < i.Chromosomes[b].ClassSN
	})
}

// 评估适应度
func (i *Individual) EvaluateFitness() (int, error) {

	classMatrix := i.toClassMatrix()

	// 初始化适应度值
	fitness := 0

	// Check if the individual is not nil
	if i == nil {
		return fitness, nil
	}

	// 遍历个体的所有基因
	// fmt.Printf("individual.Chromosomes: %d\n", len(i.Chromosomes))
	for _, chromosome := range i.Chromosomes {
		// 遍历每个基因的所有课程
		for _, gene := range chromosome.Genes {
			// 计算该基因对应的课程的适应度值

			score, err := evaluation.CalcScore(classMatrix, gene.ClassSN, gene.TeacherID, gene.VenueID, gene.TimeSlot)
			if err != nil {
				return fitness, err
			}
			fitness += score
		}
	}

	// fitness是个非负数
	if fitness < 0 {
		fitness = 0
	}

	// 返回适应度值
	return fitness, nil
}

// 打印课程表
func (i *Individual) PrintSchedule() {

	// schedule[周][节次]=科目
	schedule := make(map[int]map[int]string)

	// Fill the schedule map with the class information for each gene
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {

			day := gene.TimeSlot / constants.NUM_CLASSES
			class := gene.TimeSlot % constants.NUM_CLASSES
			classSN := gene.ClassSN
			SN, err := types.ParseSN(classSN)
			if err != nil {
				fmt.Println(err)
			}

			subject, _ := models.FindSubjectByID(SN.SubjectID)
			if _, ok := schedule[day]; !ok {
				schedule[day] = make(map[int]string)
			}
			schedule[day][class] = fmt.Sprintf("%s(%d)", subject.Name, gene.TimeSlot)
		}
	}

	// Print the schedule
	fmt.Println("课程表:")
	fmt.Println("   |", strings.Join(getWeekdays(), " | "), "|")
	fmt.Println("---+-------------------------------------------")
	for c := 0; c < constants.NUM_CLASSES; c++ {
		fmt.Printf("%-2d |", c+1)
		for d := 0; d < constants.NUM_DAYS; d++ {
			class, ok := schedule[d][c]
			if !ok {
				class = ""
			}
			fmt.Printf(" %-16s |", class)
		}
		fmt.Println()
		fmt.Println("---+-------------------------------------------")
	}
}

// 验证个体是否满足所有约束条件
// 返回值map[int]int 1 满足约束条件，0 不满足约束条件，-1 表示未知或未检查
func (i *Individual) ValidateConstraints() (int, map[int]int, error) {

	score := 0
	constraintsMet := make(map[int]int)
	classMatrix := i.toClassMatrix()

	// 遍历个体的所有基因
	for _, chromosome := range i.Chromosomes {
		// 遍历每个基因的所有课程
		for _, gene := range chromosome.Genes {
			// 计算该基因对应的课程的适应度值
			score, err := evaluation.CalcScore(classMatrix, gene.ClassSN, gene.TeacherID, gene.VenueID, gene.TimeSlot)
			if err != nil {
				return score, constraintsMet, err
			}

			// 检查每个约束条件是否满足
			// 约束条件1: 一年级(1)班 语文 王老师 第1节 固排
			if gene.ClassSN == "1_1_1" && gene.TeacherID == 1 && gene.TimeSlot == 0 {
				constraintsMet[1] = 1
			}

			// 约束条件2: 三年级(1)班 第7节 禁排 班会
			if gene.ClassSN == "14_3_1" && gene.TimeSlot != 6 {
				constraintsMet[2] = 1
			}

			// 3. 三年级(2)班 第8节 禁排 班会
			if gene.ClassSN == "14_3_2" && gene.TimeSlot != 7 {
				constraintsMet[3] = 1
			}

			// 4. 四年级 第8节 禁排 班会
			if gene.ClassSN == "14_3_2" && gene.TimeSlot != 7 {
				constraintsMet[3] = 1
			}

			// 5. 四年级(1)班 语文 王老师 第1节 禁排
			if gene.ClassSN == "1_4_1" && gene.TeacherID == 1 && gene.TimeSlot != 0 {
				constraintsMet[3] = 1
			}

			// 6. 五年级 数学 李老师 第2节 固排
			if gene.ClassSN == "2_5_0" && gene.TeacherID == 2 && gene.TimeSlot == 1 {
				constraintsMet[3] = 1
			}

			//
			// ... 继续检查其他约束条件
		}
	}

	// 初始化未检查的约束条件为true
	for i := 1; i <= 38; i++ {
		if _, ok := constraintsMet[i]; !ok {
			constraintsMet[i] = -1
		}
	}
	return score, constraintsMet, nil
}

// 打印个体满足的约束条件
func (i *Individual) PrintConstraints() {

	_, constraintsMet, err := i.ValidateConstraints()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("个体满足的约束条件:")
	for constraint, met := range constraintsMet {
		if met == 1 {
			fmt.Printf("约束条件%d: 满足\n", constraint)
		} else if met == 0 {
			fmt.Printf("约束条件%d: 不满足\n", constraint)
		} else if met == -1 {
			fmt.Printf("约束条件%d: 未知\n", constraint)
		}
	}
}

// =================================

func getWeekdays() []string {
	return []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
}
