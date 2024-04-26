// individual.go
package genetic_algorithm

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// Individual 个体结构体，代表一个完整的课表排课方案
type Individual struct {
	Chromosomes []*Chromosome // 染色体序列
	Fitness     int           // 适应度
}

// 生成个体
// classMatrix 课班适应性矩阵
// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Val
// key: [9][13][9][40],
func newIndividual(classMatrix *types.ClassMatrix, classHours map[int]int) (*Individual, error) {

	// fmt.Println("================ classMatrix =====================")
	// printClassMatrix(classMatrix)

	// 所有课班选择点位完毕后即可得到一个随机课表，作为种群一个个体
	individual := &Individual{
		// 种群的个体中每个课班选择作为一个染色体
		Chromosomes: []*Chromosome{},
	}

	totalGenes := 0
	for sn, classMap := range classMatrix.Elements {

		// 种群的个体中每个课班选择作为一个染色体
		chromosome := Chromosome{
			ClassSN: sn,
			// 将每个课班的时间、教室、老师作为染色体上的基因
			Genes: []Gene{},
		}

		numGenesInChromosome := 0

		for teacherID, teacherMap := range classMap {
			for venueID, venueMap := range teacherMap {
				for timeSlot, element := range venueMap {

					if element.Val.Used == 1 {

						// 将每个课班的时间、教室、老师作为染色体上的基因
						gene := Gene{
							ClassSN:           sn,
							TeacherID:         teacherID,
							VenueID:           venueID,
							TimeSlot:          timeSlot,
							PassedConstraints: element.GetPassedConstraints(),
							FailedConstraints: element.GetFailedConstraints(),
						}
						chromosome.Genes = append(chromosome.Genes, gene)
						numGenesInChromosome++
						totalGenes++
					}
				}
			}
		}

		// log.Printf("Chromosome for class %s has %d genes\n", sn, numGenesInChromosome)

		// 种群的个体中每个课班选择作为一个染色体
		individual.Chromosomes = append(individual.Chromosomes, &chromosome)
	}

	log.Printf("Total number of chromosomes: %d\n", len(individual.Chromosomes))
	log.Printf("Total number of genes: %d\n", totalGenes)
	individual.SortChromosomes()

	// 检查个体是否有时间段冲突
	conflictExists, conflictDetails := individual.HasTimeSlotConflicts()
	if conflictExists {
		return nil, fmt.Errorf("individual has time slot conflicts: %v", conflictDetails)
	}

	// 设置适应度
	fitness, err := individual.EvaluateFitness(classMatrix, classHours)
	if err != nil {
		return nil, err
	}
	individual.Fitness = fitness

	return individual, nil
}

// Copy 复制一个 Individual 实例
func (i *Individual) Copy() *Individual {

	copiedChromosomes := make([]*Chromosome, len(i.Chromosomes))
	for j, chromosome := range i.Chromosomes {
		copiedGenes := make([]Gene, len(chromosome.Genes))
		copy(copiedGenes, chromosome.Genes)
		copiedChromosomes[j] = &Chromosome{
			ClassSN: chromosome.ClassSN,
			Genes:   copiedGenes,
		}
	}
	return &Individual{
		Chromosomes: copiedChromosomes,
		Fitness:     i.Fitness,
	}
}

// UniqueId 生成唯一的标识符字符串
func (i *Individual) UniqueId() string {

	// 为了确保生成的标识符是唯一的，我们首先对 Chromosomes 切片进行排序
	sortedChromosomes := make([]*Chromosome, len(i.Chromosomes))

	for i, chromosome := range i.Chromosomes {
		sortedChromosomes[i] = chromosome.Copy()
	}
	// copy(sortedChromosomes, i.Chromosomes)
	sort.Slice(sortedChromosomes, func(i, j int) bool {
		return sortedChromosomes[i].ClassSN < sortedChromosomes[j].ClassSN
	})

	// 将排序后的 Chromosomes 转换为 JSON 字符串
	jsonData, err := json.Marshal(sortedChromosomes)
	if err != nil {

		log.Printf("ERROR: json marshal failed. %s", err.Error())
		return ""

	}

	// Hash the resulting string to generate a fixed-length identifier
	hasher := sha256.New()
	hasher.Write([]byte(jsonData))

	uniqueId := fmt.Sprintf("%x", hasher.Sum(nil))
	lastFour := uniqueId[len(uniqueId)-4:]
	return lastFour
}

// 将个体反向转换为科班适应性矩阵
// 目的是公用课班适应性矩阵的约束计算,以此计算个体的适应度
func (i *Individual) toClassMatrix() *types.ClassMatrix {
	// 汇总课班集合
	classes1 := types.InitClasses()

	// 初始化课班适应性矩阵
	classMatrix := types.NewClassMatrix()
	classMatrix.Init(classes1)

	for _, chromosome := range i.Chromosomes {
		for i, gene := range chromosome.Genes {

			// 根据基因信息更新矩阵内部元素约束,得分,占用状态
			element := classMatrix.Elements[gene.ClassSN][gene.TeacherID][gene.VenueID][gene.TimeSlot]
			element.Val.Used = 1
			fixedRules := constraint.GetFixedRules()
			dynamicRules := constraint.GetDynamicRules()
			classMatrix.UpdateElementScore(element, fixedRules, dynamicRules)

			// 修改基因内的约束状态信息
			chromosome.Genes[i].PassedConstraints = element.GetPassedConstraints()
			chromosome.Genes[i].FailedConstraints = element.GetFailedConstraints()
		}
	}
	score := classMatrix.SumUsedElementsScore()
	classMatrix.Score = score

	return classMatrix
}

// SortChromosomes 对个体中的染色体进行排序
func (i *Individual) SortChromosomes() {
	sort.Slice(i.Chromosomes, func(a, b int) bool {
		return i.Chromosomes[a].ClassSN < i.Chromosomes[b].ClassSN
	})
}

// 评估适应度
func (i *Individual) EvaluateFitness(classMatrix *types.ClassMatrix, classHours map[int]int) (int, error) {

	// 初始化适应度值
	fitness := 0

	// 计算矩阵内已占用元素得分
	fitness = classMatrix.Score

	// 计算科目分散度得分
	subjectDispersionScore, err := i.calcSubjectDispersionScore(true, constants.PERIOD_THRESHOLD)
	if err != nil {
		return fitness, nil
	}

	// 计算教师分散度得分
	teacherDispersionScore := i.calcTeacherDispersionScore()

	// 乘以系数并转为整数,
	subjectDispersionScoreInt := int(math.Round(subjectDispersionScore * 10))
	teacherDispersionScoreInt := int(math.Round(teacherDispersionScore * 10))

	fitness += subjectDispersionScoreInt
	fitness += teacherDispersionScoreInt

	// log.Printf("科目分散度: %.2f, 教师分散度: %.2f\n", subjectDispersionScore, teacherDispersionScore)

	// 返回适应度值
	return fitness, nil
}

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

// 修复时间段冲突总数, 修复是否段明细, 错误信息
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

		// log.Printf("=== 有冲突, 开始修复 ============\n")
		// log.Printf("冲突总数: %d, 冲突时间段与冲突次数 conflictsMap: %#v, 未占用的时间段: unusedTimeSlots: %v\n", conflictCount, conflictsMap, unusedTimeSlots)

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

			log.Printf("unusedTimeSlots: %#v\n", unusedTimeSlots)
			return conflictCount, repairs, fmt.Errorf("still have conflicts: timeslot %d has %d conflicts remaining", conflictSlot, conflictNum)
		}
	}

	// 返回冲突总数、修复情况、是否已修复的标记
	return conflictCount, repairs, nil
}

// 计算各个课程班级的分散度
func (i *Individual) calcSubjectStandardDeviation() (map[string]float64, error) {
	subjectTimeSlots := make(map[string][]int) // 记录每个班级的每个科目的课时数
	subjectCount := make(map[string]int)       // 记录每个班级的每个科目在每个时间段内的课时数

	// 遍历每个基因，统计每个班级的每个科目在每个时间段的排课情况
	for _, chromosome := range i.Chromosomes {
		classSN := chromosome.ClassSN
		for _, gene := range chromosome.Genes {
			subjectTimeSlots[classSN] = append(subjectTimeSlots[classSN], gene.TimeSlot)
		}
		subjectCount[classSN] = len(chromosome.Genes)
	}

	return calcStandardDeviation(subjectTimeSlots, subjectCount)
}

// 计算各个教师的分散度
func (i *Individual) calcTeacherStandardDeviation() (map[string]float64, error) {
	teacherTimeSlots := make(map[string][]int)
	teacherCount := make(map[string]int)

	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			teacherID := cast.ToString(gene.TeacherID)
			teacherTimeSlots[teacherID] = append(teacherTimeSlots[teacherID], gene.TimeSlot)
			teacherCount[teacherID]++
		}
	}

	return calcStandardDeviation(teacherTimeSlots, teacherCount)
}

func calcStandardDeviation(timeSlotsMap map[string][]int, countMap map[string]int) (map[string]float64, error) {
	stdDevMap := make(map[string]float64)

	// Calculate the standard deviation for each subject or teacher
	for key, timeSlots := range timeSlotsMap {
		mean := float64(len(timeSlots)) / float64(constants.NUM_TIMESLOTS)
		variance := 0.0
		for _, timeSlot := range timeSlots {
			variance += math.Pow(float64(timeSlot)-mean, 2)
		}
		stdDev := math.Sqrt(variance / float64(constants.NUM_TIMESLOTS))
		stdDevMap[key] = stdDev
	}

	return stdDevMap, nil
}

// 计算一个个体（全校所有年级所有班级的课程表）的科目分散度
func (i Individual) calcSubjectDispersionScore(punishSamePeriod bool, samePeriodThreshold int) (float64, error) {
	// 调用 calcSubjectStandardDeviation 方法计算每个班级的科目分散度
	classSubjectStdDev, err := i.calcSubjectStandardDeviation()
	if err != nil {
		return 0.0, err
	}

	// 计算所有班级的科目分散度的平均值
	totalStdDev := 0.0
	numClasses := len(classSubjectStdDev)
	for _, stdDev := range classSubjectStdDev {
		totalStdDev += stdDev
	}
	if numClasses > 0 {
		totalStdDev /= float64(numClasses)
	}

	// 统计每节课出现的课程数量
	periodCount := make(map[int]int)
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			period := gene.TimeSlot % constants.NUM_CLASSES
			periodCount[period]++
		}
	}

	// 计算惩罚项
	punishment := 0.0
	if punishSamePeriod {
		// threshold := 2 // 阈值，超过此阈值则惩罚
		for _, count := range periodCount {
			if count > samePeriodThreshold {
				punishment += math.Pow(float64(count-samePeriodThreshold), 2)
			}
		}

		// 将惩罚项缩放到一个合适的数量级
		punishment /= 100.0
	}

	// fmt.Printf("totalStdDev: %0.2f, punishment: %0.2f\n", totalStdDev, punishment)
	// 返回总分散度得分，包括平均分散度和惩罚项
	return totalStdDev - punishment, nil
}

// 计算教师分散度得分
// 通过计算信息熵来计算
func (i *Individual) calcTeacherDispersionScore() float64 {

	teacherDispersion := make(map[int]map[int]bool) // 记录每个教师在每个时间段是否已经排课
	teacherCount := make(map[int]int)               // 记录每个教师的课时数

	// 遍历每个基因，统计每个教师在每个时间段的排课情况
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			teacherID := gene.TeacherID
			teacherCount[teacherID]++
			if teacherDispersion[teacherID] == nil {
				teacherDispersion[teacherID] = make(map[int]bool)
				for i := 0; i < constants.NUM_TIMESLOTS; i++ {
					teacherDispersion[teacherID][i] = false
				}
			}
			teacherDispersion[teacherID][gene.TimeSlot] = true
		}
	}

	totalTeacherCount := 0 // 总课时数
	for _, count := range teacherCount {
		totalTeacherCount += count
	}

	dispersionScore := 0.0
	// 计算每个教师的分散度得分
	for teacher, timeSlots := range teacherDispersion {
		// 计算每个时间段的概率
		timeSlotProb := make(map[int]float64)
		numTimeSlots := float64(len(timeSlots))
		for timeSlot := range timeSlots {
			timeSlotProb[timeSlot] = float64(teacherCount[teacher]) / float64(totalTeacherCount) / numTimeSlots
		}

		// 计算信息熵
		entropy := 0.0
		for _, prob := range timeSlotProb {
			entropy -= prob * math.Log2(prob)
		}

		// 计算分散度得分
		dispersionScore += entropy / math.Log2(numTimeSlots)
	}
	return dispersionScore
}

//

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
				log.Printf("CONFLICT! timeSlot: %d,  day: %d, period: %d\n", gene.TimeSlot, day, period)
			}
			schedule[day][period] = fmt.Sprintf("%s(%d)", subject.Name, gene.TimeSlot)
		}
	}

	log.Println("========= schedule =======")
	// log.Printf("%#v\n", schedule)

	// Print the schedule
	log.Printf("课程表: 共%d节课\n", count)
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

// 打印约束状态信息
func (i *Individual) PrintConstraints() {

	var totalConstraints int
	var totalFailedConstraints int
	var totalPassedConstraints int

	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {

			failedConstraints := gene.FailedConstraints
			passedConstraints := gene.PassedConstraints

			totalConstraints += len(failedConstraints) + len(passedConstraints)
			totalFailedConstraints += len(failedConstraints)
			totalPassedConstraints += len(passedConstraints)

			// failedStr := strings.Join(failedConstraints, ", ")
			// passedStr := strings.Join(passedConstraints, ", ")

			// fmt.Printf("SN: %s, TeacherID: %d, VenueID: %d, TimeSlot: %d, Failed Constraints: %s, Passed Constraints: %s\n",
			// 	gene.ClassSN, gene.TeacherID, gene.VenueID, gene.TimeSlot, failedStr, passedStr)
		}
	}

	fmt.Printf("\nTotal Constraints: %d, Failed Constraints: %d, Passed Constraints: %d\n",
		totalConstraints, totalFailedConstraints, totalPassedConstraints)
}

// =================================

func getWeekdays() []string {
	return []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
}
