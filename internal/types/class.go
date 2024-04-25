package types

import (
	"course_scheduler/internal/constants"
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
func InitClasses() []Class {

	var classes []Class
	// subjects := models.GetSubjects()

	// 这里根据年级,班级,科目生成课班
	for i := 0; i < constants.NUM_GRADES; i++ {
		for j := 0; j < constants.NUM_CLASSES_PER_GRADE; j++ {
			for k := 0; k < constants.NUM_SUBJECTS; k++ {

				subjectID := k + 1
				gradeID := i + 1
				classID := j + 1

				class := Class{
					SubjectID: subjectID,
					GradeID:   gradeID,
					ClassID:   classID,
					SN:        SN{SubjectID: subjectID, GradeID: gradeID, ClassID: classID},
					// Name:      fmt.Sprintf("%d年级(%d)班 %s", gradeID, classID, subjects[k].Name),
				}
				classes = append(classes, class)
			}
		}
	}
	return classes
}

// 时间集合
// 基于教师集合和教师集合确定时间集合
// TODO: 如果教师有禁止时间,教室有禁止时间,这里是否需要处理? 如何处理?
// 每周5天上学,每天8节课, 一周40节课, {0, 1, 2, ... 39}
func ClassTimeSlots(teacherIDs []int, venueIDs []int) []int {

	var timeSlots []int
	for i := 0; i < constants.NUM_DAYS; i++ {
		for j := 0; j < constants.NUM_CLASSES; j++ {
			timeSlot := i*constants.NUM_CLASSES + j
			timeSlots = append(timeSlots, timeSlot)
		}
	}
	return timeSlots
}
