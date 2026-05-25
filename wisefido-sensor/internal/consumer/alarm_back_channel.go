package consumer

// sensor → cardagg alarm 回流通道。
//
// sensor 决策出"确认 alarm"后通过本 helper publish 到 iot:alarm:stream，cardagg
// AlarmRouter 完成实际写库 / 推送。原则：
//
//   - sensor 内不写 alarm_events DB，不写 cardagg 内部 Redis state
//   - **sensor 在此处统一做 enablement gate**（按 device /128 LPM 查 spatial_config alarm.cloud_config）。
//     未启用的 alarm 在 sensor 源头丢弃，不发往 cardagg；cardagg platform-agent trust bypass 保留，
//     视为 second-line（产 + 消两端都不漏 + 不重复 gate）。
//   - envelope.event_status = "start"（确认态）
//
// **state-anchored zonealarm**：sensor.Supervisor 不订阅 ZoneEvent，Tick 周期从 StreamPublisher
// 拿 RoomState/BedState 快照按 anchor + KeepCounting 条件 + 阈值判 fire；fire-once-until-
// anchor-invalidates。本 channel 只发 confirmed PublishAlarmFire；PR1 时代的
// PublishPendingArm / PublishPendingCancel sentinel 已删（cardagg 端从未消费；E 收口 dead code）。

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	rediscommon "owl-common/redis"

	redislib "github.com/go-redis/redis/v8"
)

// EnablementGate sensor 决定 alarm 是否启用的回调（main wiring 传 closure 注入；
// 接口而非 *service.AlarmEnablementCache，避免 consumer→service 导入环）。
type EnablementGate func(ctx context.Context, deviceAddr, alarmType string) bool

// EventStatusStart envelope.event_status 标记本条是 alarm 起始事件（confirmed）。
// state-anchored zonealarm 后无 pending 概念，本 channel 只发 confirmed alarm；
// PR1 时代的 PublishPendingArm/PublishPendingCancel 已删（E 收口 dead code）。
const EventStatusStart = "start"

// AgentIdentity sensor 进程作为 platform agent 的 IPv6 + UID 身份。
// 详见 owlBack/doc/platform_agent_addressing.md §3。
type AgentIdentity struct {
	IPv6      netip.Addr // /128, slot fd00:0:fff1::/48
	AgentName string     // 'wisefido-sensor'
}

// AlarmBackChannel sensor → cardagg 回流 publisher。
//
// 内部持 enablement gate callback：所有 publish 在源头按 device /128 + alarmType 查使能；
// 未启用直接 drop（不发流不入 cardagg）。gate=nil 时退化为 always-allow
// （用于早期 boot 期 / 测试，生产 wiring 应注入）。
type AlarmBackChannel struct {
	client   *redislib.Client
	identity AgentIdentity
	seqKey   string // Redis INCR key for envelope.sequence_number
	gateFn   EnablementGate
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

// SetEnablement 注入 enablement gate callback（sensor main wiring 调）。
// 不调时退化为 always-allow（兼容老 wiring + 单测）。
func (a *AlarmBackChannel) SetEnablement(gate EnablementGate) {
	if a == nil {
		return
	}
	a.gateFn = gate
}

// gate 在 publish 入口查 device 的 alarm 使能；未启用返回 false → caller 直接退出。
// gateFn=nil 或 deviceAddr 无效时返回 true（allow，避免 boot 期所有 alarm 静默丢）。
//
// 注：deviceAddr 是 alarm 归因的物理设备 /128；alarmType 已是 owl-common/alarm 常量字符串。
// 设备类报警（Offline / SignalPoor 等）也走这条 gate —— 客户禁用了就不该报；若需强制审计
// 不受 enablement 影响的报警，由 caller 不走 BackChannel 而走另一专用 publisher。
func (a *AlarmBackChannel) gate(ctx context.Context, deviceAddr netip.Addr, alarmType string) bool {
	if a == nil || a.gateFn == nil {
		return true
	}
	if !deviceAddr.IsValid() {
		return true // 无 device 归因时不阻断（兜底兼容）
	}
	return a.gateFn(ctx, deviceAddr.String(), alarmType)
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
	if !a.gate(ctx, deviceAddr, eventName) {
		return "", nil // 未启用 → 静默丢弃
	}
	data := mergeTriggerData(eventName, EventStatusStart, level, triggerData)
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

// mergeTriggerData 将 envelope 必带字段（event_name / event_status / alarm_level）
// 合并到 trigger data map，cardagg alarm_handler 读取时按 key 取值。
//
// duration_sec / upgrade_to / event_since_ms 三字段 PR1 时代为 PublishPendingArm sentinel 服务，
// v2 zonealarm.Supervisor 内部维护 pending（不外发 sentinel）后这些字段无 producer → 一并删除。
func mergeTriggerData(eventName, eventStatus, level string, extra map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, 3+len(extra))
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
	return out
}
