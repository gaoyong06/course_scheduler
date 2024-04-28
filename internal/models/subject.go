// subject.go
package models

import (
	"fmt"

	"github.com/samber/lo"
)

type Subject struct {
	SubjectID       int    // 科目id
	Name            string // 名称
	SubjectGroupIDs []int  // 科目分组id
}

func GetSubjects() []Subject {

	subjects := []Subject{
		{SubjectID: 1, Name: "语文", SubjectGroupIDs: []int{1}},
		{SubjectID: 2, Name: "数学", SubjectGroupIDs: []int{1}},
		{SubjectID: 3, Name: "英语", SubjectGroupIDs: []int{1}},
		{SubjectID: 4, Name: "音乐", SubjectGroupIDs: []int{3}},
		{SubjectID: 5, Name: "美术", SubjectGroupIDs: []int{3}},
		{SubjectID: 6, Name: "体育", SubjectGroupIDs: []int{3}},
		{SubjectID: 7, Name: "物理", SubjectGroupIDs: []int{2}},
		{SubjectID: 8, Name: "化学", SubjectGroupIDs: []int{2}},
		{SubjectID: 9, Name: "政治", SubjectGroupIDs: []int{2}},
		{SubjectID: 10, Name: "历史", SubjectGroupIDs: []int{3}},
		{SubjectID: 11, Name: "生物", SubjectGroupIDs: []int{3}},
		{SubjectID: 12, Name: "地理", SubjectGroupIDs: []int{3}},
		{SubjectID: 13, Name: "劳技", SubjectGroupIDs: []int{3}},
		{SubjectID: 14, Name: "活动", SubjectGroupIDs: []int{3}},
		{SubjectID: 15, Name: "班会", SubjectGroupIDs: []int{3}},
	}

	return subjects
}

// 获取课时信息
// 每周各个科目,分别ji
func GetClassHours() map[int]int {

	// 课时
	// subjectID: 课时
	// 共39
	classHours := map[int]int{
		1:  5, // 语文
		2:  4, // 数学
		3:  4, // 英语
		4:  3, // 音乐
		5:  3, // 美术
		6:  3, // 体育
		7:  3, // 物理
		8:  3, // 化学
		9:  2, // 政治
		10: 2, // 历史
		11: 2, // 生物
		12: 2, // 地理
		13: 1, // 劳技
		14: 1, // 活动
		15: 1, // 班会
	}

	return classHours
}

// 根据科目id查找科目对象
func FindSubjectByID(subjectID int) (*Subject, error) {

	subjects := GetSubjects()
	for _, subject := range subjects {
		if subject.SubjectID == subjectID {
			return &subject, nil
		}
	}
	return nil, fmt.Errorf("subject not found")
}

// 根据科目组id查找科目
func FindSubjectsByGroupID(groupID int) ([]*Subject, error) {

	var subjectsByGroupID []*Subject

	subjects := GetSubjects()
	for _, subject := range subjects {
		if lo.Contains(subject.SubjectGroupIDs, groupID) {
			subjectsByGroupID = append(subjectsByGroupID, &subject)
		}
	}

	if len(subjectsByGroupID) > 0 {
		return subjectsByGroupID, nil
	}
	return nil, fmt.Errorf("no subjects found for group ID %d", groupID)
}
