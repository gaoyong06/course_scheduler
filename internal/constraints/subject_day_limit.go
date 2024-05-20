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
	Object    string `json:"object"  mapstructure:"object"`        // 对象，可选项为"科目"、"教师"
	SubjectID int    `json:"subject_id" mapstructure:"subject_id"` // 科目ID
	TeacherID int    `json:"teacher_id" mapstructure:"teacher_id"` // 教师ID
	Weekday   int    `json:"weekday" mapstructure:"weekday"`       // 周几，可选项为"0: 每天"、"1: 星期一"、"2: 星期二"、"3: 星期三"、"4: 星期四"、"5: 星期五"
	Type      string `json:"type"  mapstructure:"type"`            // 限制类型，可选项为"fixed: 固定"、"min: 最少"、"max: 最多"
	Count     int    `json:"count"  mapstructure:"count"`          // 课节数，可选项为0、1、2、···
}

// 生成字符串
func (s *SubjectDayLimit) String() string {
	return fmt.Sprintf("ID: %d, Object: %s, Weekday: %d, Type: %s, Count: %d", s.ID, s.Object, s.Weekday, s.Type, s.Count)
}

// 获取班级固排禁排规则
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
		Score:    s.getPoints(),
		Penalty:  s.getPoints(),
		Weight:   1,
		Priority: 1,
	}
}

// 生成规则校验方法
func (s *SubjectDayLimit) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

		totalClassesPerDay := schedule.GetTotalClassesPerDay()
		classSN := element.GetClassSN()
		teacherID := element.GetTeacherID()
		venueID := element.GetVenueID()
		subjectID := element.SubjectID
		timeSlots := element.GetTimeSlots()

		preCheckPassed := false
		isReward := false
		count := 0

		weekdayCountMap := countDayClasses(classMatrix, classSN, teacherID, venueID, schedule)

		// 这里使用第1个时间段
		weekday := timeSlots[0]/totalClassesPerDay + 1
		count = weekdayCountMap[weekday]

		if (s.Weekday == weekday || s.Weekday == 0) && (s.TeacherID == teacherID || s.SubjectID == subjectID) {
			preCheckPassed = true
		}

		// 固定次数
		if s.Type == "fixed" && count == s.Count {
			isReward = true
		}

		// 最多
		if s.Type == "max" && count <= s.Count {
			isReward = true
		}

		// 最少
		if s.Type == "min" && count >= s.Count {
			isReward = true
		}

		return preCheckPassed, isReward, nil
	}
}

// 奖励分,惩罚分
func (s *SubjectDayLimit) getPoints() int {
	switch s.Type {

	case "fixed":
		return 6

	case "min":
		return 3

	case "max":
		return 1

	default:
		return 0
	}
}

// countDayClasses 计算每天的科目数量
func countDayClasses(classMatrix *types.ClassMatrix, sn string, teacherID, venueID int, schedule *models.Schedule) map[int]int {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// key: 星期几, val: 数量
	weekdayCountMap := make(map[int]int)
	for timeSlotStr, element := range classMatrix.Elements[sn][teacherID][venueID] {

		timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
		for _, timeSlot := range timeSlots {
			if element.Val.Used == 1 {
				weekday := timeSlot/totalClassesPerDay + 1
				// 星期几
				weekdayCountMap[weekday]++
			}
		}
	}
	return weekdayCountMap
}
