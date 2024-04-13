package genetic_algorithm

type Gene struct {
	ClassSN   string // 课班信息，如美术一班、美术二班等
	TeacherID int    // 教师id
	VenueID   int    // 教室id
	TimeSlot  int    // 时间段
}
