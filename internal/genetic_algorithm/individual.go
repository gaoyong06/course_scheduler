// individual.go
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

	"github.com/samber/lo"
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
func newIndividual(classMatrix map[string]map[int]map[int]map[int]types.Val) (*Individual, error) {

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

	// 检查个体是否有时间段冲突
	conflictExists, conflictDetails := individual.HasTimeSlotConflicts()
	if conflictExists {
		return nil, fmt.Errorf("individual has time slot conflicts: %v", conflictDetails)
	}

	return individual, nil
}

// Copy 复制一个 Individual 实例
func (i *Individual) Copy() *Individual {

	copiedChromosomes := make([]Chromosome, len(i.Chromosomes))
	for j, chromosome := range i.Chromosomes {
		copiedGenes := make([]Gene, len(chromosome.Genes))
		copy(copiedGenes, chromosome.Genes)
		copiedChromosomes[j] = Chromosome{
			ClassSN: chromosome.ClassSN,
			Genes:   copiedGenes,
		}
	}
	return &Individual{
		Chromosomes: copiedChromosomes,
		Fitness:     i.Fitness,
	}
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
			fitness += score.FinalScore
		}
	}

	// fitness是个非负数
	if fitness < 0 {
		fitness = 0
	}

	// 返回适应度值
	return fitness, nil
}

// 检查是否有时间段冲突
// 检查是否有时间段冲突
// 检查是否有时间段冲突
func (i *Individual) HasTimeSlotConflicts() (bool, []int) {

	// 记录冲突的时间段
	var conflicts []int

	// 创建一个用于记录已使用时间段的 map
	usedTimeSlots := make(map[int]bool)

	// 检查每个基因的时间段是否有冲突
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			if usedTimeSlots[gene.TimeSlot] {
				conflicts = append(conflicts, gene.TimeSlot)
			} else {
				usedTimeSlots[gene.TimeSlot] = true
			}
		}
	}

	// 判断是否有时间段冲突
	if len(conflicts) == 0 {
		return false, nil
	} else {
		return true, conflicts
	}
}

// 修复时间段冲突，并返回是否已修复的标记
func (individual *Individual) RepairTimeSlotConflicts() (int, [][]int, error) {
	// 记录冲突的总数
	var conflictCount int
	// 修复的时间段对[[a,b],[c,d]], a修复为b, c修复为d
	var repairs [][]int

	// 找出已经占用的时间段和未占用的时间段
	usedTimeSlots := make(map[int]bool)
	unusedTimeSlots := lo.Range(constants.NUM_TIMESLOTS)

	// 冲突时间段:冲突次数
	conflictsMap := make(map[int]int)

	// 标记所有已占用的时间段
	for _, chromosome := range individual.Chromosomes {
		for _, gene := range chromosome.Genes {
			if usedTimeSlots[gene.TimeSlot] {
				conflictsMap[gene.TimeSlot]++
				conflictCount++
			} else {
				usedTimeSlots[gene.TimeSlot] = true
			}
		}
	}

	// 有冲突, 开始修复
	if conflictCount > 0 {

		// 生成未占用的时间段
		unusedTimeSlots = lo.Reject(unusedTimeSlots, func(x int, index int) bool {
			return usedTimeSlots[x]
		})

		// fmt.Printf("=== 有冲突, 开始修复 ============\n")
		// fmt.Printf("冲突总数: %d, 冲突时间段与冲突次数 conflictsMap: %#v, 未占用的时间段: unusedTimeSlots: %v\n", conflictCount, conflictsMap, unusedTimeSlots)

		// 遍历冲突 冲突时间段:冲突次数
		for conflictSlot, conflictNum := range conflictsMap {
			for i := 0; i < conflictNum; i++ {

				// 在修复冲突时，可能会存在多个基因占用同一个时间段，导致修复时重复计算冲突数量
				repaired := false
				for _, chromosome := range individual.Chromosomes {
					for j := 0; j < len(chromosome.Genes); j++ {
						if conflictSlot == chromosome.Genes[j].TimeSlot && !repaired {
							// 从未占用的时间段中随机选择一个时间段
							newTimeSlot := lo.Sample(unusedTimeSlots)
							if !usedTimeSlots[newTimeSlot] {
								// 将其中一个基因的时间段调整到新选择的时间段
								chromosome.Genes[j].TimeSlot = newTimeSlot
								usedTimeSlots[newTimeSlot] = true
								unusedTimeSlots = lo.Filter(unusedTimeSlots, func(x int, index int) bool {
									return x != newTimeSlot
								})
								repairs = append(repairs, []int{conflictSlot, newTimeSlot})
								// 修复了一个冲突，冲突次数减一
								conflictsMap[conflictSlot]--
								repaired = true
							}
						}
					}
				}
			}
		}
	}

	// 检查是否所有冲突都已修复
	for conflictSlot, conflictNum := range conflictsMap {
		if conflictNum > 0 {

			fmt.Printf("unusedTimeSlots: %#v\n", unusedTimeSlots)
			return conflictCount, repairs, fmt.Errorf("still have conflicts: timeslot %d has %d conflicts remaining", conflictSlot, conflictNum)
		}
	}

	// 返回冲突总数、修复情况、是否已修复的标记
	return conflictCount, repairs, nil
}

// 打印课程表
func (i *Individual) PrintSchedule() {

	// schedule[周][节次]=科目
	schedule := make(map[int]map[int]string)
	count := 0

	// Fill the schedule map with the class information for each gene
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			count++
			day := gene.TimeSlot / constants.NUM_CLASSES
			period := gene.TimeSlot % constants.NUM_CLASSES
			classSN := gene.ClassSN
			SN, err := types.ParseSN(classSN)
			if err != nil {
				fmt.Println(err)
			}

			subject, _ := models.FindSubjectByID(SN.SubjectID)
			if _, ok := schedule[day]; !ok {
				schedule[day] = make(map[int]string)
			}

			if schedule[day][period] != "" {
				fmt.Printf("CONFLICT! timeSlot: %d,  day: %d, period: %d\n", gene.TimeSlot, day, period)
			}
			schedule[day][period] = fmt.Sprintf("%s(%d)", subject.Name, gene.TimeSlot)
		}
	}

	fmt.Println("========= schedule =======")
	fmt.Printf("%#v\n", schedule)

	// Print the schedule
	fmt.Printf("课程表: 共%d节课\n", count)
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

// =================================

func getWeekdays() []string {
	return []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
}
