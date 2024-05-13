package types

// 课班适应性矩阵中的一个元素
type Element struct {
	ClassSN   string // 科目_年级_班级
	SubjectID int    // 科目
	GradeID   int    // 年级
	ClassID   int    // 班级
	TeacherID int    // 教师
	VenueID   int    // 教室
	TimeSlot  int    // 时间段
	Val       Val    // 分数
}

func NewElement(classSN string, subjectID, gradeID, classID, teacherID, venueID, timeSlot int) *Element {
	return &Element{
		ClassSN:   classSN,
		SubjectID: subjectID,
		GradeID:   gradeID,
		ClassID:   classID,
		TeacherID: teacherID,
		VenueID:   venueID,
		TimeSlot:  timeSlot,
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
