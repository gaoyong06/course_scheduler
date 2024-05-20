// schedule_input.go
package base

import (
	"course_scheduler/internal/constraints"
	"course_scheduler/internal/models"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/viper"
)

// 排课输入信息
type ScheduleInput struct {
	Schedule             *models.Schedule              `json:"schedule" mapstructure:"schedule"`                             // 排课方案
	TeachTaskAllocations []*models.TeachTaskAllocation `json:"teach_task_allocations" mapstructure:"teach_task_allocations"` // 教学任务
	Teachers             []*models.Teacher             `json:"teachers" mapstructure:"teachers"`                             // 教师信息
	Subjects             []*models.Subject             `json:"subjects" mapstructure:"subjects"`                             // 科目信息
	// Venues                        []*models.Venue                  `json:"venues" mapstructure:"venues"`                                                             // 教学场地
	SubjectVenueMap               map[string][]int                  `json:"subject_venue_map" mapstructure:"subject_venue_map"`                               // 教学场地 key: sn(科目id_年级id_班级id) value: 教室id
	Grades                        []*models.Grade                   `json:"grades"`                                                                           // 年级信息
	ClassConstraints              []*constraints.Class              `json:"class_constraints" mapstructure:"class_constraints"`                               // 班级固排禁排约束条件
	SubjectMutexConstraints       []*constraints.SubjectMutex       `json:"subject_mutex_constraints" mapstructure:"subject_mutex_constraints"`               // 科目互斥限制约束条件
	SubjectOrderConstraints       []*constraints.SubjectOrder       `json:"subject_order_constraints" mapstructure:"subject_order_constraints"`               // 科目顺序限制约束条件
	SubjectDayLimitConstraints    []*constraints.SubjectDayLimit    `json:"subject_day_limit_constraints" mapstructure:"subject_day_limit_constraints"`       // 科目顺序限制约束条件
	SubjectConstraints            []*constraints.Subject            `json:"subject_constraints" mapstructure:"subject_constraints"`                           // 科目优先排禁排约束条件
	TeacherMutexConstraints       []*constraints.TeacherMutex       `json:"teacher_mutex_constraints" mapstructure:"teacher_mutex_constraints"`               // 教师互斥限制约束条件
	TeacherNoonBreakConstraints   []*constraints.TeacherNoonBreak   `json:"teacher_noon_break_constraints" mapstructure:"teacher_noon_break_constraints"`     // 教师不跨中午约束条件
	TeacherPeriodLimitConstraints []*constraints.TeacherPeriodLimit `json:"teacher_period_limit_constraints" mapstructure:"teacher_period_limit_constraints"` // 教师节数限制条件
	TeacherRangeLimitConstraints  []*constraints.TeacherRangeLimit  `json:"teacher_range_limit_constraints" mapstructure:"teacher_range_limit_constraints"`   // 教师时间段限制条件
	TeacherConstraints            []*constraints.Teacher            `json:"teacher_constraints" mapstructure:"teacher_constraints"`                           // 教师固排禁排约束条件
}

// 检查教学计划是否正确
func (s *ScheduleInput) CheckTeachTaskAllocation() (bool, error) {

	// 1. 检查每周总课时数是否超过总课时数
	totalClassesPerWeek := s.Schedule.GetTotalClassesPerDay() * s.Schedule.NumWorkdays

	// 按照年级和班级统计课程数量
	classCount := make(map[string]int)

	// 按照年级、班级和科目统计上课次数
	subjectCount := make(map[string]int)
	for _, task := range s.TeachTaskAllocations {
		classKey := fmt.Sprintf("%d_%d", task.GradeID, task.ClassID)
		classSubjectKey := fmt.Sprintf("%d_%d_%d", task.GradeID, task.ClassID, task.SubjectID)
		classCount[classKey] += task.NumClassesPerWeek - task.NumConnectedClassesPerWeek
		subjectCount[classSubjectKey] += task.NumClassesPerWeek - task.NumConnectedClassesPerWeek
	}

	for key, count := range classCount {
		if count > totalClassesPerWeek {
			return false, fmt.Errorf("%s total course Classes %d exceed maximum weekly Classes %d", key, count, totalClassesPerWeek)
		}
	}

	// 2. 检查每个科目每周上课总次数是否正确
	// 每周工作5天,所以最多上5次课
	// 会出现,有一天两节课，不需要连堂, 所以去掉下面的判断
	// for key, time := range subjectCount {
	// 	if time > s.Schedule.NumWorkdays {
	// 		return false, fmt.Errorf("subject %s has invalid weekly Classes count", key)
	// 	}
	// }
	return true, nil
}

func (s *ScheduleInput) ConvertConstraints() map[string]interface{} {

	constraints := make(map[string]interface{})
	constraints["Class"] = s.ClassConstraints
	constraints["Subject"] = s.SubjectConstraints
	constraints["Teacher"] = s.TeacherConstraints
	constraints["SubjectMutex"] = s.SubjectMutexConstraints
	constraints["SubjectOrder"] = s.SubjectOrderConstraints
	constraints["SubjectDayLimit"] = s.SubjectDayLimitConstraints
	constraints["TeacherMutex"] = s.TeacherMutexConstraints
	constraints["TeacherNoonBreak"] = s.TeacherNoonBreakConstraints
	constraints["TeacherPeriodLimit"] = s.TeacherPeriodLimitConstraints
	constraints["TeacherRangeLimit"] = s.TeacherRangeLimitConstraints

	return constraints
}

// 加载yaml测试数据
func LoadTestData() (*ScheduleInput, error) {
	var config ScheduleInput

	// 设置配置文件名和类型
	viper.SetConfigType("yaml")
	// viper.SetConfigName("testdata")
	viper.SetConfigName("test1")
	// viper.SetConfigName("wen_dao")

	// 添加配置文件搜索路径
	viper.AddConfigPath("../../testdata")

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

// 从 JSON 字符串解析出排课输入数据
func ParseScheduleInputFromJSON(jsonstr string) (*ScheduleInput, error) {
	var input ScheduleInput

	// 使用 json.Unmarshal 函数将 JSON 格式的字符串解析为排课输入结构体
	err := json.Unmarshal([]byte(jsonstr), &input)
	if err != nil {
		return nil, fmt.Errorf("fatal error unmarshaling json: %s", err)
	}

	// 对 Courses 属性的值按照 NumClassesPerWeek 排序
	sort.Slice(input.TeachTaskAllocations, func(i, j int) bool {
		return input.TeachTaskAllocations[i].NumClassesPerWeek > input.TeachTaskAllocations[j].NumClassesPerWeek
	})

	return &input, nil
}
