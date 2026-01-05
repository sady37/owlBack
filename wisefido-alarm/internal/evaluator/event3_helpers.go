package evaluator

import (
	"math"
	"strings"
	"time"
	"wisefido-alarm/internal/consumer"
	"wisefido-alarm/internal/models"
)

// checkBathroomRoomName 检查房间名是否匹配 bathroom（模块化，可复用）
// 匹配规则（不区分大小写）：
// - 完整词：Bathroom, restRoom
// - 子串：bath, rest, toilet
func checkBathroomRoomName(roomName string) bool {
	if roomName == "" {
		return false
	}
	roomNameLower := strings.ToLower(roomName)
	return strings.Contains(roomNameLower, "bathroom") ||
		strings.Contains(roomNameLower, "restroom") ||
		strings.Contains(roomNameLower, "bath") ||
		strings.Contains(roomNameLower, "rest") ||
		strings.Contains(roomNameLower, "toilet")
}

// checkStandingPosture 检查是否是站立状态（模块化，可复用）
// 返回：是否是站立状态
func checkStandingPosture(posture *models.Posture) bool {
	if posture == nil {
		return false
	}
	// 站立姿态码：102538003 (STANDING) 或 10904000 (ORTHOSTATIC)
	// TODO: 需要确认实际使用的 SNOMED 编码
	return posture.PostureCode == "102538003" || posture.PostureCode == "10904000"
}

// checkPositionChange 检查位置变化（模块化，可复用）
// 返回：位置变化距离（厘米），以及是否 >= 阈值
func checkPositionChange(
	state *consumer.Event3State,
	posture *models.Posture,
	threshold float64, // 阈值（厘米），默认 10cm
) (distance float64, exceeded bool) {
	if state.LastPosition == nil || posture == nil {
		return 0, false
	}

	if posture.PositionX == nil || posture.PositionY == nil {
		return 0, false
	}

	// 计算欧几里得距离
	dx := float64(*posture.PositionX) - state.LastPosition.X
	dy := float64(*posture.PositionY) - state.LastPosition.Y
	distance = math.Sqrt(dx*dx + dy*dy)

	exceeded = distance >= threshold
	return distance, exceeded
}

// updatePositionInEvent3State 更新 Event3State 中的位置信息（模块化，可复用）
func updatePositionInEvent3State(
	state *consumer.Event3State,
	posture *models.Posture,
) {
	if posture.PositionX != nil && posture.PositionY != nil {
		now := time.Now().Unix()
		if state.LastPosition == nil {
			state.LastPosition = &struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			}{}
		}
		state.LastPosition.X = float64(*posture.PositionX)
		state.LastPosition.Y = float64(*posture.PositionY)
		state.LastPositionTime = &now
	}
}

// resetTimers 重置计时器（当位置变化 >= 10cm 时）
func resetTimers(state *consumer.Event3State) {
	now := time.Now().Unix()
	// 重置站姿计时（如果之前是站立，重新开始计时）
	if state.StandingTime != nil {
		state.StandingTime = &now
	}
	// 重置位置不动计时
	if state.LastPositionTime != nil {
		state.LastPositionTime = &now
	}
	state.StandingDuration = 0
	state.StillDuration = 0
}

// calculateDurations 计算持续时间（模块化，可复用）
func calculateDurations(state *consumer.Event3State) (standingDuration, stillDuration int64) {
	now := time.Now().Unix()

	// 计算站姿持续时间
	if state.StandingTime != nil {
		standingDuration = now - *state.StandingTime
	}

	// 计算位置不动持续时间
	if state.LastPositionTime != nil {
		stillDuration = now - *state.LastPositionTime
	}

	return standingDuration, stillDuration
}

// findTrackInBathroom 在 bathroom 中查找 track_id（模块化，可复用）
// 返回：找到的 posture 和是否找到
func findTrackInBathroom(
	realtimeData *models.RealtimeData,
	trackID string,
) (*models.Posture, bool) {
	for _, posture := range realtimeData.Postures {
		if posture.TrackingID == trackID {
			return &posture, true
		}
	}
	return nil, false
}

// selectTrackIDInBathroom 选择要跟踪的 track_id（模块化，可复用）
// 返回：选中的 track_id，如果没有找到返回空字符串
func selectTrackIDInBathroom(realtimeData *models.RealtimeData) string {
	// 选择第一个 track_id（仅1个 track_id 时）
	if len(realtimeData.Postures) == 1 {
		return realtimeData.Postures[0].TrackingID
	}
	return ""
}

