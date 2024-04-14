package genetic_algorithm

type Gene struct {
	ClassSN   string // 课班信息，科目_年级_班级 如:美术_一年级_1班
	TeacherID int    // 教师id
	VenueID   int    // 教室id
	TimeSlot  int    // 时间段
}
