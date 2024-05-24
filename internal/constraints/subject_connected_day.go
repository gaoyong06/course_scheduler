package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
	"fmt"
	"log"
)

// 连堂课各天次数限制
type SubjectConnectedDay struct {
	ID        int `json:"id" mapstructure:"id"`                 // 自增ID
	GradeID   int `json:"grade_id" mapstructure:"grade_id"`     // 年级ID
	ClassID   int `json:"class_id" mapstructure:"class_id"`     // 班级ID, 可以为空
	SubjectID int `json:"subject_id" mapstructure:"subject_id"` // 科目ID
	TeacherID int `json:"teacher_id" mapstructure:"teacher_id"` // 教师ID
	Weekday   int `json:"weekday" mapstructure:"weekday"`       // 周几，可选项为"0: 每天"、"1: 星期一"、"2: 星期二"、"3: 星期三"、"4: 星期四"、"5: 星期五"
	Count     int `json:"count"  mapstructure:"count"`          // 连堂课次数
}

// 生成字符串
func (s *SubjectConnectedDay) String() string {
	return fmt.Sprintf("ID: %d, GradeID: %d, ClassID: %d, SubjectID: %d, TeacherID: %d,  Weekday: %d, Count: %d", s.ID, s.GradeID, s.ClassID, s.SubjectID, s.TeacherID, s.Weekday, s.Count)
}

// 获取班级固排禁排规则
func GetSubjectConnectedDayRules(constraints []*SubjectConnectedDay) []*types.Rule {
	// constraints := loadSubjectMutexConstraintsFromDB()
	var rules []*types.Rule
	for _, c := range constraints {
		rule := c.genRule()
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (s *SubjectConnectedDay) genRule() *types.Rule {
	fn := s.genConstraintFn()
	return &types.Rule{
		Name:     "subjectConnectedDay",
		Type:     "fixed",
		Fn:       fn,
		Score:    s.getPoints(),
		Penalty:  s.getPoints(),
		Weight:   1,
		Priority: 1,
	}
}

// 生成规则校验方法
func (s *SubjectConnectedDay) genConstraintFn() types.ConstraintFn {

	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, taskAllocs []*models.TeachTaskAllocation) (bool, bool, error) {

		totalClassesPerDay := schedule.GetTotalClassesPerDay()
		classSN := element.GetClassSN()
		teacherID := element.GetTeacherID()
		venueID := element.GetVenueID()
		gradeID := element.GradeID
		classID := element.ClassID
		subjectID := element.SubjectID
		timeSlots := element.GetTimeSlots()

		preCheckPassed := false
		isReward := false
		count := 0
		var weekdayConnectedCountMap map[int]int

		// 这里使用第1个时间段
		weekday := timeSlots[0]/totalClassesPerDay + 1
		// count = weekdayConnectedCountMap[weekday]

		if element.IsConnected && gradeID == s.GradeID && (classID == s.ClassID || s.ClassID == 0) && (weekday == s.Weekday || s.Weekday == 0) && (teacherID == s.TeacherID || subjectID == s.SubjectID) {

			preCheckPassed = true

			weekdayConnectedCountMap = countDayConnectedClasses(classMatrix, classSN, teacherID, venueID, schedule)
			count = weekdayConnectedCountMap[weekday]
		}

		// 固定次数
		if preCheckPassed && count < s.Count {
			isReward = true
		}

		if element.IsConnected {
			log.Printf("subject connected day constraint, sn: %s, element.TimeSlots: %#v, (gradeID: %d, s.GradeID: %d), (classID: %d, s.ClassID: %d), (weekday: %d, s.Weekday: %d),(teacherID: %d, s.TeacherID: %d), (subjectID: %d, s.SubjectID: %d), count: %d, preCheckPassed: %#v, isReward: %#v \n", classSN, element.TimeSlots, gradeID, s.GradeID, classID, s.ClassID, weekday, s.Weekday, teacherID, s.TeacherID, subjectID, s.SubjectID, count, preCheckPassed, isReward)
		}
		return preCheckPassed, isReward, nil
	}
}

// 奖励分,惩罚分
func (s *SubjectConnectedDay) getPoints() int {

	return 6
}

// countDayClasses 计算每天的科目数量
func countDayConnectedClasses(classMatrix *types.ClassMatrix, sn string, teacherID, venueID int, schedule *models.Schedule) map[int]int {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	// key: 星期几, val: 连堂课数量
	weekdayConnectedCountMap := make(map[int]int)
	for timeSlotStr, element := range classMatrix.Elements[sn][teacherID][venueID] {

		timeSlots := utils.ParseTimeSlotStr(timeSlotStr)
		if element.Val.Used == 1 && element.IsConnected {
			weekday := timeSlots[0]/totalClassesPerDay + 1
			// 星期几
			weekdayConnectedCountMap[weekday]++
		}

	}
	return weekdayConnectedCountMap
}
