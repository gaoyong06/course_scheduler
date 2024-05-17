package models

// 课务安排课程
// 每周的教学任务
// x老师教y班,每周z节课
type TeachTaskAllocation struct {
	ID                         int    `json:"id" mapstructure:"id"`                                                             // 唯一id
	GradeID                    int    `json:"grade_id" mapstructure:"grade_id"`                                                 // 年级id
	ClassID                    int    `json:"class_id" mapstructure:"class_id"`                                                 // 班级id
	SubjectID                  int    `json:"subject_id" mapstructure:"subject_id"`                                             // 科目id
	TeacherID                  int    `json:"teacher_id" mapstructure:"teacher_id"`                                             // 教师id
	NumClassesPerWeek          int    `json:"num_classes_per_week" mapstructure:"num_classes_per_week"`                         // 每周几节课
	NumConnectedClassesPerWeek int    `json:"num_connected_classes_per_week" mapstructure:"num_connected_classes_per_week"`     // 每周几次连堂课 1连堂课=2节课
	WeekType                   string `json:"week_type,omitempty" mapstructure:"week_type,omitempty"`                           // 单双周类型: single 表示单周，double 表示双周, 默认为空, 不做设置
	SubjectIDForWeek           int    `json:"subject_id_for_week,omitempty" mapstructure:"subject_id_for_week,omitempty"`       // 单双周轮换科目
	SubjectIDOnDiffDay         int    `json:"subject_id_on_diff_day,omitempty" mapstructure:"subject_id_on_diff_day,omitempty"` // 不同天上课科目id TODO: 这个和科目互斥有重复?
	CourseType                 string `json:"course_type" mapstructure:"course_type"`                                           // 课程类型: class_specific 表示班级特殊课, grade_shared 表示年级统一课
}

// NewCourse 创建一个新的课程
// 根据参数创建一个新的课程对象
func NewTeachTaskAllocation(id, gradeID, classID, subjectID, teacherID, numClassesPerWeek, numConnectedClassesPerWeek int, weekType string, subjectIDForWeek, subjectIDOnDiffDay int, courseType string) *TeachTaskAllocation {
	course := &TeachTaskAllocation{
		ID:                         id,
		GradeID:                    gradeID,
		ClassID:                    classID,
		SubjectID:                  subjectID,
		TeacherID:                  teacherID,
		NumClassesPerWeek:          numClassesPerWeek,
		NumConnectedClassesPerWeek: numConnectedClassesPerWeek,
		WeekType:                   weekType,
		SubjectIDForWeek:           subjectIDForWeek,
		SubjectIDOnDiffDay:         subjectIDOnDiffDay,
		CourseType:                 courseType,
	}
	return course
}

// 获取一个科目的周课时
func GetNumClassesPerWeek(gradeID, classID, subjectID int, teachAllocs []*TeachTaskAllocation) int {

	count := 0
	for _, task := range teachAllocs {

		if task.GradeID == gradeID && task.ClassID == classID && task.SubjectID == subjectID {
			count = task.NumClassesPerWeek
			break
		}
	}
	return count
}

// 获取一个年级,一个班级，一个科目的所有老师
func GetTeacherIDs(gradeID, classID, subjectID int, teachAllocs []*TeachTaskAllocation) []int {

	var teacherIDs []int
	for _, task := range teachAllocs {

		if task.GradeID == gradeID && task.ClassID == classID && task.SubjectID == subjectID {
			teacherIDs = append(teacherIDs, task.TeacherID)
		}
	}
	return teacherIDs
}
