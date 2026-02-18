package subscriber

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"owl-common/alarm"
	constDevice "owl-common/const"

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
		WifiRSSI:     -120, // 默认值（最坏情况）
		AcceleraRaw:  "",
		InstallStyle: 0,
	}

	// 提取 wifi_rssi
	if wifiVal, ok := props["wifi_rssi"]; ok {
		if wifiInt, ok := wifiVal.(int); ok {
			health.WifiRSSI = wifiInt
		}
	}

	// 提取 accelera
	if accVal, ok := props["accelera"]; ok {
		if accStr, ok := accVal.(string); ok {
			health.AcceleraRaw = accStr
		}
	}

	// 提取 radar_install_style
	if styleVal, ok := props["radar_install_style"]; ok {
		if styleInt, ok := styleVal.(int); ok {
			health.InstallStyle = styleInt
		}
	}

	// 验证和计算状态
	m.validateDeviceHealth(health, deviceUID)

	// 发布设备状态（无论是否异常，只要已验证就发布）
	statuses := map[string]int{
		constDevice.StatusFieldOffline: 0, // 0=在线（语义：1=异常/离线, 0=正常/在线）
	}

	// 如果有异常，添加异常标志
	if health.SignalPoor == 1 {
		statuses[constDevice.StatusFieldSignalPoor] = 1
		log.Printf("⚠️ Device %s signal poor detected: wifi_rssi=%d dBm",
			deviceUID, health.WifiRSSI)
	}

	if health.AngleAbnormal == 1 {
		statuses[constDevice.StatusFieldAngleAbnormal] = 1
		log.Printf("⚠️ Device %s angle abnormal detected",
			deviceUID)
	}

	// 发布状态到stream
	if health.SignalPoor == 1 || health.AngleAbnormal == 1 {
		m.logger.Warn("Device health issue detected",
			zap.String("device_uid", deviceUID),
			zap.Int("signal_poor", health.SignalPoor),
			zap.Int("angle_abnormal", health.AngleAbnormal),
		)
	}

	go m.streamPublisher.PublishDeviceStatus(ctx, deviceID, deviceType, tenantID, deviceUID, statuses)
}

// validateDeviceHealth 验证设备健康状态，计算 SignalPoor 和 AngleAbnormal
func (m *DeviceSubscriptionManager) validateDeviceHealth(health *DeviceHealthStatus, deviceUID string) {
	health.SignalPoor = 0
	health.AngleAbnormal = 0

	// 1. 验证 wifi_rssi
	// 正常范围: -88 ~ -20 dBm
	// <= -88 或 > -20 为异常
	if health.WifiRSSI <= -88 || health.WifiRSSI > -20 {
		health.SignalPoor = 1
		m.logger.Debug("Device signal poor",
			zap.String("device_uid", deviceUID),
			zap.Int("wifi_rssi", health.WifiRSSI),
		)
	}

	// 2. 验证 accelera（倾角）
	x, y, z, calibrated := m.parseAccelerator(health.AcceleraRaw)
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


// publishDeviceAlarmAuto 自动查使能表后发布设备报警（供 device_subscription_manager 调用）
func (m *DeviceSubscriptionManager) publishDeviceAlarmAuto(ctx context.Context, tenantID, deviceID, deviceUID, alarmType, statusFieldValue string) {
	if tenantID == "" || m.streamPublisher == nil {
		return
	}
	enablementItems, err := m.deviceRepo.GetAlarmEnablement(ctx, tenantID, deviceUID)
	if err != nil {
		m.logger.Warn("Failed to get alarm enablement",
			zap.String("device_uid", deviceUID),
			zap.Error(err))
		return
	}
	m.publishDeviceAlarm(ctx, tenantID, deviceID, deviceUID, alarmType, statusFieldValue, enablementItems)
}


// publishDeviceAlarm 发布设备报警到 iot:alarm:stream（统一格式）
// alarmType: "SignalPoor", "AngleException" 等
// statusFieldValue: "1"=异常, "0"=恢复
// enablementItems: 使能表，用于查找 alarm_level
func (m *DeviceSubscriptionManager) publishDeviceAlarm(ctx context.Context, tenantID, deviceID, deviceUID, alarmType, statusFieldValue string, enablementItems []alarm.AlarmEnablementItem) {
	if m.streamPublisher == nil {
		return
	}

	// 从使能表查找该 alarmType 是否启用及其 alarm_level
	var alarmLevel string
	enabled := false
	for _, item := range enablementItems {
		if item.AlarmType == alarmType && item.IsEnabled == 1 {
			enabled = true
			alarmLevel = item.AlarmLevel
			break
		}
	}
	if !enabled {
		return
	}

	// 统一格式: category = "LEVEL.AlarmType", data_value[0] = {category, StatusFieldValue, device_uid}
	alarmCategory := alarmLevel + "." + alarmType
	dataValue := []interface{}{
		map[string]interface{}{
			"category":         alarm.GetFHIRCategory(alarmType),
			"StatusFieldValue": statusFieldValue,
			"device_uid":       deviceUID,
		},
	}
	encodedData := m.streamPublisher.BuildEncodedData(
		"", // cardID 由 cardagg 通过 device_id 查找
		tenantID,
		deviceID,
		"alarm",
		alarmCategory,
		dataValue,
	)
	streamName := m.streamPublisher.GetOutputStreamName("alarm")
	if _, err := m.streamPublisher.PublishToStream(ctx, streamName, encodedData); err != nil {		
		m.logger.Warn("Failed to publish device alarm from health check",
			zap.String("alarm_type", alarmType),
			zap.String("device_id", deviceID),
			zap.Error(err))
	} else {
		m.logger.Info("Health check → iot:alarm:stream",
			zap.String("device_id", deviceID),
			zap.String("alarm_category", alarmCategory),
			zap.String("status_field_value", statusFieldValue))
	}
}
