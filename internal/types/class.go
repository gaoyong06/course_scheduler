package types

import (
	"course_scheduler/internal/models"
	"fmt"
)

// 课班
// 表示科目班级如:美术一班, 如果增加年级如: 美术三年级一班
type Class struct {
	SN        SN     // 序列号
	SubjectID int    // 科目id
	GradeID   int    // 年级id
	ClassID   int    // 班级id
	Name      string // 名称
}

func (c *Class) String() string {
	return fmt.Sprintf("%s (%d-%d-%d)", c.Name, c.SubjectID, c.GradeID, c.ClassID)
}

// 初始化课班
func InitClasses(teachAllocs []*models.TeachTaskAllocation) []Class {

	var classes []Class

	// 这里根据年级,班级,科目生成课班
	for _, task := range teachAllocs {

		subjectID := task.SubjectID
		gradeID := task.GradeID
		classID := task.ClassID

		class := Class{
			SubjectID: subjectID,
			GradeID:   gradeID,
			ClassID:   classID,
			SN:        SN{SubjectID: subjectID, GradeID: gradeID, ClassID: classID},
			Name:      fmt.Sprintf("gradeID: %d, classID: %d, subjectID: %d", gradeID, classID, subjectID),
		}
		classes = append(classes, class)
	}
	return classes
}

// 时间集合
// 基于教师集合和教师集合确定时间集合
// TODO: 如果教师有禁止时间,教室有禁止时间,这里是否需要处理? 如何处理?
// 每周5天上学,每天8节课, 一周40节课, {0, 1, 2, ... 39}
//
// 2024.4.29 从总可用的时间段列表内,过滤掉教师禁止时间,教室禁止时间
// 如果多个老师,或者多个场地的禁止时间都不同,则返回类似map的结构体
// 根据前一个逻辑选择的教师,和教室,给定可选的时间段
func ClassTimeSlots(schedule *models.Schedule, teacherIDs []int, venueIDs []int) []int {

	var timeSlots []int
	totalClassesPerDay := schedule.GetTotalClassesPerDay()

	for i := 0; i < schedule.NumWorkdays; i++ {
		for j := 0; j < totalClassesPerDay; j++ {
			timeSlot := i*totalClassesPerDay + j
			timeSlots = append(timeSlots, timeSlot)
		}
	}
	return timeSlots
}
