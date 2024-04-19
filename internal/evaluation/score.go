// score.go
package evaluation

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"math"
	"sort"

	"github.com/samber/lo"
)

// 得分明细
// 得分规则的名称、得分和惩罚
type ScoreDetail struct {
	Name    string
	Score   int
	Penalty int
}

// 计算得分的结果
// 包含最终得分和得分明细
type CalcScoreResult struct {
	FinalScore int
	Details    []ScoreDetail
}

// 匹配结果值
// 将: 匹配结果值越大越好，匹配结果值为“-1”表示课班不可用当前适应性矩阵的元素下标
// 修改为: 匹配结果值越大越好，匹配结果值也可能会是负数
func CalcScore(classMatrix map[string]map[int]map[int]map[int]types.Val, classHours map[int]int, sn string, teacherID, venueID, timeSlot int) (*CalcScoreResult, error) {

	scoreDetails := []ScoreDetail{}

	// 检查缓存
	if cachedScore, ok := getCachedScore(sn, teacherID, venueID, timeSlot); ok {
		return cachedScore, nil
	}

	score := 0   // 得分
	penalty := 0 // 惩罚分

	if sn == "" {
		return &CalcScoreResult{FinalScore: 0, Details: scoreDetails}, fmt.Errorf("sn is empty")
	}

	SN, err := types.ParseSN(sn)
	if err != nil {
		return &CalcScoreResult{FinalScore: 0, Details: scoreDetails}, err
	}

	teacher, err := models.FindTeacherByID(teacherID)
	if err != nil {
		return &CalcScoreResult{FinalScore: 0, Details: scoreDetails}, err
	}

	// 周几
	day := timeSlot/constants.NUM_CLASSES + 1

	// 课节数
	lesson := timeSlot%constants.NUM_CLASSES + 1

	// 班级固排禁排
	score, penalty, scoreDetails = classFixedAndForbidden(SN, teacherID, lesson, score, penalty, scoreDetails)
	if penalty == math.MaxInt32 {
		return &CalcScoreResult{FinalScore: score - penalty, Details: scoreDetails}, nil
	}

	// 教师固排禁排
	score, penalty, scoreDetails = teacherFixedAndForbidden(teacher, day, lesson, score, penalty, scoreDetails)
	if penalty == math.MaxInt32 {
		return &CalcScoreResult{FinalScore: score - penalty, Details: scoreDetails}, nil
	}

	// 科目优先排禁排
	subject, _ := models.FindSubjectByID(SN.SubjectID)
	score, penalty, scoreDetails = subjectPreferAndForbidden(subject, day, lesson, score, penalty, scoreDetails)
	if penalty == math.MaxInt32 {
		return &CalcScoreResult{FinalScore: score - penalty, Details: scoreDetails}, nil
	}

	// 教师时间段限制
	score, penalty, scoreDetails = teacherTimeLimit(teacherID, classMatrix, score, penalty, scoreDetails)

	// 教师节数限制
	score, penalty, scoreDetails = teacherClassLimit(teacherID, day, lesson, classMatrix, score, penalty, scoreDetails)

	// 教师互斥限制
	score, penalty, scoreDetails = teacherMutualExclusion(teacherID, classMatrix, score, penalty, scoreDetails)

	// 教师不跨中午
	score, penalty, scoreDetails = teacherNotAcrossNoon(teacherID, classMatrix, score, penalty, scoreDetails)

	// 科目互斥限制
	score, penalty, scoreDetails, err = subjectMutualExclusion(subject.SubjectID, classMatrix, score, penalty, scoreDetails)
	if err != nil {
		return &CalcScoreResult{FinalScore: 0, Details: scoreDetails}, err
	}

	// 科目顺序限制
	score, penalty, scoreDetails, err = subjectOrder(subject.SubjectID, classMatrix, score, penalty, scoreDetails)
	if err != nil {
		return &CalcScoreResult{FinalScore: 0, Details: scoreDetails}, err
	}

	// 科目课时小于天数
	if classHours[subject.SubjectID] <= constants.NUM_DAYS {
		ret, err := isSubjectSameDay(subject.SubjectID, classMatrix)
		if err != nil {
			return &CalcScoreResult{FinalScore: 0, Details: scoreDetails}, err
		}
		if ret {
			name := fmt.Sprintf("subject_same_day_%d", subject.SubjectID)
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: math.MaxInt32})
			return &CalcScoreResult{FinalScore: score - penalty, Details: scoreDetails}, nil
		}
	} else {
		ret, err := isSubjectConsecutive(subject.SubjectID, classMatrix)
		if err != nil {
			return &CalcScoreResult{FinalScore: 0, Details: scoreDetails}, err
		}
		if !ret {
			name := fmt.Sprintf("subject_not_consecutive_%d", subject.SubjectID)
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: math.MaxInt32})
			return &CalcScoreResult{FinalScore: score - penalty, Details: scoreDetails}, nil
		}
	}

	// 计算最终得分
	finalScore := score - penalty
	calcScoreResult := CalcScoreResult{FinalScore: finalScore, Details: scoreDetails}

	// 保存缓存
	saveCachedScore(sn, teacherID, venueID, timeSlot, calcScoreResult)
	return &calcScoreResult, nil
}

// 班级固排禁排
// 前缀_科目_年级_班级_老师_场地_节次
func classFixedAndForbidden(SN *types.SN, teacherID, lesson int, score, penalty int, scoreDetails []ScoreDetail) (int, int, []ScoreDetail) {

	// 1. 一年级(1)班 语文 王老师 第1节 固排
	if SN.GradeID == 1 && SN.ClassID == 1 && SN.SubjectID == 1 && teacherID == 1 && lesson == 1 {
		score += 2
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "class_fixed_1_1_1_1_0_1", Score: 2, Penalty: 0})
	}

	// 2. 三年级(1)班 第7节 禁排 班会
	if SN.GradeID == 3 && SN.ClassID == 1 && lesson == 7 {
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "class_forbidden_0_3_1_0_0_7", Score: 0, Penalty: math.MaxInt32})
		return score, penalty, scoreDetails
	}

	// 3. 三年级(2)班 第8节 禁排 班会
	if SN.GradeID == 3 && SN.ClassID == 2 && lesson == 8 {
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "class_forbidden_0_3_2_0_0_8", Score: 0, Penalty: math.MaxInt32})
		return score, penalty, scoreDetails
	}

	// 4. 四年级 第8节 禁排 班会
	if SN.GradeID == 4 && lesson == 8 {
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "class_forbidden_0_4_0_0_0_8", Score: 0, Penalty: math.MaxInt32})
		return score, penalty, scoreDetails
	}

	// 5. 四年级(1)班 语文 王老师 第1节 禁排
	if SN.GradeID == 4 && SN.SubjectID == 1 && teacherID == 1 && lesson == 1 {
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "class_forbidden_0_4_1_1_0_1", Score: 0, Penalty: math.MaxInt32})
		return score, penalty, scoreDetails
	}

	// 6. 五年级 数学 李老师 第2节 固排
	if SN.GradeID == 5 && SN.SubjectID == 2 && teacherID == 1 && lesson == 2 {
		score += 2
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "class_fixed_2_5_0_2_0_2", Score: 2, Penalty: 0})
	}

	// 7. 五年级 数学 李老师 第3节 尽量排
	if SN.GradeID == 5 && SN.SubjectID == 2 && teacherID == 2 && lesson == 3 {
		score++
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "class_prefer_2_5_0_2_0_3", Score: 1, Penalty: 0})
	}

	// 8. 五年级 数学 李老师 第5节 固排
	if SN.GradeID == 5 && SN.SubjectID == 2 && teacherID == 2 && lesson == 5 {
		score += 2
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "class_fixed_2_5_0_2_0_5", Score: 2, Penalty: 0})
	}

	return score, penalty, scoreDetails
}

// 教师固排禁排
// 前缀_教室组ID_教师ID_周几_节次
func teacherFixedAndForbidden(teacher *models.Teacher, day, lesson int, score, penalty int, scoreDetails []ScoreDetail) (int, int, []ScoreDetail) {

	// 9. 数学组 周一 第4节 禁排 教研会
	if teacher.TeacherGroupIDs[0] == 2 && day == 1 && lesson == 4 {
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "teacher_forbidden_2_0_1_4", Score: 0, Penalty: math.MaxInt32})
		return score, penalty, scoreDetails
	}

	// 10. 刘老师 周一 第4节 禁排 教研会
	if teacher.TeacherID == 3 && day == 1 && lesson == 4 {
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "teacher_forbidden_0_3_1_4", Score: 0, Penalty: math.MaxInt32})
		return score, penalty, scoreDetails
	}

	// 11. 行政领导 周二 第7节 禁排 例会
	if teacher.TeacherGroupIDs[0] == 3 && day == 2 && lesson == 7 {
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "teacher_forbidden_3_0_2_7", Score: 0, Penalty: math.MaxInt32})
		return score, penalty, scoreDetails
	}

	// 12. 马老师 周二 第7节 禁排 例会
	if teacher.TeacherID == 5 && day == 2 && lesson == 7 {
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "teacher_forbidden_0_5_2_7", Score: 0, Penalty: math.MaxInt32})
		return score, penalty, scoreDetails
	}

	// 13. 王老师 周二 第2节 固排
	if teacher.TeacherID == 1 && day == 2 && lesson == 2 {
		score += 2
		scoreDetails = append(scoreDetails, ScoreDetail{Name: "teacher_fixed_0_1_2_2", Score: 2, Penalty: 0})
	}

	return score, penalty, scoreDetails
}

// 科目优先排禁排
// 前缀_科目组ID_科目ID_周几_节次
func subjectPreferAndForbidden(subject *models.Subject, day, lesson int, score, penalty int, scoreDetails []ScoreDetail) (int, int, []ScoreDetail) {

	// 14. 语数英 周一~周五 第1节 优先排
	// 15. 语数英 周一~周五 第2节 优先排
	// 16. 语数英 周一~周五 第3节 优先排

	if lo.Contains(subject.SubjectGroupIDs, 1) && (lesson == 1 || lesson == 2 || lesson == 3) {
		score++
		name := fmt.Sprintf("subject_prefer_1_%d_%d_%d", subject.SubjectID, day, lesson)
		scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 1, Penalty: 0})
	}

	// 17. 主课 周一~周五 第8节 禁排
	if lo.Contains(subject.SubjectGroupIDs, 2) && lesson == 8 {

		name := fmt.Sprintf("subject_forbidden_2_%d_%d_%d", subject.SubjectID, day, lesson)
		scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: math.MaxInt32})
		return score, penalty, scoreDetails
	}

	// 18. 主课 周一~周五 第7节 尽量不排
	if lo.Contains(subject.SubjectGroupIDs, 2) && lesson == 7 {
		penalty++
		name := fmt.Sprintf("subject_avoid_2_%d_%d_%d", subject.SubjectID, day, lesson)
		scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
	}

	// 19. 已删除

	return score, penalty, scoreDetails
}

// 场地优先排禁排
// 20. 机房1 一年级(1)班 计算机 周一~周五 后2节 固排
// 21. 机房2 一年级(2)班 计算机 周一~周五 第8节 优先排
// 22. 操场 周一 第 7,8 节 禁排

// 教师时间段限制
// 前缀_teacherID_startPeriod_endPeriod_limit_count
func teacherTimeLimit(teacherID int, classMatrix map[string]map[int]map[int]map[int]types.Val, score, penalty int, scoreDetails []ScoreDetail) (int, int, []ScoreDetail) {

	// 23. 王老师 上午 最多1节
	if teacherID == 1 {
		count := countTeacherClassesInRange(1, 1, 4, classMatrix)
		if count > 1 {
			penalty++
			name := fmt.Sprintf("teacher_time_limit_1_1_4_1_%d", count)
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}

		// 24. 王老师 下午 最多2节
		count = countTeacherClassesInRange(1, 5, 8, classMatrix)
		if count > 2 {
			penalty++
			name := fmt.Sprintf("teacher_time_limit_1_5_8_2_%d", count)
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}

		// 25. 王老师 全天(不含晚自习) 最多3节
		count = countTeacherClassesInRange(1, 1, 8, classMatrix)
		if count > 3 {
			penalty++
			name := fmt.Sprintf("teacher_time_limit_1_1_8_3_%d", count)
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}

		// 26. 王老师 晚自习 最多1节
	}

	return score, penalty, scoreDetails
}

// 教师节数限制
// 前缀_teacherID_period_limit_count
func teacherClassLimit(teacherID, day, lesson int, classMatrix map[string]map[int]map[int]map[int]types.Val, score, penalty int, scoreDetails []ScoreDetail) (int, int, []ScoreDetail) {

	// 27. 王老师 上午第4节 最多3次
	if teacherID == 1 {
		count := countTeacherClassInPeriod(1, 4, classMatrix)
		if count > 3 {
			penalty++
			name := fmt.Sprintf("teacher_class_limit_1_4_3_%d", count)
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}
	}

	// 28. 李老师 上午第4节 最多3次
	if teacherID == 2 {
		count := countTeacherClassInPeriod(2, 4, classMatrix)
		if count > 3 {
			penalty++
			name := fmt.Sprintf("teacher_class_limit_2_4_3_%d", count)
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}
	}

	// 29. 刘老师 上午第4节 最多3次
	if teacherID == 3 {
		count := countTeacherClassInPeriod(3, 4, classMatrix)
		if count > 3 {
			penalty++
			name := fmt.Sprintf("teacher_class_limit_3_4_3_%d", count)
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}
	}
	// 30. 张老师 上午第4节 最多3次
	if teacherID == 4 {
		count := countTeacherClassInPeriod(4, 4, classMatrix)
		if count > 3 {
			penalty++
			name := fmt.Sprintf("teacher_class_limit_4_4_3_%d", count)
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}
	}
	return score, penalty, scoreDetails
}

// 教师互斥限制
// 前缀_教师AID_教师BID
func teacherMutualExclusion(teacherID int, classMatrix map[string]map[int]map[int]map[int]types.Val, score, penalty int, scoreDetails []ScoreDetail) (int, int, []ScoreDetail) {

	// 31. 王老师 马老师
	if teacherID == 1 || teacherID == 5 {
		if isTeacherSameDay(1, 5, classMatrix) {
			penalty++
			name := "teacher_mutual_exclusion_1_5"
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}
	}

	// 32. 李老师 黄老师
	if teacherID == 2 || teacherID == 6 {
		if isTeacherSameDay(2, 6, classMatrix) {
			penalty++
			name := "teacher_mutual_exclusion_2_6"
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}
	}

	return score, penalty, scoreDetails
}

// 教师不跨中午(教师排了上午最后一节就不排下午第一节)
// teacherID
func teacherNotAcrossNoon(teacherID int, classMatrix map[string]map[int]map[int]map[int]types.Val, score, penalty int, scoreDetails []ScoreDetail) (int, int, []ScoreDetail) {

	// 33. 王老师
	if teacherID == 1 {
		if isTeacherInBothPeriods(1, 4, 5, classMatrix) {
			penalty++
			name := "teacher_not_across_noon_1"
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}
	}

	// 34. 李老师
	if teacherID == 2 {

		if isTeacherInBothPeriods(2, 4, 5, classMatrix) {
			penalty++
			name := "teacher_not_across_noon_2"
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}
	}
	return score, penalty, scoreDetails
}

// 教师连堂限制(选中的教师不排连堂课)
// 35. 王老师
// 36. 李老师

// 科目互斥限制(科目A与科目B不排在同一天)
// 37. 活动 体育
func subjectMutualExclusion(subjectID int, classMatrix map[string]map[int]map[int]map[int]types.Val, score, penalty int, scoreDetails []ScoreDetail) (int, int, []ScoreDetail, error) {

	if subjectID == 14 || subjectID == 6 {
		ret, err := isSubjectsSameDay(14, 6, classMatrix)
		if err != nil {
			return score, penalty, scoreDetails, err
		}
		if ret {
			penalty++
			name := "subject_mutual_exclusion_14_6"
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}
	}

	return score, penalty, scoreDetails, nil

}

// 科目顺序限制(体育课不排在数学课前)
// 38. 体育 数学
func subjectOrder(subjectID int, classMatrix map[string]map[int]map[int]map[int]types.Val, score, penalty int, scoreDetails []ScoreDetail) (int, int, []ScoreDetail, error) {

	if subjectID == 6 || subjectID == 2 {

		ret, err := isSubjectABeforeSubjectB(6, 2, classMatrix)
		if err != nil {
			return score, penalty, scoreDetails, err
		}
		if ret {
			penalty++
			name := "subject_order_6_2"
			scoreDetails = append(scoreDetails, ScoreDetail{Name: name, Score: 0, Penalty: 1})
		}

	}

	return score, penalty, scoreDetails, nil
}

// 统计教师时间段节次
func countTeacherClassesInRange(teacherID int, startPeriod, endPeriod int, classMatrix map[string]map[int]map[int]map[int]types.Val) int {
	count := 0
	for _, classMap := range classMatrix {
		for id, teacherMap := range classMap {
			if teacherID == id {
				for _, timeSlotMap := range teacherMap {
					if timeSlotMap == nil {
						continue
					}
					for timeSlot, val := range timeSlotMap {
						if val.Used == 1 && timeSlot >= startPeriod && timeSlot <= endPeriod {
							count++
						}
					}
				}
			}
		}
	}
	return count
}

// 统计教师节数
func countTeacherClassInPeriod(teacherID int, period int, classMatrix map[string]map[int]map[int]map[int]types.Val) int {
	count := 0
	for _, classMap := range classMatrix {
		for id, teacherMap := range classMap {
			if teacherID == id {
				for _, timeSlotMap := range teacherMap {
					if val, ok := timeSlotMap[period]; ok && val.Used == 1 {
						count++
					}
				}
			}
		}
	}
	return count
}

// 判断教师A,教师B是否同一天都有课
func isTeacherSameDay(teacherAID, teacherBID int, classMatrix map[string]map[int]map[int]map[int]types.Val) bool {
	teacher1Days := make(map[int]bool)
	teacher2Days := make(map[int]bool)
	for _, classMap := range classMatrix {
		for id, teacherMap := range classMap {
			if id == teacherAID {
				for _, timeSlotMap := range teacherMap {
					for timeSlot, val := range timeSlotMap {
						if val.Used == 1 {
							teacher1Days[timeSlot/constants.NUM_CLASSES] = true // 将时间段转换为天数
						}
					}
				}
			} else if id == teacherBID {
				for _, timeSlotMap := range teacherMap {
					for timeSlot, val := range timeSlotMap {
						if val.Used == 1 {
							teacher2Days[timeSlot/constants.NUM_CLASSES] = true // 将时间段转换为天数
						}
					}
				}
			}
		}
	}
	for day := 0; day < constants.NUM_DAYS; day++ {
		if teacher1Days[day] && teacher2Days[day] {
			return true
		}
	}
	return false
}

// 判断教师是否在两个节次都有课
func isTeacherInBothPeriods(teacherID int, period1, period2 int, classMatrix map[string]map[int]map[int]map[int]types.Val) bool {
	for _, classMap := range classMatrix {
		for id, teacherMap := range classMap {
			if id == teacherID {
				for _, periodMap := range teacherMap {
					if val1, ok := periodMap[period1]; ok && val1.Used == 1 {
						if val2, ok := periodMap[period2]; ok && val2.Used == 1 {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// 判断活动课和体育课是否在同一天
func isSubjectsSameDay(subjectAID, subjectBID int, classMatrix map[string]map[int]map[int]map[int]types.Val) (bool, error) {

	subjectADays := make(map[int]bool)
	subjectBDays := make(map[int]bool)
	for sn, classMap := range classMatrix {

		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}

		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for timeSlot, val := range venueMap {
					if val.Used == 1 {
						if SN.SubjectID == subjectAID {
							subjectADays[timeSlot/constants.NUM_CLASSES] = true // 将时间段转换为天数
						} else if SN.SubjectID == subjectBID {
							subjectBDays[timeSlot/constants.NUM_CLASSES] = true // 将时间段转换为天数
						}
					}
				}
			}
		}
	}

	for day := 0; day < constants.NUM_DAYS; day++ {
		if subjectADays[day] && subjectBDays[day] {
			return true, nil
		}
	}
	return false, nil
}

// 判断体育课后是否就是数学课
// 判断课程A是在课程B之前
func isSubjectABeforeSubjectB(subjectAID, subjectBID int, classMatrix map[string]map[int]map[int]map[int]types.Val) (bool, error) {
	// 遍历课程表，同时记录课程A和课程B的上课时间段
	var timeSlotsA, timeSlotsB []int
	for sn, classMap := range classMatrix {
		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}
		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for timeSlot, val := range venueMap {
					if val.Used == 1 {
						if SN.SubjectID == subjectAID {
							timeSlotsA = append(timeSlotsA, timeSlot)
						} else if SN.SubjectID == subjectBID {
							timeSlotsB = append(timeSlotsB, timeSlot)
						}
					}
				}
			}
		}
	}
	// 如果课程A或课程B没有上课时间段，则返回false
	if len(timeSlotsA) == 0 || len(timeSlotsB) == 0 {
		return false, nil
	}
	// 对上课时间段进行排序
	sort.Ints(timeSlotsA)
	sort.Ints(timeSlotsB)
	// 检查课程B是否在课程A之后
	for _, timeSlotA := range timeSlotsA {
		for _, timeSlotB := range timeSlotsB {
			if timeSlotB == timeSlotA+1 {
				return true, nil
			}
		}
	}
	// 如果没有找到课程B在课程A之后的上课时间，则返回false
	return false, nil
}

// 检查同一科目是否在同一天安排了多次
func isSubjectSameDay(subjectID int, classMatrix map[string]map[int]map[int]map[int]types.Val) (bool, error) {

	// 科目每天排课次数 科目:次数
	subjectDays := make(map[int]int)
	for sn, classMap := range classMatrix {
		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}
		if SN.SubjectID != subjectID {
			continue
		}
		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for timeSlot, val := range venueMap {
					if val.Used == 1 {
						day := timeSlot / constants.NUM_CLASSES
						subjectDays[day]++
					}
				}
			}
		}
	}
	for _, count := range subjectDays {
		if count > 1 {
			return true, nil
		}
	}
	return false, nil
}

// 检查科目是否连续排课
func isSubjectConsecutive(subjectID int, classMatrix map[string]map[int]map[int]map[int]types.Val) (bool, error) {
	subjectTimeSlots := make([]int, 0)
	for sn, classMap := range classMatrix {
		SN, err := types.ParseSN(sn)
		if err != nil {
			return false, err
		}
		if SN.SubjectID != subjectID {
			continue
		}
		for _, teacherMap := range classMap {
			for _, venueMap := range teacherMap {
				for timeSlot, val := range venueMap {
					if val.Used == 1 {
						subjectTimeSlots = append(subjectTimeSlots, timeSlot)
					}
				}
			}
		}
	}
	sort.Ints(subjectTimeSlots)
	for i := 0; i < len(subjectTimeSlots)-1; i++ {
		if subjectTimeSlots[i]+1 == subjectTimeSlots[i+1] {
			return true, nil
		}
	}
	return false, nil
}
