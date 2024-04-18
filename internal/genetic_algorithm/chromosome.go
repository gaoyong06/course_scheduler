package genetic_algorithm

// Chromosome 染色体结构体，代表一个课班的排课信息
type Chromosome struct {
	ClassSN string // 课班 科目_年级_班级
	Genes   []Gene // 基因序列
}

// Copy 复制一个 Chromosome 实例
func (c *Chromosome) Copy() *Chromosome {
	newGenes := make([]Gene, len(c.Genes))
	copy(newGenes, c.Genes)

	return &Chromosome{
		ClassSN: c.ClassSN,
		Genes:   newGenes,
	}
}
