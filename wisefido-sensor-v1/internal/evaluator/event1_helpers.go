package evaluator

import (
	"context"
	"fmt"
	"math"
	"time"
	"wisefido-sensor-v1/internal/consumer"
	"wisefido-sensor-v1/internal/models"

	"go.uber.org/zap"
)

// checkExitConditions 检查退出条件（模块化，可复用）
// 返回 true 表示满足退出条件，应该退出检测
func checkExitConditions(
	realtimeData *models.RealtimeData,
	state *consumer.Event1State,
	posture *models.Posture,
) bool {
	// 退出条件1：SleepPad 有 HR/RR
	if realtimeData.Heart != nil || realtimeData.Breath != nil {
		return true
	}

	// 退出条件2：SleepPad 有上床事件
	if realtimeData.BedStatus != nil && *realtimeData.BedStatus == "on_bed" {
		return true
	}

	// 退出条件3：姿态非 lying
	// lying 姿态码：248218002 (Lying)
	// 注意：如果 posture 为 nil，无法判断姿态，不退出
	if posture != nil && posture.PostureCode != "248218002" {
		return true // 姿态非 lying，退出
	}

	// 退出条件4：有移动（位置变化超过30cm）
	if state.LastPosition != nil && posture != nil {
		if posture.PositionX != nil && posture.PositionY != nil {
			dx := float64(*posture.PositionX - state.LastPosition.X)
			dy := float64(*posture.PositionY - state.LastPosition.Y)
			distance := math.Sqrt(dx*dx + dy*dy)
			if distance > 30 { // 移动超过30cm
				return true // 有移动，退出
			}
		}
	}

	return false
}

// checkTrackDisappeared 检查 track_id 是否消失（单独的function，可复用）
// 返回：是否应该触发 Fall 报警，以及是否需要继续监控
func checkTrackDisappeared(
	ctx context.Context,
	realtimeData *models.RealtimeData,
	state *consumer.Event1State,
	logger *zap.Logger,
) (shouldAlarm bool, shouldContinue bool) {
	// 检查 track_id 是否在当前数据中存在
	trackExists := false
	for _, posture := range realtimeData.Postures {
		if posture.TrackingID == state.TrackID {
			trackExists = true
			break
		}
	}

	// 如果 track 消失
	if !trackExists {
		now := time.Now()

		// 如果之前没有记录消失时间，记录当前时间
		if state.TrackDisappearTime == nil {
			disappearTime := now.Unix()
			state.TrackDisappearTime = &disappearTime
			logger.Info("Track disappeared, starting 30s monitoring",
				zap.String("track_id", state.TrackID),
				zap.Int64("disappear_time", disappearTime),
			)
			// 第一次消失，不立即报警，等待30秒
			return false, true // 不报警，继续监控
		}

		// 如果已经记录了消失时间，检查是否超过30秒
		disappearTime := time.Unix(*state.TrackDisappearTime, 0)
		elapsed := now.Sub(disappearTime)

		if elapsed >= 30*time.Second {
			// 检查当前 RealtimeData 中是否有新的可移动 track 出现（除了消失的 track_id）
			hasNewTrack := false
			for _, posture := range realtimeData.Postures {
				if posture.TrackingID != state.TrackID {
					// 有新的 track 出现（不是消失的 track_id）
					hasNewTrack = true
					break
				}
			}

			if !hasNewTrack {
				// 30秒内没有新的可移动 track，确认跌倒
				logger.Info("Track disappeared for 30s with no new track, confirming fall",
					zap.String("track_id", state.TrackID),
					zap.Duration("elapsed", elapsed),
				)
				return true, false // 报警，不继续监控
			} else {
				// 有新的可移动 track，说明人正常移动了，退出检测
				logger.Info("New movable track appeared, exiting fall detection",
					zap.String("track_id", state.TrackID),
				)
				state.NewTrackAppeared = true
				return false, false // 不报警，不继续监控（退出）
			}
		}

		// 还在30秒内，继续监控
		return false, true // 不报警，继续监控
	} else {
		// track_id 仍然存在，清除消失时间记录
		if state.TrackDisappearTime != nil {
			state.TrackDisappearTime = nil
			state.NewTrackAppeared = false
		}
	}

	return false, true // 不报警，继续监控
}

// checkHeightDrop 检查高度是否降低（模块化，可复用）
// 返回：是否高度降低，以及降低的高度值
func checkHeightDrop(
	state *consumer.Event1State,
	posture *models.Posture,
) (bool, float64) {
	if state.LyingHeight == nil || posture.PositionZ == nil {
		return false, 0
	}

	currentHeight := float64(*posture.PositionZ)
	lyingHeight := *state.LyingHeight

	// 高度降低判断条件：
	// 1. current_height < lying_height - 30cm（相对于基线降低超过30cm）
	// 2. 或者 current_height <= 30cm（绝对高度低于30cm）
	heightDropped := (currentHeight < lyingHeight-30.0) || (currentHeight <= 30.0)

	if heightDropped {
		dropAmount := lyingHeight - currentHeight
		return true, dropAmount
	}

	return false, 0
}

// checkLyingDuration 检查 lying 状态维持时间（模块化，可复用）
// 从 stream 中获取 lying 状态开始时间，计算维持了多少秒
func checkLyingDuration(
	state *consumer.Event1State,
	realtimeData *models.RealtimeData,
	posture *models.Posture,
) (int64, error) {
	// 检查是否是 lying 状态
	if posture == nil || posture.PostureCode != "248218002" {
		return 0, fmt.Errorf("not in lying state")
	}

	// 如果状态中有 LyingTime，使用它
	if state.LyingTime != nil {
		lyingTime := time.Unix(*state.LyingTime, 0)
		duration := time.Since(lyingTime).Seconds()
		return int64(duration), nil
	}

	// 如果没有 LyingTime，从 realtimeData 的时间戳计算
	// 注意：这里需要从 stream 中获取 lying 状态开始时间
	// 当前实现：使用当前时间作为近似值（不准确，需要改进）
	return 0, fmt.Errorf("lying time not available in state")
}

// findTrackInBedArea 在床区域查找 track_id（模块化，可复用）
// 返回：找到的 posture 和是否找到
func findTrackInBedArea(
	realtimeData *models.RealtimeData,
	bedAreaID int,
	trackID string,
) (*models.Posture, bool) {
	for _, posture := range realtimeData.Postures {
		if posture.TrackingID == trackID {
			// 验证 area_id 是否匹配床区域
			if posture.AreaID != nil && *posture.AreaID == bedAreaID {
				return &posture, true
			}
		}
	}
	return nil, false
}

// selectTrackID 选择要跟踪的 track_id（模块化，可复用）
// 返回：选中的 track_id，如果没有找到返回空字符串
func selectTrackID(
	realtimeData *models.RealtimeData,
	bedAreaID int,
) string {
	// 选择在床区域的第一个 track_id
	for _, posture := range realtimeData.Postures {
		if posture.AreaID != nil && *posture.AreaID == bedAreaID {
			return posture.TrackingID
		}
	}
	return ""
}

// updatePositionInState 更新状态中的位置信息（模块化，可复用）
func updatePositionInState(
	state *consumer.Event1State,
	posture *models.Posture,
) {
	if posture.PositionX != nil && posture.PositionY != nil {
		now := time.Now().Unix()
		if state.LastPosition == nil {
			state.LastPosition = &struct {
				X int `json:"x"`
				Y int `json:"y"`
			}{}
		}
		state.LastPosition.X = *posture.PositionX
		state.LastPosition.Y = *posture.PositionY
		state.LastPositionTime = &now
	}
}

