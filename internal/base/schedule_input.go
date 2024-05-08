package base

import (
	"course_scheduler/internal/constraint"
	"course_scheduler/internal/models"
	"fmt"
	"sort"

	"github.com/spf13/viper"
)

// 排课输入信息
type ScheduleInput struct {
	Schedule             *models.Schedule              `json:"schedule" mapstructure:"schedule"`                             // 排课方案
	TeachTaskAllocations []*models.TeachTaskAllocation `json:"teach_task_allocations" mapstructure:"teach_task_allocations"` // 教学任务
	Teachers             []*models.Teacher             `json:"teachers" mapstructure:"teachers"`                             // 教师
	Subjects             []*models.Subject             `json:"subjects" mapstructure:"subjects"`                             // 科目
	// Venues                        []*models.Venue                  `json:"venues" mapstructure:"venues"`                                                             // 教学场地
	SubjectVenueMap               map[string][]int                 `json:"venue_map" mapstructure:"venue_map"`                                                       // 教学场地 key: sn(科目id_年级id_班级id) value: 教室id
	Grades                        []*models.Grade                  `json:"grades"`                                                                                   // 年级
	ClassConstraints              []*constraint.Class              `json:"class_constraints" mapstructure:"class_constraints"`                                       // 班级固排禁排约束
	SubjectMutexConstraints       []*constraint.SubjectMutex       `json:"subject_mutex_constraints" mapstructure:"subject_mutex_constraints"`                       // 科目互斥限制约束
	SubjectOrderConstraints       []*constraint.SubjectOrder       `json:"subject_order_constraints" mapstructure:"subject_order_constraints"`                       // 科目顺序限制约束
	SubjectConstraints            []*constraint.Subject            `json:"subject_constraints" mapstructure:"subject_constraints"`                                   // 科目优先排禁排约束
	TeacherMutexConstraints       []*constraint.TeacherMutex       `json:"teacher_mutex_constraints" mapstructure:"teacher_mutex_constraints"`                       // 教师互斥限制约束
	TeacherNoonBreakConstraints   []*constraint.TeacherNoonBreak   `json:"teacher_noon_break_constraints" mapstructure:"teacher_noon_break_constraints"`             // 教师不跨中午约束
	TeacherPeriodLimitConstraints []*constraint.TeacherPeriodLimit `json:"teacher_period_constraints" mapstructure:"teacher_period_constraints"`                     // 教师节数限制
	TeacherRangeLimitConstraints  []*constraint.TeacherRangeLimit  `json:"teacher_connected_lesson_constraints" mapstructure:"teacher_connected_lesson_constraints"` // 教师时间段限制
	TeacherConstraints            []*constraint.Teacher            `json:"teacher_constraints" mapstructure:"teacher_constraints"`                                   // 教师固排禁排约束
}

// 检查教学计划是否正确
func (s *ScheduleInput) CheckTeachTaskAllocation() (bool, error) {

	// 1. 检查每周总课时数是否超过总课时数
	totalClassesPerWeek := s.Schedule.GetTotalClassesPerDay() * s.Schedule.NumWorkdays
	count := 0
	// key: 科目ID, value: 每周上课总次数
	subjectClasses := make(map[int]int)
	for _, task := range s.TeachTaskAllocations {
		count += task.NumClassesPerWeek
		// 上课次数 = 科目每周课时 - 每周连堂课课时
		subjectClasses[task.SubjectID] += task.NumClassesPerWeek - task.NumConnectedClassesPerWeek
	}
	if count > totalClassesPerWeek {
		return false, fmt.Errorf("total course Classes %d exceed maximum weekly Classes %d", count, totalClassesPerWeek)
	}

	// 2. 检查每个科目每周上课总次数是否正确
	// 每周工作5天,所以最多上5次课
	for subjectID, time := range subjectClasses {
		if time > s.Schedule.NumWorkdays {
			return false, fmt.Errorf("subject %d has invalid weekly Classes count", subjectID)
		}
	}
	return true, nil
}

// 加载yaml测试数据
func LoadTestData() (*ScheduleInput, error) {
	var config ScheduleInput

	// 设置配置文件名和类型
	viper.SetConfigType("yaml")
	viper.SetConfigName("testdata")

	// 添加配置文件搜索路径
	viper.AddConfigPath("../testdata")

	// 为 viper 添加自定义解析函数
	viper.SetConfigType("yaml")

	// 读取并解析配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fatal error loading testdata: %s", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("fatal error unmarshaling testdata: %s", err)
	}

	// 对 Courses 属性的值按照 NumClassesPerWeek 排序
	sort.Slice(config.TeachTaskAllocations, func(i, j int) bool {
		return config.TeachTaskAllocations[i].NumClassesPerWeek > config.TeachTaskAllocations[j].NumClassesPerWeek
	})

	return &config, nil
}
