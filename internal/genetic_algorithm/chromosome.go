package genetic_algorithm

// Chromosome 染色体结构体，代表一个课班的排课信息
type Chromosome struct {
	ClassSN string // 课班 科目_年级_班级
	Genes   []Gene // 基因序列
}
