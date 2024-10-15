// subject.go
// 科目优先排禁排

package constraints

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"math"

	"github.com/samber/lo"
)

// ##### 科目优先排禁排
// - 科目自定义分组(语数英,主课,副课)
// | 科目分组 | 科目 | 时间               | 限制     | 描述 |
// | -------- | ---- | ------------------ | -------- | ---- |
// | 语数英   |      | 周一~周五, 第 1 节 | 优先排   |      |
// | 语数英   |      | 周一~周五, 第 2 节 | 优先排   |      |
// | 语数英   |      | 周一~周五, 第 3 节 | 优先排   |      |
// | 主课     |      | 周一~周五, 第 8 节 | 禁排     |      |
// | 副课     |      | 周一~周五, 第 7 节 | 尽量不排 |      |
// |          | 语文 | 周一~周五, 第 1 节 | 禁排     |      |

// SubjectGroupConstraint 科目优先排禁排约束
type Subject struct {
	ID             int    `json:"id" mapstructure:"id"`                             // 自增ID
	SubjectGroupID int    `json:"subject_group_id" mapstructure:"subject_group_id"` // 科目分组id
	SubjectID      int    `json:"subject_id" mapstructure:"subject_id"`             // 科目id
	TimeSlots      []int  `json:"time_slots" mapstructure:"time_slots"`             // 时间段集合
	Limit          string `json:"limit" mapstructure:"limit"`                       // 限制 固定排: fixed, 优先排: prefer, 禁排: not, 尽量不排: avoid
	Desc           string `json:"desc" mapstructure:"desc"`                         // 描述
}

// 生成字符串
func (s *Subject) String() string {
	return fmt.Sprintf("ID: %d, SubjectGroupID: %d, SubjectID: %d, TimeSlots: %v, Limit: %s, Desc: %s", s.ID,
		s.SubjectGroupID, s.SubjectID, s.TimeSlots, s.Limit, s.Desc)
}

// 获取科目优先排禁排规则
func GetSubjectRules(subjects []*models.Subject, constraints []*Subject) []*types.Rule {
	// constraints := loadSubjectConstraintsFromDB()
	var rules []*types.Rule
	for _, s := range constraints {
		rule := s.genRule(subjects)
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (s *Subject) genRule(subjects []*models.Subject) *types.Rule {
	fn := s.genConstraintFn(subjects)
	return &types.Rule{
		Name:     "subject",
		Type:     "fixed",
		Fn:       fn,
		Score:    s.getScore(),
		Penalty:  s.getPenalty(),
		Weight:   1,
		Priority: 1,
	}
}

// 加载班级固排禁排规则
func loadSubjectConstraintsFromDB() []*Subject {
	var constraints []*Subject
	return constraints
}

// 生成规则校验方法
func (s *Subject) genConstraintFn(subjects []*models.Subject) types.ConstraintFn {
	return func(classMatrix *types.ClassMatrix, element types.Element, schedule *models.Schedule, teachingTasks []*models.TeachingTask) (bool, bool, error) {

		subjectID := element.SubjectID
		subject, err := models.FindSubjectByID(subjectID, subjects)
		if err != nil {
			return false, false, err
		}

		preCheckPassed := false
		isReward := false

		// 当前时间段,是否包含在约束时间段内
		intersect := lo.Intersect(s.TimeSlots, element.TimeSlots)
		isContain := len(intersect) > 0

		// 固排,优先排是: 排了有奖励,不排有处罚
		if s.Limit == "fixed" || s.Limit == "prefer" {
			preCheckPassed = isContain
			isReward = preCheckPassed && (s.SubjectGroupID == 0 || lo.Contains(subject.SubjectGroupIDs, s.SubjectGroupID)) && (s.SubjectID == 0 || s.SubjectID == subjectID)
		}

		// 禁排,尽量不排是: 不排没关系, 排了就处罚
		if s.Limit == "not" || s.Limit == "avoid" {
			preCheckPassed = isContain && (s.SubjectGroupID == 0 || lo.Contains(subject.SubjectGroupIDs, s.SubjectGroupID)) && (s.SubjectID == 0 || s.SubjectID == subjectID)
			isReward = false
		}

		// if element.ClassSN == "1_9_1" && s.SubjectID == 1 {
		// 	log.Printf("subject constraint, sn: %s, timeSlots: %v, limit: %s,  preCheckPassed: %v, isReward: %v\n", element.ClassSN, element.TimeSlots, s.Limit, preCheckPassed, isReward)
		// }

		return preCheckPassed, isReward, nil
	}
}

// 奖励分
func (s *Subject) getScore() int {
	score := 0
	if s.Limit == "fixed" {
		score = math.MaxInt32
	} else if s.Limit == "prefer" {
		score = 4
	}
	return score
}

// 惩罚分
func (s *Subject) getPenalty() int {
	penalty := 0
	if s.Limit == "not" {
		penalty = math.MaxInt32
	} else if s.Limit == "avoid" {
		penalty = 4
	}
	return penalty
}

// 判断subjectGroupID的课程是否已经排完
func isSubjectGroupScheduled(classMatrix *types.ClassMatrix, gradeID, classID, subjectGroupID int, subjects []*models.Subject, teachingTasks []*models.TeachingTask) (bool, error) {

	// 根据科目分组得到所有的科目
	subjects, err := models.FindSubjectsByGroupID(subjectGroupID, subjects)
	if err != nil {
		return false, err
	}

	// 根据科目得到该科目的一周课时
	// classHours := models.GetClassHours()

	for _, subject := range subjects {
		subjectID := subject.SubjectID
		subjectClassHours := models.GetNumClassesPerWeek(gradeID, classID, subjectID, teachingTasks)
		totalScheduledHours := 0

		for sn, classMap := range classMatrix.Elements {
			SN, err := types.ParseSN(sn)
			if err != nil {
				return false, err
			}

			if SN.SubjectID == subjectID {
				for _, teacherMap := range classMap {
					for _, venueMap := range teacherMap {
						for _, e := range venueMap {
							if e.Val.Used == 1 {
								totalScheduledHours++
							}
						}
					}
				}
			}
		}

		if subjectClassHours != totalScheduledHours {

			// fmt.Printf("subjectID: %d, subjectClassHours: %d, totalScheduledHours: %d\n", subjectID, subjectClassHours, totalScheduledHours)
			return false, nil
		}
	}

	return true, nil
}
