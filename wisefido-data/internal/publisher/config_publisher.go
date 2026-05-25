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

// ConfigPublisher — config:* 流的统一 publish 入口。
type ConfigPublisher struct {
	redisClient *redis.Client
	logger      *zap.Logger
	db          *sql.DB
}

func NewConfigPublisher(redisClient *redis.Client, logger *zap.Logger) *ConfigPublisher {
	return &ConfigPublisher{
		redisClient: redisClient,
		logger:      logger,
	}
}

func (p *ConfigPublisher) SetDB(db *sql.DB) {
	p.db = db
}

// PublishConfigChanged — config:card:stream 唯一 publish 入口（统一替代历史 5 个 variant）。
//
//	op = "reset"  → startup 重对账信号；cards/devAddrs/devUIDs 应为空
//	op = "delete" → 卡物理删除（cardagg DeleteCardState + DDNS unregister 等清理）
//	op = "update" → 其他变化（spatial / monitor / bind / resident / name / alarm config）
//
// scope 三数组按 affected 范围给：
//   - cards       : card_id CIDR  → cardagg.metaCache.Remove + UnitPicker.InvalidateUnit
//   - deviceAddrs : device_addr /128 → cardagg.enablement.InvalidateDevices
//   - deviceUIDs  : device_uid → qinglan/sleepace.InvalidateByDeviceUID
//
// "card 变更很少"设计哲学：不做字段级 incremental hint；evict 范围 + lazy reload。
func (p *ConfigPublisher) PublishConfigChanged(
	ctx context.Context,
	op string,
	cards, deviceAddrs, deviceUIDs []string,
) error {
	msg := rediscommon.BuildConfigChangedMessage("wisefido-data", op, cards, deviceAddrs, deviceUIDs)
	stream := rediscommon.StreamConfigCard.Name
	maxLen, retention := rediscommon.GetStreamConfig(rediscommon.StreamConfigCard, nil)

	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, stream, msg, int64(maxLen), retention)
	if err != nil {
		p.logger.Error("publish config.changed",
			zap.String("op", op),
			zap.Int("cards", len(cards)),
			zap.Int("device_addrs", len(deviceAddrs)),
			zap.Int("device_uids", len(deviceUIDs)),
			zap.Error(err))
		return fmt.Errorf("publish config.changed: %w", err)
	}
	p.logger.Info("config.changed",
		zap.String("op", op),
		zap.Int("cards", len(cards)),
		zap.Int("device_addrs", len(deviceAddrs)),
		zap.Int("device_uids", len(deviceUIDs)),
		zap.String("stream_id", streamID))
	return nil
}

// PublishAlarmProcessMessage — alarm 处置通知（config:alarmProcess:stream）。
// 跟 PublishConfigChanged 是不同流不同 schema，不合并。
func (p *ConfigPublisher) PublishAlarmProcessMessage(
	ctx context.Context,
	tenantID, cardID, deviceAddr, alarmLevel, alarmType, processType, eventID string,
	alarmTimestamp int64,
) error {
	_ = deviceAddr
	msg := rediscommon.BuildAlarmProcessMessage("wisefido-data", cardID, alarmLevel, alarmType, processType, eventID, alarmTimestamp)
	stream := rediscommon.StreamConfigAlarmProcess.Name
	maxLen, retention := rediscommon.GetStreamConfig(rediscommon.StreamConfigAlarmProcess, nil)

	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, stream, msg, int64(maxLen), retention)
	if err != nil {
		p.logger.Error("publish alarm.process",
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.String("alarm_level", alarmLevel),
			zap.String("process_type", processType),
			zap.Error(err))
		return fmt.Errorf("publish alarm.process: %w", err)
	}
	p.logger.Info("alarm.process",
		zap.String("tenant_id", tenantID),
		zap.String("card_id", cardID),
		zap.String("alarm_level", alarmLevel),
		zap.String("alarm_type", alarmType),
		zap.String("process_type", processType),
		zap.String("stream_id", streamID))
	return nil
}

// PublishAlarmDeviceMessage — 设备 alarm 使能配置变更（config:alarmDevice:stream）。
// sensor / cardagg 端 enablement cache 按 device_addr 失效；与 config:card 是不同流。
func (p *ConfigPublisher) PublishAlarmDeviceMessage(
	ctx context.Context,
	tenantID, deviceAddr, deviceUID, settingType string,
	settingData map[string]interface{},
) error {
	addr := p.lookupDeviceAddr(ctx, deviceAddr)
	if !addr.IsValid() {
		p.logger.Warn("PublishAlarmDeviceMessage skipped: cannot resolve device_addr",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", deviceAddr),
			zap.String("device_uid", deviceUID))
		return nil
	}
	msg := rediscommon.BuildAlarmDeviceMessage("wisefido-data", addr, settingType, settingData)
	stream := rediscommon.StreamConfigAlarmDevice.Name
	maxLen, retention := rediscommon.GetStreamConfig(rediscommon.StreamConfigAlarmDevice, nil)

	streamID, err := rediscommon.PublishJSONToStream(ctx, p.redisClient, stream, msg, int64(maxLen), retention)
	if err != nil {
		p.logger.Error("publish alarm.device",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", addr.String()),
			zap.String("setting_type", settingType),
			zap.Error(err))
		return fmt.Errorf("publish alarm.device: %w", err)
	}
	p.logger.Info("alarm.device",
		zap.String("tenant_id", tenantID),
		zap.String("device_addr", addr.String()),
		zap.String("setting_type", settingType),
		zap.String("stream_id", streamID))
	return nil
}

// lookupDeviceAddr 校验 device_addr 存在并 normalize；失败返回零值。
func (p *ConfigPublisher) lookupDeviceAddr(ctx context.Context, deviceAddr string) netip.Addr {
	if p.db == nil || deviceAddr == "" {
		return netip.Addr{}
	}
	var addrStr string
	err := p.db.QueryRowContext(ctx,
		`SELECT host(d.device_addr)::text FROM devices d WHERE d.device_addr = $1::INET LIMIT 1`,
		deviceAddr).Scan(&addrStr)
	if err != nil {
		return netip.Addr{}
	}
	a, _ := netip.ParseAddr(addrStr)
	return a
}

// LookupDeviceAddrsByPrefix 查 INET CIDR 下所有 devices.device_addr host text；供 caller 派生 affected 范围。
func (p *ConfigPublisher) LookupDeviceAddrsByPrefix(ctx context.Context, prefix string) []string {
	if p.db == nil || prefix == "" {
		return nil
	}
	rows, err := p.db.QueryContext(ctx,
		`SELECT DISTINCT host(d.device_addr)::text FROM devices d WHERE d.device_addr <<= $1::inet`,
		prefix)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]string, 0, 8)
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err == nil && addr != "" {
			out = append(out, addr)
		}
	}
	return out
}
