// subject.go
package models

import (
	"fmt"

	"github.com/samber/lo"
)

type Subject struct {
	SubjectID       int    `json:"subject_id" mapstructure:"subject_id"`               // 科目id
	Name            string `json:"name" mapstructure:"name"`                           // 名称
	SubjectGroupIDs []int  `json:"subject_group_ids" mapstructure:"subject_group_ids"` // 科目分组id
	Priority        int    `json:"priority" mapstructure:"priority"`                   // 排课的优先级, 优先级高的优先排课
}

// 读取科目信息
// 启动遗传算法之前，从数据库中读取必要的基本数据并将其输入到排课程序中
// 优点:
// 1.性能更好: 从数据库中阅读一次数据，然后在内存中处理它比在遗传算法执行期间重复查询数据库要快
// 2.一致性: 通过加载数据一次，可以确保数据在遗传算法的整个执行过程中保持一致
// 3.简单性: 一次读取数据并将其传递给遗传算法，简化了算法的实现。只需要处理一次数据加载和预处理，而不是在每次迭代中处理它
func GetSubjectsFromDB() []*Subject {

	subjects := []*Subject{
		// {SubjectID: 1, Name: "语文", SubjectGroupIDs: []int{1}},
		// {SubjectID: 2, Name: "数学", SubjectGroupIDs: []int{1}},
		// {SubjectID: 3, Name: "英语", SubjectGroupIDs: []int{1}},
		// {SubjectID: 4, Name: "音乐", SubjectGroupIDs: []int{3}},
		// {SubjectID: 5, Name: "美术", SubjectGroupIDs: []int{3}},
		// {SubjectID: 6, Name: "体育", SubjectGroupIDs: []int{3}},
		// {SubjectID: 7, Name: "物理", SubjectGroupIDs: []int{2}},
		// {SubjectID: 8, Name: "化学", SubjectGroupIDs: []int{2}},
		// {SubjectID: 9, Name: "政治", SubjectGroupIDs: []int{2}},
		// {SubjectID: 10, Name: "历史", SubjectGroupIDs: []int{3}},
		// {SubjectID: 11, Name: "生物", SubjectGroupIDs: []int{3}},
		// {SubjectID: 12, Name: "地理", SubjectGroupIDs: []int{3}},
		// {SubjectID: 13, Name: "劳技", SubjectGroupIDs: []int{3}},
		// {SubjectID: 14, Name: "活动", SubjectGroupIDs: []int{3}},
		// {SubjectID: 15, Name: "班会", SubjectGroupIDs: []int{3}},
	}
	return subjects
}

// 根据科目id查找科目对象
// subjectID 科目
// subjects 全部科目信息
func FindSubjectByID(subjectID int, subjects []*Subject) (*Subject, error) {

	for _, s := range subjects {
		if s.SubjectID == subjectID {
			return s, nil
		}
	}

	return nil, fmt.Errorf("subject not found")
}

// 根据科目组id查找科目
// subjectID 科目
// subjects 全部科目信息
func FindSubjectsByGroupID(groupID int, subjects []*Subject) ([]*Subject, error) {

	var subjectsByGroupID []*Subject
	for _, subject := range subjects {
		if lo.Contains(subject.SubjectGroupIDs, groupID) {
			subjectsByGroupID = append(subjectsByGroupID, subject)
		}
	}

	if len(subjectsByGroupID) > 0 {
		return subjectsByGroupID, nil
	}
	return nil, fmt.Errorf("no subjects found for group ID %d", groupID)
}

// ToMap 将Subject切片转换为map，key为SubjectID
func (s *Subject) ToMap(subjects []*Subject) map[int]*Subject {
	subjectMap := lo.KeyBy(subjects, func(subject *Subject) int {
		return subject.SubjectID
	})
	return subjectMap
}
