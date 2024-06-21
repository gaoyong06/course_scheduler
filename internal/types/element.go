package types

import "course_scheduler/internal/models"

// 课班适应性矩阵中的一个元素
type Element struct {
	ClassSN     string // 科目_年级_班级
	SubjectID   int    // 科目
	GradeID     int    // 年级
	ClassID     int    // 班级
	TeacherID   int    // 教师
	VenueID     int    // 教室
	TimeSlots   []int  // 连堂课: 时间段1,时间段2, 普通课：时间段1
	IsConnected bool   // 是否是连堂课
	Val         Val    // 分数
}

func NewElement(classSN string, subjectID, gradeID, classID, teacherID, venueID int, timeSlots []int) *Element {

	isConnected := len(timeSlots) == 2
	return &Element{
		ClassSN:     classSN,
		SubjectID:   subjectID,
		GradeID:     gradeID,
		ClassID:     classID,
		TeacherID:   teacherID,
		VenueID:     venueID,
		TimeSlots:   timeSlots,
		IsConnected: isConnected,
		Val: Val{
			ScoreInfo: ScoreInfo{
				Score:          0,
				FixedScore:     0,
				DynamicScore:   0,
				FixedPassed:    []string{},
				FixedFailed:    []string{},
				FixedSkipped:   []string{},
				DynamicPassed:  []string{},
				DynamicFailed:  []string{},
				DynamicSkipped: []string{},
			},
			Used: 0,
		},
	}
}

func (e *Element) GetClassSN() string {
	return e.ClassSN
}

func (e *Element) GetTeacherID() int {
	return e.TeacherID
}

func (e *Element) GetVenueID() int {
	return e.VenueID
}

func (e *Element) GetTimeSlots() []int {
	return e.TimeSlots
}

func (e *Element) GetPassedConstraints() []string {

	fixedPassed := e.Val.ScoreInfo.FixedPassed
	dynamicPassed := e.Val.ScoreInfo.DynamicPassed

	passedConstraints := append(fixedPassed, dynamicPassed...)
	return passedConstraints
}

func (e *Element) GetFailedConstraints() []string {

	fixedFailed := e.Val.ScoreInfo.FixedFailed
	dynamicFailed := e.Val.ScoreInfo.DynamicFailed

	failedConstraints := append(fixedFailed, dynamicFailed...)
	return failedConstraints

}

func (e *Element) GetSkippedConstraints() []string {

	fixedSkipped := e.Val.ScoreInfo.FixedSkipped
	dynamicSkipped := e.Val.ScoreInfo.DynamicSkipped

	skippedConstraints := append(fixedSkipped, dynamicSkipped...)
	return skippedConstraints
}

// 当前元素排课信息的节次
func GetElementPeriods(element Element, schedule *models.Schedule) []int {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	currTimeSlots := element.TimeSlots

	// 当前元素,课程所在的节次
	var currPeriods []int
	for _, currTimeSlot := range currTimeSlots {
		currPeriod := currTimeSlot % totalClassesPerDay
		currPeriods = append(currPeriods, currPeriod)
	}

	return currPeriods
}
