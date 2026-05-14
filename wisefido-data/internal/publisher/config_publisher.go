package publisher

import (
	"context"
	"database/sql"
	"fmt"
	"net/netip"

	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ConfigPublisher 配置变更消息发布器
// 统一管理所有 config:* 消息的发送
type ConfigPublisher struct {
	redisClient *redis.Client
	logger      *zap.Logger
	db          *sql.DB // 反查 device_ipv6 用 (PublishAlarmDeviceMessage 内部 lookup)
}

// NewConfigPublisher 创建配置变更消息发布器
func NewConfigPublisher(redisClient *redis.Client, logger *zap.Logger) *ConfigPublisher {
	return &ConfigPublisher{
		redisClient: redisClient,
		logger:      logger,
	}
}

// SetDB 注入 DB 连接（用于内部 device_id → device_ipv6 反查）
func (p *ConfigPublisher) SetDB(db *sql.DB) {
	p.db = db
}

// PublishAlarmProcessMessage 发送报警处理消息到 config:alarmProcess:stream，供 cardagg 消费用于更新告警显示。
//
// device_ipv6 单程票后 cardID 是路由 key，device 字段不再传（cardagg 消费侧也只读 cardID/alarmLevel/alarmType/eventID）。
// tenantID 入参保留作日志，不入 payload。
func (p *ConfigPublisher) PublishAlarmProcessMessage(
	ctx context.Context,
	tenantID, cardID, deviceID, alarmLevel, alarmType, processType, eventID string,
	alarmTimestamp int64,
) error {
	_ = deviceID // 单程票后不入 payload；保留入参兼容 caller 签名
	alarmProcessMsg := rediscommon.BuildAlarmProcessMessage(
		"wisefido-data",
		cardID,
		alarmLevel,
		alarmType,
		processType, // 如 "ack"
		eventID,
		alarmTimestamp,
	)

	// 发布到 config:alarmProcess:stream
	streamName := rediscommon.StreamConfigAlarmProcess.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigAlarmProcess, nil)

	streamID, err := rediscommon.PublishJSONToStream(
		ctx,
		p.redisClient,
		streamName,
		alarmProcessMsg,
		int64(maxLen),
		retentionSeconds,
	)

	if err != nil {
		p.logger.Error("Failed to publish alarm process message",
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.String("alarm_level", alarmLevel),
			zap.String("process_type", processType),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish alarm process message: %w", err)
	}

	p.logger.Info("Published alarm process message",
		zap.String("tenant_id", tenantID),
		zap.String("card_id", cardID),
		zap.String("alarm_level", alarmLevel),
		zap.String("alarm_type", alarmType),
		zap.String("process_type", processType),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}

// PublishAlarmDeviceMessage 发送设备告警配置变更消息到 config:alarmDevice:stream，供 cardagg 消费用于失效 enablement 缓存。
//
// device_ipv6 单程票后 addr 是 cache key；deviceID/deviceUID 入参保留兼容签名 (R-002 外部 API caller 仍传 UUID)，
// 内部反查 device_ipv6 后填 payload。
func (p *ConfigPublisher) PublishAlarmDeviceMessage(
	ctx context.Context,
	tenantID, deviceID, deviceUID, settingType string,
	settingData map[string]interface{},
) error {
	addr := p.lookupDeviceAddr(ctx, deviceID)
	if !addr.IsValid() {
		p.logger.Warn("PublishAlarmDeviceMessage skipped: cannot resolve device_addr",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_uid", deviceUID))
		return nil
	}
	settingMsg := rediscommon.BuildAlarmDeviceMessage(
		"wisefido-data",
		addr,
		settingType,
		settingData,
	)

	streamName := rediscommon.StreamConfigAlarmDevice.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigAlarmDevice, nil)

	streamID, err := rediscommon.PublishJSONToStream(
		ctx,
		p.redisClient,
		streamName,
		settingMsg,
		int64(maxLen),
		retentionSeconds,
	)

	if err != nil {
		p.logger.Error("Failed to publish alarm device message",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", addr.String()),
			zap.String("setting_type", settingType),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish alarm device message: %w", err)
	}

	p.logger.Info("Published alarm device message",
		zap.String("tenant_id", tenantID),
		zap.String("device_addr", addr.String()),
		zap.String("setting_type", settingType),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}

// lookupDeviceAddr 由 deviceID UUID 反查 device_ipv6（device_ipv6 单程票内部 helper）。
// 失败返回零值 netip.Addr；caller 用 .IsValid() 判断。
func (p *ConfigPublisher) lookupDeviceAddr(ctx context.Context, deviceID string) netip.Addr {
	if p.db == nil || deviceID == "" {
		return netip.Addr{}
	}
	var addrStr string
	err := p.db.QueryRowContext(ctx,
		`SELECT host(d.device_ipv6)::text FROM devices d WHERE d.device_id = $1::uuid LIMIT 1`,
		deviceID).Scan(&addrStr)
	if err != nil {
		return netip.Addr{}
	}
	a, _ := netip.ParseAddr(addrStr)
	return a
}

// PublishCardChangeMessage 发送卡片变更消息到 config:card:stream
// 供 qinglan 和其他服务消费用于更新卡片相关配置
func (p *ConfigPublisher) PublishCardChangeMessage(
	ctx context.Context,
	tenantID, cardID, unitID, branchID string,
) error {
	return p.PublishCardChangeMessageWithExtra(ctx, tenantID, cardID, unitID, branchID, nil)
}

// PublishCardChangeMessageWithExtra 发送卡片变更消息到 config:card:stream（支持额外字段）
// 用于处理 card 变化导致的卡片变更（消息类型为 config.card）
// extraData 可以包含额外字段
func (p *ConfigPublisher) PublishCardChangeMessageWithExtra(
	ctx context.Context,
	tenantID, cardID, unitID, branchID string,
	extraData map[string]interface{},
) error {
	return p.PublishCardChangeMessageWithExtraAndType(ctx, tenantID, cardID, unitID, branchID, rediscommon.ConfigCardChanged, extraData)
}

// PublishCardChangeForDevice 发送 config.card，data 含 device_id、change_type；可选 deviceUIDs 写入 affected_device_uids 供网关精确失效。
func (p *ConfigPublisher) PublishCardChangeForDevice(ctx context.Context, tenantID, deviceID, changeType string, deviceUIDs ...string) error {
	if p == nil || deviceID == "" {
		return nil
	}
	extra := map[string]interface{}{
		"device_id":   deviceID,
		"change_type": changeType,
	}
	if u := compactDeviceUIDs(deviceUIDs); len(u) > 0 {
		extra["affected_device_uids"] = u
	}
	return p.PublishCardChangeMessageWithExtraAndType(ctx, tenantID, "", "", "", rediscommon.ConfigCardChanged, extra)
}

func compactDeviceUIDs(uids []string) []string {
	if len(uids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(uids))
	out := make([]string, 0, len(uids))
	for _, s := range uids {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// PublishConfigCardReset 发送 reset 通知到 config:card:stream。
func (p *ConfigPublisher) PublishConfigCardReset(ctx context.Context) error {
	resetMsg := rediscommon.BuildCardChangeMessageWithExtraAndType(
		"wisefido-data", "", "", rediscommon.ConfigCardChanged,
		map[string]interface{}{"op": "reset"},
	)
	streamName := rediscommon.StreamConfigCard.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigCard, nil)
	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, streamName, resetMsg, int64(maxLen), retentionSeconds)
	if err != nil {
		p.logger.Error("Failed to publish configCard reset", zap.String("stream", streamName), zap.Error(err))
		return fmt.Errorf("failed to publish configCard reset: %w", err)
	}
	p.logger.Info("Published configCard reset", zap.String("stream", streamName), zap.String("stream_id", streamID))
	return nil
}

// PublishCardChanged — 结构变更（card 增/删；v2 cards 表行的产生与消亡）。
//
// 当前实现：仍发 ConfigCardChanged stream 单 type（消费者已识别）；语义解耦体现在 extras.op：
//
//	op = "create" : card 新建（首个 device bind 触发 / admission 触发 active_bed 新建）
//	op = "delete" : card 删除（spatial_prefix 下所有 device unbind 触发；保留 resident_id 历史归属由 view 反查）
//
// 与 PublishCardResidentChanged 的区别：本方法表示 cards 表 INSERT/DELETE；resident 流入/流出
// 同一 card 不动 cards 行的 op，应该用 PublishCardResidentChanged。
//
// 详 [doc/cards_v2_migration_checklist.md] § 一.7。
func (p *ConfigPublisher) PublishCardChanged(
	ctx context.Context,
	tenantID, cardID, op, spatialPrefix, dnsShortName string,
) error {
	extra := map[string]interface{}{
		"op":             op,
		"spatial_prefix": spatialPrefix,
		"dns_short_name": dnsShortName,
	}
	return p.PublishCardChangeMessageWithExtraAndType(ctx, tenantID, cardID, "", "", rediscommon.ConfigCardChanged, extra)
}

// PublishCardResidentChanged — resident 在 cards.resident_id 指针的迁移（admission/discharge/transfer）。
//
// 不动 cards 表行本身；只动 resident_id 列。供 cardagg 失效 resident 视图缓存 + 历史归属反查侧 trigger。
//
//	op = "admission" : 空床 → resident（prevHoA=""）
//	op = "discharge" : resident → 空床（newHoA=""）
//	op = "transfer"  : 同一 resident 跨 prefix（prevHoA/newHoA 同 HoA，不同 spatial_prefix）
//
// 详 [doc/cards_v2_migration_checklist.md] § 一.7。
func (p *ConfigPublisher) PublishCardResidentChanged(
	ctx context.Context,
	tenantID, cardID, op, prevHoA, newHoA, spatialPrefix string,
) error {
	extra := map[string]interface{}{
		"op":               op,
		"prev_resident_id": prevHoA,
		"new_resident_id":  newHoA,
		"spatial_prefix":   spatialPrefix,
	}
	return p.PublishCardChangeMessageWithExtraAndType(ctx, tenantID, cardID, "", "", rediscommon.ConfigCardChanged, extra)
}

// PublishCardChangeMessageWithExtraAndType 发送卡片变更消息到 config:card:stream（支持额外字段和自定义 type 字段）
//
// device_ipv6 单程票后 owl-common 的 BuildCardChangeMessageWithExtraAndType 简化为 (source, cardID, spatialPrefix, messageType, extra)。
// tenantID/unitID/branchID 入参保留兼容 caller 签名；tenantID 走 extra 透传给 cardagg 用于按 tenant 失效；unitID/branchID 弃用。
func (p *ConfigPublisher) PublishCardChangeMessageWithExtraAndType(
	ctx context.Context,
	tenantID, cardID, unitID, branchID, messageType string,
	extraData map[string]interface{},
) error {
	if extraData == nil {
		extraData = make(map[string]interface{})
	}
	if tenantID != "" {
		extraData["tenant_id"] = tenantID
	}
	if unitID != "" {
		extraData["unit_id"] = unitID
	}
	if branchID != "" {
		extraData["branch_id"] = branchID
	}
	// 构建卡片变更消息（spatial_prefix 由 caller 在 extras 中提供，老 caller 暂传空）
	spatialPrefix, _ := extraData["spatial_prefix"].(string)
	cardMsg := rediscommon.BuildCardChangeMessageWithExtraAndType(
		"wisefido-data",
		cardID,
		spatialPrefix,
		messageType,
		extraData,
	)

	// 发布到 config:card:stream
	streamName := rediscommon.StreamConfigCard.Name
	maxLen, retentionSeconds := rediscommon.GetStreamConfig(rediscommon.StreamConfigCard, nil)

	streamID, err := rediscommon.PublishJSONToStream(
		ctx,
		p.redisClient,
		streamName,
		cardMsg,
		int64(maxLen),
		retentionSeconds,
	)

	if err != nil {
		p.logger.Error("Failed to publish card change message",
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.String("unit_id", unitID),
			zap.String("branch_id", branchID),
			zap.String("message_type", messageType),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish card change message: %w", err)
	}

	p.logger.Info("Published card change message",
		zap.String("tenant_id", tenantID),
		zap.String("card_id", cardID),
		zap.String("unit_id", unitID),
		zap.String("branch_id", branchID),
		zap.String("message_type", messageType),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)

	return nil
}
