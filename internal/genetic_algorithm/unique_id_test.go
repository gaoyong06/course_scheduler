package genetic_algorithm

import (
	"fmt"
	"testing"
)

func TestUniqueId(t *testing.T) {

	individ1 := &Individual{
		Chromosomes: []*Chromosome{
			&Chromosome{
				ClassSN: "1_9_1",
				Genes: []*Gene{
					&Gene{
						ClassSN:            "1_9_1",
						TeacherID:          1,
						VenueID:            901,
						TimeSlots:          []int{10, 11},
						IsConnected:        true, // 假设这是连堂课
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
				},
			},
		},
	}

	individ2 := &Individual{
		Chromosomes: []*Chromosome{
			&Chromosome{
				ClassSN: "1_9_1",
				Genes: []*Gene{
					&Gene{
						ClassSN:            "1_9_1",
						TeacherID:          2, // 这里有修改
						VenueID:            901,
						TimeSlots:          []int{10, 11},
						IsConnected:        true, // 假设这是连堂课
						FailedConstraints:  nil,
						PassedConstraints:  nil,
						SkippedConstraints: nil,
					},
				},
			},
		},
	}

	uniqueId1_1 := individ1.genUniqueId()
	uniqueId1_2 := individ1.genUniqueId()
	uniqueId2_1 := individ2.genUniqueId()
	uniqueId2_2 := individ2.genUniqueId()

	fmt.Printf("uniqueId1_1: %s, uniqueId1_2: %s, uniqueId2_1: %s, uniqueId2_2: %s\n", uniqueId1_1, uniqueId1_2, uniqueId2_1, uniqueId2_2)

}
