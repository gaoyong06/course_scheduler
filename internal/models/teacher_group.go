package models

var (
	teacherGroup = []TeacherGroup{
		{TeacherGroupID: 1, Name: "语文组"},
		{TeacherGroupID: 2, Name: "数学组"},
		{TeacherGroupID: 3, Name: "行政领导"},
	}
)

type TeacherGroup struct {
	TeacherGroupID int    // 分组id
	Name           string // 分组名称
}
