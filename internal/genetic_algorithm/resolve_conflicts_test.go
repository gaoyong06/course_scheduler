package genetic_algorithm

import (
	"course_scheduler/internal/base"
	"course_scheduler/internal/constraints"
	"course_scheduler/internal/utils"
	"fmt"
	"testing"
)

func getIndivid() *Individual {

	// 构造Individual
	individ := &Individual{
		Chromosomes: []*Chromosome{
			&Chromosome{
				ClassSN: "1_9_1",
				Genes: []*Gene{
					&Gene{
						ClassSN:            "1_9_1",
						TeacherID:          1,
						VenueID:            901,
						TimeSlots:          []int{10, 11},
						IsConnected:        true, // 假设这是连堂课
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "1_9_1",
						TeacherID:          1,
						VenueID:            901,
						TimeSlots:          []int{16},
						IsConnected:        false,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "1_9_1",
						TeacherID:          1,
						VenueID:            901,
						TimeSlots:          []int{0},
						IsConnected:        false,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "1_9_1",
						TeacherID:          1,
						VenueID:            901,
						TimeSlots:          []int{35},
						IsConnected:        false,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "1_9_1",
						TeacherID:          1,
						VenueID:            901,
						TimeSlots:          []int{29, 30},
						IsConnected:        true,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
				},
			},
			&Chromosome{
				ClassSN: "2_9_1",
				Genes: []*Gene{
					&Gene{
						ClassSN:            "2_9_1",
						TeacherID:          8,
						VenueID:            901,
						TimeSlots:          []int{36, 37},
						IsConnected:        true,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "2_9_1",
						TeacherID:          8,
						VenueID:            901,
						TimeSlots:          []int{16, 17},
						IsConnected:        true,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "2_9_1",
						TeacherID:          8,
						VenueID:            901,
						TimeSlots:          []int{11},
						IsConnected:        false,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "2_9_1",
						TeacherID:          8,
						VenueID:            901,
						TimeSlots:          []int{2},
						IsConnected:        false,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "2_9_1",
						TeacherID:          8,
						VenueID:            901,
						TimeSlots:          []int{29},
						IsConnected:        false,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
				},
			},
			&Chromosome{
				ClassSN: "3_9_1",
				Genes: []*Gene{
					&Gene{
						ClassSN:            "3_9_1",
						TeacherID:          16,
						VenueID:            901,
						TimeSlots:          []int{18, 19},
						IsConnected:        true,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "3_9_1",
						TeacherID:          16,
						VenueID:            901,
						TimeSlots:          []int{30},
						IsConnected:        false,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "3_9_1",
						TeacherID:          16,
						VenueID:            901,
						TimeSlots:          []int{32, 33},
						IsConnected:        true,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
					&Gene{
						ClassSN:            "3_9_1",
						TeacherID:          16,
						VenueID:            901,
						TimeSlots:          []int{6},
						IsConnected:        false,
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
				},
			},
		},
		Fitness: 0, // 填入适应度值
	}

	return individ
}

func TestResolveConflicts(t *testing.T) {

	input, _ := base.LoadTestData()
	constraintMap := input.Constraints()
	schedule := input.Schedule
	constr1 := constraintMap["Class"].([]*constraints.Class)
	constr2 := constraintMap["Teacher"].([]*constraints.Teacher)
	teachers := input.Teachers

	i := getIndivid()

	// 打印构造的 Individual
	fmt.Printf("====== 构造的 Individual 排课信息 ======\n")
	for _, chromosome := range i.Chromosomes {
		for _, gene := range chromosome.Genes {
			fmt.Printf("sn: %s, teacherID: %d, venueID: %d, timeSlots: %v\n", gene.ClassSN, gene.TeacherID, gene.VenueID, gene.TimeSlots)
		}
	}

	// 已经使用的连堂课, 普通课时间段
	connectedGenes, normalGenes := i.getClassTimeSlots(false)

	fmt.Println("==== 已经使用的连堂课 =====")
	for classKey, list := range connectedGenes {
		for _, gene := range list {
			fmt.Printf("individual %p connectedGenes classKey: %s, sn: %s, timeSlots: %v\n", i, classKey, gene.ClassSN, gene.TimeSlots)
		}
	}

	fmt.Println("==== 已经使用的普通课 =====")
	for classKey, list := range normalGenes {
		for _, gene := range list {
			fmt.Printf("individual %p normalGenes classKey: %s, sn: %s, timeSlots: %v\n", i, classKey, gene.ClassSN, gene.TimeSlots)
		}
	}

	// 班级时间段冲突
	classConnectedConflict, classNormalConflict := i.getClassTimeSlots(true)

	// 教师时间段冲突
	teacherConnectedConflict, teacherNormalConflict := i.getTeacherTimeSlots(true)

	fmt.Println("==== 班级连堂课冲突 =====")
	for classKey, conflictList := range classConnectedConflict {
		for _, gene := range conflictList {
			fmt.Printf("individual %p classConnectedConflict classKey: %s, sn: %s, timeSlots: %v\n", i, classKey, gene.ClassSN, gene.TimeSlots)
		}
	}

	fmt.Println("==== 班级普通冲突 =====")
	for classKey, conflictList := range classNormalConflict {
		for _, gene := range conflictList {
			fmt.Printf("individual %p classNormalConflict classKey: %s, sn: %s, timeSlots: %v\n", i, classKey, gene.ClassSN, gene.TimeSlots)
		}
	}

	fmt.Println("==== 教师连堂课冲突 =====")
	for teacherID, conflictList := range teacherConnectedConflict {
		for _, gene := range conflictList {
			fmt.Printf("individual %p teacherConnectedConflict teacherID: %s, sn: %s, timeSlots: %v\n", i, teacherID, gene.ClassSN, gene.TimeSlots)
		}
	}

	fmt.Println("==== 教师普通冲突 =====")
	for teacherID, conflictList := range teacherNormalConflict {
		for _, gene := range conflictList {
			fmt.Printf("individual %p teacherNormalConflict teacherID: %s, sn: %s, timeSlots: %v\n", i, teacherID, gene.ClassSN, gene.TimeSlots)
		}
	}

	// 全部连堂课时间
	allConnected := utils.GetAllConnectedTimeSlots(schedule)

	// 全部的普通课时间
	allNormal := utils.GetAllNormalTimeSlots(schedule)

	fmt.Println("====== 全部连堂课时间 ======")
	fmt.Printf("allConnected: %v\n", allConnected)

	fmt.Println("====== 全部的普通课时间 ======")
	fmt.Printf("allNormal: %v\n", allNormal)

	// 班级可用时间段
	classConnected, classNormal := i.getClassValidTimeSlots(schedule, constr1)
	fmt.Println("==== 班级可用时间段 =====")
	fmt.Printf("individual %p classConnected: %v\n", i, classConnected)
	fmt.Printf("individual %p classNormal: %v\n", i, classNormal)

	// 教师可用时间段
	teacherConnected, teacherNormal, _ := i.getTeacherValidTimeSlots(schedule, teachers, constr2)
	fmt.Println("==== 教师可用时间段 =====")
	fmt.Printf("individual %p teacherConnected: %v\n", i, teacherConnected)
	fmt.Printf("individual %p teacherNormal: %v\n", i, teacherNormal)

	// 冲突去重
	// 从教师冲突中去重, 即在班级冲突中存在又在教师冲突中存在的基因
	i.rejectConflictGenes(teacherConnectedConflict, classConnectedConflict)
	i.rejectConflictGenes(teacherNormalConflict, classNormalConflict)

	// 修复班级连堂课,普通课冲突
	count1, err1 := i.resolveClassConflict(classConnectedConflict, classConnected, teacherConnected)
	count2, err2 := i.resolveClassConflict(classNormalConflict, classNormal, teacherNormal)

	// 修复教师连堂课,普通课冲突
	count3, err3 := i.resolveTeacherConflict(teacherConnectedConflict, teacherConnected, classConnected)
	count4, err4 := i.resolveTeacherConflict(teacherNormalConflict, teacherNormal, classNormal)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		fmt.Printf("resolve conflicts failed. err1: %v, err2: %v, err3: %v, err4: %v\n", err1, err2, err3, err4)
	}
	count := count1 + count2 + count3 + count4
	fmt.Printf("resolve conflicts success. count: %d\n", count)

}
