// schedule.go
package models

import (
	"github.com/samber/lo"
)

// 课程表结构
// 课表方案
// 每周上课天数, 每天上课节数
type Schedule struct {
	Name                     string `json:"name" mapstructure:"name"`                                               // 课表名称, 如: 心远中学2023年第一学期课程表
	Workdays                 int    `json:"workdays" mapstructure:"workdays"`                                       // 一周工作日, 默认:5天
	DaysOff                  int    `json:"days_off" mapstructure:"days_off"`                                       // 一周休息日, 默认:2天
	NumMorningReadingClasses int    `json:"num_morning_reading_classes" mapstructure:"num_morning_reading_classes"` // 早读 几节课, 默认: 0节
	NumForenoonClasses       int    `json:"num_forenoon_classes" mapstructure:"num_forenoon_classes"`               // 上午 几节课, 默认: 0节
	// 下面和约束条件中的教师时间段约束不一致,教师时间段约束中没有中午,容易出现混乱,先去掉
	// NumNoonClasses           int    `json:"num_noon_classes" mapstructure:"num_noon_classes"`                       // 中午 几节课, 默认: 4节
	NumAfternoonClasses int `json:"num_afternoon_classes" mapstructure:"num_afternoon_classes"` // 下午 几节课, 默认: 4节
	NumNightClasses     int `json:"num_night_classes" mapstructure:"num_night_classes"`         // 晚自习 几节课, 默认: 0节
}

// 获取每天的总课时数，包括早读、上午、中午、下午和晚自习的课时数
func (s *Schedule) GetTotalClassesPerDay() int {
	// return s.NumMorningReadingClasses + s.NumForenoonClasses + s.NumNoonClasses + s.NumAfternoonClasses + s.NumNightClasses
	return s.NumMorningReadingClasses + s.NumForenoonClasses + s.NumAfternoonClasses + s.NumNightClasses
}

// 根据时间区间获取每天的节次值
// startPeriod 起始节次
// endPeriod 截止节次
func (s *Schedule) GetPeriodWithRange(r string) (int, int) {
	timeRanges := map[string]int{
		"morning_reading": s.NumMorningReadingClasses,
		"forenoon":        s.NumForenoonClasses,
		"afternoon":       s.NumAfternoonClasses,
		"night":           s.NumNightClasses,
	}

	totalClasses := 0
	startPeriod, endPeriod := 0, 0

	for rangeName, numClasses := range timeRanges {
		if rangeName == r {
			if numClasses > 0 {
				startPeriod = totalClasses
				endPeriod = startPeriod + numClasses - 1
			}
			break
		}
		totalClasses += numClasses
	}

	return startPeriod, endPeriod
}

// 生成一周课程时间段
func (s *Schedule) GenWeekTimeSlots() []int {

	// 每天总课时
	dayTotalClasses := s.GetTotalClassesPerDay()
	// 每周总课时
	weekTotalClasses := s.Workdays * dayTotalClasses

	timeSlots := lo.Range(weekTotalClasses)

	return timeSlots
}
