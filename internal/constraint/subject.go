// subject.go
// 科目优先排禁排

package constraint

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"

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
	TimeSlot       int    `json:"time_slot" mapstructure:"time_slot"`               // 时间段
	Limit          string `json:"limit" mapstructure:"limit"`                       // 限制 固定排: fixed, 优先排: prefer, 禁排: not, 尽量不排: avoid
	Desc           string `json:"desc" mapstructure:"desc"`                         // 描述
}

// 生成字符串
func (s *Subject) String() string {
	return fmt.Sprintf("ID: %d, SubjectGroupID: %d, SubjectID: %d, TimeSlot: %d, Limit: %s, Desc: %s", s.ID,
		s.SubjectGroupID, s.SubjectID, s.TimeSlot, s.Limit, s.Desc)
}

// 获取科目优先排禁排规则
func GetSubjectRules() []*types.Rule {
	constraints := loadSubjectConstraintsFromDB()
	var rules []*types.Rule
	for _, s := range constraints {
		rule := s.genRule()
		rules = append(rules, rule)
	}
	return rules
}

// 生成规则
func (s *Subject) genRule() *types.Rule {
	fn := s.genConstraintFn()
	return &types.Rule{
		Name:     s.String(),
		Type:     "fixed",
		Fn:       fn,
		Score:    s.getScore(),
		Penalty:  s.getPenalty(),
		Weight:   s.getWeight(),
		Priority: s.getPriority(),
	}
}

// 加载班级固排禁排规则
func loadSubjectConstraintsFromDB() []*Subject {
	var constraints []*Subject
	return constraints
}

// 生成规则校验方法
func (s *Subject) genConstraintFn() types.ConstraintFn {
	return func(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {

		classSN := element.GetClassSN()
		timeSlot := element.GetTimeSlot()

		SN, err := types.ParseSN(classSN)
		if err != nil {
			return false, false, err
		}
		subject, err := models.FindSubjectByID(SN.SubjectID)
		if err != nil {
			return false, false, err
		}

		preCheckPassed := timeSlot == s.TimeSlot
		isValid := (s.SubjectGroupID == 0 || lo.Contains(subject.SubjectGroupIDs, s.SubjectGroupID)) && (s.SubjectID == 0 || s.SubjectID == SN.SubjectID)
		return preCheckPassed, isValid, nil
	}
}

// 奖励分
func (s *Subject) getScore() int {
	score := 0
	if s.Limit == "fixed" {
		score = 3
	} else if s.Limit == "prefer" {
		score = 2
	}
	return score
}

// 惩罚分
func (s *Subject) getPenalty() int {
	penalty := 0
	if s.Limit == "not" {
		penalty = 3
	} else if s.Limit == "avoid" {
		penalty = 2
	}
	return penalty
}

// 权重
func (s *Subject) getWeight() int {
	return 1
}

// 优先级
func (s *Subject) getPriority() int {
	return 1
}

// // 14. 语数英 周一~周五 第1节 优先排
// // 15. 语数英 周一~周五 第2节 优先排
// // 16. 语数英 周一~周五 第3节 优先排
// var SRule1 = &types.Rule{
// 	Name:     "SRule1",
// 	Type:     "fixed",
// 	Fn:       sRule1Fn,
// 	Score:    2,
// 	Penalty:  0,
// 	Weight:   1,
// 	Priority: 1,
// }

// // 副课 安排在第1,2,3节 扣分
// var SRule2 = &types.Rule{
// 	Name:     "SRule2",
// 	Type:     "fixed",
// 	Fn:       sRule2Fn,
// 	Score:    0,
// 	Penalty:  2,
// 	Weight:   1,
// 	Priority: 2,
// }

// // 17. 主课 周一~周五 第8节 禁排
// var SRule3 = &types.Rule{
// 	Name:     "SRule3",
// 	Type:     "fixed",
// 	Fn:       sRule3Fn,
// 	Score:    0,
// 	Penalty:  config.MaxPenaltyScore,
// 	Weight:   1,
// 	Priority: 1,
// }

// // 18. 主课 周一~周五 第7节 尽量不排
// var SRule4 = &types.Rule{
// 	Name:     "SRule4",
// 	Type:     "fixed",
// 	Fn:       sRule4Fn,
// 	Score:    0,
// 	Penalty:  1,
// 	Weight:   1,
// 	Priority: 1,
// }

// // 14. 语数英 周一~周五 第1节 优先排
// // 15. 语数英 周一~周五 第2节 优先排
// // 16. 语数英 周一~周五 第3节 优先排
// func sRule1Fn(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {

// 	subjectGroupID := 1
// 	classSN := element.GetClassSN()
// 	timeSlot := element.GetTimeSlot()

// 	SN, _ := types.ParseSN(classSN)
// 	subject, err := models.FindSubjectByID(SN.SubjectID)
// 	if err != nil {
// 		return false, false, err
// 	}
// 	day := timeSlot/config.NumClasses + 1
// 	period := timeSlot%config.NumClasses + 1

// 	// 判断subjectGroupID是否已经排课完成
// 	isSubjectGroupScheduled, err := isSubjectGroupScheduled(classMatrix, subjectGroupID)
// 	if err != nil {
// 		return false, false, err
// 	}
// 	preCheckPassed := (period == 1 || period == 2 || period == 3) && (day >= 1 && day <= 5)

// 	// FindAvailableSubjectsByGroupID
// 	shouldPenalize := preCheckPassed && !lo.Contains(subject.SubjectGroupIDs, subjectGroupID) && !isSubjectGroupScheduled

// 	// fmt.Printf("sRule1Fn sn: %s, timeSlot: %d, subjectGroupIDs: %d\n", classSN, timeSlot, subject.SubjectGroupIDs)
// 	return preCheckPassed, !shouldPenalize, nil
// }

// // 副课 安排在第1,2,3节 扣分
// // 满足该条件扣分, 不满足该该条件, 不增加分数, 也不扣分
// func sRule2Fn(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {

// 	classSN := element.GetClassSN()
// 	timeSlot := element.GetTimeSlot()

// 	SN, _ := types.ParseSN(classSN)
// 	subject, err := models.FindSubjectByID(SN.SubjectID)
// 	if err != nil {
// 		return false, false, err
// 	}
// 	period := timeSlot%config.NumClasses + 1

// 	preCheckPassed := period == 1 || period == 2 || period == 3

// 	shouldPenalize := preCheckPassed && lo.Contains(subject.SubjectGroupIDs, 3)
// 	return preCheckPassed, !shouldPenalize, nil
// }

// // 17. 主课 周一~周五 第8节 禁排
// func sRule3Fn(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {

// 	classSN := element.GetClassSN()
// 	timeSlot := element.GetTimeSlot()

// 	SN, _ := types.ParseSN(classSN)
// 	subject, err := models.FindSubjectByID(SN.SubjectID)
// 	if err != nil {
// 		return false, false, err
// 	}
// 	period := timeSlot%config.NumClasses + 1
// 	preCheckPassed := period == 8

// 	shouldPenalize := preCheckPassed && lo.Contains(subject.SubjectGroupIDs, 2)
// 	// fmt.Printf("sRule3Fn sn: %s, timeSlot: %d, shouldPenalize: %v\n", classSN, timeSlot, shouldPenalize)
// 	return preCheckPassed, !shouldPenalize, nil
// }

// // 18. 主课 周一~周五 第7节 尽量不排
// func sRule4Fn(classMatrix *types.ClassMatrix, element types.Element) (bool, bool, error) {

// 	classSN := element.GetClassSN()
// 	timeSlot := element.GetTimeSlot()

// 	SN, _ := types.ParseSN(classSN)
// 	subject, err := models.FindSubjectByID(SN.SubjectID)
// 	if err != nil {
// 		return false, false, err
// 	}
// 	period := timeSlot%config.NumClasses + 1
// 	preCheckPassed := period == 7

// 	shouldPenalize := preCheckPassed && lo.Contains(subject.SubjectGroupIDs, 2)
// 	return preCheckPassed, !shouldPenalize, nil
// }

// 判断subjectGroupID的课程是否已经排完
// Check if all courses in the subject group are scheduled
func isSubjectGroupScheduled(classMatrix *types.ClassMatrix, subjectGroupID int) (bool, error) {

	// 根据科目分组得到所有的科目
	subjects, err := models.FindSubjectsByGroupID(subjectGroupID)
	if err != nil {
		return false, err
	}

	// 根据科目得到该科目的一周课时
	classHours := models.GetClassHours()

	for _, subject := range subjects {
		subjectID := subject.SubjectID
		subjectClassHours := classHours[subjectID]
		totalScheduledHours := 0

		for sn, classMap := range classMatrix.Elements {
			SN, err := types.ParseSN(sn)
			if err != nil {
				return false, err
			}

			if SN.SubjectID == subjectID {
				for _, teacherMap := range classMap {
					for _, venueMap := range teacherMap {
						for _, element := range venueMap {
							if element.Val.Used == 1 {
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
