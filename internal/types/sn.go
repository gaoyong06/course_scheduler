package types

import (
	"fmt"
	"strconv"
	"strings"
)

type SN struct {
	SubjectID int
	GradeID   int
	ClassID   int
}

// Generate 生成 SN
func (sn *SN) Generate() string {
	return fmt.Sprintf("%d_%d_%d", sn.SubjectID, sn.GradeID, sn.ClassID)
}

// Parse 解析 SN
func ParseSN(sn string) (*SN, error) {
	parts := strings.Split(sn, "_")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid SN format: %s", sn)
	}

	subjectID, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid subject ID: %s", parts[0])
	}

	gradeID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid grade ID: %s", parts[1])
	}

	classID, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid class ID: %s", parts[2])
	}

	return &SN{
		SubjectID: subjectID,
		GradeID:   gradeID,
		ClassID:   classID,
	}, nil
}
