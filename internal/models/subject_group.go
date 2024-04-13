package models

var (
	subjectGroup = []SubjectGroup{
		{SubjectGroupID: 1, Name: "语数英"},
		{SubjectGroupID: 2, Name: "主课"},
		{SubjectGroupID: 3, Name: "副课"},
	}
)

type SubjectGroup struct {
	SubjectGroupID int    // 分组id
	Name           string // 分组名称
}
