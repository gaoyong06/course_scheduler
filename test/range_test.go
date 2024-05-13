package test

import (
	"course_scheduler/internal/models"
	"fmt"
	"testing"
)

func TestGetPeriodWithRange(t *testing.T) {

	schedule := models.Schedule{
		Name:                     "心远中学2023年第一学期",
		NumWorkdays:              5,
		NumDaysOff:               2,
		NumMorningReadingClasses: 1,
		NumForenoonClasses:       4,
		NumAfternoonClasses:      4,
		NumNightClasses:          0,
	}

	s1, e1 := schedule.GetPeriodWithRange("morning_reading")
	s2, e2 := schedule.GetPeriodWithRange("forenoon")
	s3, e3 := schedule.GetPeriodWithRange("afternoon")
	s4, e4 := schedule.GetPeriodWithRange("night")

	fmt.Printf("morning_reading %d-%d\n", s1, e1)
	fmt.Printf("forenoon %d-%d\n", s2, e2)
	fmt.Printf("afternoon %d-%d\n", s3, e3)
	fmt.Printf("night %d-%d\n", s4, e4)

}
