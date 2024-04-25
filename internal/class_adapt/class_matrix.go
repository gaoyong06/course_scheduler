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
	Elements map[string]map[int]map[int]map[int]types.Val
}

// 新建课班适应性矩阵
func NewClassMatrix() *ClassMatrix {
	return &ClassMatrix{
		Elements: make(map[string]map[int]map[int]map[int]types.Val),
	}
}

// 课班适应性矩阵
// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Val
// key: [9][13][9][40]
func (cm *ClassMatrix) Init(classes []Class) {

	for i := 0; i < len(classes); i++ {
		class := classes[i]
		teacherIDs := models.ClassTeacherIDs(class.SN)
		venueIDs := models.ClassVenueIDs(class.SN)
		timeSlots := ClassTimeSlots(teacherIDs, venueIDs)
		sn := class.SN.Generate()

		cm.Elements[sn] = make(map[int]map[int]map[int]types.Val)
		for j := 0; j < len(teacherIDs); j++ {
			teacherID := teacherIDs[j]
			cm.Elements[sn][teacherID] = make(map[int]map[int]types.Val)
			for k := 0; k < len(venueIDs); k++ {
				venueID := venueIDs[k]
				cm.Elements[sn][teacherID][venueID] = make(map[int]types.Val)
				for l := 0; l < len(timeSlots); l++ {
					timeSlot := timeSlots[l]
					scoreInfo := &types.ScoreInfo{
						Score:         0,
						FixedScore:    0,
						DynamicScore:  0,
						FixedPassed:   []string{},
						FixedFailed:   []string{},
						DynamicPassed: []string{},
						DynamicFailed: []string{},
					}
					cm.Elements[sn][teacherID][venueID][timeSlot] = types.Val{ScoreInfo: scoreInfo, Used: 0}
				}
			}
		}
	}
}

// 计算课班适应性矩阵的所有元素, 固定约束条件下的得分
func (cm *ClassMatrix) CalcFixedScores() error {
	return cm.calcScores(cm.calcFixedScore)
}

// 计算元素element固定约束得分和动态约束得分
func (cm *ClassMatrix) CalcScore(element *types.Element) {

	cm.calcFixedScore(element)
	cm.calcDynamicScore(element)

	// 更新val.ScoreInfo.Score
	tempVal := cm.Elements[element.ClassSN][element.TeacherID][element.VenueID][element.TimeSlot]
	tempVal.ScoreInfo.Score = tempVal.ScoreInfo.FixedScore + tempVal.ScoreInfo.DynamicScore
	cm.Elements[element.ClassSN][element.TeacherID][element.VenueID][element.TimeSlot] = tempVal
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

				temp := cm.Elements[sn][teacherID][venueID][timeSlot]
				temp.Used = 1
				cm.Elements[sn][teacherID][venueID][timeSlot] = temp

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

	for sn, teacherMap := range cm.Elements {
		SN, err := types.ParseSN(sn)
		if err != nil {
			return err
		}
		for teacherID, venueMap := range teacherMap {
			for venueID, timeSlotMap := range venueMap {
				for timeSlot, val := range timeSlotMap {
					element := &types.Element{
						ClassSN:   sn,
						SubjectID: SN.SubjectID,
						GradeID:   SN.GradeID,
						ClassID:   SN.ClassID,
						TeacherID: teacherID,
						VenueID:   venueID,
						TimeSlot:  timeSlot,
					}
					calcFunc(element)
					val.ScoreInfo.Score = val.ScoreInfo.DynamicScore + val.ScoreInfo.FixedScore
					cm.Elements[sn][teacherID][venueID][timeSlot] = val
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
			for t, val := range timeSlotMap {
				if timeTable.Used[t] {
					continue
				}
				valScore := val.ScoreInfo.Score
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
	newVal := cm.Elements[classSN][teacherID][venueID][timeSlot]
	newVal.ScoreInfo.FixedPassed = []string{}
	newVal.ScoreInfo.FixedFailed = []string{}

	for _, rule := range rules {
		if rule.Type == "fixed" {
			if preCheckPassed, result, err := rule.Fn(cm.Elements, element); preCheckPassed && err == nil {

				if result {
					score += rule.Score * rule.Weight
					// tempVal.ScoreInfo.Passed = append(tempVal.ScoreInfo.Passed, rule)
					newVal.ScoreInfo.FixedPassed = append(newVal.ScoreInfo.FixedPassed, rule.Name)
				} else {
					penalty += rule.Penalty * rule.Weight
					// tempVal.ScoreInfo.Failed = append(tempVal.ScoreInfo.Failed, rule)
					newVal.ScoreInfo.FixedFailed = append(newVal.ScoreInfo.FixedFailed, rule.Name)
				}

			}
		}
	}

	// 固定约束条件得分
	finalScore := score - penalty
	newVal.ScoreInfo.FixedScore = finalScore
	cm.Elements[classSN][teacherID][venueID][timeSlot] = newVal
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
	newVal := cm.Elements[classSN][teacherID][venueID][timeSlot]
	newVal.ScoreInfo.DynamicPassed = []string{}
	newVal.ScoreInfo.DynamicFailed = []string{}

	for _, rule := range rules {
		if rule.Type == "dynamic" {
			if preCheckPassed, result, err := rule.Fn(cm.Elements, element); preCheckPassed && err == nil {

				if result {
					score += rule.Score * rule.Weight
					newVal.ScoreInfo.DynamicPassed = append(newVal.ScoreInfo.DynamicPassed, rule.Name)
				} else {
					penalty += rule.Penalty * rule.Weight
					newVal.ScoreInfo.DynamicFailed = append(newVal.ScoreInfo.DynamicFailed, rule.Name)
				}
			}
		}
	}

	// 动态约束条件得分
	oldDynamicScore := newVal.ScoreInfo.DynamicScore
	finalScore := score - penalty
	if oldDynamicScore != finalScore {
		newVal.ScoreInfo.DynamicScore = finalScore
		// log.Printf("Updated dynamic score: sn=%s, teacherID=%d, venueID=%d, TimeSlot=%d, oldDynamicScore=%d, currentDynamicScore=%d", element.ClassSN, element.TeacherID, element.VenueID, element.TimeSlot, oldDynamicScore, finalScore)
	}

	cm.Elements[classSN][teacherID][venueID][timeSlot] = newVal
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
				for timeSlot, val := range timeSlotMap {

					if val.Used == 1 && (len(val.ScoreInfo.FixedFailed) > 0 || len(val.ScoreInfo.DynamicFailed) > 0) {
						fixedStr := strings.Join(val.ScoreInfo.FixedFailed, ",")
						dynamicStr := strings.Join(val.ScoreInfo.DynamicFailed, ",")
						fmt.Printf("%p sn: %s, teacherID: %d, venueID: %d, timeSlot: %d failed rules: %s, %s, score: %d\n", cm, sn, teacherID, venueID, timeSlot, fixedStr, dynamicStr, val.ScoreInfo.Score)
					}
				}
			}
		}
	}
	// }
}
