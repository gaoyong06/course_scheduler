// class_matrix.go
package types

import (
	"course_scheduler/internal/models"
	"fmt"
	"log"
	"math"
	"strings"
)

// 课班适应性矩阵
type ClassMatrix struct {
	// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
	Elements map[string]map[int]map[int]map[int]*Element

	// 已占用元素总分数
	Score int
}

// 新建课班适应性矩阵
func NewClassMatrix() *ClassMatrix {
	return &ClassMatrix{
		Elements: make(map[string]map[int]map[int]map[int]*Element),
	}
}

// 课班适应性矩阵
// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Val
// key: [9][13][9][40]
// 课班适应性矩阵
// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
// key: [9][13][9][40]
func (cm *ClassMatrix) Init(classes []Class) {

	for i := 0; i < len(classes); i++ {
		class := classes[i]

		subjectID := class.SN.SubjectID
		classID := class.SN.ClassID

		teacherIDs := models.ClassTeacherIDs(subjectID)
		venueIDs := models.ClassVenueIDs(classID)
		timeSlots := ClassTimeSlots(teacherIDs, venueIDs)
		sn := class.SN.Generate()

		cm.Elements[sn] = make(map[int]map[int]map[int]*Element)
		for j := 0; j < len(teacherIDs); j++ {
			teacherID := teacherIDs[j]
			cm.Elements[sn][teacherID] = make(map[int]map[int]*Element)
			for k := 0; k < len(venueIDs); k++ {
				venueID := venueIDs[k]
				cm.Elements[sn][teacherID][venueID] = make(map[int]*Element)
				for l := 0; l < len(timeSlots); l++ {
					timeSlot := timeSlots[l]
					element := NewElement(sn, class.SubjectID, class.GradeID, class.ClassID, teacherID, venueID, timeSlot)
					cm.Elements[sn][teacherID][venueID][timeSlot] = element
				}
			}
		}
	}
}

// 计算课班适应性矩阵的所有元素, 固定约束条件下的得分
func (cm *ClassMatrix) CalcElementFixedScores(rules []*Rule) error {
	return cm.updateElementScores(cm.calcElementFixedScore, rules)
}

// 计算一个元素的得分(含固定约束和动态约束)
func (cm *ClassMatrix) UpdateElementScore(element *Element, fixedRules, dynamicRules []*Rule) {

	fixedVal := cm.calcElementFixedScore(element, fixedRules)
	dynamicVal := cm.calcElementDynamicScore(element, dynamicRules)

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
func (cm *ClassMatrix) Allocate(classSNs []string, classHours map[int]int, rules []*Rule) (int, error) {

	var numAssignedClasses int

	timeTable := initTimeTable()

	for _, sn := range classSNs {
		SN, err := ParseSN(sn)
		if err != nil {
			return numAssignedClasses, err
		}

		subjectID := SN.SubjectID
		numClassHours := classHours[subjectID]

		for i := 0; i < numClassHours; i++ {
			teacherID, venueID, timeSlot, _, err := cm.findBestTimeSlot(sn, timeTable)
			if err != nil {
				return numAssignedClasses, err
			}

			if teacherID > 0 && venueID > 0 && timeSlot >= 0 {

				temp := cm.Elements[sn][teacherID][venueID][timeSlot].Val
				temp.Used = 1
				cm.Elements[sn][teacherID][venueID][timeSlot].Val = temp

				timeTable.Used[timeSlot] = true

				cm.updateElementDynamicScores(rules)

				// updateTimeTableAndClassMatrix(sn, teacherID, venueID, timeSlot, cm.data, timeTable)
				numAssignedClasses++
			} else {
				return numAssignedClasses, fmt.Errorf("failed sn: %s, classHour: %d,  numClassHours: %d", sn, i+1, numClassHours)
			}
		}
	}

	// 对已占用的矩阵元素的求和
	cm.Score = cm.SumUsedElementsScore()
	return numAssignedClasses, nil
}

// 对已占用的矩阵元素的求和
// 只对元素score分数大于0的元素score求和
func (cm *ClassMatrix) SumUsedElementsScore() int {
	score := 0
	for sn, teacherMap := range cm.Elements {
		for teacherID, venueMap := range teacherMap {
			for venueID, timeSlotMap := range venueMap {
				for timeSlot, element := range timeSlotMap {
					if element.Val.Used == 1 && element.Val.ScoreInfo.Score >= 0 {
						elementScore := element.Val.ScoreInfo.Score
						score += elementScore
						log.Printf("Used element: SN=%s, TeacherID=%d, VenueID=%d, TimeSlot=%d, Score=%d\n",
							sn, teacherID, venueID, timeSlot, elementScore)
					}
				}
			}
		}
	}
	log.Printf("ClassMatrix: %p, Sum of used elements score: %d\n", cm, score)
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

// 查找当前课程的最佳可用时间段
func (cm *ClassMatrix) findBestTimeSlot(sn string, timeTable *TimeTable) (int, int, int, int, error) {

	maxScore := math.MinInt32
	teacherID, venueID, timeSlot := -1, -1, -1

	for tid, venueMap := range cm.Elements[sn] {
		for vid, timeSlotMap := range venueMap {
			for t, element := range timeSlotMap {
				if timeTable.Used[t] {
					continue
				}
				valScore := element.Val.ScoreInfo.Score
				if valScore > maxScore {
					maxScore = valScore
					teacherID = tid
					venueID = vid
					timeSlot = t
				}
			}
		}
	}

	return teacherID, venueID, timeSlot, maxScore, nil
}

// 计算固定约束条件得分
func (cm *ClassMatrix) calcElementFixedScore(element ClassUnit, rules []*Rule) Val {
	return cm.calcElementScore(element, rules, "fixed")
}

// 计算动态约束条件得分
func (cm *ClassMatrix) calcElementDynamicScore(element ClassUnit, rules []*Rule) Val {
	return cm.calcElementScore(element, rules, "dynamic")
}

// 计算元素的约束条件得分
func (cm *ClassMatrix) calcElementScore(element ClassUnit, rules []*Rule, scoreType string) Val {
	score := 0
	penalty := 0

	classSN := element.GetClassSN()
	teacherID := element.GetTeacherID()
	venueID := element.GetVenueID()
	timeSlot := element.GetTimeSlot()

	// 先清空
	elementVal := cm.Elements[classSN][teacherID][venueID][timeSlot].Val
	if scoreType == "fixed" {
		elementVal.ScoreInfo.FixedPassed = []string{}
		elementVal.ScoreInfo.FixedFailed = []string{}
	} else {
		elementVal.ScoreInfo.DynamicPassed = []string{}
		elementVal.ScoreInfo.DynamicFailed = []string{}
	}

	for _, rule := range rules {
		if rule.Type == scoreType {
			if preCheckPassed, result, err := rule.Fn(cm, element); preCheckPassed && err == nil {
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
func (cm *ClassMatrix) updateElementDynamicScores(rules []*Rule) error {
	return cm.updateElementScores(cm.calcElementDynamicScore, rules)
}

// 更新课班适应性矩阵所有元素的得分
func (cm *ClassMatrix) updateElementScores(calcFunc func(element ClassUnit, rules []*Rule) Val, rules []*Rule) error {

	for sn, teacherMap := range cm.Elements {
		for teacherID, venueMap := range teacherMap {
			for venueID, timeSlotMap := range venueMap {
				for timeSlot, element := range timeSlotMap {
					elementVal := calcFunc(element, rules)
					cm.Elements[sn][teacherID][venueID][timeSlot].Val = elementVal
					element.Val.ScoreInfo.Score = element.Val.ScoreInfo.DynamicScore + element.Val.ScoreInfo.FixedScore
				}
			}
		}
	}
	return nil
}
