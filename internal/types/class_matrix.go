// class_matrix.go
package types

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/utils"
	"errors"
	"fmt"
	"math"
	"strings"
)

// 课班适应性矩阵
type ClassMatrix struct {
	// key: [课班(科目_年级_班级)][教师][教室][时间段1_时间段2], value: Element
	Elements map[string]map[int]map[int]map[string]*Element

	// 已占用元素总分数
	Score int
}

// 新建课班适应性矩阵
func NewClassMatrix() *ClassMatrix {
	return &ClassMatrix{
		Elements: make(map[string]map[int]map[int]map[string]*Element),
	}
}

// 课班适应性矩阵
// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Val
// key: [9][13][9][40]
// 课班适应性矩阵
// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
// key: [9][13][9][40]
func (cm *ClassMatrix) Init(classes []Class, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, teachers []*models.Teacher, subjectVenueMap map[string][]int) error {

	if len(classes) == 0 {
		return errors.New("classes cannot be empty")
	}
	if schedule == nil {
		return errors.New("schedule cannot be nil")
	}
	if len(teachers) == 0 {
		return errors.New("teachers cannot be empty")
	}

	for _, class := range classes {

		subjectID := class.SN.SubjectID
		gradeID := class.SN.GradeID
		classID := class.SN.ClassID

		teacherIDs := models.ClassTeacherIDs(gradeID, classID, subjectID, teachers)
		if len(teacherIDs) == 0 {
			return fmt.Errorf("no teacher available for class subjectID: %d, gradeID: %d, classID: %d", subjectID, gradeID, classID)
		}

		venueIDs := models.ClassVenueIDs(gradeID, classID, subjectID, subjectVenueMap)
		if len(venueIDs) == 0 {
			return fmt.Errorf("no venue available for class subjectID: %d, gradeID: %d, classID: %d", subjectID, gradeID, classID)
		}

		timeSlotStrs, err := ClassTimeSlots(schedule, taskAllocs, gradeID, classID, subjectID, teacherIDs, venueIDs)
		if err != nil {
			return err
		}

		if len(timeSlotStrs) == 0 {
			return fmt.Errorf("no time slot available for class subjectID: %d, gradeID: %d, classID: %d", subjectID, gradeID, classID)
		}

		sn := class.SN.Generate()

		cm.Elements[sn] = make(map[int]map[int]map[string]*Element)
		for _, teacherID := range teacherIDs {
			cm.Elements[sn][teacherID] = make(map[int]map[string]*Element)
			for _, venueID := range venueIDs {
				cm.Elements[sn][teacherID][venueID] = make(map[string]*Element)

				for _, timeSlotStr := range timeSlotStrs {

					timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
					element := NewElement(sn, class.SubjectID, class.GradeID, class.ClassID, teacherID, venueID, timeSlots)
					cm.Elements[sn][teacherID][venueID][timeSlotStr] = element
				}
			}
		}
	}

	return nil
}

// 计算课班适应性矩阵的所有元素, 固定约束条件下的得分
func (cm *ClassMatrix) CalcElementFixedScores(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, rules []*Rule) error {
	return cm.updateElementScores(cm.calcElementFixedScore, schedule, taskAllocs, rules)
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
func (cm *ClassMatrix) Allocate(classes []Class, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, rules []*Rule) (int, error) {

	// 已分配的课时数量
	var numAssignedClasses int

	// 班级课时占用标记表
	classTimeTableMap := make(map[string]*TimeTable)

	// 教师课时占用标记标
	teacherTimeTableMap := make(map[int]*TimeTable)

	for _, class := range classes {

		sn := class.SN.Generate()
		gradeID := class.SN.GradeID
		classID := class.SN.ClassID
		subjectID := class.SN.SubjectID
		numClassesPerWeek := models.GetNumClassesPerWeek(gradeID, classID, subjectID, taskAllocs)
		numConnectedClassesPerWeek := models.GetNumConnectedClassesPerWeek(gradeID, classID, subjectID, taskAllocs)
		teacherIDs := models.GetTeacherIDs(gradeID, classID, subjectID, taskAllocs)

		// 初始化班级课时占用标记表
		key := fmt.Sprintf("%d_%d", gradeID, classID)
		if _, ok := classTimeTableMap[key]; !ok {
			classTimeTableMap[key] = initTimeTable(schedule)
		}

		// 初始化教师课时占用标记表
		for _, id := range teacherIDs {
			if _, ok := teacherTimeTableMap[id]; !ok {
				teacherTimeTableMap[id] = initTimeTable(schedule)
			}
		}

		// 普通课,连堂课分配
		count := numClassesPerWeek - numConnectedClassesPerWeek
		connectedCount := numConnectedClassesPerWeek
		isConnected := false
		for i := 0; i < count; i++ {
			isConnected = connectedCount > 0
			teacherID, venueID, timeSlotStr, score := cm.findBestTimeSlot(sn, isConnected, classTimeTableMap, teacherTimeTableMap)
			if teacherID > 0 && venueID > 0 && len(timeSlotStr) > 0 {

				temp := cm.Elements[sn][teacherID][venueID][timeSlotStr].Val
				temp.Used = 1
				cm.Elements[sn][teacherID][venueID][timeSlotStr].Val = temp

				timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
				for _, timeSlot := range timeSlots {
					classTimeTableMap[key].Used[timeSlot] = true
					teacherTimeTableMap[teacherID].Used[timeSlot] = true
				}
				cm.updateElementDynamicScores(schedule, taskAllocs, rules)
				connectedCount--
				numAssignedClasses++
			} else {

				return numAssignedClasses, fmt.Errorf("class matrix allocate failed, sn: %s, current class hour: %d, subject num classes per week: %d, teacher ID: %d, venue ID: %d, time slot str: %s, score: %d", sn, i+1, numClassesPerWeek, teacherID, venueID, timeSlotStr, score)
			}
		}
	}

	// 对已占用的矩阵元素的求和
	cm.Score = cm.SumUsedElementsScore()
	return numAssignedClasses, nil
}

// 对已占用的矩阵元素的求和
// 只对元素score分数大于math.MinInt32 的元素score求和
func (cm *ClassMatrix) SumUsedElementsScore() int {
	score := 0
	for _, teacherMap := range cm.Elements {
		for _, venueMap := range teacherMap {
			for _, timeSlotMap := range venueMap {
				for _, element := range timeSlotMap {
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

	for sn, teacherMap := range cm.Elements {

		// if sn == "6_1_1" {
		for teacherID, venueMap := range teacherMap {
			for venueID, timeSlotMap := range venueMap {
				for timeSlot, element := range timeSlotMap {

					if element.Val.Used == 1 && (len(element.Val.ScoreInfo.FixedFailed) > 0 || len(element.Val.ScoreInfo.DynamicFailed) > 0) {
						fixedStr := strings.Join(element.Val.ScoreInfo.FixedFailed, ",")
						dynamicStr := strings.Join(element.Val.ScoreInfo.DynamicFailed, ",")
						fmt.Printf("%p sn: %s, teacherID: %d, venueID: %d, timeSlot: %d failed rules: %s, %s, score: %d\n", cm, sn, teacherID, venueID, timeSlot, fixedStr, dynamicStr, element.Val.ScoreInfo.Score)
					}
				}
			}
		}
	}
	// }
}

// 打印cm的所以key，和val的长度
func (cm *ClassMatrix) PrintKeysAndLength() {

	for sn, teacherMap := range cm.Elements {
		fmt.Printf("Key: %s, Length: %d ", sn, len(teacherMap))
		for teacherID, venueMap := range teacherMap {
			for venueID, timeSlotMap := range venueMap {
				fmt.Printf("teacherID: %d, venueID: %d: len(timeSlotMap): %d\n", teacherID, venueID, len(timeSlotMap))
			}
		}
	}
}

// ==================

// 查找当前课程的最佳可用时间段
// 返回值: teacherID, venueID, timeSlot, score
func (cm *ClassMatrix) findBestTimeSlot(sn string, isConnected bool, classTimeTableMap map[string]*TimeTable, teacherTimeTableMap map[int]*TimeTable) (int, int, string, int) {

	maxScore := math.MinInt32
	teacherID, venueID := -1, -1
	timeSlotStr := ""

	SN, _ := ParseSN(sn)
	key := fmt.Sprintf("%d_%d", SN.GradeID, SN.ClassID)

	for tid, venueMap := range cm.Elements[sn] {
		for vid, timeSlotMap := range venueMap {
			for t, element := range timeSlotMap {

				timeSlots := utils.ParseTimeSlotStr(t)
				if (isConnected && len(timeSlots) != 2) || (!isConnected && len(timeSlots) != 1) {
					continue
				}

				for _, timeSlot := range timeSlots {

					if classTimeTableMap[key].Used[timeSlot] || teacherTimeTableMap[tid].Used[timeSlot] {
						continue
					}
				}

				valScore := element.Val.ScoreInfo.Score
				if valScore > maxScore {
					maxScore = valScore
					teacherID = tid
					venueID = vid
					timeSlotStr = t
				}
			}
		}
	}

	return teacherID, venueID, timeSlotStr, maxScore
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

	for sn, teacherMap := range cm.Elements {
		for teacherID, venueMap := range teacherMap {
			for venueID, timeSlotMap := range venueMap {
				for timeSlot, element := range timeSlotMap {
					elementVal := calcFunc(schedule, taskAllocs, *element, rules)
					cm.Elements[sn][teacherID][venueID][timeSlot].Val = elementVal
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
