// individual.go
package genetic_algorithm

import (
	"course_scheduler/config"
	"course_scheduler/internal/constraints"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
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
func newIndividual(classMatrix *types.ClassMatrix, schedule *models.Schedule, subjects []*models.Subject, teachers []*models.Teacher, constraintMap map[string]interface{}) (*Individual, error) {

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
			Genes: []*Gene{},
		}
		numGenesInChromosome := 0
		for teacherID, teacherMap := range classMap {
			for venueID, venueMap := range teacherMap {
				for timeSlotStr, e := range venueMap {
					if e.Val.Used == 1 {
						timeSlots := utils.ParseTimeSlotStr(timeSlotStr)

						// 将每个课班的时间、教室、老师作为染色体上的基因
						gene := &Gene{
							ClassSN:            sn,
							TeacherID:          teacherID,
							VenueID:            venueID,
							TimeSlots:          timeSlots,
							IsConnected:        len(timeSlots) == 2,
							PassedConstraints:  e.GetPassedConstraints(),
							FailedConstraints:  e.GetFailedConstraints(),
							SkippedConstraints: e.GetSkippedConstraints(),
						}
						chromosome.Genes = append(chromosome.Genes, gene)
						numGenesInChromosome++
						totalGenes++
					}
				}
			}
		}

		// 对染色体中的基因按照timeSlots排序,连堂课排在前面,普通课排在后面
		sort.Slice(chromosome.Genes, func(i, j int) bool {
			if chromosome.Genes[i].IsConnected && !chromosome.Genes[j].IsConnected {
				return true
			} else if !chromosome.Genes[i].IsConnected && chromosome.Genes[j].IsConnected {
				return false
			} else {
				ts1 := chromosome.Genes[i].TimeSlots
				ts2 := chromosome.Genes[j].TimeSlots
				return ts1[0] < ts2[0] || (ts1[0] == ts2[0] && ts1[1] < ts2[1])
			}
		})

		// 种群的个体中每个课班选择作为一个染色体
		individual.Chromosomes = append(individual.Chromosomes, &chromosome)
	}

	log.Printf("Total number of chromosomes: %d\n", len(individual.Chromosomes))
	log.Printf("Total number of genes: %d\n", totalGenes)
	individual.sortChromosomes()

	// 检查个体是否有时间段冲突
	conflictExists, conflictDetails := individual.HasTimeSlotConflicts()
	if conflictExists {
		return nil, fmt.Errorf("new individual failed. individual has time slot conflicts: %v", conflictDetails)
	}

	// 设置适应度
	fitness, err := individual.evaluateFitness(classMatrix, schedule, subjects, teachers, constraintMap)
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
		copiedChromosomes[j] = chromosome.Copy()
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

// 将个体反向转换为科班适应性矩阵,计算矩阵中已占用元素的得分,矩阵的总得分
// 目的是公用课班适应性矩阵的约束计算,以此计算个体的适应度
func (i *Individual) toClassMatrix(schedule *models.Schedule, teachAllocs []*models.TeachTaskAllocation, subjects []*models.Subject, teachers []*models.Teacher, subjectVenueMap map[string][]int, constraintMap map[string]interface{}) (*types.ClassMatrix, error) {

	// 初始化课班适应性矩阵
	classMatrix, err := types.NewClassMatrix(schedule, teachAllocs, subjects, teachers, subjectVenueMap)
	if err != nil {
		return nil, err
	}

	err = classMatrix.Init()
	if err != nil {
		return nil, err
	}

	// 先标记占用情况
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {

			// 根据基因信息更新矩阵内部元素约束,得分,占用状态
			timeSlotsStr := utils.TimeSlotsToStr(gene.TimeSlots)
			element := classMatrix.Elements[gene.ClassSN][gene.TeacherID][gene.VenueID][timeSlotsStr]
			element.Val.Used = 1
		}
	}

	// 在计算冲突情况,因为冲突是根据现有标记的已占用情况来计算的, 不然这里会出现计算错误
	for _, chromosome := range i.Chromosomes {
		for i, gene := range chromosome.Genes {

			timeSlotsStr := utils.TimeSlotsToStr(gene.TimeSlots)
			element := classMatrix.Elements[gene.ClassSN][gene.TeacherID][gene.VenueID][timeSlotsStr]
			fixedRules := constraints.GetFixedRules(subjects, teachers, constraintMap)
			dynamicRules := constraints.GetDynamicRules(schedule, constraintMap)
			classMatrix.UpdateElementScore(schedule, teachAllocs, element, fixedRules, dynamicRules)

			// 修改基因内的约束状态信息
			chromosome.Genes[i].PassedConstraints = element.GetPassedConstraints()
			chromosome.Genes[i].FailedConstraints = element.GetFailedConstraints()
			chromosome.Genes[i].SkippedConstraints = element.GetSkippedConstraints()
		}
	}
	score := classMatrix.SumUsedElementsScore()
	classMatrix.Score = score

	return classMatrix, nil
}

// SortChromosomes 对个体中的染色体进行排序
func (i *Individual) sortChromosomes() {
	sort.Slice(i.Chromosomes, func(a, b int) bool {
		return i.Chromosomes[a].ClassSN < i.Chromosomes[b].ClassSN
	})
}

// 评估适应度
// 适应度评估使用三个参数, 课班适应度矩阵的分数(归一化后),教师分散度,科目分散度
// 适应度 = 课班适应性矩阵的总分数 * 100 + 教师分散度*10 + 科目分散度 * 10
// 其中影响比较大的几个参数是:
// 1. 矩阵元素的最大惩罚得分
// 2. 矩阵元素的最大奖励得分
// 3. 上面1,2的分值范围, 不能太大, 例如惩罚得分是math.MinInt32,奖励得分是30, 这会导致归一化的值是1.0, 就让这个课班适应度矩阵的分数在计算个体适应度值时失去了意义
// 4. 现在计算的值是：
// Total score: 33
// Min score: -50, Max score: 13
// Normalized score: 1.317460
// Subject dispersion score: 4.939426
// eacher dispersion score: 1.707025
// Fitness: 198
// 给normalizedScore乘以100,目的是为了提升normalizedScore的重要性
// 给subjectDispersionScore, teacherDispersionScore 乘以10, 目的是把数据归到同一个数量级和提升两者的重要度
func (i *Individual) evaluateFitness(classMatrix *types.ClassMatrix, schedule *models.Schedule, subjects []*models.Subject, teachers []*models.Teacher, constraintMap map[string]interface{}) (int, error) {

	// Calculate the total score of the class matrix
	totalScore := classMatrix.Score

	minScore := constraints.GetElementsMinScore(schedule, subjects, teachers, constraintMap)
	maxScore := constraints.GetElementsMaxScore(schedule, subjects, teachers, constraintMap)

	// log.Printf("Min score: %d, Max score: %d\n", minScore, maxScore)

	// Normalize the total score
	normalizedScore := (float64(totalScore) - float64(minScore)) / (float64(maxScore) - float64(minScore))
	// log.Printf("Normalized score: %f\n", normalizedScore)

	// Calculate the subject dispersion score
	subjectDispersionScore, err := i.calcSubjectDispersionScore(schedule, true, config.SubjectPeriodLimitThreshold)
	if err != nil {
		return 0, err
	}
	// log.Printf("Subject dispersion score: %f\n", subjectDispersionScore)

	// Calculate the teacher dispersion score
	teacherDispersionScore := i.calcTeacherDispersionScore(schedule)
	// log.Printf("Teacher dispersion score: %f\n", teacherDispersionScore)

	// Calculate the fitness by multiplying the normalized score by a weight and adding the dispersion scores
	fitness := int(normalizedScore*100 + float64(subjectDispersionScore)*10 + float64(teacherDispersionScore)*10)
	// log.Printf("Fitness: %d\n", fitness)

	return fitness, nil
}

// 检查是否有时间段冲突
// 时间段冲突是指,同一个时间段有多个排课信息
func (i *Individual) HasTimeSlotConflicts() (bool, []string) {

	// 记录冲突的时间段
	var conflicts []string

	// 创建一个用于记录已使用时间段的 map
	// key: gradeID_classID_timeSlot, val: bool
	usedClassTimeSlots := make(map[string]bool)

	// 创建一个用于记录教师已使用时间段的 map
	// key: teacherID_timeSlot, val: bool
	usedTeacherTimeSlots := make(map[string]bool)

	// 检查每个基因的时间段是否有冲突
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {

			sn := gene.ClassSN
			SN, _ := types.ParseSN(sn)

			// 年级
			gradeID := SN.GradeID
			classID := SN.ClassID

			// 教师
			teacherID := gene.TeacherID

			for _, timeSlot := range gene.TimeSlots {
				// 构造 key
				classKey := fmt.Sprintf("gradeID(%d)_classID(%d)_timeSlot(%d)", gradeID, classID, timeSlot)
				teacherKey := fmt.Sprintf("teacherID(%d)_timeSlot(%d)", teacherID, timeSlot)

				if usedClassTimeSlots[classKey] {
					conflicts = append(conflicts, classKey)
				} else {
					usedClassTimeSlots[classKey] = true
				}

				if usedTeacherTimeSlots[teacherKey] {
					conflicts = append(conflicts, teacherKey)
				} else {
					usedTeacherTimeSlots[teacherKey] = true
				}
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

// 获取个体中的课时数量
func (i *Individual) GetTimeSlotsCount() int {

	count := 0
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			count = count + len(gene.TimeSlots)
		}
	}
	return count
}

// 修复函数
// 获取班级已使用的时间段列表(连堂,普通),或者冲突的(连堂,普通)时间段列表
//
//  1. 获取班级 已经使用的时间段列表(连堂,普通)
//     遍历个体中，统计班级已经使用, 但没有冲突的连堂课 时间段
//     遍历个体中，统计班级已经使用, 但没有冲突的普通课 时间段
//  5. 获取班级 冲突的 时间段列表
//     遍历个体中，统计班级已经使用, 但有冲突的连堂课 时间段
//     遍历个体中，统计班级已经使用, 但有冲突的普通课 时间段
//
// 获取班级时间段
func (i *Individual) getClassTimeSlots(conflict bool) (map[string][]*Gene, map[string][]*Gene) {

	classKeyFunc := func(gene *Gene) string {
		SN, _ := types.ParseSN(gene.ClassSN)
		return fmt.Sprintf("%d_%d", SN.GradeID, SN.ClassID)
	}
	connected, normal := i.getTimeSlots(conflict, classKeyFunc)
	return connected, normal
}

// 获取教师已使用的时间段列表(连堂,普通),或者冲突的(连堂,普通)时间段列表
//
//  2. 获取教师 已经使用的时间段列表(连堂,普通)
//     遍历个体中，统计教师已经使用, 但没有冲突的连堂课 时间段
//     遍历个体中，统计教师已经使用, 但没有冲突的普通课 时间段
//  6. 获取教师 冲突的 时间段列表
//     遍历个体中，统计教师已经使用, 但有冲突的连堂课 时间段
//     遍历个体中，统计教师已经使用, 但有冲突的普通课 时间段
//
// 获取教师时间段
func (i *Individual) getTeacherTimeSlots(conflict bool) (map[string][]*Gene, map[string][]*Gene) {

	teacherKeyFunc := func(gene *Gene) string {
		return cast.ToString(gene.TeacherID)
	}
	connected, normal := i.getTimeSlots(conflict, teacherKeyFunc)
	return connected, normal
}

// 获取key函数
type KeyFunc func(gene *Gene) string

// 获取已经使用,和冲突的时间段
func (i *Individual) getTimeSlots(conflict bool, keyFunc KeyFunc) (map[string][]*Gene, map[string][]*Gene) {

	// 已使用的时间段列表(连堂,普通)
	usageConnected := make(map[string][]*Gene)
	usageNormal := make(map[string][]*Gene)

	// 冲突的(连堂,普通)时间段列表
	conflictConnected := make(map[string][]*Gene)
	conflictNormal := make(map[string][]*Gene)

	// 已使用
	usage := make(map[string]map[int]bool)

	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {

			key := keyFunc(gene)
			if usage[key] == nil {
				usage[key] = make(map[int]bool)
			}

			if gene.IsConnected {
				ts0, ts1 := gene.TimeSlots[0], gene.TimeSlots[1]
				if usage[key][ts0] || usage[key][ts1] {
					conflictConnected[key] = append(conflictConnected[key], gene)
				} else {
					usageConnected[key] = append(usageConnected[key], gene)
				}
				usage[key][ts0], usage[key][ts1] = true, true
			} else {
				ts0 := gene.TimeSlots[0]
				if usage[key][ts0] {
					conflictNormal[key] = append(conflictNormal[key], gene)
				} else {
					usageNormal[key] = append(usageNormal[key], gene)
				}
				usage[key][ts0] = true
			}

		}
	}

	if conflict {
		return conflictConnected, conflictNormal
	} else {
		return usageConnected, usageNormal
	}
}

//  3. 获取班级 可用的时间段(连堂,普通)
//     全部的连堂课时间段 剔除 班级已经使用的连堂课时间段, 剔除 班级已经使用的普通课时间段
//     全部的普通课时间段 剔除 班级已经使用的连堂课时间段, 剔除 班级已经使用的普通课时间段
func (i *Individual) getClassValidTimeSlots(schedule *models.Schedule, constr []*constraints.Class) (map[string][]string, map[string][]string) {

	// 全部时间段
	allConnected := utils.GetAllConnectedTimeSlots(schedule)
	allNormal := utils.GetAllNormalTimeSlots(schedule)

	// 班级已使用的时间段
	connectedGenes, normalGenes := i.getClassTimeSlots(false)

	// 获取已经使用的时间段列表(连堂,普通)
	usageConnected := getTimeSlotsFromGenes(connectedGenes)
	usageNormal := getTimeSlotsFromGenes(normalGenes)

	// 从全部的普通课中,去掉已用的连堂课时间段, 已用的普通课时间段
	normal := getDifference(usageNormal, usageConnected, allNormal)
	// 从全部的连堂课中,去掉已用的普通课时间段, 已用的连堂课时间段
	connected := getDifference(usageConnected, usageNormal, allConnected)

	// 过滤掉班级禁排时间段
	connected = i.filterClassTimeSlots(connected, constr)
	normal = i.filterClassTimeSlots(normal, constr)

	return connected, normal
}

//  4. 获取教师 可用的时间段(连堂,普通)
//     全部的连堂课时间段 剔除 教师已经使用的连堂课时间段
//     全部的普通课时间段 剔除 教师已经使用的连堂课时间段
func (i *Individual) getTeacherValidTimeSlots(schedule *models.Schedule, teachers []*models.Teacher, constr []*constraints.Teacher) (map[string][]string, map[string][]string, error) {

	// 全部时间段
	allConnected := utils.GetAllConnectedTimeSlots(schedule)
	allNormal := utils.GetAllNormalTimeSlots(schedule)

	// 教师已使用的时间段
	connectedGenes, normalGenes := i.getTeacherTimeSlots(false)

	// 获取已经使用的时间段列表(连堂,普通)
	usageConnected := getTimeSlotsFromGenes(connectedGenes)
	usageNormal := getTimeSlotsFromGenes(normalGenes)

	// 从全部的普通课中,去掉已用的连堂课时间段, 已用的普通课时间段
	normal := getDifference(usageNormal, usageConnected, allNormal)
	// 从全部的连堂课中,去掉已用的普通课时间段, 已用的连堂课时间段
	connected := getDifference(usageConnected, usageNormal, allConnected)

	// 过滤掉教师禁排时间段
	connected, err := i.filterTeacherTimeSlots(connected, teachers, constr)
	if err != nil {
		return nil, nil, err
	}

	normal, err = i.filterTeacherTimeSlots(normal, teachers, constr)
	if err != nil {
		return nil, nil, err
	}

	return connected, normal, nil
}

// 从教师可用时间中过滤掉教师的禁排时间
// timeSlotsMap key: teacherID, value: 时间段列表
func (i *Individual) filterTeacherTimeSlots(timeSlotsMap map[string][]string, teachers []*models.Teacher, constr []*constraints.Teacher) (map[string][]string, error) {

	for key, items := range timeSlotsMap {

		teacherID := cast.ToInt(key)
		notTimeSlots, err := constraints.GetTeacherNotTimeSlots(teacherID, teachers, constr)
		if err != nil {
			return nil, err
		}

		// 如果items中有一项,在timeSlots中,则移除该items项
		var newItems []string
		for _, str := range items {

			timeSlots := utils.ParseTimeSlotStr(str)
			intersect := lo.Intersect(notTimeSlots, timeSlots)
			if len(intersect) == 0 {
				newItems = append(newItems, str)
			}
		}
		timeSlotsMap[key] = newItems
	}

	return timeSlotsMap, nil
}

// 从班级可用时间中,过滤掉班级的禁排时间
// timeSlotsMap key: gradeID_classID, value: 时间段列表
func (i *Individual) filterClassTimeSlots(timeSlotsMap map[string][]string, constr []*constraints.Class) map[string][]string {

	for key, items := range timeSlotsMap {

		parts := strings.Split(key, "_")
		gradeID := cast.ToInt(parts[0])
		classID := cast.ToInt(parts[1])

		notTimeSlots := constraints.GetClassNotTimeSlots(gradeID, classID, constr)

		// 如果items中有一项,在timeSlots中,则移除该items项
		var newItems []string
		for _, str := range items {

			timeSlots := utils.ParseTimeSlotStr(str)
			intersect := lo.Intersect(notTimeSlots, timeSlots)
			if len(intersect) == 0 {
				newItems = append(newItems, str)
			}
		}
		timeSlotsMap[key] = newItems
	}

	return timeSlotsMap
}

//  7. 修复班级, 教师 冲突的 时间段列表 (当前基因冲突的时间点，修改为即是,该基因对应班级可用的时间段,又是,该基因对应教师可用的时间段)
//     a. 遍历 班级冲突的时间段列表, 如果是连堂课, 则从班级 可用的连堂时间段取一个
//     b. 判断该时间段, 是否在教师可用的时间段(连堂,普通)内,
//     c. 如果在，则修复, 修复后, 将该时间段, 从班级 可用的时间段(连堂,普通) 中剔除, 从教师 可用的时间段(连堂,普通) 中剔除, 然后继续迭代步骤a
//     d. 如果不在,则从可用的连堂时间段中, 迭代取下一个时间段，重复b,c,d步骤
//     e. 如果最后依然无法修复,则表示无法修复，则退出
//
//     班级 普通课 冲突时间段修复 类似
//     教师 连堂课 冲突时间段修复 类似
//     教师 普通课 冲突时间段修复 类似
func (i *Individual) resolveConflicts(schedule *models.Schedule, teachers []*models.Teacher, constr1 []*constraints.Class, constr2 []*constraints.Teacher) (int, error) {

	fmt.Println("===== 修复前")
	i.PrintTimeSlots(false)

	// 记录冲突的总数
	var count int
	// 错误信息
	var err error

	// 班级可用时间段
	classConnected, classNormal := i.getClassValidTimeSlots(schedule, constr1)
	fmt.Println("==== 班级可用时间段 ===")
	fmt.Printf("classConnected: %v\n", classConnected)
	fmt.Printf("classNormal: %v\n", classNormal)

	// 教师可用时间段
	teacherConnected, teacherNormal, err := i.getTeacherValidTimeSlots(schedule, teachers, constr2)
	if err != nil {
		return 0, err
	}
	// fmt.Println("==== 教师可用时间段 ===")
	// fmt.Printf("teacherConnected: %v\n", teacherConnected)
	// fmt.Printf("teacherNormal: %v\n", teacherNormal)

	// 班级时间段冲突
	classConnectedConflictGenes, classNormalConflictGenes := i.getClassTimeSlots(true)
	classConnectedConflict := getTimeSlotsFromGenes(classConnectedConflictGenes)
	classNormalConflict := getTimeSlotsFromGenes(classNormalConflictGenes)

	fmt.Println("==== 班级时间段冲突 ===")
	fmt.Printf("classConnectedConflictGenes: %v\n", classConnectedConflictGenes)
	fmt.Printf("classNormalConflictGenes: %v\n", classNormalConflictGenes)
	fmt.Printf("classConnectedConflict: %v\n", classConnectedConflict)
	fmt.Printf("classNormalConflict: %v\n", classNormalConflict)

	// 教师时间段冲突
	teacherConnectedConflictGenes, teacherNormalConflictGenes := i.getTeacherTimeSlots(true)
	// teacherConnectedConflict := getTimeSlotsFromGenes(teacherConnectedConflictGenes)
	// teacherNormalConflict := getTimeSlotsFromGenes(teacherNormalConflictGenes)
	// fmt.Println("==== 教师时间段冲突 ===")
	// fmt.Printf("teacherConnectedConflict: %v\n", teacherConnectedConflict)
	// fmt.Printf("teacherNormalConflict: %v\n", teacherNormalConflict)

	// 冲突去重
	// 从教师冲突中去重, 即在班级冲突中存在又在教师冲突中存在的基因
	i.rejectConflictGenes(teacherConnectedConflictGenes, classConnectedConflictGenes)
	i.rejectConflictGenes(teacherNormalConflictGenes, classNormalConflictGenes)

	// fmt.Println("==== 从教师冲突中去重, 即在班级冲突中存在又在教师冲突中存在的基因 ===")
	// fmt.Printf("teacherConnectedConflictGenes: %v\n", teacherConnectedConflictGenes)
	// fmt.Printf("teacherNormalConflictGenes: %v\n", teacherNormalConflictGenes)

	// 班级可用时间段
	classValidTime := make(map[string][]string)
	teacherValidTime := make(map[string][]string)
	for key, list := range classConnected {
		classValidTime[key] = append(classValidTime[key], list...)
	}

	for key, list := range classNormal {
		classValidTime[key] = append(classValidTime[key], list...)
	}

	for key, list := range teacherConnected {
		teacherValidTime[key] = append(teacherValidTime[key], list...)
	}

	for key, list := range teacherNormal {
		teacherValidTime[key] = append(teacherValidTime[key], list...)
	}

	// 修复班级连堂课,普通课冲突
	fmt.Printf("===== %p 开始修复 班级连堂课\n", i)
	count1, err1 := i.resolveClassConflict(classConnectedConflictGenes, classValidTime, teacherValidTime)
	if err1 != nil {
		return 0, fmt.Errorf("resolve class connected conflicts failed. err1: %v", err1)
	}
	fmt.Printf("===== %p 修复 班级连堂课成功: %d\n", i, count1)

	fmt.Printf("===== %p 开始修复 班级普通课\n", i)
	count2, err2 := i.resolveClassConflict(classNormalConflictGenes, classValidTime, teacherValidTime)
	if err2 != nil {
		return 0, fmt.Errorf("resolve class normal conflicts failed. err2: %v", err2)
	}
	fmt.Printf("===== %p 修复 班级普通课成功: %d\n", i, count2)

	// 修复教师连堂课,普通课冲突
	fmt.Printf("===== %p 开始修复 教师连堂课\n", i)
	count3, err3 := i.resolveTeacherConflict(teacherConnectedConflictGenes, classValidTime, teacherValidTime)
	if err3 != nil {
		return 0, fmt.Errorf("resolve teacher connected conflicts failed. err3: %v", err3)
	}
	fmt.Printf("===== %p 修复 教师连堂课成功: %d\n", i, count3)

	fmt.Printf("===== %p 开始修复 教师普通课\n", i)
	count4, err4 := i.resolveTeacherConflict(teacherNormalConflictGenes, classValidTime, teacherValidTime)
	if err4 != nil {
		return 0, fmt.Errorf("resolve teacher normal conflicts failed. err4: %v", err4)
	}
	fmt.Printf("===== %p 修复 教师普通课成功: %d\n", i, count4)

	count = count1 + count2 + count3 + count4

	fmt.Printf("===== Success! 冲突修复成功, 修复冲突数量: %d\n", count)

	// 检查是否还有冲突
	hasConflicts, conflicts := i.HasTimeSlotConflicts()
	if hasConflicts {

		fmt.Printf("===== Fuck! 冲突修复成功, 但是检测到冲突: %s\n", conflicts)
		fmt.Println("===== 修复后")
		i.PrintTimeSlots(false)
		return 0, fmt.Errorf("check individual conflict failed. conflicts detail %s", conflicts)
	}

	return count, nil
}

// resolveClassConflict 用于解决班级的课程表冲突
func (i *Individual) resolveClassConflict(conflictMap map[string][]*Gene, classValidTime map[string][]string, teacherValidTime map[string][]string) (int, error) {

	fmt.Printf("开始执行 resolve class conflict, conflictMap: %v, classValidTime: %v, teacherValidTime: %v\n", conflictMap, classValidTime, teacherValidTime)

	count := 0
	for key, conflictList := range conflictMap {
		for _, gene := range conflictList {
			repaired := false
			teacherIDStr := cast.ToString(gene.TeacherID)
			// 年级可用时间段
			classValidList := classValidTime[key]
			// 教师可用时间段
			teacherValidList := teacherValidTime[teacherIDStr]
			conflictTimeSlots := gene.TimeSlots

			fmt.Printf("准备修复: resolve class conflict, key: %s, conflictTimeSlots: %v, classValidList: %v, teacherValidList: %v\n", key, conflictTimeSlots, classValidList, teacherValidList)

			for _, str := range classValidList {

				// 找到一个班级可用的时间段，并且教师也可用
				ts := utils.ParseTimeSlotStr(str)
				if ((gene.IsConnected && len(ts) == 2) || (!gene.IsConnected && len(ts) == 1)) && lo.Contains(teacherValidList, str) {

					// 更新班级和教师的可用时间段
					newTimeSlots := utils.ParseTimeSlotStr(str)

					gene.TimeSlots = newTimeSlots
					repaired = true
					count++

					fmt.Printf("开始修复 resolveClassConflict %v -> %v\n", conflictTimeSlots, newTimeSlots)

					// 从班级可用时间段中移除
					fmt.Printf("删除 classValidTime key: %s, 删除前: %v", key, classValidTime[key])
					classValidTime[key] = utils.RemoveRelatedItems(classValidTime[key], str)
					fmt.Printf(" 删除后: %v\n", classValidTime[key])

					// 从教师可用时间段中移除
					fmt.Printf("删除 teacherValidTime teacherIDStr: %s, 删除前: %v", key, teacherValidTime[teacherIDStr])
					teacherValidTime[teacherIDStr] = utils.RemoveRelatedItems(teacherValidTime[teacherIDStr], str)
					fmt.Printf(" 删除后: %v\n", teacherValidTime[teacherIDStr])

					// // 如果是连堂课,则将连堂课对应的普通课时间段删掉
					// if len(ts) == 2 {

					// 	tsStr0 := cast.ToString(ts[0])
					// 	tsStr1 := cast.ToString(ts[1])

					// 	// 从班级可用时间段中移除
					// 	fmt.Printf("删除 classValidTime key: %s, 删除前: %v", key, classValidTime[key])
					// 	classValidTime[key] = utils.RemoveStr(classValidTime[key], tsStr0)
					// 	classValidTime[key] = utils.RemoveStr(classValidTime[key], tsStr1)
					// 	fmt.Printf(" 删除后: %v\n", classValidTime[key])

					// 	// 从教师可用时间段中移除
					// 	fmt.Printf("删除 teacherValidTime teacherIDStr: %s, 删除前: %v", key, teacherValidTime[teacherIDStr])
					// 	teacherValidTime[teacherIDStr] = utils.RemoveStr(teacherValidTime[teacherIDStr], tsStr0)
					// 	teacherValidTime[teacherIDStr] = utils.RemoveStr(teacherValidTime[teacherIDStr], tsStr1)
					// 	fmt.Printf(" 删除后: %v\n", teacherValidTime[teacherIDStr])
					// }

					break
				}
			}

			// 如果冲突无法修复
			if !repaired {
				return count, fmt.Errorf("resolve class conflict failed. gene: %#v", gene)
			}
		}
	}

	return count, nil
}

// resolveTeacherConflict 用于解决教师的课程表冲突
func (i *Individual) resolveTeacherConflict(conflictMap map[string][]*Gene, teacherValidTime map[string][]string, classValidTime map[string][]string) (int, error) {

	count := 0
	for key, conflictList := range conflictMap {

		for _, gene := range conflictList {

			SN, _ := types.ParseSN(gene.ClassSN)
			gradeID := SN.GradeID
			classID := SN.ClassID
			classKey := fmt.Sprintf("%d_%d", gradeID, classID)

			repaired := false
			teacherIDStr := cast.ToString(gene.TeacherID)
			teacherValidList := teacherValidTime[key]
			classValidList := classValidTime[classKey]
			for _, str := range teacherValidList {

				// 找到一个可教师用的时间段，并且班级也可用
				if lo.Contains(classValidList, str) {

					// 更新基因的时间段
					gene.TimeSlots = utils.ParseTimeSlotStr(str)
					repaired = true
					count++

					// 从班级可用时间段中移除
					classValidTime[classKey] = utils.RemoveRelatedItems(classValidTime[classKey], str)
					// 从教师可用时间段中移除
					teacherValidTime[teacherIDStr] = utils.RemoveRelatedItems(teacherValidTime[teacherIDStr], str)
					break
				}
			}

			// 如果冲突无法修复
			if !repaired {
				return count, fmt.Errorf("resolve teacher conflict failed. gene: %#v", gene)
			}
		}
	}

	return count, nil
}

// 冲突去重
// 从教师冲突中去重, 即如果既在班级冲突中存在, 又在教师冲突中存在的基因, 则从教师冲突中删除
func (individual *Individual) rejectConflictGenes(teacherConflictGenes map[string][]*Gene, classConflictGenes map[string][]*Gene) {

	for tid, teacherGenes := range teacherConflictGenes {
		for _, teacherGene := range teacherGenes {
			for _, classGenes := range classConflictGenes {
				for _, classGene := range classGenes {
					if teacherGene.Equal(classGene) {
						teacherConflictGenes[tid] = lo.Reject(teacherConflictGenes[tid], func(x *Gene, _ int) bool {
							return x.Equal(teacherGene)
						})
					}
				}
			}
		}
	}
}

// 计算各个课程班级的分散度
func (i *Individual) calcSubjectStandardDeviation(schedule *models.Schedule) (map[string]float64, error) {
	subjectTimeSlots := make(map[string][]int) // 记录每个班级的每个科目的课时数
	subjectCount := make(map[string]int)       // 记录每个班级的每个科目在每个时间段内的课时数

	// 遍历每个基因，统计每个班级的每个科目在每个时间段的排课情况
	for _, chromosome := range i.Chromosomes {
		classSN := chromosome.ClassSN
		for _, gene := range chromosome.Genes {
			subjectTimeSlots[classSN] = append(subjectTimeSlots[classSN], gene.TimeSlots...)
		}
		subjectCount[classSN] = len(chromosome.Genes)
	}

	return calcStandardDeviation(schedule, subjectTimeSlots, subjectCount)
}

// 计算各个教师的分散度
func (i *Individual) calcTeacherStandardDeviation(schedule *models.Schedule) (map[string]float64, error) {
	teacherTimeSlots := make(map[string][]int)
	teacherCount := make(map[string]int)

	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			teacherID := cast.ToString(gene.TeacherID)
			for _, timeSlot := range gene.TimeSlots {
				teacherTimeSlots[teacherID] = append(teacherTimeSlots[teacherID], timeSlot)
				teacherCount[teacherID]++
			}
		}
	}

	return calcStandardDeviation(schedule, teacherTimeSlots, teacherCount)
}

func calcStandardDeviation(schedule *models.Schedule, timeSlotsMap map[string][]int, countMap map[string]int) (map[string]float64, error) {
	stdDevMap := make(map[string]float64)
	totalClassesPerWeek := schedule.TotalClassesPerWeek()

	// Calculate the standard deviation for each subject or teacher
	for key, timeSlots := range timeSlotsMap {

		mean := float64(len(timeSlots)) / float64(totalClassesPerWeek)
		variance := 0.0
		for _, timeSlot := range timeSlots {
			variance += math.Pow(float64(timeSlot)-mean, 2)
		}
		stdDev := math.Sqrt(variance / float64(totalClassesPerWeek))
		stdDevMap[key] = stdDev
	}

	return stdDevMap, nil
}

// 计算一个个体（全校所有年级所有班级的课程表）的科目分散度
func (i Individual) calcSubjectDispersionScore(schedule *models.Schedule, punishSamePeriod bool, samePeriodThreshold int) (float64, error) {
	// 调用 calcSubjectStandardDeviation 方法计算每个班级的科目分散度
	classSubjectStdDev, err := i.calcSubjectStandardDeviation(schedule)
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
	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	periodCount := make(map[int]int)
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			for _, timeSlot := range gene.TimeSlots {
				period := timeSlot % totalClassesPerDay
				periodCount[period]++
			}
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
func (i *Individual) calcTeacherDispersionScore(schedule *models.Schedule) float64 {

	teacherDispersion := make(map[int]map[int]bool) // 记录每个教师在每个时间段是否已经排课
	teacherCount := make(map[int]int)               // 记录每个教师的课时数
	totalClassesPerWeek := schedule.TotalClassesPerWeek()

	// 遍历每个基因，统计每个教师在每个时间段的排课情况
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			teacherID := gene.TeacherID
			teacherCount[teacherID]++
			if teacherDispersion[teacherID] == nil {
				teacherDispersion[teacherID] = make(map[int]bool)
				for i := 0; i < totalClassesPerWeek; i++ {
					teacherDispersion[teacherID][i] = false
				}
			}
			for _, timeSlot := range gene.TimeSlots {
				teacherDispersion[teacherID][timeSlot] = true
			}
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

// 打印课程表
func (i *Individual) PrintSchedule(schedule *models.Schedule, subjects []*models.Subject) {

	// schedule[年级班级][周][节次]=科目
	scheduleMap := make(map[string]map[int]map[int]string) // 使用字符串键来表示年级和班级

	// 课节数
	countMap := make(map[string]int)

	// 一周总课时
	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// 一周工作日
	numWorkdays := schedule.NumWorkdays

	// Fill the schedule map with the class information for each gene
	for _, chromosome := range i.Chromosomes {
		classSN := chromosome.ClassSN
		SN, err := types.ParseSN(classSN)
		if err != nil {
			fmt.Println(err)
			continue
		}

		gradeAndClass := fmt.Sprintf("%d_%d", SN.GradeID, SN.ClassID)

		if _, ok := scheduleMap[gradeAndClass]; !ok {
			scheduleMap[gradeAndClass] = make(map[int]map[int]string)
			countMap[gradeAndClass] = 0
		}

		for _, gene := range chromosome.Genes {
			for _, timeSlot := range gene.TimeSlots {
				countMap[gradeAndClass]++
				day := timeSlot / totalClassesPerDay
				period := timeSlot % totalClassesPerDay

				subject, err := models.FindSubjectByID(SN.SubjectID, subjects)
				if err != nil {
					log.Printf("Error finding subject with ID %d: %v", SN.SubjectID, err)
					continue
				}

				if _, ok := scheduleMap[gradeAndClass][day]; !ok {
					scheduleMap[gradeAndClass][day] = make(map[int]string)
				}

				// 如果不为空, 则说明之前赋值过,这里会出现覆盖,这是因为同一个时间段有多个排课
				if scheduleMap[gradeAndClass][day][period] != "" {
					log.Printf("CONFLICT! timeSlot: %d,  day: %d, period: %d\n", timeSlot, day, period)
				}

				// 连堂课后面多一个+号
				if gene.IsConnected {
					scheduleMap[gradeAndClass][day][period] = fmt.Sprintf("%s(%d)+", subject.Name, timeSlot)
				} else {
					scheduleMap[gradeAndClass][day][period] = fmt.Sprintf("%s(%d)", subject.Name, timeSlot)
				}

			}
		}
	}

	log.Println("========= schedule =======")
	// log.Printf("%#v\n", schedule)

	// 按照年级和班级的组合字符串排序
	var keys []string
	for key := range scheduleMap {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		// 将字符串按照 "_" 分割成两个部分
		partsI := strings.Split(keys[i], "_")
		partsJ := strings.Split(keys[j], "_")

		// 将两个部分都转换成整数
		numI := cast.ToInt(partsI[0])
		numJ := cast.ToInt(partsJ[0])
		numI2 := cast.ToInt(partsI[1])
		numJ2 := cast.ToInt(partsJ[1])

		// 自定义排序逻辑
		return numI*100+numI2 < numJ*100+numJ2
	})

	// 按照年级和班级打印课程表
	for _, gradeAndClass := range keys {

		log.Printf("课程表(%s): 共%d节课\n", gradeAndClass, countMap[gradeAndClass])
		fmt.Println("   |", strings.Join(getWeekdays(), " | "), "|")
		fmt.Println("---+-------------------------------------------")
		for c := 0; c < totalClassesPerDay; c++ {
			fmt.Printf("%-2d |", c+1)
			for d := 0; d < numWorkdays; d++ {
				class, ok := scheduleMap[gradeAndClass][d][c]
				if !ok {
					class = ""
				}
				fmt.Printf(" %-16s |", class)
			}
			fmt.Println()
			fmt.Println("---+-------------------------------------------")
		}
		fmt.Println()
	}
}

// 打印时间段
func (i *Individual) PrintTimeSlots(sorted bool) {

	genes := make([]*Gene, 0)
	for _, chromosome := range i.Chromosomes {
		genes = append(genes, chromosome.Genes...)
	}

	if sorted {
		sort.Slice(genes, func(i, j int) bool {
			return genes[i].TimeSlots[0] < genes[j].TimeSlots[0]
		})
	}

	for _, gene := range genes {
		fmt.Printf("SN: %s\tTeacherID: %d\tVenueID: %d\tTimeSlots: %v\n",
			gene.ClassSN, gene.TeacherID, gene.VenueID, gene.TimeSlots)
	}
}

// 打印约束状态信息
func (i *Individual) PrintConstraints() {

	var totalConstraints int
	var totalFailedConstraints int
	var totalPassedConstraints int
	var totalSkippedConstraints int

	// Merge genes from all chromosomes into a single slice
	genes := make([]*Gene, 0)
	for _, chromosome := range i.Chromosomes {
		genes = append(genes, chromosome.Genes...)
	}

	// Sort genes by TimeSlot in ascending order
	sort.Slice(genes, func(i, j int) bool {
		return genes[i].TimeSlots[0] < genes[j].TimeSlots[0]
	})

	for _, gene := range genes {
		failedConstraints := gene.FailedConstraints
		passedConstraints := gene.PassedConstraints
		skippedConstraints := gene.SkippedConstraints

		totalConstraints += len(failedConstraints) + len(passedConstraints) + len(skippedConstraints)
		totalFailedConstraints += len(failedConstraints)
		totalPassedConstraints += len(passedConstraints)
		totalSkippedConstraints += len(skippedConstraints)

		failedStr := strings.Join(failedConstraints, ", ")
		passedStr := strings.Join(passedConstraints, ", ")

		fmt.Printf("SN: %s\tTeacherID: %d\tVenueID: %d\tTimeSlots: %v\tFailed Constraints: [%s]\tPassed Constraints: [%s]\n",
			gene.ClassSN, gene.TeacherID, gene.VenueID, gene.TimeSlots, failedStr, passedStr)
	}

	fmt.Printf("\nTotal Constraints: %d, Failed Constraints: %d, Passed Constraints: %d, Skipped Constraints: %d\n",
		totalConstraints, totalFailedConstraints, totalPassedConstraints, totalSkippedConstraints)
}

// =================================

func getWeekdays() []string {
	return []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
}

// 取数组的差集
func getDifference(usageMap map[string][]string, usageMap1 map[string][]string, allTimeSlots []string) map[string][]string {

	resultMap := make(map[string][]string)
	for key, items := range usageMap {

		// 从全部的xx课中,去掉已用的xx课时间段(时间段完全相同)
		resultMap[key], _ = lo.Difference(allTimeSlots, items)
		var removeItems []string

		// 剔除已经使用的时间段(包含在连堂课中)
		// 剔除已经使用的时间段(包含在普通课中)
		for _, tsStr1 := range append(usageMap[key], usageMap1[key]...) {
			for _, tsStr := range resultMap[key] {
				ts1 := utils.ParseTimeSlotStr(tsStr1)
				ts := utils.ParseTimeSlotStr(tsStr)
				intersect := lo.Intersect(ts1, ts)

				if len(intersect) > 0 {
					removeItems = append(removeItems, tsStr)
				}
			}
		}

		// 剔除已经使用的普通课时间段
		resultMap[key], _ = lo.Difference(resultMap[key], removeItems)

		// 对结果进行排序
		sort.Slice(resultMap[key], func(i, j int) bool {
			ts1 := utils.ParseTimeSlotStr(resultMap[key][i])
			ts2 := utils.ParseTimeSlotStr(resultMap[key][j])
			return ts1[0] < ts2[0] || (ts1[0] == ts2[0] && ts1[1] < ts2[1])
		})
	}
	return resultMap
}

// 获取geneMap中的时间段列表(连堂,普通)
func getTimeSlotsFromGenes(genesMap map[string][]*Gene) map[string][]string {

	timeSlotMap := make(map[string][]string)
	for key, genes := range genesMap {
		for _, gene := range genes {
			tsStr := utils.TimeSlotsToStr(gene.TimeSlots)
			timeSlotMap[key] = append(timeSlotMap[key], tsStr)
		}

		// 对结果进行排序
		sort.Slice(timeSlotMap[key], func(i, j int) bool {
			ts1 := utils.ParseTimeSlotStr(timeSlotMap[key][i])
			ts2 := utils.ParseTimeSlotStr(timeSlotMap[key][j])
			return ts1[0] < ts2[0] || (ts1[0] == ts2[0] && ts1[1] < ts2[1])
		})
	}
	return timeSlotMap
}
