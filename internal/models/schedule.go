// schedule.go
package models

import (
	"errors"
	"fmt"

	"github.com/samber/lo"
)

// 课表方案
// 课程表结构
// 每周上课天数, 每天上课节数
type Schedule struct {
	Name                     string `json:"name" mapstructure:"name"`                                               // 课表名称, 如: 心远中学2023年第一学期课程表
	NumWorkdays              int    `json:"num_workdays" mapstructure:"num_workdays"`                               // 一周工作日, 默认:5天
	NumDaysOff               int    `json:"num_days_off" mapstructure:"num_days_off"`                               // 一周休息日, 默认:2天
	NumMorningReadingClasses int    `json:"num_morning_reading_classes" mapstructure:"num_morning_reading_classes"` // 早读 几节课, 默认: 0节
	NumForenoonClasses       int    `json:"num_forenoon_classes" mapstructure:"num_forenoon_classes"`               // 上午 几节课, 默认: 0节
	// 下面和约束条件中的教师时间段约束不一致,教师时间段约束中没有中午,容易出现混乱,先去掉
	// NumNoonClasses           int    `json:"num_noon_classes" mapstructure:"num_noon_classes"`                       // 中午 几节课, 默认: 4节
	NumAfternoonClasses int `json:"num_afternoon_classes" mapstructure:"num_afternoon_classes"` // 下午 几节课, 默认: 4节
	NumNightClasses     int `json:"num_night_classes" mapstructure:"num_night_classes"`         // 晚自习 几节课, 默认: 0节
}

func (s *Schedule) Check() error {

	if s.Name == "" {
		return errors.New("schedule name cannot be empty")
	}

	if s.NumWorkdays <= 0 || s.NumWorkdays > 30 {
		return errors.New("invalid num_workdays, must be in range (1, 30]")
	}

	if s.NumDaysOff < 0 {
		return errors.New("invalid num_days_off, must be non-negative")
	}

	if s.NumMorningReadingClasses < 0 {
		return errors.New("invalid num_morning_reading_classes, must be non-negative")
	}

	if s.NumForenoonClasses < 0 {
		return errors.New("invalid num_forenoon_classes, must be non-negative")
	}

	if s.NumAfternoonClasses < 0 {
		return errors.New("invalid num_afternoon_classes, must be non-negative")
	}

	if s.NumNightClasses < 0 {
		return errors.New("invalid num_night_classes, must be non-negative")
	}

	if s.NumMorningReadingClasses+s.NumForenoonClasses+s.NumAfternoonClasses+s.NumNightClasses <= 0 {
		return fmt.Errorf("invalid schedule, the sum of morning reading, forenoon, afternoon and night classes, must be positive")
	}

	return nil
}

// 获取每天的总课时数，包括早读、上午、中午、下午和晚自习的课时数
func (s *Schedule) GetTotalClassesPerDay() int {
	// return s.NumMorningReadingClasses + s.NumForenoonClasses + s.NumNoonClasses + s.NumAfternoonClasses + s.NumNightClasses
	return s.NumMorningReadingClasses + s.NumForenoonClasses + s.NumAfternoonClasses + s.NumNightClasses
}

// 根据时间区间获取每天的节次值(前闭后开)
// startPeriod 起始节次
// endPeriod 截止节次
// startPeriod, endPeriod 前闭后闭, -1 表示不存在
func (s *Schedule) GetPeriodWithRange(r string) (int, int) {
	timeRanges := []struct {
		name       string
		numClasses int
	}{
		{"morning_reading", s.NumMorningReadingClasses},
		{"forenoon", s.NumForenoonClasses},
		{"afternoon", s.NumAfternoonClasses},
		{"night", s.NumNightClasses},
	}

	totalClasses := 0
	startPeriod, endPeriod := -1, -1

	for _, rangeInfo := range timeRanges {
		if rangeInfo.name == r && rangeInfo.numClasses > 0 {
			startPeriod = totalClasses
			endPeriod = startPeriod + rangeInfo.numClasses - 1
			break
		}
		totalClasses += rangeInfo.numClasses
	}

	return startPeriod, endPeriod
}

// 生成一周课程时间段
func (s *Schedule) GenWeekTimeSlots() []int {

	// // 每天总课时
	// dayTotalClasses := s.GetTotalClassesPerDay()
	// // 每周总课时
	// weekTotalClasses := s.NumWorkdays * dayTotalClasses
	totalClassesPerWeek := s.TotalClassesPerWeek()

	timeSlots := lo.Range(totalClassesPerWeek)

	return timeSlots
}

// 每周总课时数
// TODO: 这个名字要统计下
func (s *Schedule) TotalClassesPerWeek() int {

	// 每天总课时
	dayTotalClasses := s.GetTotalClassesPerDay()
	// 每周总课时
	totalClassesPerWeek := s.NumWorkdays * dayTotalClasses
	return totalClassesPerWeek
}
