package types

import (
	"course_scheduler/internal/models"
	"fmt"
	"log"
	"math/rand"
	"sort"

	"github.com/samber/lo"
)

// 课班
// 表示科目班级如:美术一班, 如果增加年级如: 美术三年级一班
type SubjectClass struct {
	SN        SN     // 序列号
	SubjectID int    // 科目id
	GradeID   int    // 年级id
	ClassID   int    // 班级id
	Name      string // 名称
	Priority  int    // 排课的优先级, 优先级高的优先排课
}

func (c *SubjectClass) String() string {
	return fmt.Sprintf("%s (%d-%d-%d)", c.Name, c.SubjectID, c.GradeID, c.ClassID)
}

// 初始化课班
func InitSubjectClasses(teachAllocs []*models.TeachTaskAllocation, subjects []*models.Subject) ([]SubjectClass, error) {

	// 测试
	// fmt.Println("打印subjects START")
	// for _, subject := range subjects {
	// 	spew.Dump(subject)
	// }
	// fmt.Println("打印subjects END")
	// fmt.Println("\n")

	var subjectClasses []SubjectClass

	// 这里根据年级,班级,科目生成课班
	for _, task := range teachAllocs {

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

	shuffleSubjectClassesOrder(subjectClasses)

	return subjectClasses, nil
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
func SubjectClassTimeSlots(schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation, gradeID, classID, subjectID int, teacherIDs []int, venueIDs []int) ([]string, error) {

	var timeSlotStrs []string
	timeSlots := schedule.GenWeekTimeSlots()
	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	waitAllocCount := len(timeSlots)

	// 周课时
	numClassesPerWeek := models.GetNumClassesPerWeek(gradeID, classID, subjectID, taskAllocs)

	// 周连堂课次数
	numConnectedClassesPerWeek := models.GetNumConnectedClassesPerWeek(gradeID, classID, subjectID, taskAllocs)

	// 连堂课次数
	waitAllocConnectedCount := numConnectedClassesPerWeek

	// 普通课次数
	waitAllocNormalCount := numClassesPerWeek - (numConnectedClassesPerWeek * 2)

	// 将timeSlots随机打乱
	timeSlots = lo.Shuffle(timeSlots)
	// timeSlot是否占用
	timeSlotsMap := make(map[int]bool)

	// 将时间段数组中的各个项作为键，值都为 false 地添加到 timeSlotsMap 中
	for _, timeSlot := range timeSlots {
		timeSlotsMap[timeSlot] = false
	}

	for waitAllocCount > 0 {

		// 按照普通课和连堂课的排课数量要求, 将timeSlots分成多个组, 有的组内有1个元素表示普通课, 有的组内有2个元素表示连堂课
		for timeSlot, used := range timeSlotsMap {

			if !used {
				// 检查连堂课是否排完
				// 如果连堂课未排完, 检查timeSlot+1是否是false, 检查timeSlot, 和timeSlot+1, 是否在同一天, 并且都同在上午, 或者同在下午
				if waitAllocConnectedCount > 0 && !timeSlotsMap[timeSlot+1] {

					day1 := timeSlot / totalClassesPerDay
					day2 := (timeSlot + 1) / totalClassesPerDay

					period1 := timeSlot % totalClassesPerDay
					period2 := (timeSlot + 1) % totalClassesPerDay

					// 上午
					amStartPeriod, amEndPeriod := schedule.GetPeriodWithRange("forenoon")
					// 下午
					pmStartPeriod, pmEndPeriod := schedule.GetPeriodWithRange("afternoon")

					if day1 == day2 && ((period1 >= amStartPeriod && period1 <= amEndPeriod && period2 >= amStartPeriod && period2 <= amEndPeriod) ||
						(period1 >= pmStartPeriod && period1 <= pmEndPeriod && period2 >= pmStartPeriod && period2 <= pmEndPeriod)) {

						// 如果是, 则排一次连堂课
						timeSlotStrs = append(timeSlotStrs, fmt.Sprintf("%d_%d", timeSlot, timeSlot+1))
						timeSlotsMap[timeSlot] = true
						timeSlotsMap[timeSlot+1] = true
						waitAllocConnectedCount--
						waitAllocCount -= 2
						continue
					}
				}

				// 如果不是, 则检查普通课，是否排完
				if waitAllocNormalCount > 0 || (waitAllocNormalCount == 0 && waitAllocConnectedCount > 0) {

					// 如果普通课未排完, 则排一次普通课
					timeSlotStrs = append(timeSlotStrs, fmt.Sprintf("%d", timeSlot))
					timeSlotsMap[timeSlot] = true
					if waitAllocNormalCount > 0 {
						waitAllocNormalCount--
					}
					waitAllocCount--
					continue
				}

				// 如果普通课已经排完,则继续检查下一个timeSlot
				// 直至连堂课和普通课都排完, 则继续进入下一次迭代
				if waitAllocNormalCount == 0 && waitAllocConnectedCount == 0 {
					// 恢复连堂课和普通课的排课数量
					waitAllocConnectedCount = numConnectedClassesPerWeek
					waitAllocNormalCount = numClassesPerWeek - (numConnectedClassesPerWeek * 2)
				}

				// 如果最后,出现普通课未排完, 或者连课堂未排完,但是还有timeSlotsMap[timeSlot]是false,则将该timeSlotsMap[timeSlot]排为普通课
			}
		}
	}

	return timeSlotStrs, nil
}

// 先按照优先级排序，再随机打乱课程顺序
func shuffleSubjectClassesOrder(subjectClasses []SubjectClass) {

	sort.Slice(subjectClasses, func(i, j int) bool {
		return subjectClasses[i].Priority > subjectClasses[j].Priority
	})
	for i := len(subjectClasses) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		// 只有当优先级相同时，才随机打乱顺序
		if subjectClasses[i].Priority == subjectClasses[j].Priority {
			subjectClasses[i], subjectClasses[j] = subjectClasses[j], subjectClasses[i]
		}
	}
	log.Println("Class order shuffled")
}
