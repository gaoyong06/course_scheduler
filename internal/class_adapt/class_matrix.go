// class_matrix.go
package class_adapt

import (
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"log"
	"math/rand"
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
					classMatrix[sn][teacherID][venueID][timeSlot] = types.Val{Score: 0, Used: 0}
				}
			}
		}
	}

	return classMatrix
}

// 课班适应性矩阵各元素, 计算固定约束条件下的得分
func CalcFixedScores(classMatrix map[string]map[int]map[int]map[int]types.Val, classHours map[int]int) error {
	return calcScores(classMatrix, classHours, constraint.CalcFixed, false)
}

// // 课班适应性矩阵各元素, 计算固定约束条件下的得分
// func CalcFixedScore(classMatrix map[string]map[int]map[int]map[int]types.Val, classHours map[int]int) error {

// 	log.Println("Calculating fixed scores...")
// 	for sn, teacherMap := range classMatrix {

// 		SN, err := types.ParseSN(sn)
// 		if err != nil {
// 			return err
// 		}

// 		for teacherID, venueMap := range teacherMap {
// 			for venueID, timeSlotMap := range venueMap {
// 				for timeSlot, val := range timeSlotMap {

// 					// start := time.Now() // 记录开始时间
// 					// score, err := evaluation.CalcScore(classMatrix, classHours, sn, teacherID, venueID, timeSlot)
// 					element := constraint.Element{
// 						ClassSN:   sn,
// 						SubjectID: SN.SubjectID,
// 						GradeID:   SN.GradeID,
// 						ClassID:   SN.ClassID,
// 						TeacherID: teacherID,
// 						VenueID:   venueID,
// 						TimeSlot:  timeSlot,
// 					}
// 					score, err := constraint.CalcFixed(classMatrix, element)
// 					if err != nil {
// 						return err
// 					}
// 					// duration := time.Since(start) // 计算耗时
// 					// val.Score = score.FinalScore
// 					val.Score = score
// 					classMatrix[sn][teacherID][venueID][timeSlot] = val
// 					// log.Printf("Match score for (%s, %d, %d, %d) calculated: %d (took %v)\n", sn, teacherID, venueID, timeSlot, score, duration)
// 				}
// 			}
// 		}
// 	}
// 	log.Println("Fixed scores calculation completed.")
// 	return nil
// }

// // 实时更新动态约束条件得分
// func updateDynamicScore(classMatrix map[string]map[int]map[int]map[int]types.Val, classHours map[int]int) error {

// 	log.Println("Calculating fixed scores...")
// 	for sn, teacherMap := range classMatrix {

// 		SN, err := types.ParseSN(sn)
// 		if err != nil {
// 			return err
// 		}

// 		for teacherID, venueMap := range teacherMap {
// 			for venueID, timeSlotMap := range venueMap {
// 				for timeSlot, val := range timeSlotMap {

// 					// start := time.Now() // 记录开始时间
// 					element := constraint.Element{
// 						ClassSN:   sn,
// 						SubjectID: SN.SubjectID,
// 						GradeID:   SN.GradeID,
// 						ClassID:   SN.ClassID,
// 						TeacherID: teacherID,
// 						VenueID:   venueID,
// 						TimeSlot:  timeSlot,
// 					}
// 					score, err := constraint.CalcDynamic(classMatrix, element)
// 					if err != nil {
// 						return err
// 					}
// 					// duration := time.Since(start) // 计算耗时

// 					// 总分=固定得分+动态得分
// 					val.Score = val.Score + score
// 					classMatrix[sn][teacherID][venueID][timeSlot] = val
// 					// log.Printf("Match score for (%s, %d, %d, %d) calculated: %d (took %v)\n", sn, teacherID, venueID, timeSlot, score, duration)
// 				}
// 			}
// 		}
// 	}
// 	log.Println("Fixed scores calculation completed.")
// 	return nil
// }

// 课班适应性矩阵分配
// 循环迭代各个课班，根据匹配结果值, 为每个课班选择课班适应性矩阵中可用的点位，并记录，下个课班选择点位时会避免冲突(一个点位可以引起多点位冲突)
// AllocateClassMatrix allocates class hours based on the class adaptability matrix.
func AllocateClassMatrix(classSNs []string, classHours map[int]int, classMatrix map[string]map[int]map[int]map[int]types.Val) (int, error) {

	timeTable := initTimeTable()

	var numAssignedClasses int
	var numBestTimeSlots, numRandomTimeSlots int

	for _, sn := range classSNs {

		// log.Printf("assignClassMatrix sn: %s\n", sn)
		SN, err := types.ParseSN(sn)
		if err != nil {
			return numAssignedClasses, err
		}

		subjectID := SN.SubjectID
		numClassHours := classHours[subjectID]

		// Loop through class hours.
		for i := 0; i < numClassHours; i++ {

			// Find the best available time slot.
			selectedTeacherID, selectedVenueID, selectedTimeSlot, maxScore, err := findBestAvailableTimeSlot(sn, classMatrix, timeTable)
			if err != nil {
				return numAssignedClasses, err
			}

			// If no available time slot found, try to find a random available one.
			if maxScore == 0 {
				selectedTeacherID, selectedVenueID, selectedTimeSlot, err = findRandomAvailableTimeSlot(sn, classMatrix, timeTable)
				if err != nil {
					return numAssignedClasses, err
				}
				numRandomTimeSlots++
			} else {
				numBestTimeSlots++
			}

			// Update the time table and class matrix.
			if selectedTeacherID > 0 && selectedVenueID > 0 && selectedTimeSlot >= 0 {
				updateTimeTableAndClassMatrix(sn, selectedTeacherID, selectedVenueID, selectedTimeSlot, classMatrix, timeTable)
				numAssignedClasses++
			} else {
				return numAssignedClasses, fmt.Errorf("failed sn: %s, classHour: %d,  numClassHours: %d", sn, i+1, numClassHours)
			}
		}
	}

	log.Printf("Number of best time slots assigned: %d\n", numBestTimeSlots)
	log.Printf("Number of random time slots assigned: %d\n", numRandomTimeSlots)

	return numAssignedClasses, nil
}

// findBestAvailableTimeSlot finds the best available time slot for a given class SN.
func findBestAvailableTimeSlot(sn string, classMatrix map[string]map[int]map[int]map[int]types.Val, timeTable *TimeTable) (int, int, int, int, error) {
	maxScore := 0
	selectedTeacherID, selectedVenueID, selectedTimeSlot := -1, -1, -1

	for teacherID, venueMap := range classMatrix[sn] {
		for venueID, timeSlotMap := range venueMap {
			for timeSlot, val := range timeSlotMap {
				if timeTable.Used[timeSlot] {
					continue
				}
				score := val.Score
				if score > maxScore {
					maxScore = score
					selectedTeacherID = teacherID
					selectedVenueID = venueID
					selectedTimeSlot = timeSlot
				}
			}
		}
	}

	return selectedTeacherID, selectedVenueID, selectedTimeSlot, maxScore, nil
}

// findRandomAvailableTimeSlot finds a random available time slot for a given class SN.
func findRandomAvailableTimeSlot(sn string, classMatrix map[string]map[int]map[int]map[int]types.Val, timeTable *TimeTable) (int, int, int, error) {

	SN, err := types.ParseSN(sn)
	if err != nil {
		return -1, -1, -1, err
	}

	teacherIDs := models.ClassTeacherIDs(SN)
	venueIDs := models.ClassVenueIDs(SN)

	var availableTimeSlots []int
	for timeSlot, used := range timeTable.Used {
		if !used {
			availableTimeSlots = append(availableTimeSlots, timeSlot)
		}
	}

	// Shuffle availableTimeSlots to make the selection more random and dispersed.
	rand.Shuffle(len(availableTimeSlots), func(i, j int) {
		availableTimeSlots[i], availableTimeSlots[j] = availableTimeSlots[j], availableTimeSlots[i]
	})

	total := len(teacherIDs) * len(venueIDs) * len(availableTimeSlots)
	for j := 0; j < total; j++ {
		teacherID := teacherIDs[rand.Intn(len(teacherIDs))]
		venueID := venueIDs[rand.Intn(len(venueIDs))]
		timeSlot := availableTimeSlots[j%len(availableTimeSlots)] // Use modulo to ensure all available time slots are traversed.

		val := classMatrix[sn][teacherID][venueID][timeSlot]
		if !timeTable.Used[timeSlot] && val.Score == 0 {
			return teacherID, venueID, timeSlot, nil
		}
	}

	return -1, -1, -1, fmt.Errorf("no available time slot found")
}

// CalcScores 计算固定得分和动态得分
func calcScores(classMatrix map[string]map[int]map[int]map[int]types.Val, classHours map[int]int, calcFunc func(map[string]map[int]map[int]map[int]types.Val, constraint.Element) (int, error), isAddScore bool) error {

	// log.Println("Calculating scores...")
	for sn, teacherMap := range classMatrix {
		SN, err := types.ParseSN(sn)
		if err != nil {
			return err
		}
		for teacherID, venueMap := range teacherMap {
			for venueID, timeSlotMap := range venueMap {
				for timeSlot, val := range timeSlotMap {
					element := constraint.Element{
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
						val.Score = val.Score + score
					} else {
						val.Score = score
					}
					classMatrix[sn][teacherID][venueID][timeSlot] = val
				}
			}
		}
	}
	// log.Println("Scores calculation completed.")
	return nil
}

// UpdateDynamicScore 计算动态得分
func updateDynamicScores(classMatrix map[string]map[int]map[int]map[int]types.Val, classHours map[int]int) error {
	return calcScores(classMatrix, classHours, constraint.CalcDynamic, true)
}

// updateTimeTableAndClassMatrix updates the time table and class matrix with the selected teacher, venue, and time slot.
func updateTimeTableAndClassMatrix(sn string, teacherID, venueID, timeSlot int, classMatrix map[string]map[int]map[int]map[int]types.Val, timeTable *TimeTable) {

	// 更新科班适应性矩阵元素分配信息
	temp := classMatrix[sn][teacherID][venueID][timeSlot]
	temp.Used = 1
	classMatrix[sn][teacherID][venueID][timeSlot] = temp

	// TODO: 这个参数要修改
	classHours := models.GetClassHours()

	// 更新科班适应性矩阵元素动态元素条件得分
	updateDynamicScores(classMatrix, classHours)

	timeTable.Used[timeSlot] = true
}
