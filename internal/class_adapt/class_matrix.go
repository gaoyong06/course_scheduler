// class_matrix.go
package class_adapt

import (
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"math"
	"strings"
)

// 课班适应性矩阵
type ClassMatrix struct {
	// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Element
	Elements map[string]map[int]map[int]map[int]*types.Element
}

// 新建课班适应性矩阵
func NewClassMatrix() *ClassMatrix {
	return &ClassMatrix{
		Elements: make(map[string]map[int]map[int]map[int]*types.Element),
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
		teacherIDs := models.ClassTeacherIDs(class.SN)
		venueIDs := models.ClassVenueIDs(class.SN)
		timeSlots := ClassTimeSlots(teacherIDs, venueIDs)
		sn := class.SN.Generate()

		cm.Elements[sn] = make(map[int]map[int]map[int]*types.Element)
		for j := 0; j < len(teacherIDs); j++ {
			teacherID := teacherIDs[j]
			cm.Elements[sn][teacherID] = make(map[int]map[int]*types.Element)
			for k := 0; k < len(venueIDs); k++ {
				venueID := venueIDs[k]
				cm.Elements[sn][teacherID][venueID] = make(map[int]*types.Element)
				for l := 0; l < len(timeSlots); l++ {
					timeSlot := timeSlots[l]
					element := types.NewElement(sn, class.SubjectID, class.GradeID, class.ClassID, teacherID, venueID, timeSlot)
					cm.Elements[sn][teacherID][venueID][timeSlot] = element
				}
			}
		}
	}
}

// func (cm *ClassMatrix) Init(classes []Class) {

// 	for i := 0; i < len(classes); i++ {
// 		class := classes[i]
// 		teacherIDs := models.ClassTeacherIDs(class.SN)
// 		venueIDs := models.ClassVenueIDs(class.SN)
// 		timeSlots := ClassTimeSlots(teacherIDs, venueIDs)
// 		sn := class.SN.Generate()

// 		cm.Elements[sn] = make(map[int]map[int]map[int]types.Val)
// 		for j := 0; j < len(teacherIDs); j++ {
// 			teacherID := teacherIDs[j]
// 			cm.Elements[sn][teacherID] = make(map[int]map[int]types.Val)
// 			for k := 0; k < len(venueIDs); k++ {
// 				venueID := venueIDs[k]
// 				cm.Elements[sn][teacherID][venueID] = make(map[int]types.Val)
// 				for l := 0; l < len(timeSlots); l++ {
// 					timeSlot := timeSlots[l]
// 					scoreInfo := &types.ScoreInfo{
// 						Score:         0,
// 						FixedScore:    0,
// 						DynamicScore:  0,
// 						FixedPassed:   []string{},
// 						FixedFailed:   []string{},
// 						DynamicPassed: []string{},
// 						DynamicFailed: []string{},
// 					}
// 					cm.Elements[sn][teacherID][venueID][timeSlot] = types.Val{ScoreInfo: scoreInfo, Used: 0}
// 				}
// 			}
// 		}
// 	}
// }

// // GetElement 获取课班适应性矩阵中指定元素
// func (cm *ClassMatrix) GetElement(classSN string, teacherID, venueID, timeSlot int) (*types.Element, bool) {

// 	if val, ok := cm.Elements[classSN][teacherID][venueID][timeSlot]; ok {
// 		SN, err := types.ParseSN(classSN)
// 		if err != nil {
// 			fmt.Println(err)
// 			return nil, false
// 		}

// 		return &types.Element{
// 			ClassSN:   classSN,
// 			SubjectID: SN.SubjectID,
// 			GradeID:   SN.GradeID,
// 			ClassID:   SN.ClassID,
// 			TeacherID: teacherID,
// 			VenueID:   venueID,
// 			TimeSlot:  timeSlot,
// 			Val:       &val,
// 		}, true
// 	}
// 	return nil, false
// }

// // SetElementVal 修改矩阵中某个元素的 Val 值
// func (cm *ClassMatrix) SetElementVal(classSN string, teacherID, venueID, timeSlot int, val types.Val) {
// 	cm.Elements[classSN][teacherID][venueID][timeSlot] = val
// }

// 计算课班适应性矩阵的所有元素, 固定约束条件下的得分
func (cm *ClassMatrix) CalcFixedScores() error {
	return cm.calcScores(cm.calcFixedScore)
}

// 计算元素element固定约束得分和动态约束得分
// func (cm *ClassMatrix) CalcScore(element *types.Element) {

// 	cm.calcFixedScore(element)
// 	cm.calcDynamicScore(element)

// 	// 更新val.ScoreInfo.Score
// 	tempVal := cm.Elements[element.ClassSN][element.TeacherID][element.VenueID][element.TimeSlot]
// 	tempVal.ScoreInfo.Score = tempVal.ScoreInfo.FixedScore + tempVal.ScoreInfo.DynamicScore
// 	cm.Elements[element.ClassSN][element.TeacherID][element.VenueID][element.TimeSlot] = tempVal
// }

func (cm *ClassMatrix) CalcScore(element *types.Element) {

	cm.calcFixedScore(element)
	cm.calcDynamicScore(element)

	// 更新 element.Val.Score
	element.Val.ScoreInfo.Score = element.Val.ScoreInfo.FixedScore + element.Val.ScoreInfo.DynamicScore
}

// 根据班级适应性矩阵分配课时
// 循环迭代各个课班，根据匹配结果值, 为每个课班选择课班适应性矩阵中可用的点位，并记录，下个课班选择点位时会避免冲突(一个点位可以引起多点位冲突)
func (cm *ClassMatrix) Allocate(classSNs []string, classHours map[int]int) (int, error) {

	var numAssignedClasses int

	timeTable := initTimeTable()

	for _, sn := range classSNs {
		SN, err := types.ParseSN(sn)
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

				cm.reCalcDynamicScores()

				// updateTimeTableAndClassMatrix(sn, teacherID, venueID, timeSlot, cm.data, timeTable)
				numAssignedClasses++
			} else {
				return numAssignedClasses, fmt.Errorf("failed sn: %s, classHour: %d,  numClassHours: %d", sn, i+1, numClassHours)
			}
		}
	}

	return numAssignedClasses, nil
}

// 计算课班适应性矩阵中,各个元素的动态约束条件下的得分
func (cm *ClassMatrix) reCalcDynamicScores() error {
	return cm.calcScores(cm.calcDynamicScore)
}

// 计算课班适应性矩阵所有元素的得分
func (cm *ClassMatrix) calcScores(calcFunc func(types.ClassUnit)) error {

	for _, teacherMap := range cm.Elements {
		for _, venueMap := range teacherMap {
			for _, timeSlotMap := range venueMap {
				for _, element := range timeSlotMap {
					calcFunc(element)
					element.Val.ScoreInfo.Score = element.Val.ScoreInfo.DynamicScore + element.Val.ScoreInfo.FixedScore
				}
			}
		}
	}
	return nil
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

// CalcFixed 计算固定约束条件得分
func (cm *ClassMatrix) calcFixedScore(element types.ClassUnit) {

	rules := constraint.GetFixedRules()
	score := 0
	penalty := 0

	classSN := element.GetClassSN()
	teacherID := element.GetTeacherID()
	venueID := element.GetVenueID()
	timeSlot := element.GetTimeSlot()

	// 先清空
	elementVal := cm.Elements[classSN][teacherID][venueID][timeSlot].Val
	elementVal.ScoreInfo.FixedPassed = []string{}
	elementVal.ScoreInfo.FixedFailed = []string{}

	for _, rule := range rules {
		if rule.Type == "fixed" {
			if preCheckPassed, result, err := rule.Fn(cm.Elements, element); preCheckPassed && err == nil {

				if result {
					score += rule.Score * rule.Weight
					elementVal.ScoreInfo.FixedPassed = append(elementVal.ScoreInfo.FixedPassed, rule.Name)
				} else {
					penalty += rule.Penalty * rule.Weight
					elementVal.ScoreInfo.FixedFailed = append(elementVal.ScoreInfo.FixedFailed, rule.Name)
				}

			}
		}
	}

	// 固定约束条件得分
	finalScore := score - penalty
	elementVal.ScoreInfo.FixedScore = finalScore
	cm.Elements[classSN][teacherID][venueID][timeSlot].Val = elementVal
}

// CalcDynamic 计算动态约束条件得分
func (cm *ClassMatrix) calcDynamicScore(element types.ClassUnit) {

	rules := constraint.GetDynamicRules()
	score := 0
	penalty := 0

	classSN := element.GetClassSN()
	teacherID := element.GetTeacherID()
	venueID := element.GetVenueID()
	timeSlot := element.GetTimeSlot()

	// 先清空
	elementVal := cm.Elements[classSN][teacherID][venueID][timeSlot].Val
	elementVal.ScoreInfo.DynamicPassed = []string{}
	elementVal.ScoreInfo.DynamicFailed = []string{}

	for _, rule := range rules {
		if rule.Type == "dynamic" {
			if preCheckPassed, result, err := rule.Fn(cm.Elements, element); preCheckPassed && err == nil {

				if result {
					score += rule.Score * rule.Weight
					elementVal.ScoreInfo.DynamicPassed = append(elementVal.ScoreInfo.DynamicPassed, rule.Name)
				} else {
					penalty += rule.Penalty * rule.Weight
					elementVal.ScoreInfo.DynamicFailed = append(elementVal.ScoreInfo.DynamicFailed, rule.Name)
				}
			}
		}
	}

	// 动态约束条件得分
	oldDynamicScore := elementVal.ScoreInfo.DynamicScore
	finalScore := score - penalty
	if oldDynamicScore != finalScore {
		elementVal.ScoreInfo.DynamicScore = finalScore
		// log.Printf("Updated dynamic score: sn=%s, teacherID=%d, venueID=%d, TimeSlot=%d, oldDynamicScore=%d, currentDynamicScore=%d", element.ClassSN, element.TeacherID, element.VenueID, element.TimeSlot, oldDynamicScore, finalScore)
	}

	cm.Elements[classSN][teacherID][venueID][timeSlot].Val = elementVal
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
