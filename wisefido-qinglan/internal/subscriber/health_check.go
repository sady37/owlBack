package subscriber

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/observation"
	"owl-common/redis"

	"go.uber.org/zap"
)

// DeviceHealthStatus 设备健康状态（属性值）
type DeviceHealthStatus struct {
	WifiRSSI      int    // wifi信号强度(dBm)
	AcceleraRaw   string // 原始格式: X:Y:Z:V
	InstallStyle  int    // 安装方式: 0=顶装, 1=侧装, 2=角装
	SignalPoor    int    // 0=正常, 1=信号差
	AngleAbnormal int    // 0=正常, 1=倾角异常
	LastCheckTime time.Time
}

// healthCheckMonitor 设备健康检查（每10分钟执行一次）
func (m *DeviceSubscriptionManager) healthCheckMonitor(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	log.Println("Device health check monitor started")

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.checkAllDevicesHealth(ctx)
		}
	}
}

// checkAllDevicesHealth 检查所有在线设备的健康状态
func (m *DeviceSubscriptionManager) checkAllDevicesHealth(ctx context.Context) {
	m.mu.RLock()
	subs := make([]*DeviceSubscription, 0, len(m.subscriptionsByUID))
	for _, sub := range m.subscriptionsByUID {
		subs = append(subs, sub)
	}
	m.mu.RUnlock()

	for _, sub := range subs {
		sub.mu.RLock()
		deviceUID := sub.DeviceUID
		deviceID := sub.DeviceID
		status := sub.Status
		deviceType := sub.DeviceType
		tenantID := sub.TenantID
		sub.mu.RUnlock()

		// 仅检查在线设备
		if status != "online" {
			continue
		}

		// 发送read命令获取属性值
		go m.checkDeviceHealth(ctx, deviceUID, deviceID, deviceType, tenantID)
	}
}

// checkDeviceHealth 检查单个设备的健康状态
// 1. 通过 RadarService 读取 wifi_rssi, accelera, radar_install_style
// 2. 验证值是否正常
// 3. 发送异常状态到stream
func (m *DeviceSubscriptionManager) checkDeviceHealth(ctx context.Context, deviceUID, deviceID, deviceType, tenantID string) {
	m.logger.Info("Checking device health",
		zap.String("device_uid", deviceUID),
		zap.String("device_id", deviceID),
	)

	tid := strings.TrimSpace(tenantID)
	did := strings.TrimSpace(deviceID)
	if tid == "" || did == "" {
		if m.deviceRepo == nil {
			m.logger.Warn("health check skip stream publish: no deviceRepo", zap.String("device_uid", deviceUID))
			return
		}
		ds, err := m.deviceRepo.GetDeviceStoreInfo(ctx, deviceUID)
		if err != nil || ds == nil {
			m.logger.Warn("health check skip stream publish: device_store lookup failed",
				zap.String("device_uid", deviceUID), zap.Error(err))
			return
		}
		tid = strings.TrimSpace(ds.TenantID)
		did = strings.TrimSpace(ds.DeviceID)
		if tid == "" || did == "" {
			m.logger.Warn("health check skip stream publish: device_store missing tenant_id or device_id",
				zap.String("device_uid", deviceUID))
			return
		}
	}

	// 检查 RadarService 是否已设置
	if m.radarService == nil {
		m.logger.Warn("RadarService not set, skipping health check",
			zap.String("device_uid", deviceUID),
		)
		return
	}

	// 通过 RadarService 读取设备属性
	keys := []string{"wifi_rssi", "accelera", "radar_install_style"}
	props, err := m.radarService.GetDeviceProperties(ctx, deviceUID, keys)
	if err != nil {
		m.logger.Warn("Failed to get device properties",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		return
	}

	// 解析属性值
	health := &DeviceHealthStatus{
		AcceleraRaw:  "",
		InstallStyle: 0,
	}

	// 提取 wifi_rssi（JSON 解析后可能为 float64）
	if wifiVal, ok := props["wifi_rssi"]; ok {
		health.WifiRSSI = toIntFromInterface(wifiVal)
	}

	if accVal, ok := props["accelera"]; ok {
		switch v := accVal.(type) {
		case string:
			health.AcceleraRaw = v
		default:
			health.AcceleraRaw = fmt.Sprintf("%v", accVal)
		}
	}

	// 提取 radar_install_style（JSON 解析后可能为 float64）
	if styleVal, ok := props["install_model"]; ok {
		health.InstallStyle = toIntFromInterface(styleVal)
	} else if styleVal, ok := props["radar_install_style"]; ok {
		health.InstallStyle = toIntFromInterface(styleVal)
	}

	// 验证和计算状态
	m.validateDeviceHealth(health, deviceUID)

	if health.SignalPoor == 1 || health.AngleAbnormal == 1 {
		log.Printf("⚠️ Device %s health: signal_poor=%d(rssi=%d) angle_abnormal=%d",
			deviceUID, health.SignalPoor, health.WifiRSSI, health.AngleAbnormal)
	}

	// 按 iot:alarm:stream 标准格式：每个 status 一条 alarm，category/eventName 用 dataCategoryFromFieldAndValue 的准确值
	m.publishDeviceAlarm(ctx, tid, did, deviceUID, observation.FieldOffline, 0) // DeviceRecover
	m.publishDeviceAlarm(ctx, tid, did, deviceUID, observation.FieldSignalPoor, health.SignalPoor)
	m.publishDeviceAlarm(ctx, tid, did, deviceUID, observation.FieldAngleAbnormal, health.AngleAbnormal)
}

// validateDeviceHealth 验证设备健康状态，计算 SignalPoor 和 AngleAbnormal
func (m *DeviceSubscriptionManager) validateDeviceHealth(health *DeviceHealthStatus, deviceUID string) {
	health.SignalPoor = 0
	health.AngleAbnormal = 0

	if health.WifiRSSI <= -88 {
		health.SignalPoor = 1
	}

	x, y, z, calibrated := m.parseAccelerator(health.AcceleraRaw)
	log.Printf("[HEALTH_DEBUG] device=%s accelera=%q install_style=%d x=%.2f y=%.2f calibrated=%v",
		deviceUID, health.AcceleraRaw, health.InstallStyle, x, y, calibrated)
	if !calibrated {
		// V == 0，未校准
		health.AngleAbnormal = 1
		m.logger.Debug("Device angle not calibrated",
			zap.String("device_uid", deviceUID),
		)
	} else {
		// V == 1，已校准，根据安装方式验证角度范围
		isValid := m.validateAngle(x, y, health.InstallStyle)
		if !isValid {
			health.AngleAbnormal = 1
			m.logger.Debug("Device angle abnormal",
				zap.String("device_uid", deviceUID),
				zap.Float64("x", x),
				zap.Float64("y", y),
				zap.Float64("z", z),
				zap.Int("install_style", health.InstallStyle),
			)
		}
	}
}

// parseAccelerator 解析 accelera 字符串格式: X:Y:Z:V
// 返回: x, y, z, calibrated(V==1)
func (m *DeviceSubscriptionManager) parseAccelerator(accRaw string) (float64, float64, float64, bool) {
	accRaw = strings.TrimSpace(accRaw)
	parts := strings.Split(accRaw, ":")
	if len(parts) != 4 {
		return 0, 0, 0, false
	}

	x, errX := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	y, errY := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	z, errZ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	v, errV := strconv.ParseInt(strings.TrimSpace(parts[3]), 10, 32)

	if errX != nil || errY != nil || errZ != nil || errV != nil {
		return 0, 0, 0, false
	}

	calibrated := v == 1
	return x, y, z, calibrated
}

// validateAngle 根据安装方式验证角度是否在正常范围内
// installStyle: 0=顶装, 1=侧装, 2=角装
// 返回: true=正常, false=异常
func (m *DeviceSubscriptionManager) validateAngle(x, y float64, installStyle int) bool {
	switch installStyle {
	case 0: // 顶装
		// X和Y应在±10°以内
		return absFloat(x) <= 10 && absFloat(y) <= 10

	case 1, 2: // 侧装或角装
		// X应在±10°以内，Y应在-60°到-90°之间
		return absFloat(x) <= 10 && y >= -90 && y <= -60

	default:
		return false
	}
}

// absFloat 返回浮点数的绝对值
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// toIntFromInterface 将 interface{} 转为 int（兼容 JSON 解析后的 float64）
func toIntFromInterface(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case int64:
		return int(n)
	}
	return 0
}

// publishDeviceAlarm 发布设备类：离线/恢复、SensorDetached/恢复 → iot:alarm:stream（直接影响报警）；信号差、倾角 → iot:event:stream（按使能落库）。
// tenantID、deviceID（UUID）须由调用方从认证/订阅或 device_store 解析后传入；本函数不再查库。
// fieldKey：observation 字段（FieldOffline/FieldSignalPoor/FieldAngleAbnormal/FieldDetached）。
func (m *DeviceSubscriptionManager) publishDeviceAlarm(ctx context.Context, tenantID, deviceID, deviceUID, fieldKey string, value int) {
	if m.streamPublisher == nil {
		return
	}
	tid := strings.TrimSpace(tenantID)
	did := strings.TrimSpace(deviceID)
	if tid == "" || did == "" {
		m.logger.Warn("skip publish device alarm: empty tenant_id or device_id",
			zap.String("device_uid", deviceUID))
		return
	}
	eventName := dataCategoryFromFieldAndValue(fieldKey, value)
	if eventName == "" {
		return
	}
	cid := ""
	ts := time.Now().UnixMilli()
	eventStatus := "start"
	if value == 0 {
		eventStatus = "end"
	}
	item := observation.EventItem{
		DataCategory: eventName,
		EventName:    eventName,
		EventSince:   ts,
		EventStatus:  eventStatus,
		TrackID:      observation.TrackDevice,
		EventValue:   int64(value),
	}
	data, err := observation.EventItemToDataMap(&item)
	if err != nil {
		m.logger.Warn("EventItemToDataMap failed", zap.String("device_uid", deviceUID), zap.Error(err))
		return
	}
	if data == nil {
		data = make(map[string]interface{})
	}
	data[fieldKey] = value
	directAlarm := eventName == alarm.AlarmTypeOffline || eventName == alarm.AlarmTypeOfflineRecover ||
		eventName == alarm.SensorDetached || eventName == alarm.SensorDetachedRecover
	topicType := "event"
	if directAlarm {
		topicType = "alarm"
	}
	msg := redis.NewSingleItemMessage(tid, cid, deviceUID, did, "Radar", ts, topicType, eventName, data)
	if directAlarm {
		if err := m.streamPublisher.PublishAlarm(ctx, msg); err != nil {
			m.logger.Warn("Failed to publish device alarm", zap.String("event_name", eventName), zap.String("device_uid", deviceUID), zap.Error(err))
		} else {
			m.logger.Info("Published device alarm → iot:alarm:stream", zap.String("device_uid", deviceUID), zap.String("event_name", eventName), zap.String("field", fieldKey), zap.Int("value", value))
		}
	} else {
		if err := m.streamPublisher.PublishEvent(ctx, msg); err != nil {
			m.logger.Warn("Failed to publish device event", zap.String("event_name", eventName), zap.String("device_uid", deviceUID), zap.Error(err))
		} else {
			m.logger.Info("Published device event → iot:event:stream", zap.String("device_uid", deviceUID), zap.String("event_name", eventName), zap.String("field", fieldKey), zap.Int("value", value))
		}
	}
}

// dataCategoryFromFieldAndValue 根据 observation 字段与取值推导 event_name/dataCategory（0=恢复，1=告警）。
func dataCategoryFromFieldAndValue(fieldKey string, value int) string {
	switch fieldKey {
	case observation.FieldOffline:
		if value == 0 {
			return alarm.AlarmTypeOfflineRecover
		}
		return alarm.AlarmTypeOffline
	case observation.FieldSignalPoor:
		if value == 0 {
			return alarm.SingalPoorRecover
		}
		return alarm.SignalPoor
	case observation.FieldAngleAbnormal:
		if value == 0 {
			return alarm.AngleExceptionRecover
		}
		return alarm.AngleException
	case observation.FieldDetached:
		if value == 0 {
			return alarm.SensorDetachedRecover
		}
		return alarm.SensorDetached
	default:
		return ""
	}
}
