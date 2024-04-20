package utils

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/types"
	"fmt"
)

// PrintClassMatrix 以Markdown格式打印classMatrix
func PrintClassMatrix(classMatrix map[string]map[int]map[int]map[int]types.Val) {

	fmt.Println("| Class SN | Teacher ID | Venue ID | Time Slot | Score | Period | Used |")
	fmt.Println("| --- | --- | --- | --- | --- | --- | --- |")

	for classSN, teacherMap := range classMatrix {
		if classSN == "1_1_1" {
			for teacherID, venueMap := range teacherMap {
				for venueID, timeSlotMap := range venueMap {
					for timeSlot, val := range timeSlotMap {
						period := timeSlot % constants.NUM_CLASSES
						usedStr := "true"
						if val.Used == 0 {
							usedStr = "false"
						}
						fmt.Printf("| %s | %d | %d | %d | %d | %d | %s |\n", classSN, teacherID, venueID, timeSlot, val.Score, period, usedStr)
					}
				}
			}

		}

	}
}
