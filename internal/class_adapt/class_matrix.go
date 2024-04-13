package class_adapt

import (
	"course_scheduler/internal/evaluation"
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
	teachers := models.GetTeachers()

	for i := 0; i < len(classes); i++ {
		class := classes[i]
		teacherIDs := models.ClassTeacherIDs(class.SN)
		venueIDs := models.ClassVenueIDs(class.SN)
		timeSlots := ClassTimeSlots(teacherIDs, venueIDs)
		sn := class.SN.Generate()

		// fmt.Printf("initClassMatrix sn: %s, len(teacherIDs): %d, len(venueIDs): %d, len(timeSlots): %d\n", sn, len(teacherIDs), len(venueIDs), len(timeSlots))

		classMatrix[sn] = make(map[int]map[int]map[int]types.Val)
		for j := 0; j < len(teachers); j++ {
			teacher := teachers[j]
			teacherID := teacher.TeacherID
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

// 匹配结果值
func MatchScore(classMatrix map[string]map[int]map[int]map[int]types.Val) error {

	log.Println("Calculating match scores...")
	for sn, teacherMap := range classMatrix {
		for teacherID, venueMap := range teacherMap {
			for venueID, timeSlotMap := range venueMap {
				for timeSlot, val := range timeSlotMap {

					// start := time.Now() // 记录开始时间
					score, err := evaluation.CalcScore(classMatrix, sn, teacherID, venueID, timeSlot)
					if err != nil {
						return err
					}
					// duration := time.Since(start) // 计算耗时
					val.Score = score
					classMatrix[sn][teacherID][venueID][timeSlot] = val
					// log.Printf("Match score for (%s, %d, %d, %d) calculated: %d (took %v)\n", sn, teacherID, venueID, timeSlot, score, duration)
				}
			}
		}
	}
	log.Println("Match scores calculation completed.")
	return nil
}

// 课班适应性矩阵分配
// 循环迭代各个课班，根据匹配结果值, 为每个课班选择课班适应性矩阵中可用的点位，并记录，下个课班选择点位时会避免冲突(一个点位可以引起多点位冲突)
func AssignClassMatrix(classeSNs []string, classHours map[int]int, classMatrix map[string]map[int]map[int]map[int]types.Val) (int, error) {

	timeTable := initTimeTable()

	var numAssignedClasses int
	for _, sn := range classeSNs {

		fmt.Printf("assignClassMatrix sn: %s\n", sn)
		SN, err := types.ParseSN(sn)
		if err != nil {
			return numAssignedClasses, err
		}

		subjectID := SN.SubjectID
		numClassHours := classHours[subjectID]

		// 遍历课时
		for i := 0; i < numClassHours; i++ {
			maxScore := 0

			// 设置初始值
			selectedTeacherID, selectedVenueID, selectedTimeSlot := -1, -1, -1

			// 匹配结果值
			for teacherID, venueMap := range classMatrix[sn] {
				for venueID, timeSlotMap := range venueMap {
					for timeSlot, val := range timeSlotMap {

						// 检查时间段是否已被占用
						if timeTable.Used[timeSlot] {
							continue // 跳过已被占用的时间段
						}

						// 是适应性矩阵中可用的点位, 课班可用当前适应性矩阵的元素下标
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

			// 如果maxScore为0，随机取一个可用的Teacher, Classroom, TimeSlot
			// 选取可用的teacherIDs,venueIDs,timeSlots
			teacherIDs := models.ClassTeacherIDs(SN)
			venueIDs := models.ClassVenueIDs(SN)
			// timeSlots := classTimeSlots(teacherIDs, venueIDs)

			// 过滤掉已经被占用的时间段
			var availableTimeSlots []int
			for timeSlot, used := range timeTable.Used {
				if !used {
					availableTimeSlots = append(availableTimeSlots, timeSlot)
				}
			}

			// fmt.Printf("availableTimeSlots: %#v\n", availableTimeSlots)

			var teacherID, venueID, timeSlot int

			if maxScore == 0 {
				// 添加一个计数器，避免无限循环
				total := len(teacherIDs) * len(venueIDs) * len(availableTimeSlots)
				fmt.Printf("total: %d\n", total)
				for j := 0; j < total; j++ {

					teacherID = teacherIDs[rand.Intn(len(teacherIDs))]
					venueID = venueIDs[rand.Intn(len(venueIDs))]

					// TODO: 这是有错误
					// timeSlot = availableTimeSlots[rand.Intn(len(availableTimeSlots))]
					timeSlot = availableTimeSlots[j]

					val := classMatrix[sn][teacherID][venueID][timeSlot]
					// fmt.Printf("timeSlot: %d, timeTable.Used[timeSlot]: %v, val.Score: %d\n", timeSlot, timeTable.Used[timeSlot], val.Score)
					if !timeTable.Used[timeSlot] && val.Score == 0 {
						selectedTeacherID = teacherID
						selectedVenueID = venueID
						selectedTimeSlot = timeSlot
						break
					}
				}
			}

			if selectedTeacherID > 0 && selectedVenueID > 0 && selectedTimeSlot >= 0 {
				// 打印选择的老师、教室和时间段
				// fmt.Printf("Class: %s, Hour: %d, Teacher: %s, Classroom: %s, TimeSlot: %s, Score: %d\n", class, i+1, selectedTeacher, selectedClassroom, selectedTimeSlot, maxScore)

				// 更新适应性矩阵
				temp := classMatrix[sn][selectedTeacherID][selectedVenueID][selectedTimeSlot]
				temp.Used = 1
				// fmt.Printf("sn: %s, selectedTeacherID: %d, selectedVenueID: %d, selectedTimeSlot: %d\n", sn, selectedTeacherID, selectedVenueID, selectedTimeSlot)
				classMatrix[sn][selectedTeacherID][selectedVenueID][selectedTimeSlot] = temp
				timeTable.Used[selectedTimeSlot] = true
				numAssignedClasses++
			} else {

				fmt.Println("============= classMatrix ==============")
				fmt.Printf("============= sn: %s, teacherID: %d, venueID: %d\n", sn, teacherID, venueID)
				fmt.Printf("%#v\n", classMatrix[sn][teacherID][venueID])

				return numAssignedClasses, fmt.Errorf("failed sn: %s, classHour: %d,  numClassHours: %d", sn, i+1, numClassHours)
			}
		}
	}
	return numAssignedClasses, nil
}

// 打印classMatrix
func PrintClassMatrix(classMatrix map[string]map[int]map[int]map[int]types.Val) {

	for classSN, teacherMap := range classMatrix {
		fmt.Printf("%s:\n", classSN)
		for teacherID, venueMap := range teacherMap {
			fmt.Printf("\tTeacher ID: %d\n", teacherID)
			for venueID, timeSlotMap := range venueMap {
				fmt.Printf("\t\tVenue ID: %d\n", venueID)
				for timeSlot, val := range timeSlotMap {
					if val.Used == 1 {
						fmt.Printf("\t\t\tTime Slot: %d, Score: %d, Used: %t\n", timeSlot, val.Score, val.Used == 1)
					}

				}
			}
		}
	}
}
