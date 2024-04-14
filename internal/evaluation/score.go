package evaluation

import (
	"course_scheduler/internal/constants"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"

	"github.com/samber/lo"
)

// 缓存计算结果，key为sn+teacherID+venueID+timeSlot，value为计算结果
var calcScoreCache = make(map[string]int)
var calcScoreCacheLock sync.RWMutex

// 获取缓存计算结果
func getCachedScore(sn string, teacherID, venueID, timeSlot int) (int, bool) {
	key := sn + "_" + strconv.Itoa(teacherID) + "_" + strconv.Itoa(venueID) + "_" + strconv.Itoa(timeSlot)
	calcScoreCacheLock.RLock()
	score, ok := calcScoreCache[key]
	calcScoreCacheLock.RUnlock()
	return score, ok
}

// 保存缓存计算结果
func saveCachedScore(sn string, teacherID, venueID, timeSlot, score int) {
	key := sn + "_" + strconv.Itoa(teacherID) + "_" + strconv.Itoa(venueID) + "_" + strconv.Itoa(timeSlot)
	calcScoreCacheLock.Lock()
	defer calcScoreCacheLock.Unlock() // 在函数退出时释放锁
	calcScoreCache[key] = score
}

// 匹配结果值
// 将: 匹配结果值越大越好，匹配结果值为“-1”表示课班不可用当前适应性矩阵的元素下标
// 修改为: 匹配结果值越大越好，匹配结果值也可能会是负数
func CalcScore(classMatrix map[string]map[int]map[int]map[int]types.Val, sn string, teacherID, venueID, timeSlot int) (int, error) {

	// 检查缓存
	if cachedScore, ok := getCachedScore(sn, teacherID, venueID, timeSlot); ok {
		return cachedScore, nil
	}

	score := 0   // 得分
	penalty := 0 // 惩罚分

	if sn == "" {
		return 0, fmt.Errorf("sn is empty")
	}

	SN, err := types.ParseSN(sn)
	if err != nil {
		return 0, err
	}

	teacher, err := models.FindTeacherByID(teacherID)
	if err != nil {
		return 0, err
	}

	// 周几
	day := timeSlot/constants.NUM_CLASSES + 1

	// 课节数
	lesson := timeSlot%constants.NUM_CLASSES + 1

	// 班级固排禁排
	// 1. 一年级(1)班 语文 王老师 第1节 固排
	if SN.GradeID == 1 && SN.ClassID == 1 && SN.SubjectID == 1 && teacherID == 1 && lesson == 1 {
		score += 2
	}

	// 2. 三年级(1)班 第7节 禁排 班会
	if SN.GradeID == 3 && SN.ClassID == 1 && lesson == 7 {
		penalty = math.MaxInt32
		return score - penalty, nil
	}

	// 3. 三年级(2)班 第8节 禁排 班会
	if SN.GradeID == 3 && SN.ClassID == 2 && lesson == 8 {
		penalty = math.MaxInt32
		return score - penalty, nil
	}

	// 4. 四年级 第8节 禁排 班会
	if SN.GradeID == 4 && lesson == 8 {
		penalty = math.MaxInt32
		return score - penalty, nil
	}

	// 5. 四年级(1)班 语文 王老师 第1节 禁排
	if SN.GradeID == 4 && SN.SubjectID == 1 && teacherID == 1 && lesson == 1 {
		penalty = math.MaxInt32
		return score - penalty, nil
	}

	// 6. 五年级 数学 李老师 第2节 固排
	if SN.GradeID == 5 && SN.SubjectID == 2 && teacherID == 1 && lesson == 2 {
		score += 2
	}

	// 7. 五年级 数学 李老师 第3节 尽量排
	if SN.GradeID == 5 && SN.SubjectID == 2 && teacherID == 2 && lesson == 3 {
		score++
	}

	// 8. 五年级 数学 李老师 第5节 固排
	if SN.GradeID == 5 && SN.SubjectID == 2 && teacherID == 2 && lesson == 5 {
		score += 2
	}

	// 教师固排禁排
	// 9. 数学组 周一 第4节 禁排 教研会
	if lo.Contains(teacher.TeacherGroupIDs, 2) && day == 1 && lesson == 4 {
		penalty = math.MaxInt32
		return score - penalty, nil
	}

	// 10. 刘老师 周一 第4节 禁排 教研会
	if teacherID == 3 && day == 1 && lesson == 4 {
		penalty = math.MaxInt32
		return score - penalty, nil
	}

	// 11. 行政领导 周二 第7节 禁排 例会
	if lo.Contains(teacher.TeacherGroupIDs, 3) && day == 2 && lesson == 7 {
		penalty = math.MaxInt32
		return score - penalty, nil
	}

	// 12. 马老师 周二 第7节 禁排 例会
	if teacherID == 5 && day == 2 && lesson == 7 {
		penalty = math.MaxInt32
		return score - penalty, nil
	}

	// 13. 王老师 周二 第2节 固排
	if teacherID == 1 && day == 2 && lesson == 2 {
		score += 2
	}

	// 科目优先排禁排
	// 14. 语数英 周一~周五 第1节 优先排
	// 15. 语数英 周一~周五 第2节 优先排
	// 16. 语数英 周一~周五 第3节 优先排
	subject, _ := models.FindSubjectByID(SN.SubjectID)
	if lo.Contains(subject.SubjectGroupIDs, 1) && (lesson == 1 || lesson == 2 || lesson == 3) {
		score++
	}

	// 17. 主课 周一~周五 第8节 禁排
	if lo.Contains(subject.SubjectGroupIDs, 2) && lesson == 8 {
		penalty = math.MaxInt32
		return score - penalty, nil
	}
	// 18. 主课 周一~周五 第7节 尽量不排
	if lo.Contains(subject.SubjectGroupIDs, 2) && lesson == 7 {
		penalty++
	}

	// 场地优先排禁排
	// 20. 机房1 一年级(1)班 计算机 周一~周五 后2节 固排
	// 21. 机房2 一年级(2)班 计算机 周一~周五 第8节 优先排
	// 22. 操场 周一 第 7,8 节 禁排

	// 教师时间段限制
	// 23. 王老师 上午 最多1节
	count := countTeacherClassesInRange(1, 1, 4, classMatrix)
	if count > 1 {
		penalty++
	}

	// 24. 王老师 下午 最多2节
	count = countTeacherClassesInRange(1, 5, 8, classMatrix)
	if count > 2 {
		penalty++
	}

	// 25. 王老师 全天(不含晚自习) 最多3节
	count = countTeacherClassesInRange(1, 1, 8, classMatrix)
	if count > 3 {
		penalty++
	}

	// 26. 王老师 晚自习 最多1节

	// 教师节数限制
	// 27. 王老师 上午第4节 最多3次
	count = countTeacherClassInPeriod(1, 4, classMatrix)
	if count > 3 {
		penalty++
	}

	// 28. 李老师 上午第4节 最多3次
	count = countTeacherClassInPeriod(2, 4, classMatrix)
	if count > 3 {
		penalty++
	}

	// 29. 刘老师 上午第4节 最多3次
	count = countTeacherClassInPeriod(3, 4, classMatrix)
	if count > 3 {
		penalty++
	}
	// 30. 张老师 上午第4节 最多3次
	count = countTeacherClassInPeriod(4, 4, classMatrix)
	if count > 3 {
		penalty++
	}

	// 教师互斥限制(教师A与教师B不排在同一天)
	// 31. 王老师 马老师
	if isTeacherSameDay(1, 5, classMatrix) {
		penalty++
	}
	// 32. 李老师 黄老师
	if isTeacherSameDay(2, 6, classMatrix) {
		penalty++
	}

	// 教师不跨中午(教师排了上午最后一节就不排下午第一节)
	// 33. 王老师
	if isTeacherInBothPeriods(1, 4, 5, classMatrix) {
		penalty++
	}
	// 34. 李老师
	if isTeacherInBothPeriods(2, 4, 5, classMatrix) {
		penalty++
	}

	// 教师连堂限制(选中的教师不排连堂课)
	// 35. 王老师
	// 36. 李老师

	// 科目互斥限制(科目A与科目B不排在同一天)
	// 37. 活动 体育
	ret, err := isSubjectsSameDay(14, 6, classMatrix)
	if err != nil {
		return 0, err
	}

	if ret {
		penalty++
	}

	// 科目顺序限制(体育课不排在数学课前)
	// 38. 体育 数学
	ret, err = isSubjectABeforeSubjectB(6, 2, classMatrix)
	if err != nil {
		return 0, err
	}

	if ret {
		penalty++
	}

	// 计算最终得分
	finalScore := score - penalty

	// 保存缓存
	saveCachedScore(sn, teacherID, venueID, timeSlot, finalScore)

	return finalScore, nil
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
