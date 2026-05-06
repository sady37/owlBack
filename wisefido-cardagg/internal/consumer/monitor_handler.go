package consumer

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"owl-common/card"
	"owl-common/observation"
	"owl-common/redis"
	"wisefido-cardagg/internal/service"

	"go.uber.org/zap"
)

type MonitorHandler struct {
	buffer      *service.MonitorBuffer
	writer      *card.Writer
	metaCache   *service.DeviceMetaCache
	bedCoord    *service.BedEventCoordinator
	state       *service.StateService
	alarms      *service.AlarmService
	aiOverrides *service.AIOverrideCache // PR6: AI 派生 verdict 合并 + tid=88 清理
	logger      *zap.Logger
}

const (
	// MonitorFieldTTL: PruneFields 删除 buffer 中超过此时长未刷新的字段。设为 12s 以覆盖 sleepace 10s 上报间隔，避免 HR/RR 在两帧之间被清空导致前端"假断流"。
	MonitorFieldTTL = 12000
	// MonitorMaxInboundAgeMs: 入站消息时间戳超过此时长则丢弃，防 MQTT 积压陈旧数据 / 设备时钟漂移污染。与 buffer TTL 解耦：buffer 容忍 12s，但拒绝接收 > 6s 旧的新消息。
	MonitorMaxInboundAgeMs = 6000
)

// 实时流不做报警处理：雷达/sleepace 已有阀值，立即报警型由网关直接发 iot:alarm:stream。
// bedCoord/state/alarms 可选：非 nil 时每条 monitor 写入后尝试解析该卡 bed pending（对齐 2～3s 采样，不等 1s Tick）。
func NewMonitorHandler(buffer *service.MonitorBuffer, writer *card.Writer, metaCache *service.DeviceMetaCache, bedCoord *service.BedEventCoordinator, state *service.StateService, alarms *service.AlarmService, aiOverrides *service.AIOverrideCache, logger *zap.Logger) *MonitorHandler {
	return &MonitorHandler{buffer: buffer, writer: writer, metaCache: metaCache, bedCoord: bedCoord, state: state, alarms: alarms, aiOverrides: aiOverrides, logger: logger}
}

// RunLoop 与 monitor 相关的定时：1 秒发 snap，6 秒 PruneFields。不包含 derive。
// 写优先：PruneFields 用写锁、Flush 用读锁；两次加锁之间释放锁，Handle.Write 可插入。
func (h *MonitorHandler) RunLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	tick := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick++
			nowMs := time.Now().UnixMilli()
			if tick%4 == 0 {
				h.buffer.PruneFields(nowMs, MonitorFieldTTL)
				// 获取所有活跃的卡片ID，然后为每个卡片单独处理在线状态
				activeCardIDs := h.buffer.ActiveCardIDs()
				for _, cardID := range activeCardIDs {
					h.buffer.AdvancePruneTick(cardID)
				}
			}
			// Flush 用 RLock，与 Write 不互斥；PruneFields 与 Write 之间已释放锁，写可优先
			snapshots := h.buffer.Flush(nowMs)
			for _, snap := range snapshots {
				if err := publishRealtimeSnap(ctx, h.writer, snap, h.logger); err != nil {
					h.logger.Warn("publish realtime", zap.String("cid", snap.CardID), zap.Error(err))
				}
			}
		}
	}
}

func publishRealtimeSnap(ctx context.Context, w *card.Writer, snap service.CardSnapshot, logger *zap.Logger) error {
	data := make(map[string]any, len(snap.Devices))
	for _, dev := range snap.Devices {
		data[dev.DeviceID] = dev.Tracks
	}
	if err := w.PublishMonitor(ctx, snap.CardID, "", data); err != nil {
		return err
	}
	return nil
}

func (h *MonitorHandler) Handle(ctx context.Context, msg interface{}) error {
	raw, ok := msg.(map[string]interface{})
	if !ok {
		return nil
	}

	m, err := ParseMessage(raw)
	if err != nil {
		h.logger.Warn("parse monitor", zap.Error(err))
		return nil
	}

	// #region agent log
	dbg := strings.Contains(strings.ToLower(m.SubjectEntity), "d2ba029d") || m.DeviceUID == "BM87224700978" ||
		strings.EqualFold(m.DeviceUID, "E598A2ACD523") || strings.EqualFold(m.DeviceUID, "E598A2ACD5F7") ||
		strings.EqualFold(m.DeviceID, "E598A2ACD523") || strings.EqualFold(m.DeviceID, "E598A2ACD5F7")
	if dbg {
		agentDebugLog("H0", "monitor_handler.Handle:parsed", "monitor parsed", map[string]any{"cardID": m.SubjectEntity, "deviceID": m.DeviceID, "deviceUID": m.DeviceUID, "msgTs": m.Timestamp})
	}
	// #endregion

	if m.DeviceUID == "" {
		// #region agent log
		if dbg {
			agentDebugLog("H1", "monitor_handler.Handle:drop", "empty device_uid", map[string]any{"cardID": m.SubjectEntity})
		}
		// #endregion
		return nil
	}
	deviceKey := h.metaCache.ResolveDeviceID(ctx, m.SubjectEntity, m.DeviceID, m.DeviceUID)
	if deviceKey == "" {
		// #region agent log
		if dbg {
			agentDebugLog("H1", "monitor_handler.Handle:drop", "ResolveDeviceID empty", map[string]any{"cardID": m.SubjectEntity, "deviceID": m.DeviceID, "deviceUID": m.DeviceUID})
		}
		// #endregion
		return nil
	}

	nowMs := time.Now().UnixMilli()
	age := nowMs - m.Timestamp
	if age > MonitorMaxInboundAgeMs {
		// #region agent log
		if dbg {
			agentDebugLog("H2", "monitor_handler.Handle:drop", "stale timestamp", map[string]any{"ageMs": age, "limitMs": MonitorMaxInboundAgeMs, "msgTs": m.Timestamp, "nowMs": nowMs})
		}
		// #endregion
		return nil
	}

	fields := redis.FirstDataValue(m.DataValue)
	if fields == nil {
		// #region agent log
		if dbg {
			agentDebugLog("H3", "monitor_handler.Handle:drop", "FirstDataValue nil", map[string]any{"cardID": m.SubjectEntity, "deviceKey": deviceKey})
		}
		// #endregion
		return nil
	}
	trackID := resolveTrackID(fields)
	// #region agent log
	if dbg {
		agentDebugLog("H4", "monitor_handler.Handle:write", "buffer Write", map[string]any{"cardID": m.SubjectEntity, "deviceKey": deviceKey, "trackID": trackID, "trackInvalid": trackID == observation.TrackInvalid})
	}
	// #endregion
	// PR6: AI override 缓存合并 / tid=88 清理（仅 radar 来源；sleepad 不参与 ghost verdict）
	if h.aiOverrides != nil && m.DeviceUID != "" {
		if trackID == observation.TrackInvalid {
			// firmware no-target heartbeat → 该 device 所有 track 都退场，清空 verdicts
			h.aiOverrides.ClearDevice(m.DeviceUID)
			// SleepStage 同步清：device 退场即"事件源消失"，stale Awake 必须清掉
			if h.state != nil && m.SubjectEntity != "" {
				_ = h.state.ClearBedStateSleepStage(ctx, m.SubjectEntity)
			}
		} else {
			// 有效 track → 合并 AI verdict 到 fields（release 模式才真覆写 track_confidence）
			h.aiOverrides.Apply(m.DeviceUID, trackID, fields)
		}
	}
	h.buffer.Write(m.SubjectEntity, deviceKey, strconv.Itoa(trackID), fields, m.Timestamp)
	if h.bedCoord != nil && h.state != nil {
		h.bedCoord.TryResolveAfterMonitorWrite(ctx, h.state, h.alarms, h.metaCache, h.buffer, m.SubjectEntity, h.logger)
	}
	return nil
}

// resolveTrackID 从 payload 取 track_id。缺省、类型无法解析或与 88 一样走无效桶，避免 string/缺字段误落 key "0" 与真人轨迹合并成双目标。
func resolveTrackID(fields map[string]interface{}) int {
	if fields == nil {
		return observation.TrackInvalid
	}
	v, ok := fields["track_id"]
	if !ok || v == nil {
		return observation.TrackInvalid
	}
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			return observation.TrackInvalid
		}
		return int(n)
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return observation.TrackInvalid
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return observation.TrackInvalid
		}
		return n
	default:
		return observation.TrackInvalid
	}
}
