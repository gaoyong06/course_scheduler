package utils

import (
	"course_scheduler/internal/models"
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// 将字符串a_b,转为[a,b]
func ParseTimeSlotStr(timeSlotStr string) []int {

	parts := strings.Split(timeSlotStr, "_")
	timeSlots := make([]int, len(parts))
	for i, part := range parts {
		timeSlot := cast.ToInt(part)
		timeSlots[i] = timeSlot
	}
	return timeSlots
}

// 将数组[a,b]转为字符串a_b
func TimeSlotsToStr(timeSlot []int) string {

	// 将时间段切片中的元素转换为字符串
	strs := make([]string, len(timeSlot))
	for i, t := range timeSlot {
		strs[i] = strconv.Itoa(t)
	}

	// 连接字符串并返回
	return strings.Join(strs, "_")
}

// 从availableSlots获取一个可用的连堂课时间
func GetConnectedTimeSlots(schedule *models.Schedule, availableSlots []int) (int, int) {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()

	// 找出所有的连堂课时间段
	pairs := make([][]int, 0)
	for i := 0; i < len(availableSlots)-1; i++ {
		if availableSlots[i]+1 == availableSlots[i+1] {
			pairs = append(pairs, []int{availableSlots[i], availableSlots[i+1]})
		}
	}

	// 上午的时间段
	forenoonStartPeriod, forenoonEndPeriod := schedule.GetPeriodWithRange("forenoon")

	// 下午的时间段
	afternoonStartPeriod, afternoonEndPeriod := schedule.GetPeriodWithRange("afternoon")

	// 遍历所有的连堂课时间段，找出一个可用的
	for _, pair := range pairs {

		day1 := pair[0] / totalClassesPerDay
		day2 := pair[1] / totalClassesPerDay

		period0 := pair[0] % totalClassesPerDay
		period1 := pair[1] % totalClassesPerDay

		if day1 == day2 && (period0 >= forenoonStartPeriod && period1 <= forenoonEndPeriod) ||
			(period0 >= afternoonStartPeriod && period1 <= afternoonEndPeriod) {
			return pair[0], pair[1]
		}
	}

	// 没有找到可用的连堂课时间，返回-1，-1
	return -1, -1
}

// 从timeSlotStr切片中获取连堂课时间段,组成切片返回
func GetConnectedTimeSlotStrs(timeSlotStrs []string) []string {

	var strs []string
	for _, str := range timeSlotStrs {

		parts := strings.Split(str, "_")
		if len(parts) == 2 {
			strs = append(strs, str)
		}
	}
	return strs
}

// 全部的连堂课时间段
func GetAllConnectedTimeSlots(schedule *models.Schedule) []string {

	var timeSlotStrs []string

	timeSlots := schedule.GenWeekTimeSlots()
	totalClassesPerDay := schedule.GetTotalClassesPerDay()

	// 设置连堂课的时间是上午和下午
	amStart, amEnd := schedule.GetPeriodWithRange("forenoon")
	pmStart, pmEnd := schedule.GetPeriodWithRange("afternoon")

	for _, timeSlot := range timeSlots {

		if !lo.Contains(timeSlots, timeSlot+1) {
			continue
		}

		day1, day2 := timeSlot/totalClassesPerDay, (timeSlot+1)/totalClassesPerDay
		period1, period2 := timeSlot%totalClassesPerDay, (timeSlot+1)%totalClassesPerDay

		if day1 == day2 && ((period1 >= amStart && period1 <= amEnd && period2 >= amStart && period2 <= amEnd) ||
			(period1 >= pmStart && period1 <= pmEnd && period2 >= pmStart && period2 <= pmEnd)) {
			timeSlotStrs = append(timeSlotStrs, fmt.Sprintf("%d_%d", timeSlot, timeSlot+1))
		}
	}

	return timeSlotStrs
}

// 全部的普通课时间段
func GetAllNormalTimeSlots(schedule *models.Schedule) []string {

	var timeSlotStrs []string
	timeSlots := schedule.GenWeekTimeSlots()
	for _, timeSlot := range timeSlots {
		timeSlotStrs = append(timeSlotStrs, fmt.Sprintf("%d", timeSlot))
	}
	return timeSlotStrs
}
