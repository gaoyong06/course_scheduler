package models

import (
	"fmt"

	"github.com/samber/lo"
)

type Teacher struct {
	TeacherID       int            `json:"teacher_id" mapstructure:"teacher_id"`               // 教室id
	Name            string         `json:"name" mapstructure:"name"`                           // 教师姓名
	TeacherGroupIDs []int          `json:"teacher_group_ids" mapstructure:"teacher_group_ids"` // 教师分组id, 一个老师会在多个分组中
	ClassSubjects   []ClassSubject `json:"class_subjects" mapstructure:"class_subjects"`       // 老师教授科目信息
}

func GetTeachersFromDB() []*Teacher {

	teachers := []*Teacher{
		// {TeacherID: 1, Name: "王老师", TeacherGroupIDs: []int{1, 3}, SubjectIDs: []int{1}},
		// {TeacherID: 2, Name: "李老师", TeacherGroupIDs: []int{2}, SubjectIDs: []int{2}},
		// {TeacherID: 3, Name: "刘老师", TeacherGroupIDs: []int{3}, SubjectIDs: []int{3}},
		// {TeacherID: 4, Name: "张老师", TeacherGroupIDs: []int{4}, SubjectIDs: []int{4}},
		// {TeacherID: 5, Name: "马老师", TeacherGroupIDs: []int{5}, SubjectIDs: []int{5}},
		// {TeacherID: 6, Name: "黄老师", TeacherGroupIDs: []int{6}, SubjectIDs: []int{6}},
		// {TeacherID: 7, Name: "常远", TeacherGroupIDs: []int{7}, SubjectIDs: []int{7}},
		// {TeacherID: 8, Name: "王成思", TeacherGroupIDs: []int{8}, SubjectIDs: []int{8}},
		// {TeacherID: 9, Name: "许文赫", TeacherGroupIDs: []int{9}, SubjectIDs: []int{9}},
		// {TeacherID: 10, Name: "高飞", TeacherGroupIDs: []int{10}, SubjectIDs: []int{10}},
		// {TeacherID: 11, Name: "谢娜", TeacherGroupIDs: []int{11}, SubjectIDs: []int{11}},
		// {TeacherID: 12, Name: "黄凯", TeacherGroupIDs: []int{12}, SubjectIDs: []int{12}},
		// {TeacherID: 13, Name: "孟都", TeacherGroupIDs: []int{13}, SubjectIDs: []int{13}},
		// {TeacherID: 14, Name: "刘非", TeacherGroupIDs: []int{14}, SubjectIDs: []int{14}},
		// {TeacherID: 15, Name: "王麻子", TeacherGroupIDs: []int{15}, SubjectIDs: []int{15}},
	}
	return teachers
}

// ToMap 将Teacher切片转换为map，key为TeacherID
func (t *Teacher) ToMap(teachers []*Teacher) map[int]*Teacher {
	teacherMap := lo.KeyBy(teachers, func(teacher *Teacher) int {
		return teacher.TeacherID
	})
	return teacherMap
}

// 根据教师id查找教师对象
func FindTeacherByID(teacherID int, teachers []*Teacher) (*Teacher, error) {

	for _, teacher := range teachers {
		if teacher.TeacherID == teacherID {
			return teacher, nil
		}
	}
	return nil, fmt.Errorf("teacher not found")
}

// 老师集合
// 根据课班选取老师
func ClassTeacherIDs(gradeID, classID, subjectID int, teachers []*Teacher) []int {

	var teacherIDs []int

	// 根据课班选取老师
	for _, teacher := range teachers {
		for _, classSubject := range teacher.ClassSubjects {
			if classSubject.GradeID == gradeID && classSubject.ClassID == classID && lo.Contains(classSubject.SubjectIDs, subjectID) {
				teacherIDs = append(teacherIDs, teacher.TeacherID)
			}
		}
	}
	return teacherIDs
}

// 判断teacherID是否合法
func IsTeacherIDValid(teacherID int) bool {
	return teacherID > 0
}

// 老师不可排课的时间范围
func teacherUnavailableSlots() []int {
	var timeSlots []int
	return timeSlots
}
