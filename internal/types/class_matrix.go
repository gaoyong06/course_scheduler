// class_matrix.go
package types

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/utils"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/samber/lo"
)

// 课班适应性矩阵
type ClassMatrix struct {

	// 课表方案
	Schedule *models.Schedule

	// 教学计划
	TaskAllocs []*models.TeachTaskAllocation

	// 汇总课班集合
	// 课班(科目+班级), 根据taskAllocs生成
	SubjectClasses []SubjectClass

	// 科目
	Subjects []*models.Subject

	// 教师
	Teachers []*models.Teacher

	// 教学场地
	SubjectVenueMap map[string][]int

	// key: [课班(科目_年级_班级)][教师][教室][时间段1_时间段2], value: Element
	Elements map[string]map[int]map[int]map[string]*Element

	// 已占用元素总分数
	Score int
}

// 新建课班适应性矩阵
func NewClassMatrix(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, subjects []*models.Subject, teachers []*models.Teacher, subjectVenueMap map[string][]int) (*ClassMatrix, error) {

	subjectClasses, err := InitSubjectClasses(taskAllocs, subjects)
	if err != nil {
		return nil, err
	}

	return &ClassMatrix{

		Schedule:        schedule,
		TaskAllocs:      taskAllocs,
		SubjectClasses:  subjectClasses,
		Subjects:        subjects,
		Teachers:        teachers,
		SubjectVenueMap: subjectVenueMap,
		Elements:        make(map[string]map[int]map[int]map[string]*Element),
	}, nil
}

// 课班适应性矩阵
// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
// key: [9][13][9][40]
func (cm *ClassMatrix) Init() error {

	for _, subjectClass := range cm.SubjectClasses {

		subjectID := subjectClass.SN.SubjectID
		gradeID := subjectClass.SN.GradeID
		classID := subjectClass.SN.ClassID

		teacherIDs := models.ClassTeacherIDs(gradeID, classID, subjectID, cm.Teachers)
		if len(teacherIDs) == 0 {
			return fmt.Errorf("no teacher available for class subjectID: %d, gradeID: %d, classID: %d", subjectID, gradeID, classID)
		}

		venueIDs := models.ClassVenueIDs(gradeID, classID, subjectID, cm.SubjectVenueMap)
		if len(venueIDs) == 0 {
			return fmt.Errorf("no venue available for class subjectID: %d, gradeID: %d, classID: %d", subjectID, gradeID, classID)
		}

		connectedTimeSlots := getConnectedTimeSlots(cm.Schedule, cm.TaskAllocs, gradeID, classID, subjectID, teacherIDs, venueIDs)
		normalTimeSlots := getNormalTimeSlots(cm.Schedule, cm.TaskAllocs, gradeID, classID, subjectID, teacherIDs, venueIDs)
		// log.Printf("gradeID: %d, classID: %d, subjectID: %d, connectedTimeSlots: %v, normalTimeSlots: %v\n", gradeID, classID, subjectID, connectedTimeSlots, normalTimeSlots)

		if len(connectedTimeSlots) == 0 && len(normalTimeSlots) == 0 {
			return fmt.Errorf("no time slot available for class subjectID: %d, gradeID: %d, classID: %d", subjectID, gradeID, classID)
		}

		sn := subjectClass.SN.Generate()

		cm.Elements[sn] = make(map[int]map[int]map[string]*Element)
		for _, teacherID := range teacherIDs {
			cm.Elements[sn][teacherID] = make(map[int]map[string]*Element)
			for _, venueID := range venueIDs {
				cm.Elements[sn][teacherID][venueID] = make(map[string]*Element)

				// 连堂课
				for _, connectedStr := range connectedTimeSlots {

					timeSlots := utils.ParseTimeSlotStr(connectedStr)
					element := NewElement(sn, subjectClass.SubjectID, subjectClass.GradeID, subjectClass.ClassID, teacherID, venueID, timeSlots)
					cm.Elements[sn][teacherID][venueID][connectedStr] = element
				}

				// 普通课
				for _, normalStr := range normalTimeSlots {

					timeSlots := utils.ParseTimeSlotStr(normalStr)
					element := NewElement(sn, subjectClass.SubjectID, subjectClass.GradeID, subjectClass.ClassID, teacherID, venueID, timeSlots)
					cm.Elements[sn][teacherID][venueID][normalStr] = element
				}
			}
		}
	}
	return nil
}

// 计算课班适应性矩阵的所有元素, 固定约束条件下的得分
func (cm *ClassMatrix) CalcElementFixedScores(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, rules []*Rule) error {

	// for i := 0; i < len(rules); i++ {
	// 	fmt.Printf("===== CalcElementFixedScores rule[%d]: %#v\n", i, rules[i])
	// 	spew.Dump(rules[i])
	// }

	return cm.updateElementScores(cm.calcElementFixedScore, schedule, taskAllocs, rules)
}

// 计算课班适应性矩阵的所有元素, 动态约束条件下的得分
func (cm *ClassMatrix) CalcElementDynamicScores(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, rules []*Rule) error {

	return cm.updateElementScores(cm.calcElementDynamicScore, schedule, taskAllocs, rules)
}

// 计算一个元素的得分(含固定约束和动态约束)
func (cm *ClassMatrix) UpdateElementScore(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, element *Element, fixedRules, dynamicRules []*Rule) {

	fixedVal := cm.calcElementFixedScore(schedule, taskAllocs, *element, fixedRules)
	dynamicVal := cm.calcElementDynamicScore(schedule, taskAllocs, *element, dynamicRules)

	// 更新固定约束得分
	element.Val.ScoreInfo.FixedFailed = fixedVal.ScoreInfo.FixedFailed
	element.Val.ScoreInfo.FixedPassed = fixedVal.ScoreInfo.FixedPassed
	element.Val.ScoreInfo.FixedScore = fixedVal.ScoreInfo.FixedScore

	// 更新动态约束得分
	element.Val.ScoreInfo.DynamicFailed = dynamicVal.ScoreInfo.DynamicFailed
	element.Val.ScoreInfo.DynamicPassed = dynamicVal.ScoreInfo.DynamicPassed
	element.Val.ScoreInfo.DynamicScore = dynamicVal.ScoreInfo.DynamicScore

	// 更新 element.Val.Score
	element.Val.ScoreInfo.Score = element.Val.ScoreInfo.FixedScore + element.Val.ScoreInfo.DynamicScore
}

// 根据班级适应性矩阵分配课时
// 循环迭代各个课班，根据匹配结果值, 为每个课班选择课班适应性矩阵中可用的点位，并记录，下个课班选择点位时会避免冲突(一个点位可以引起多点位冲突)
// 排课顺序：
// 1: 代班多的 教师 对应科目的 大课(连堂课)
// 2. 代班少的 教师 对应科目的 大课
// 3. 代表多的 教师 对应科目的 小课(普通课)
// 4. 代表少的 教师 对应科目的 小课
func (cm *ClassMatrix) Allocate(rules []*Rule) (int, error) {

	// 已分配的课时数量
	allocateCount := 0

	// 优先分配连堂课
	for _, sc := range cm.SubjectClasses {

		sn := sc.SN.Generate()
		gradeID := sc.SN.GradeID
		classID := sc.SN.ClassID
		subjectID := sc.SN.SubjectID
		numConnectedClassesPerWeek := models.GetNumConnectedClassesPerWeek(gradeID, classID, subjectID, cm.TaskAllocs)

		// 分配课时
		connectedCount := numConnectedClassesPerWeek
		for i := 0; i < connectedCount; i++ {
			if err := cm.allocateClass(sn, true, rules); err != nil {
				return allocateCount, err
			}
			allocateCount++
		}
	}

	// 最后在分配普通课
	for _, sc := range cm.SubjectClasses {

		sn := sc.SN.Generate()
		gradeID := sc.SN.GradeID
		classID := sc.SN.ClassID
		subjectID := sc.SN.SubjectID
		numClassesPerWeek := models.GetNumClassesPerWeek(gradeID, classID, subjectID, cm.TaskAllocs)
		numConnectedClassesPerWeek := models.GetNumConnectedClassesPerWeek(gradeID, classID, subjectID, cm.TaskAllocs)

		// 分配课时
		normalCount := numClassesPerWeek - numConnectedClassesPerWeek*2

		// 然后在分配普通课
		for i := 0; i < normalCount; i++ {
			if err := cm.allocateClass(sn, false, rules); err != nil {
				return allocateCount, err
			}
			allocateCount++
		}
	}

	// 对已占用的矩阵元素的求和
	cm.Score = cm.SumUsedElementsScore()
	return allocateCount, nil
}

func (cm *ClassMatrix) allocateClass(sn string, isConnected bool, rules []*Rule) error {

	teacherID, venueID, timeSlotStr, score, err := cm.findBestTimeSlot(sn, isConnected)
	if err != nil {
		return err
	}

	temp := cm.Elements[sn][teacherID][venueID][timeSlotStr].Val
	temp.Used = 1
	cm.Elements[sn][teacherID][venueID][timeSlotStr].Val = temp

	// 打印当前选择元素信息
	log.Printf("allocate class, class matrix: %p, sn: %s, isConnected: %v, teacherID: %d, venueID: %d, timeSlotStr: %s, score: %d, ", cm, sn, isConnected, teacherID, venueID, timeSlotStr, score)

	// 动态更新元素分数
	cm.updateElementDynamicScores(cm.Schedule, cm.TaskAllocs, rules)
	return nil
}

// 对已占用的矩阵元素的求和
// 只对元素score分数大于math.MinInt32 的元素score求和
func (cm *ClassMatrix) SumUsedElementsScore() int {
	score := 0

	for _, classMap := range cm.Elements {
		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for _, element := range venueMap {
					if element.Val.Used == 1 {
						elementScore := element.Val.ScoreInfo.Score
						score += elementScore
						// log.Printf("Used element: SN=%s, TeacherID=%d, VenueID=%d, TimeSlot=%d, Score=%d\n",
						// 	sn, teacherID, venueID, timeSlot, elementScore)
					}
				}
			}
		}
	}

	// log.Printf("ClassMatrix: %p, Sum of used elements score: %d\n", cm, score)
	return score
}

// 打印有冲突的元素
// 分配课时Allocate结束后,再打印有冲突的元素查看当前矩阵匹配的冲突情况
// [重要] 再分配前,和分配过程中打印都会与最终的结果不一致
// 因为在分配课时Allocate过程中动态约束条件的计算一直在进行ScoreInfo内部数据在一直发生变化
func (cm *ClassMatrix) PrintConstraintElement() {

	fmt.Println("================== PrintConstraintElement ==================")

	for sn, classMap := range cm.Elements {
		// if sn == "1_9_1" || sn == "1_9_2" || sn == "3_9_1" {
		if strings.Contains(sn, "1_9_") {
			for teacherID, teacherMap := range classMap {
				for venueID, venueMap := range teacherMap {
					for timeSlotStr, element := range venueMap {
						fixedStr := strings.Join(element.Val.ScoreInfo.FixedFailed, ",")
						dynamicStr := strings.Join(element.Val.ScoreInfo.DynamicFailed, ",")
						log.Printf("class matrix sn: %s, teacherID: %d, venueID: %d, timeSlotStr: %s, failed rules: %s, %s, score: %d (fixedScore: %d, dynamicScore: %d))\n", sn, teacherID, venueID, timeSlotStr, fixedStr, dynamicStr, element.Val.ScoreInfo.Score, element.Val.ScoreInfo.FixedScore, element.Val.ScoreInfo.DynamicScore)

					}
				}
			}
		}
	}

}

// 打印cm的所以key，和val的长度
func (cm *ClassMatrix) PrintKeysAndLength() {

	for sn, classMap := range cm.Elements {
		fmt.Printf("Key: %s, Length: %d ", sn, len(classMap))
		for teacherID, teacherMap := range classMap {
			for venueID, venueMap := range teacherMap {
				fmt.Printf("teacherID: %d, venueID: %d: len(venueMap): %d\n", teacherID, venueID, len(venueMap))
			}
		}

	}
}

// ==================

// 查找当前课程的最佳可用时间段
// 返回值: teacherID, venueID, timeSlot, score, error
func (cm *ClassMatrix) findBestTimeSlot(sn string, isConnected bool) (int, int, string, int, error) {

	maxScore := math.MinInt32
	teacherID, venueID := -1, -1
	timeSlotStr := ""

	SN, _ := ParseSN(sn)
	gradeID := SN.GradeID
	classID := SN.ClassID
	subjectID := SN.SubjectID

	// var err error

	for snKey, classMap := range cm.Elements {

		if snKey != sn {
			continue
		}

		for teacherIDKey, teacherMap := range classMap {
			for venueIDKey, venueMap := range teacherMap {
				for timeSlotStrKey, element := range venueMap {
					// 检查当前时间段是否已经被使用
					if element.Val.Used == 1 {
						continue
					}

					// 检查当前元素的时间段,同年级同班级 是否有排课
					// 或者相同教师,在该时间段 是否有排课
					isUsed, err := cm.isTimeSlotsUsed(element.GradeID, element.ClassID, element.TeacherID, element.TimeSlots)
					if err != nil {
						return -1, -1, "", -1, err
					}

					if isUsed {
						continue
					}

					// 连堂课,普通课判断
					timeSlots := utils.ParseTimeSlotStr(timeSlotStrKey)
					if (isConnected && len(timeSlots) != 2) || (!isConnected && len(timeSlots) != 1) {
						continue
					}

					valScore := element.Val.ScoreInfo.Score
					if valScore > maxScore {
						maxScore = valScore
						teacherID = teacherIDKey
						venueID = venueIDKey
						timeSlotStr = timeSlotStrKey
					}
				}
			}
		}

	}

	// 如果没有找到可用的时间段，则返回一个错误信息
	if teacherID == -1 && venueID == -1 && timeSlotStr == "" {
		return -1, -1, "", -1, fmt.Errorf("no available time slots for gradeID %d, classID %d, subjectID %d, isConnected %t", gradeID, classID, subjectID, isConnected)
	}

	return teacherID, venueID, timeSlotStr, maxScore, nil
}

// 辅助函数：检查时间段是否已被使用
func (cm *ClassMatrix) isTimeSlotsUsed(gradeID int, classID int, teacherID int, timeSlots []int) (bool, error) {

	for sn, classMap := range cm.Elements {

		SN, err := ParseSN(sn)
		if err != nil {
			return false, err
		}

		for teacherIDKey, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for _, element := range venueMap {

					if element.Val.Used == 1 && ((SN.GradeID == gradeID && SN.ClassID == classID) || teacherIDKey == teacherID) {

						intersect := lo.Intersect(element.TimeSlots, timeSlots)
						isContain := len(intersect) > 0
						if isContain {
							return true, nil
						}
					}
				}
			}
		}
	}
	return false, nil
}

// 计算固定约束条件得分
func (cm *ClassMatrix) calcElementFixedScore(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, element Element, rules []*Rule) Val {
	return cm.calcElementScore(schedule, taskAllocs, element, rules, "fixed")
}

// 计算动态约束条件得分
func (cm *ClassMatrix) calcElementDynamicScore(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, element Element, rules []*Rule) Val {
	return cm.calcElementScore(schedule, taskAllocs, element, rules, "dynamic")
}

// 计算元素的约束条件得分
func (cm *ClassMatrix) calcElementScore(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, element Element, rules []*Rule, scoreType string) Val {

	score := 0
	penalty := 0

	classSN := element.GetClassSN()
	teacherID := element.GetTeacherID()
	venueID := element.GetVenueID()
	timeSlots := element.GetTimeSlots()
	timeSlotsStr := utils.TimeSlotsToStr(timeSlots)

	// 先清空
	elementVal := cm.Elements[classSN][teacherID][venueID][timeSlotsStr].Val
	if scoreType == "fixed" {
		elementVal.ScoreInfo.FixedPassed = []string{}
		elementVal.ScoreInfo.FixedFailed = []string{}
		elementVal.ScoreInfo.FixedSkipped = []string{}
	} else {
		elementVal.ScoreInfo.DynamicPassed = []string{}
		elementVal.ScoreInfo.DynamicFailed = []string{}
		elementVal.ScoreInfo.DynamicSkipped = []string{}
	}

	for _, rule := range rules {
		if rule.Type == scoreType {
			if preCheckPassed, result, err := rule.Fn(cm, element, schedule, taskAllocs); err == nil {
				if preCheckPassed {
					if result {
						score += rule.Score * rule.Weight
						if scoreType == "fixed" {
							elementVal.ScoreInfo.FixedPassed = append(elementVal.ScoreInfo.FixedPassed, rule.Name)
						} else {
							elementVal.ScoreInfo.DynamicPassed = append(elementVal.ScoreInfo.DynamicPassed, rule.Name)
						}
					} else {
						penalty += rule.Penalty * rule.Weight
						if scoreType == "fixed" {
							elementVal.ScoreInfo.FixedFailed = append(elementVal.ScoreInfo.FixedFailed, rule.Name)
						} else {
							elementVal.ScoreInfo.DynamicFailed = append(elementVal.ScoreInfo.DynamicFailed, rule.Name)
						}
					}
				} else {

					// Append the skipped rule to the Skipped field
					if scoreType == "fixed" {
						elementVal.ScoreInfo.FixedSkipped = append(elementVal.ScoreInfo.FixedSkipped, rule.Name)
					} else {
						elementVal.ScoreInfo.DynamicSkipped = append(elementVal.ScoreInfo.DynamicSkipped, rule.Name)
					}
				}
			}
		}
	}

	// 计算得分
	finalScore := score - penalty
	if scoreType == "fixed" {
		elementVal.ScoreInfo.FixedScore = finalScore
	} else {
		elementVal.ScoreInfo.DynamicScore = finalScore
	}

	return elementVal
}

// 更新课班适应性矩阵中,各个元素的动态约束条件下的得分
func (cm *ClassMatrix) updateElementDynamicScores(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, rules []*Rule) error {

	return cm.updateElementScores(cm.calcElementDynamicScore, schedule, taskAllocs, rules)
}

// 更新课班适应性矩阵所有元素的得分
func (cm *ClassMatrix) updateElementScores(calcFunc func(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, element Element, rules []*Rule) Val, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, rules []*Rule) error {

	for sn, classMap := range cm.Elements {
		for teacherID, teacherMap := range classMap {
			for venueID, venueMap := range teacherMap {
				for timeSlotStr, element := range venueMap {
					elementVal := calcFunc(schedule, taskAllocs, *element, rules)
					cm.Elements[sn][teacherID][venueID][timeSlotStr].Val = elementVal
					element.Val.ScoreInfo.Score = element.Val.ScoreInfo.DynamicScore + element.Val.ScoreInfo.FixedScore
				}
			}
		}
	}

	return nil
}

//============== 下面的待修改=================

// type ClassMatrixItem struct {
// 	ClassSN  string
// 	TimeSlot int
// 	Day      int
// 	Period   int
// 	Score    int
// }

// PrintClassMatrix 以Markdown格式打印classMatrix
// func PrintClassMatrix(classMatrix map[string]map[int]map[int]map[int]types.Val, schedule *models.Schedule) {

// 	fmt.Println("| Time Slot| Day | Period | Score |")
// 	fmt.Println("| --- | --- | --- | --- |")

// 	items := make([]ClassMatrixItem, 0)
// 	totalClassesPerDay := schedule.GetTotalClassesPerDay()

// 	for classSN, teacherMap := range classMatrix {
// 		if classSN == "1_1_1" {
// 			for _, venueMap := range teacherMap {
// 				for _, timeSlotMap := range venueMap {
// 					for timeSlot, val := range timeSlotMap {
// 						day := timeSlot / totalClassesPerDay
// 						period := timeSlot % totalClassesPerDay
// 						item := ClassMatrixItem{
// 							ClassSN:  classSN,
// 							TimeSlot: timeSlot,
// 							Day:      day,
// 							Period:   period,
// 							Score:    val.ScoreInfo.Score,
// 						}
// 						items = append(items, item)
// 					}
// 				}
// 			}
// 		}
// 	}

// 	sort.Slice(items, func(i, j int) bool {
// 		if items[i].Day == items[j].Day {
// 			return items[i].Period < items[j].Period
// 		}
// 		return items[i].Day < items[j].Day
// 	})

// 	for _, item := range items {
// 		fmt.Printf("| %d | %d | %d | %d |\n", item.TimeSlot, item.Day, item.Period, item.Score)
// 	}
// }
