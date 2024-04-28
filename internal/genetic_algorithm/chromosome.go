// chromosome.go
package genetic_algorithm

// Chromosome 染色体结构体，代表一个课班的排课信息
type Chromosome struct {
	ClassSN string  // 课班 科目_年级_班级
	Genes   []*Gene // 基因序列
}

// Copy 复制一个 Chromosome 实例
func (c *Chromosome) Copy() *Chromosome {
	newGenes := make([]*Gene, len(c.Genes))
	for i, gene := range c.Genes {
		newGene := &Gene{
			ClassSN:            gene.ClassSN,
			TeacherID:          gene.TeacherID,
			VenueID:            gene.VenueID,
			TimeSlot:           gene.TimeSlot,
			FailedConstraints:  make([]string, len(gene.FailedConstraints)),
			PassedConstraints:  make([]string, len(gene.PassedConstraints)),
			SkippedConstraints: make([]string, len(gene.SkippedConstraints)),
		}
		copy(newGene.FailedConstraints, gene.FailedConstraints)
		copy(newGene.PassedConstraints, gene.PassedConstraints)
		copy(newGene.SkippedConstraints, gene.SkippedConstraints)
		newGenes[i] = newGene
	}

	return &Chromosome{
		ClassSN: c.ClassSN,
		Genes:   newGenes,
	}
}
