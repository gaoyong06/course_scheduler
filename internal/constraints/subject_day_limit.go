// 每天限制
package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
	"fmt"
)

// 每天限制，即纵向时间段限制
// 格式：对象+时间段+形式+次数
// 其中对象可选项为【科目，教师】，时间段可选项为【每天、星期一、星期二、星期三、星期四、星期五】，限制类型可选项为【固定、最少、最多】，次数为【0、1、2、···】；（+10）

// | 对象 | 时间段 | 限制类型 | 课节数 |
// | ---- | ------ | ---- | ---- |
// | 科目 | 每天   | 最多 | 4    |
// | 教师 | 星期一 | 固定 | 5    |
// | 科目 | 星期三 | 最少 | 2    |

type SubjectDayLimit struct {
	ID        int    `json:"id" mapstructure:"id"`                 // 自增ID
	GradeID   int    `json:"grade_id" mapstructure:"grade_id"`     // 年级ID
	ClassID   int    `json:"class_id" mapstructure:"class_id"`     // 班级ID, 可以为空
	Object    string `json:"object"  mapstructure:"object"`        // 对象，可选项为"科目"、"教师"
	SubjectID int    `json:"subject_id" mapstructure:"subject_id"` // 科目ID
	TeacherID int    `json:"teacher_id" mapstructure:"teacher_id"` // 教师ID
	Weekday   int    `json:"weekday" mapstructure:"weekday"`       // 周几，可选项为"0: 每天"、"1: 星期一"、"2: 星期二"、"3: 星期三"、"4: 星期四"、"5: 星期五"
	Type      string `json:"type"  mapstructure:"type"`            // 限制类型，可选项为"fixed: 固定"、"min: 最少"、"max: 最多"
	Count     int    `json:"count"  mapstructure:"count"`          // 课节数，可选项为0、1、2、···
}

// 生成字符串
func (s *SubjectDayLimit) String() string {
	return fmt.Sprintf("ID: %d, GradeID: %d, ClassID: %d, Object: %s, SubjectID: %d, TeacherID: %d, Weekday: %d, Type: %s, Count: %d", s.ID, s.GradeID, s.ClassID, s.Object, s.SubjectID, s.TeacherID, s.Weekday, s.Type, s.Count)
}

// 获取规则
func GetSubjectDayLimitRules(constraints []*SubjectDayLimit) []*types.Rule {
	// constraints := loadSubjectMutexConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule()
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (s *SubjectDayLimit) genRule() *types.Rule {
	fn := s.genConstraintFn()
	return &types.Rule{
		Name:     "subjectDayLimit",
		Type:     "dynamic",
		Fn:       fn,
		Score:    s.getScore(),
		Penalty:  s.getPenalty(),
		Weight:   1,
		Priority: 1,
	}
}

// 生成规则校验方法
func (s *SubjectDayLimit) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachingTask) (bool, bool, error) {

		totalClassesPerDay := schedule.GetTotalClassesPerDay()
		teacherID := element.GetTeacherID()
		gradeID := element.GradeID
		classID := element.ClassID
		subjectID := element.SubjectID
		timeSlots := element.GetTimeSlots()

		preCheckPassed := false
		isReward := false
		count := 0

		// 每天排课数量
		// key: 星期几, value: sn排课数量
		weekdayCountMap, err := s.countElementDayClasses(classMatrix, element, schedule)
		if err != nil {
			return false, false, err
		}

		// 这里使用第1个时间段
		weekday := timeSlots[0]/totalClassesPerDay + 1
		count = weekdayCountMap[weekday]

		// 对象是科目
		// 需年级(班级)信息,科目信息一致
		if s.Object == "subject" && (gradeID == s.GradeID && (classID == s.ClassID || s.ClassID == 0) && s.SubjectID == subjectID) && (s.Weekday == weekday || s.Weekday == 0) {
			preCheckPassed = true
		}

		// 对象是教师
		// 需教师信息一致
		if s.Object == "teacher" && s.TeacherID == teacherID && (s.Weekday == weekday || s.Weekday == 0) {
			preCheckPassed = true
		}

		// 固定次数
		if preCheckPassed && s.Type == "fixed" && count == s.Count {
			isReward = true
		}

		// 最多, 不够不奖励(奖励分为0), 超过就处罚(处罚分4)
		if preCheckPassed && s.Type == "max" && count < s.Count {

			isReward = true
		}

		// 最少, 不够就奖励
		if preCheckPassed && s.Type == "min" && count <= s.Count {
			isReward = true
		}

		return preCheckPassed, isReward, nil
	}
}

// 奖励分
func (s *SubjectDayLimit) getScore() int {
	switch s.Type {

	case "fixed":
		return 6

	case "min":
		return 6

	// 最多的奖励是0, 不鼓励
	case "max":
		return 0

	default:
		return 0
	}
}

// 惩罚分
func (s *SubjectDayLimit) getPenalty() int {
	switch s.Type {

	case "fixed":
		return 6

	case "min":
		return 6

	case "max":
		return 4

	default:
		return 0
	}
}

// countSubjectDayClasses 统计每天特定科目, 特定教师的排课数量
func (s *SubjectDayLimit) countElementDayClasses(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule) (map[int]int, error) {

	var weekdayCountMap map[int]int
	var err error
	if s.Object == "subject" {
		weekdayCountMap, err = s.countElementSubjectDayClasses(classMatrix, element, schedule)

	}

	if s.Object == "teacher" {
		weekdayCountMap, err = s.countElementTeacherDayClasses(classMatrix, element, schedule)
	}

	// if len(weekdayCountMap) > 0 {
	// 	log.Printf("subject day limit, element.TimeSlots: %v, object: %s, type: %s, (s.GradeID: %d, element.GradeID: %d), (s.ClassID: %d, element.ClassID: %d), (s.SubjectID: %d, element.SubjectID: %d), (s.TeacherID: %d, element.TeacherID: %d), weekdayCountMap: %#v\n ", element.TimeSlots, s.Object, s.Type, s.GradeID, element.GradeID, s.ClassID, element.ClassID, s.SubjectID, element.SubjectID, s.TeacherID, element.TeacherID, weekdayCountMap)
	// }
	return weekdayCountMap, err
}

// countSubjectDayClasses 统计每天特定科目的排课数量
func (s *SubjectDayLimit) countElementSubjectDayClasses(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule) (map[int]int, error) {

	// 如果type是subject则统计年级(班级)科目下的每天排课数量
	// 如果type是teacher则统计教师下面的每天的排课数量
	// 此时是假设当前元素会排课,所以需要将当前元素也计算在内
	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// key: 星期几, val: 数量
	weekdayCountMap := make(map[int]int)
	// 当前元素是周几
	eleWeekday := element.TimeSlots[0]/totalClassesPerDay + 1

	// 示例: 初一 语文 每天固定 1节课
	if s.Object == "subject" {

		// 约束条件与当前元素, 年级班级信息一致, 科目信息一致
		if element.GradeID == s.GradeID && (element.ClassID == s.ClassID || s.ClassID == 0) && element.SubjectID == s.SubjectID {

			// 将当前元素计入
			weekdayCountMap[eleWeekday]++

			// 统计矩阵内的其他元素
			for sn, classMap := range classMatrix.Elements {

				SN, err := types.ParseSN(sn)
				if err != nil {
					return nil, err
				}

				// 约束条件与当前节点的年级, 班级信息一致, 科目信息一致
				if SN.GradeID == s.GradeID && (SN.ClassID == s.ClassID || s.ClassID == 0) && SN.SubjectID == s.SubjectID {

					for _, teacherMap := range classMap {
						for _, venueMap := range teacherMap {
							for timeSlotStr, e := range venueMap {
								if e.Val.Used == 1 {

									eleTimeSlots := utils.ParseTimeSlotStr(timeSlotStr)
									// 这里把连堂课,也视为1节课
									weekday := eleTimeSlots[0]/totalClassesPerDay + 1
									// 星期几
									weekdayCountMap[weekday]++
								}
							}
						}
					}
				}
			}
		}
	}

	return weekdayCountMap, nil
}

// countTeacherDayClasses 统计每天特定教师的排课数量
func (s *SubjectDayLimit) countElementTeacherDayClasses(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule) (map[int]int, error) {

	// 如果type是subject则统计年级(班级)科目下的每天排课数量
	// 如果type是teacher则统计教师下面的每天的排课数量
	// 此时是假设当前元素会排课,所以需要将当前元素也计算在内
	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// key: 星期几, val: 数量
	weekdayCountMap := make(map[int]int)
	// 当前元素是周几
	eleWeekday := element.TimeSlots[0]/totalClassesPerDay + 1

	// 示例: : 王老师 每天固定 1节课
	if s.Object == "teacher" {
		// 约束条件与当前元素, 教师信息一致
		if element.TeacherID == s.TeacherID {
			// 将当前元素计入
			weekdayCountMap[eleWeekday]++
			// 统计矩阵内的其他元素
			for _, classMap := range classMatrix.Elements {
				for teacherID, teacherMap := range classMap {
					// 约束条件与当前的教师信息一致
					if teacherID == s.TeacherID {

						for _, venueMap := range teacherMap {
							for timeSlotStr, e := range venueMap {
								if e.Val.Used == 1 {

									eleTimeSlots := utils.ParseTimeSlotStr(timeSlotStr)
									// 这里把连堂课,也视为1节课
									weekday := eleTimeSlots[0]/totalClassesPerDay + 1
									// 星期几
									weekdayCountMap[weekday]++
								}
							}
						}

					}

				}
			}
		}
	}

	return weekdayCountMap, nil
}
