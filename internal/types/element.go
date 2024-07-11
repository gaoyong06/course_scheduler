package types

import (
	"course_scheduler/internal/models"
)

// 课班适应性矩阵中的一个元素
type Element struct {
	ClassSN   string // 科目_年级_班级
	SubjectID int    // 科目
	GradeID   int    // 年级
	ClassID   int    // 班级
	TeacherID int    // 教师
	VenueID   int    // 教室
	TimeSlot  int    // 时间段
	// IsConnected bool   // 是否是连堂课
	Val Val // 分数
}

func NewElement(classSN string, subjectID, gradeID, classID, teacherID, venueID int, timeSlot int) *Element {

	// isConnected := len(timeSlots) == 2
	return &Element{
		ClassSN:   classSN,
		SubjectID: subjectID,
		GradeID:   gradeID,
		ClassID:   classID,
		TeacherID: teacherID,
		VenueID:   venueID,
		TimeSlot:  timeSlot,
		// IsConnected: isConnected,
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

func (e *Element) GetTimeSlot() int {
	return e.TimeSlot
}

// 是否是连堂课
func (e *Element) IsConnected(classMatrix *ClassMatrix) bool {

	// 如果当前课节和下一节课的科目相同, 并且和上一节课科目不相同, 则当前是连堂课
	// 当前课
	currTimeSlot := e.TimeSlot
	// 上节课
	prevTimeSlot := e.TimeSlot - 1
	// 下节课
	nextTimeSlot := e.TimeSlot + 1

	prevExist := false
	currExist := false
	nextExist := false

	for sn, classMap := range classMatrix.Elements {
		if sn == e.ClassSN {
			for _, teacherMap := range classMap {
				for _, venueMap := range teacherMap {
					for timeSlot, e := range venueMap {
						if e.Val.Used == 1 {

							if prevTimeSlot == timeSlot {
								prevExist = true
							}

							if currTimeSlot == timeSlot {
								currExist = true
							}

							if nextTimeSlot == timeSlot {
								nextExist = true
							}
						}
					}
				}
			}
		}
	}

	if !prevExist && currExist && nextExist {
		return true
	}
	return false
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
func GetElementPeriod(element Element, schedule *models.Schedule) int {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()
	currTimeSlot := element.TimeSlot

	// 当前元素,课程所在的节次
	// var currPeriods []int
	// for _, currTimeSlot := range currTimeSlots {
	// 	currPeriod := currTimeSlot % totalClassesPerDay
	// 	currPeriods = append(currPeriods, currPeriod)
	// }
	currPeriod := currTimeSlot % totalClassesPerDay

	return currPeriod
}
