package types

// 课班适应性矩阵中的一个元素
type Element struct {
	ClassSN   string // 科目_年级_班级
	SubjectID int    // 科目
	GradeID   int    // 年级
	ClassID   int    // 班级
	TeacherID int    // 教室
	VenueID   int    // 教室
	TimeSlot  int    // 时间段
}

func (e *Element) GetClassSN() string {
	return e.ClassSN
}

func (e *Element) GetTeacherID() int {
	return e.TeacherID
}

func (e *Element) GetVenueID() int {
	return e.VenueID
}

func (e *Element) GetTimeSlot() int {
	return e.TimeSlot
}
