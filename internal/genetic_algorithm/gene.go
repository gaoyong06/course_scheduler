// gene.go
package genetic_algorithm

import "course_scheduler/internal/utils"

// 基因
type Gene struct {
	ClassSN            string   // 课班信息，科目_年级_班级 如:美术_一年级_1班
	TeacherID          int      // 教师id
	VenueID            int      // 教室id
	TimeSlots          []int    // 时间段 一周5天,每天8节课,TimeSlot值是{0,1,2,3...39}
	IsConnected        bool     // 是否是连堂课
	FailedConstraints  []string // 未满足的约束条件
	PassedConstraints  []string // 已满足的约束条件
	SkippedConstraints []string // 已满足的约束条件
}

func (g *Gene) GetClassSN() string {
	return g.ClassSN
}

func (g *Gene) GetTeacherID() int {
	return g.TeacherID
}

func (g *Gene) GetVenueID() int {
	return g.VenueID
}

func (g *Gene) GetTimeSlots() []int {
	return g.TimeSlots
}

// 判断两个基因是否相同
func (g *Gene) Equal(gene *Gene) bool {

	str1 := utils.TimeSlotsToStr(g.TimeSlots)
	str2 := utils.TimeSlotsToStr(gene.TimeSlots)

	if g.ClassSN == gene.ClassSN && g.TeacherID == gene.TeacherID && g.VenueID == gene.VenueID && str1 == str2 {
		return true
	}
	return false
}
