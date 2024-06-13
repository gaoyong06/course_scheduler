package test

import (
	"course_scheduler/internal/utils"
	"fmt"
	"testing"
)

func TestRemoveStr(t *testing.T) {

	slice := []string{"a", "b", "c", "d"}
	s := "c"

	a := utils.RemoveStr(slice, s)
	fmt.Printf("%v\n", a)

}

// func TestMap(t *testing.T) {

// 	classValidTime := map[string][]string{
// 		"1_1": {"1-1", "1-2", "1-3"},
// 		"1_2": {"1-1", "1-2", "1-3"},
// 		"2_1": {"2-1", "2-2", "2-3"},
// 		"2_2": {"2-1", "2-2", "2-3"},
// 	}

// 	teacherValidTime := map[string][]string{
// 		"1": {"1-1", "1-2", "1-3", "2-1", "2-2", "2-3"},
// 		"2": {"1-1", "1-2", "1-3", "2-1", "2-2", "2-3"},
// 		"3": {"1-1", "1-2", "1-3"},
// 		"4": {"2-1", "2-2", "2-3"},
// 	}

// }
