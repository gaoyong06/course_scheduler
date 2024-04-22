// class_matrix.go
package class_adapt

import (
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"math"
)

// 课班适应性矩阵
// key: [课班(科目_年级_班级)][教师][教室][时间段], value: Val
// key: [9][13][9][40],
func InitClassMatrix(classes []Class) map[string]map[int]map[int]map[int]types.Val {

	classMatrix := make(map[string]map[int]map[int]map[int]types.Val)

	for i := 0; i < len(classes); i++ {
		class := classes[i]
		teacherIDs := models.ClassTeacherIDs(class.SN)
		venueIDs := models.ClassVenueIDs(class.SN)
		timeSlots := ClassTimeSlots(teacherIDs, venueIDs)
		sn := class.SN.Generate()

		// log.Printf("initClassMatrix sn: %s, len(teacherIDs): %d, len(venueIDs): %d, len(timeSlots): %d\n", sn, len(teacherIDs), len(venueIDs), len(timeSlots))

		classMatrix[sn] = make(map[int]map[int]map[int]types.Val)
		for j := 0; j < len(teacherIDs); j++ {
			teacherID := teacherIDs[j]
			classMatrix[sn][teacherID] = make(map[int]map[int]types.Val)
			for k := 0; k < len(venueIDs); k++ {
				venueID := venueIDs[k]
				classMatrix[sn][teacherID][venueID] = make(map[int]types.Val)
				for l := 0; l < len(timeSlots); l++ {
					timeSlot := timeSlots[l]
					scoreInfo := &types.ScoreInfo{Score: 0}
					classMatrix[sn][teacherID][venueID][timeSlot] = types.Val{ScoreInfo: scoreInfo, Used: 0}
				}
			}
		}
	}

	return classMatrix
}

// 课班适应性矩阵各元素, 计算固定约束条件下的得分
func CalcFixedScores(classMatrix map[string]map[int]map[int]map[int]types.Val) error {
	return calcScores(classMatrix, constraint.CalcFixed, false)
}

// 根据班级适应性矩阵分配课时
// 循环迭代各个课班，根据匹配结果值, 为每个课班选择课班适应性矩阵中可用的点位，并记录，下个课班选择点位时会避免冲突(一个点位可以引起多点位冲突)
func AllocateClassMatrix(classSNs []string, classHours map[int]int, classMatrix map[string]map[int]map[int]map[int]types.Val) (int, error) {

	timeTable := initTimeTable()
	var numAssignedClasses int

	for _, sn := range classSNs {

		SN, err := types.ParseSN(sn)
		if err != nil {
			return numAssignedClasses, err
		}

		subjectID := SN.SubjectID
		numClassHours := classHours[subjectID]

		// 循环科目周课时
		for i := 0; i < numClassHours; i++ {

			// 查找当前课程的最佳可用时间段
			teacherID, venueID, timeSlot, maxScore, err := findBestTimeSlot(sn, classMatrix, timeTable)
			if err != nil {
				return numAssignedClasses, err
			}

			fmt.Printf("findBestTimeSlot teacherID: %d, venueID: %d, timeSlot: %d, maxScore: %d\n", teacherID, venueID, timeSlot, maxScore)

			// 更新时间表和课班适应性矩阵
			if teacherID > 0 && venueID > 0 && timeSlot >= 0 {

				// 测试
				if maxScore < 0 {

					val := classMatrix[sn][teacherID][venueID][timeSlot]
					fmt.Printf("timeSlot: %d, sn: %s, failed rules: ", timeSlot, sn)
					for _, rule := range val.ScoreInfo.Failed {
						fmt.Printf("%s, ", rule.Name)
					}
					fmt.Println()
				}

				updateTimeTableAndClassMatrix(sn, teacherID, venueID, timeSlot, classMatrix, timeTable)
				numAssignedClasses++

			} else {
				return numAssignedClasses, fmt.Errorf("failed sn: %s, classHour: %d,  numClassHours: %d", sn, i+1, numClassHours)
			}
		}
	}

	return numAssignedClasses, nil
}

// ==================================================================

// 查找当前课程的最佳可用时间段
func findBestTimeSlot(sn string, classMatrix map[string]map[int]map[int]map[int]types.Val, timeTable *TimeTable) (int, int, int, int, error) {

	maxScore := math.MinInt32
	teacherID, venueID, timeSlot := -1, -1, -1

	for tid, venueMap := range classMatrix[sn] {
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

// CalcScores 计算固定得分和动态得分
func calcScores(classMatrix map[string]map[int]map[int]map[int]types.Val, calcFunc func(map[string]map[int]map[int]map[int]types.Val, *types.Element) (int, error), isAddScore bool) error {

	for sn, teacherMap := range classMatrix {
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

					score, err := calcFunc(classMatrix, element)
					if err != nil {
						return err
					}
					if isAddScore {
						val.ScoreInfo.Score = val.ScoreInfo.Score + score
					} else {
						val.ScoreInfo.Score = score
					}
					classMatrix[sn][teacherID][venueID][timeSlot] = val
				}
			}
		}
	}
	return nil
}

// UpdateDynamicScore 计算动态得分
func updateDynamicScores(classMatrix map[string]map[int]map[int]map[int]types.Val) error {

	return calcScores(classMatrix, constraint.CalcDynamic, true)
}

// updateTimeTableAndClassMatrix updates the time table and class matrix with the selected teacher, venue, and time slot.
func updateTimeTableAndClassMatrix(sn string, teacherID, venueID, timeSlot int, classMatrix map[string]map[int]map[int]map[int]types.Val, timeTable *TimeTable) {

	// 更新科班适应性矩阵元素分配信息
	temp := classMatrix[sn][teacherID][venueID][timeSlot]
	temp.Used = 1
	classMatrix[sn][teacherID][venueID][timeSlot] = temp
	timeTable.Used[timeSlot] = true

	// 更新科班适应性矩阵元素动态元素条件得分
	updateDynamicScores(classMatrix)
}
