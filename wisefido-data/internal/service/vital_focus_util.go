package service

import (
	"encoding/json"

	"wisefido-data/internal/models"
)

// decodeAndNormalizeFullCard 解析并规范化卡片数据（从 Handler 复制）
func decodeAndNormalizeFullCard(raw string) (models.VitalFocusCard, bool) {
	// 先用 map 解析，避免类型不一致导致整体 unmarshal 失败
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return models.VitalFocusCard{}, false
	}

	// 再把 map 转回 json，然后 unmarshal 到目标模型（容忍少数字段缺失）
	b, err := json.Marshal(m)
	if err != nil {
		return models.VitalFocusCard{}, false
	}
	var card models.VitalFocusCard
	if err := json.Unmarshal(b, &card); err != nil {
		// 兜底：如果模型解析失败，就直接返回失败
		return models.VitalFocusCard{}, false
	}

	// residents：确保 last_name 有值（前端类型标注为必填）
	for i := range card.Residents {
		if card.Residents[i].LastName == "" {
			if card.Residents[i].Nickname != "" {
				card.Residents[i].LastName = card.Residents[i].Nickname
			} else {
				card.Residents[i].LastName = "-"
			}
		}
	}

	// devices：device_type 规范化为 number（sleepace=1, radar=2）
	for i := range card.Devices {
		switch v := card.Devices[i].DeviceType.(type) {
		case string:
			card.Devices[i].DeviceType = deviceTypeToNumber(v)
		case float64:
			// json number -> float64
			card.Devices[i].DeviceType = int(v)
		default:
			// keep as-is
		}
	}

	// heart_source/breath_source：如果被写成 Sleepace/Radar，规范为 s/r/-
	if card.HeartSource != "" {
		card.HeartSource = normalizeSource(card.HeartSource)
	}
	if card.BreathSource != "" {
		card.BreathSource = normalizeSource(card.BreathSource)
	}

	return card, true
}

func deviceTypeToNumber(s string) int {
	switch s {
	case "Sleepace", "SleepPad", "Sleepad", "SleepAd":
		return 1
	case "Radar":
		return 2
	default:
		return 0
	}
}

func normalizeSource(s string) string {
	switch s {
	case "s", "r", "-":
		return s
	case "Sleepace", "SleepPad":
		return "s"
	case "Radar":
		return "r"
	default:
		return "-"
	}
}

