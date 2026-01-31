// Package converter 睡眠、姿态、床态等 code/display 的转换标准，与 radar_convert_table、sleepace_convert_table 对齐。
// 对内/前端用 display（全小写）；SNOMED 仅在与外部系统对接时使用。
package converter

import "strings"

// 睡眠：以 radar/sleepace 表为准。SNOMED 248220002=awake, 248232005=light sleep, 248233000=deep sleep, 248234006=rem sleep
// 姿态：walk/walking=1, suspected-fall/suspected fall=2, sitting/sitting position=3, stand/standing/standing position=4, fall=5, lying/lying position=6

// SleepStage 将 sleep_status（display 或 SNOMED）转为阶段数：1=awake, 2=light, 4=deep；rem→2
func SleepStage(s string) int {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "248220002", "awake":
		return 1
	case "248232005", "248218005", "248220003", "light sleep", "248234006", "rem sleep":
		return 2
	case "248233000", "248221004", "deep sleep":
		return 4
	default:
		return 0
	}
}

// SleepStateDisplay 将 sleep_status（display 或 SNOMED）转为展示用 display（全小写）
func SleepStateDisplay(s string) *string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "248220002", "awake":
		return strPtr("awake")
	case "248232005", "248218005", "248220003", "light sleep":
		return strPtr("light sleep")
	case "248233000", "248221004", "deep sleep":
		return strPtr("deep sleep")
	case "248234006", "rem sleep":
		return strPtr("rem sleep")
	default:
		return nil
	}
}

// PostureStage 将 posture_code（display 或原始，不区分大小写）转为阶段数：1=walk, 2=suspected-fall, 3=sitting, 4=stand, 5=fall, 6=lying
// 兼容 radar_convert 的 display_en：walking, lying position, sitting position, standing position, fall, suspected fall, suspect sitting on ground, bed sit up 等
func PostureStage(s string) int {
	x := strings.ToLower(strings.TrimSpace(s))
	x = strings.ReplaceAll(x, " ", "")
	switch {
	case x == "walk" || x == "walking":
		return 1
	case x == "suspected-fall" || x == "suspectedfall" || x == "suspectsittingonground":
		return 2
	case x == "sitting" || x == "sittingposition":
		return 3
	case x == "stand" || x == "standing" || x == "standingposition":
		return 4
	case x == "fall":
		return 5
	case x == "lying" || x == "lyingposition":
		return 6
	default:
		return 0
	}
}

// BedStatus 将 bed_status 转为 0=in bed, 1=out of bed（入参不区分大小写；与 radar/sleepace 表一致：in bed/on_bed/enter_bed→0，not in bed/left bed/off_bed/left_bed→1）
func BedStatus(s string) int {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "on_bed", "enter_bed", "inbed", "in bed":
		return 0
	case "off_bed", "left_bed", "notinbed", "not in bed", "leftbed", "left bed":
		return 1
	default:
		return 0
	}
}

func strPtr(s string) *string { return &s }
