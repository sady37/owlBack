package card

import "strconv"

// poseStringToInt 将 firmware pose 字符串转换为内部 int 编码。
func poseStringToInt(pose string) int {
	if pose == "" {
		return 0
	}

	if n, err := strconv.Atoi(pose); err == nil {
		if n >= 0 && n <= 11 {
			return n
		}
	}

	poseMap := map[string]int{
		"Initialization":           0,
		"Walking":                  1,
		"SuspectedFall":            2,
		"Sitting position":         3,
		"Standing position":        4,
		"Fall":                     5,
		"Lying position":           6,
		"SuspectedSittingOnGround": 7,
		"SittingOnGround":          8,
		"Sitting up in bed":        9,
		"SuspectedBedSitUp":        10,
		"BedSitUp":                 11,
	}

	if v, ok := poseMap[pose]; ok {
		return v
	}
	return 0
}
