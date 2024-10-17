package types

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/utils"
	"fmt"

	"github.com/samber/lo"
)

// 课班
// 表示科目班级， 例如:美术一班, 如果增加年级, 例如: 美术三年级一班
type SubjectClass struct {
	SN        SN     // 序列号
	SubjectID int    // 科目id
	GradeID   int    // 年级id
	ClassID   int    // 班级id
	Name      string // 名称
	Priority  int    // 课班对应科目，排课的优先级, 优先级高的优先排课
}

func (c *SubjectClass) String() string {
	return fmt.Sprintf("%s (%d-%d-%d)", c.Name, c.SubjectID, c.GradeID, c.ClassID)
}

// 初始化课班
func InitSubjectClasses(teachingTasks []*models.TeachingTask, subjects []*models.Subject) ([]SubjectClass, error) {

	// 测试
	// fmt.Println("打印subjects START")
	// for _, subject := range subjects {
	// 	spew.Dump(subject)
	// }
	// fmt.Println("打印subjects END")
	// fmt.Println("\n")

	var subjectClasses []SubjectClass

	// 这里根据年级,班级,科目生成课班
	for _, task := range teachingTasks {

		subjectID := task.SubjectID
		gradeID := task.GradeID
		classID := task.ClassID

		subject, err := models.FindSubjectByID(subjectID, subjects)
		if err != nil {
			return nil, fmt.Errorf("error finding subject with ID %d: %v", subjectID, err)

		}

		class := SubjectClass{
			SubjectID: subjectID,
			GradeID:   gradeID,
			ClassID:   classID,
			SN:        SN{SubjectID: subjectID, GradeID: gradeID, ClassID: classID},
			Name:      fmt.Sprintf("gradeID: %d, classID: %d, subjectID: %d", gradeID, classID, subjectID),
			Priority:  subject.Priority,
		}
		subjectClasses = append(subjectClasses, class)
	}

	// 随机打乱课班排课顺序
	subjectClasses = shuffleSubjectClassesOrder(subjectClasses)
	return subjectClasses, nil
}

// 连堂课时间段集合
// 基于教师集合和教室集合确定时间集合
// TODO: 如果教师有禁止时间,教室有禁止时间,这里是否需要处理? 如何处理?
// 每周5天上学,每天8节课, 一周40节课, {0, 1, 2, ... 39}
//
// 2024.4.29 从总可用的时间段列表内,过滤掉教师禁止时间,教室禁止时间
// 如果多个老师,或者多个场地的禁止时间都不同,则返回类似map的结构体
// 根据前一个逻辑选择的教师,和教室,给定可选的时间段
func getConnectedTimeSlots(schedule *models.Schedule, teachingTasks []*models.TeachingTask, gradeID, classID, subjectID int, teacherIDs []int, venueIDs []int) []string {

	var timeSlotStrs []string

	// 课班(科目班级)每周周连堂课次数
	connectedCount := models.GetNumConnectedClassesPerWeek(gradeID, classID, subjectID, teachingTasks)
	if connectedCount > 0 {

		timeSlotStrs = utils.GetAllConnectedTimeSlots(schedule)
	}
	return timeSlotStrs
}

// 普通课时间集合
// 基于教师集合和教室集合确定时间集合
// TODO: 如果教师有禁止时间,教室有禁止时间,这里是否需要处理? 如何处理?
// 每周5天上学,每天8节课, 一周40节课, {0, 1, 2, ... 39}
//
// 2024.4.29 从总可用的时间段列表内,过滤掉教师禁止时间,教室禁止时间
// 如果多个老师,或者多个场地的禁止时间都不同,则返回类似map的结构体
// 根据前一个逻辑选择的教师,和教室,给定可选的时间段
func getNormalTimeSlots(schedule *models.Schedule, teachingTasks []*models.TeachingTask, gradeID, classID, subjectID int, teacherIDs []int, venueIDs []int) []string {

	var timeSlotStrs []string

	// 班级每周的排课时间段节数
	total := schedule.TotalClassesPerWeek()

	// 课班(科目班级)每周周连堂课次数
	connectedCount := models.GetNumConnectedClassesPerWeek(gradeID, classID, subjectID, teachingTasks)

	normalCount := total - connectedCount*2
	if normalCount > 0 {
		timeSlotStrs = utils.GetAllNormalTimeSlots(schedule)
	}
	return timeSlotStrs
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
// TODO: 下面的的方法,需要废弃掉

// 先按照优先级排序，再随机打乱课程顺序
func shuffleSubjectClassesOrder(subjectClasses []SubjectClass) []SubjectClass {

	// sort.Slice(subjectClasses, func(i, j int) bool {
	// 	return subjectClasses[i].Priority > subjectClasses[j].Priority
	// })

	// for i := len(subjectClasses) - 1; i > 0; i-- {
	// 	j := rand.Intn(i + 1)
	// 	// 只有当优先级相同时，才随机打乱顺序
	// 	if subjectClasses[i].Priority == subjectClasses[j].Priority {
	// 		subjectClasses[i], subjectClasses[j] = subjectClasses[j], subjectClasses[i]
	// 	}
	// }

	randomOrder := lo.Shuffle(subjectClasses)
	return randomOrder
}
