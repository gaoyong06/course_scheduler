我在使用遗传算法做一个排课的程序，现在的思路是先根据上课的天数,每天上课节数, 科目的教学老师, 科目班级上课的班级, 科目班级上课的教室，初始化一个课班适应性矩阵

例如: 每周5天上课,每天8节课,一周共40节课,一年级1班的数学老师是王老师, 上课是在教室1, 那么这个课班适应性矩阵classMatrix用伪代码表示就是:


// 课班适应性矩阵
type ClassMatrix struct {
	Elements map[string]map[int]map[int]map[int]types.Val
}

// 新建课班适应性矩阵
func NewClassMatrix() *ClassMatrix {
	return &ClassMatrix{
		Elements: make(map[string]map[int]map[int]map[int]types.Val),
	}
}

// 课表适应性矩阵元素值
type Val struct {
	ScoreInfo *ScoreInfo `json:"score_info"` // 匹配结果值, 匹配结果值越大越好, 默认值: 0
	Used      int        `json:"used"`       // 是否占用, 0 未占用, 1 已占用, 默认值: 0
}

// ScoreInfo 存储课班适应性矩阵元素的得分详情
type ScoreInfo struct {
	Score         int      // 最终得分 值越大越好, 默认值: 0
	FixedScore    int      // 固定约束条件得分
	DynamicScore  int      // 动态约束条件得分
	FixedPassed   []string // 满足的固定约束条件
	FixedFailed   []string // 未满足的固定约束条件
	DynamicPassed []string // 满足的动态约束条件
	DynamicFailed []string // 未满足的动态约束条件
}

subjectID = 2 // 数学
gradeID = 1 // 1年级
classID = 1 // 班级id
teacherID = 3 // 王老师
venueID = 4 // 教室1

classMatrix["2_1_1"][3][4][0]=types.Val{Score: 0, Used: 0}
classMatrix["2_1_1"][3][4][2]=types.Val{Score: 0, Used: 0}
classMatrix["2_1_1"][3][4][3]=types.Val{Score: 0, Used: 0}
classMatrix["2_1_1"][3][4][4]=types.Val{Score: 0, Used: 0}
classMatrix["2_1_1"][3][4][5]=types.Val{Score: 0, Used: 0}
....
classMatrix["2_1_1"][3][4][37]=types.Val{Score: 0, Used: 0}
classMatrix["2_1_1"][3][4][38]=types.Val{Score: 0, Used: 0}
classMatrix["2_1_1"][3][4][39]=types.Val{Score: 0, Used: 0}


然后根据各种各样的约束条件,给classMatrix的每个项打分score, 最满足要求的分数最高,最不满足要求的分数最低

有些约束条件是固定的, 例如王老师, 周一上午不排课,那么周一上午的时间段都可以设置最低的score

有些约束条件是动态的, 例如数学课一周4节课, 那么如果周一已经有1节数学课, 就不希望在排一节数学课

先对上述的课班适应性矩阵对按照固定约束条件，先对该矩阵的各个元素逐个计算Val中的Score

然后按照教学任务，某个科目每周几节课，开始排课，如果在该时间段排课，则该矩阵元素的Val中的Used赋值为1,表示已占用

然后根据当前的矩阵占用情况，在计算动态约束条件下，矩阵各个元素的得分

下次循环排课时，继续按照分数最高的的时间段，标记改时间段为已占用

这样，将教学任务中需要的科目及其课时，都在矩阵中标价完毕

然后根据矩阵中的标记，生成遗传算法的个体，个体是有染色体组成的，一个染色体，是一个科目+年级+班级组成的, 例如一年级1班语文课，一周有5节课，那么个体中就会有一个染色体染色体中会有5个基因

相关的数据结构如下

// Individual 个体结构体，代表一个完整的课表排课方案
type Individual struct {
	Chromosomes []*Chromosome // 染色体序列
	Fitness     int           // 适应度
}

// Chromosome 染色体结构体，代表一个课班的排课信息
type Chromosome struct {
	ClassSN string // 课班 科目_年级_班级
	Genes   []Gene // 基因序列
}


// 基因
type Gene struct {
	ClassSN   string // 课班信息，科目_年级_班级 如:美术_一年级_1班
	TeacherID int    // 教师id
	VenueID   int    // 教室id
	TimeSlot  int    // 时间段 一周5天,每天8节课,TimeSlot值是{0,1,2,3...39}
}

现在的问题是, 之前在课班适应性矩阵计算矩阵中各个元素的分数时，是安装矩阵的结构，来计算的，现在要计算个体的适应度，也是需要计算一个分数，但是适应度计算的约束条件，大部分是个课班适应性矩阵计算的约束条件相同的，
是应该重新写一个计算个体适应度计算的逻辑代码，还是可以复用上面课班适应性矩阵计算矩阵的计算过程

你有什么更好的建议？