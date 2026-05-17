package card

import (
	"strconv"

	"owl-common/alarm"
)

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
		alarm.Initialization:           0,
		alarm.Walking:                  1,
		alarm.SuspectedFall:            2,
		"Sitting position":             3,
		"Standing position":            4,
		alarm.Fall:                     5,
		"Lying position":               6,
		alarm.SuspectedSittingOnGround: 7,
		alarm.SittingOnGround:          8,
		"Sitting up in bed":            9,
		alarm.SuspectedBedSitUp:        10,
		alarm.BedSitUp:                 11,
	}

	if v, ok := poseMap[pose]; ok {
		return v
	}
	return 0
}
