package consumer

// PR1 (A8): sensor → cardagg 回流通道。
//
// sensor 决策出"该报警 / pending 起 / pending 取消"后通过本 helper publish 到 iot:alarm:stream，
// cardagg 现有 alarm_handler.go 路由完成实际写库 / 推送 / Redis pending 增删。原则：
//
//   - sensor 内不写 alarm_events DB，不写 cardagg 内部 Redis pending hash
//   - 不绕过 cardagg 的 enablement gate（cardagg 端按 producer 判定信任级别）
//   - envelope.event_status 协议沿用：start / pending_arm / pending_cancel
//
// "pending_arm" / "pending_cancel" 是 PR1 新增的 sensor→cardagg sentinel；
// 配套 cardagg alarm_handler.go A9 分支接收，调用现有 AddPendingAlarm / RemovePendingAlarm。

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	rediscommon "owl-common/redis"

	redislib "github.com/go-redis/redis/v8"
)

// EventStatusStart / PendingArm / PendingCancel envelope.event_status 的 PR1 sentinel。
const (
	EventStatusStart         = "start"
	EventStatusPendingArm    = "pending_arm"
	EventStatusPendingCancel = "pending_cancel"
)

// AgentIdentity sensor 进程作为 platform agent 的 IPv6 + UID 身份。
// 详见 owlBack/doc/platform_agent_addressing.md §3。
type AgentIdentity struct {
	IPv6      netip.Addr // /128, slot fd00:0:fff1::/48
	AgentName string     // 'wisefido-sensor'
}

// AlarmBackChannel sensor → cardagg 回流 publisher。
type AlarmBackChannel struct {
	client   *redislib.Client
	identity AgentIdentity
	seqKey   string // Redis INCR key for envelope.sequence_number
}

// NewAlarmBackChannel 构造 publisher。identity 必须有效（IPv6 非零）；否则 publish 时退化为 anonymous。
func NewAlarmBackChannel(client *redislib.Client, identity AgentIdentity) *AlarmBackChannel {
	seqKey := ""
	if identity.IPv6.IsValid() {
		name := identity.AgentName
		if name == "" {
			name = "agent"
		}
		seqKey = fmt.Sprintf("%s:seq:%s", name, identity.IPv6.String())
	}
	return &AlarmBackChannel{
		client:   client,
		identity: identity,
		seqKey:   seqKey,
	}
}

// nextSeq 从 Redis INCR 取下一个 envelope sequence_number；失败返回 0（degrade 不阻断 alarm publish）。
// key 跨重启单调（不重置），让 parent_span/trace_id 在 process restart 后仍唯一。
func (a *AlarmBackChannel) nextSeq(ctx context.Context) int64 {
	if a == nil || a.client == nil || a.seqKey == "" {
		return 0
	}
	v, err := a.client.Incr(ctx, a.seqKey).Result()
	if err != nil {
		return 0 // degrade（不能因 Redis 暂时不可用阻断 alarm 链）
	}
	return v
}

// PublishAlarmFire 触发即时 alarm。cardagg alarm_handler 收到后走 PersistAlarmAndPublish 路径。
// triggerData 是触发上下文 JSON（可 nil），落到 alarm_events.trigger_data。
func (a *AlarmBackChannel) PublishAlarmFire(
	ctx context.Context,
	deviceAddr netip.Addr,
	subjectEntity, eventName, level string,
	tsMs int64,
	triggerData map[string]interface{},
) (string, error) {
	if a == nil || a.client == nil {
		return "", fmt.Errorf("AlarmBackChannel not initialized")
	}
	data := mergeTriggerData(eventName, EventStatusStart, level, 0, "", 0, triggerData)
	return a.publishAlarm(ctx, deviceAddr, subjectEntity, eventName, tsMs, data)
}

// PublishPendingArm 起 pending（Suspected→Real 升级 / LeftBed duration_sec timer / Stay 等）。
// 接收方为 cardagg alarm_handler.go A9 分支：根据 event_status=pending_arm 调 AddPendingAlarm。
//
//	durationSec  到期时间（秒）
//	upgradeTo    到期后升格的 eventName；空表示直接 fire 当前 eventName
//	eventSinceMs 触发时间戳，cardagg 用来计算到期点；0 时填 tsMs
func (a *AlarmBackChannel) PublishPendingArm(
	ctx context.Context,
	deviceAddr netip.Addr,
	subjectEntity, eventName, level string,
	durationSec int,
	upgradeTo string,
	eventSinceMs int64,
	tsMs int64,
	triggerData map[string]interface{},
) (string, error) {
	if a == nil || a.client == nil {
		return "", fmt.Errorf("AlarmBackChannel not initialized")
	}
	if eventSinceMs == 0 {
		eventSinceMs = tsMs
	}
	data := mergeTriggerData(eventName, EventStatusPendingArm, level, durationSec, upgradeTo, eventSinceMs, triggerData)
	return a.publishAlarm(ctx, deviceAddr, subjectEntity, eventName, tsMs, data)
}

// PublishPendingCancel 取消 pending（如 InBed 抵消 LeftBed pending）。
// cardagg alarm_handler.go A9 分支据此调 RemovePendingAlarm。
func (a *AlarmBackChannel) PublishPendingCancel(
	ctx context.Context,
	deviceAddr netip.Addr,
	subjectEntity, eventName string,
	tsMs int64,
) (string, error) {
	if a == nil || a.client == nil {
		return "", fmt.Errorf("AlarmBackChannel not initialized")
	}
	data := mergeTriggerData(eventName, EventStatusPendingCancel, "", 0, "", 0, nil)
	return a.publishAlarm(ctx, deviceAddr, subjectEntity, eventName, tsMs, data)
}

func (a *AlarmBackChannel) publishAlarm(
	ctx context.Context,
	deviceAddr netip.Addr,
	subjectEntity, eventName string,
	tsMs int64,
	data map[string]interface{},
) (string, error) {
	if tsMs == 0 {
		tsMs = time.Now().UnixMilli()
	}
	dataValueJSON, err := json.Marshal([]interface{}{data})
	if err != nil {
		return "", fmt.Errorf("marshal dataValue: %w", err)
	}
	// envelope.producer = sensor 的 platform agent /128（per platform_agent_addressing.md §4.1）；
	// 未配置 identity 时 fallback 空串 — cardagg 端会 NULL 落库（兼容首次部署 .env 未配的过渡）
	producerStr := ""
	if a.identity.IPv6.IsValid() {
		producerStr = a.identity.IPv6.String()
	}
	seq := a.nextSeq(ctx)
	values := map[string]interface{}{
		"producer":        producerStr,
		"subject_entity":  subjectEntity,
		"sequence_number": fmt.Sprintf("%d", seq),
		"device_addr":     deviceAddr.String(),
		"device_type":     "",
		"timestamp":       fmt.Sprintf("%d", tsMs),
		"topic_type":      "alarm",
		"category":        eventName,
		rediscommon.DataValueKey: string(dataValueJSON),
	}
	return rediscommon.PublishToStream(ctx, a.client, rediscommon.StreamAlarm.Name, values,
		rediscommon.StreamAlarm.MaxLen, rediscommon.StreamAlarm.RetentionSeconds)
}

// mergeTriggerData 将 envelope 必带字段（event_name / event_status / level / duration / upgrade_to / event_since_ms）
// 合并到 trigger data map，cardagg alarm_handler 读取时按 key 取值。
func mergeTriggerData(eventName, eventStatus, level string, durationSec int, upgradeTo string, eventSinceMs int64, extra map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, 6+len(extra))
	for k, v := range extra {
		out[k] = v
	}
	out["event_name"] = eventName
	if eventStatus != "" {
		out["event_status"] = eventStatus
	}
	if level != "" {
		out["alarm_level"] = level
	}
	if durationSec > 0 {
		out["duration_sec"] = durationSec
	}
	if upgradeTo != "" {
		out["upgrade_to"] = upgradeTo
	}
	if eventSinceMs > 0 {
		out["event_since_ms"] = eventSinceMs
	}
	return out
}
