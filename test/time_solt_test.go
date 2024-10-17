package test

import (
	"course_scheduler/internal/base"
	"course_scheduler/internal/utils"
	"fmt"
	"log"
	"testing"
)

func TestTimeSoltHelper(t *testing.T) {

	configFilePath := "/Users/apple/Documents/work/my/course_scheduler/testdata/grade_school.yaml"
	input, err := base.LoadTestData(configFilePath)
	if err != nil {
		log.Fatalf("load test data failed. %s", err)
	}
	schedule := input.Schedule

	// 测试ParseTimeSlotStr
	timeSlotStr1 := "2_3"
	timeSlots1 := utils.ParseTimeSlotStr(timeSlotStr1)
	fmt.Printf("utils.ParseTimeSlotStr timeSlots: %#v\n", timeSlots1)

	// 测试TimeSlotsToStr
	timeSlotStr2 := utils.TimeSlotsToStr(timeSlots1)
	fmt.Printf("utils.TimeSlotsToStr timeSlotStr2: %#v\n", timeSlotStr2)

	// 从availableSlots获取一个可用的连堂课时间
	availableSlots := []int{0, 3, 4}
	timeSlots4, timeSlots5 := utils.GetConnectedTimeSlots(schedule, availableSlots)
	fmt.Printf("utils.GetConnectedTimeSlots timeSlots4: %d, timeSlots5: %d\n", timeSlots4, timeSlots5)
}
