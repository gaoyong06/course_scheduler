package utils

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/types"
	"fmt"
	"sort"
)

// // PrintClassMatrix 以Markdown格式打印classMatrix
// func PrintClassMatrix(classMatrix map[string]map[int]map[int]map[int]types.Val) {

// 	// fmt.Println("| Class SN | Teacher ID | Venue ID | Time Slot | Day | Period | Score | Used |")
// 	// fmt.Println("| --- | --- | --- | --- | --- | --- | --- |")

// 	fmt.Println("| Day | Period | Score |")
// 	fmt.Println("| --- | --- | --- |")

// 	for classSN, teacherMap := range classMatrix {
// 		if classSN == "6_1_1" {
// 			// for teacherID, venueMap := range teacherMap {
// 			for _, venueMap := range teacherMap {
// 				// for venueID, timeSlotMap := range venueMap {
// 				for _, timeSlotMap := range venueMap {
// 					for timeSlot, val := range timeSlotMap {
// 						day := timeSlot / constants.NUM_CLASSES
// 						period := timeSlot % constants.NUM_CLASSES
// 						// usedStr := "true"
// 						// if val.Used == 0 {
// 						// 	usedStr = "false"
// 						// }
// 						// fmt.Printf("| %s | %d | %d | %d | %d | %d | %d | %s |\n", classSN, teacherID, venueID, timeSlot, day, period, val.Score, usedStr)
// 						fmt.Printf("| %d | %d | %d |\n", day, period, val.Score)
// 					}
// 				}
// 			}
// 		}
// 	}
// }

type ClassMatrixItem struct {
	ClassSN  string
	TimeSlot int
	Day      int
	Period   int
	Score    int
}

func PrintClassMatrix(classMatrix map[string]map[int]map[int]map[int]types.Val) {

	fmt.Println("| Time Slot| Day | Period | Score |")
	fmt.Println("| --- | --- | --- | --- |")

	items := make([]ClassMatrixItem, 0)
	for classSN, teacherMap := range classMatrix {
		if classSN == "6_1_1" {
			for _, venueMap := range teacherMap {
				for _, timeSlotMap := range venueMap {
					for timeSlot, val := range timeSlotMap {
						day := timeSlot / constants.NUM_CLASSES
						period := timeSlot % constants.NUM_CLASSES
						item := ClassMatrixItem{
							ClassSN:  classSN,
							TimeSlot: timeSlot,
							Day:      day,
							Period:   period,
							Score:    val.Score,
						}
						items = append(items, item)
					}
				}
			}
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Day == items[j].Day {
			return items[i].Period < items[j].Period
		}
		return items[i].Day < items[j].Day
	})

	for _, item := range items {
		fmt.Printf("| %d | %d | %d | %d |\n", item.TimeSlot, item.Day, item.Period, item.Score)
	}
}
