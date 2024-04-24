package types

// 一个课堂单元，包括科目、年级、班级、教师、教室和一个节次
type ClassUnit interface {
	GetClassSN() string
	GetTeacherID() int
	GetVenueID() int
	GetTimeSlot() int
}
