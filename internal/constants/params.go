package constants

const (
	NUM_DAYS              = 5                      // 每周上课天数
	NUM_CLASSES           = 8                      // 每天上课节数
	NUM_SUBJECTS          = 15                     // 课程数
	NUM_GRADES            = 1                      // 年级数
	NUM_CLASSES_PER_GRADE = 1                      // 每年级班级数
	NUM_TIMESLOTS         = NUM_DAYS * NUM_CLASSES // 每周课程表格子数
)

const (
	PERIOD_THRESHOLD = 2 // 相同节次排课数量限制
)

const (
	MAX_PENALTY_SCORE = 3 // 表示ClassMatrix中的元素可以具有的最大可能得分, 这个得分很重要,会直接影响适应度计算的结果, 一般和最高的奖励分是相同的
)
