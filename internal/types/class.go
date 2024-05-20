package types

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/utils"
	"fmt"

	"github.com/spf13/cast"
)

// 课班
// 表示科目班级如:美术一班, 如果增加年级如: 美术三年级一班
type Class struct {
	SN        SN     // 序列号
	SubjectID int    // 科目id
	GradeID   int    // 年级id
	ClassID   int    // 班级id
	Name      string // 名称
	Priority  int    // 排课的优先级, 优先级高的优先排课
}

func (c *Class) String() string {
	return fmt.Sprintf("%s (%d-%d-%d)", c.Name, c.SubjectID, c.GradeID, c.ClassID)
}

// 初始化课班
func InitClasses(teachAllocs []*models.TeachTaskAllocation, subjects []*models.Subject) ([]Class, error) {

	// 测试
	// fmt.Println("打印subjects START")
	// for _, subject := range subjects {
	// 	spew.Dump(subject)
	// }
	// fmt.Println("打印subjects END")
	// fmt.Println("\n")

	var classes []Class

	// 这里根据年级,班级,科目生成课班
	for _, task := range teachAllocs {

		subjectID := task.SubjectID
		gradeID := task.GradeID
		classID := task.ClassID

		subject, err := models.FindSubjectByID(subjectID, subjects)
		if err != nil {
			return nil, fmt.Errorf("error finding subject with ID %d: %v", subjectID, err)

		}

		class := Class{
			SubjectID: subjectID,
			GradeID:   gradeID,
			ClassID:   classID,
			SN:        SN{SubjectID: subjectID, GradeID: gradeID, ClassID: classID},
			Name:      fmt.Sprintf("gradeID: %d, classID: %d, subjectID: %d", gradeID, classID, subjectID),
			Priority:  subject.Priority,
		}
		classes = append(classes, class)
	}
	return classes, nil
}

// 时间集合
// 基于教师集合和教室集合确定时间集合
// TODO: 如果教师有禁止时间,教室有禁止时间,这里是否需要处理? 如何处理?
// 每周5天上学,每天8节课, 一周40节课, {0, 1, 2, ... 39}
//
// 2024.4.29 从总可用的时间段列表内,过滤掉教师禁止时间,教室禁止时间
// 如果多个老师,或者多个场地的禁止时间都不同,则返回类似map的结构体
// 根据前一个逻辑选择的教师,和教室,给定可选的时间段
//
// 2024.5.20 将一周的课时，根据教学安排,分成多个组,每个组内的普通课课时和连堂课课时和教学计划相同,最后不足以分成一个组的,都按照普通课处理
func ClassTimeSlots(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, gradeID, classID, subjectID int, teacherIDs []int, venueIDs []int) ([]string, error) {

	var timeSlotStrs []string
	timeSlots := schedule.GenWeekTimeSlots()

	// 周课时
	numClassesPerWeek := models.GetNumClassesPerWeek(gradeID, classID, subjectID, taskAllocs)

	// 周连堂课次数
	numConnectedClassesPerWeek := models.GetNumConnectedClassesPerWeek(gradeID, classID, subjectID, taskAllocs)

	// 将时间段,根据连堂课的要求,将时间段分割为一个一组,或者两个一组,然后拼接为字符串
	// 按照连堂课的次数,和普通课的次数将可用时间段分成多个组,每个组内有都有按照教学计划的连堂课次数,和普通课次数
	// 因为单次课会多余连堂课, 不够分组的,多出的时间段都按照普通课处理
	groups := utils.SplitSlice(timeSlots, numClassesPerWeek)
	for _, group := range groups {

		if len(group) == numClassesPerWeek {
			pairs, err := utils.GroupByPairs(group, numConnectedClassesPerWeek)
			if err != nil {
				return nil, err
			}

			pairStrs, err := utils.GroupedIntsToString(pairs)
			if err != nil {
				return nil, err
			}
			timeSlotStrs = append(timeSlotStrs, pairStrs...)

		} else {

			for i := 0; i < len(group); i++ {
				groupStr := cast.ToString(group[i])
				timeSlotStrs = append(timeSlotStrs, groupStr)
			}
		}
	}

	return timeSlotStrs, nil
}
