package utils

import (
	"course_scheduler/config"
	"course_scheduler/internal/types"
	"fmt"
	"sort"
)

type ClassMatrixItem struct {
	ClassSN  string
	TimeSlot int
	Day      int
	Period   int
	Score    int
}

// PrintClassMatrix 以Markdown格式打印classMatrix
func PrintClassMatrix(classMatrix map[string]map[int]map[int]map[int]types.Val) {

	fmt.Println("| Time Slot| Day | Period | Score |")
	fmt.Println("| --- | --- | --- | --- |")

	items := make([]ClassMatrixItem, 0)
	for classSN, teacherMap := range classMatrix {
		if classSN == "1_1_1" {
			for _, venueMap := range teacherMap {
				for _, timeSlotMap := range venueMap {
					for timeSlot, val := range timeSlotMap {
						day := timeSlot / config.NumClasses
						period := timeSlot % config.NumClasses
						item := ClassMatrixItem{
							ClassSN:  classSN,
							TimeSlot: timeSlot,
							Day:      day,
							Period:   period,
							Score:    val.ScoreInfo.Score,
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
