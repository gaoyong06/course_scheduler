// gene.go
package genetic_algorithm

type Gene struct {
	ClassSN   string // 课班信息，科目_年级_班级 如:美术_一年级_1班
	TeacherID int    // 教师id
	VenueID   int    // 教室id
	TimeSlot  int    // 时间段 一周5天,每天8节课,TimeSlot值是{0,1,2,3...39}
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

func (g *Gene) GetTimeSlot() int {
	return g.TimeSlot
}
