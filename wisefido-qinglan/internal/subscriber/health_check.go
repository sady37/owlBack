package subscriber

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"owl-common/alarm"
	"owl-common/observation"
	"owl-common/redis"
	"wisefido-qinglan/internal/consumer"

	"go.uber.org/zap"
)

// DeviceHealthStatus 设备健康状态（属性值）
//
// 该结构既用作单次健康检查的临时计算结果（含 WifiRSSI/AcceleraRaw 原始值），
// 也用作 manager.prevHealth 缓存——记录上次"已发布"的状态供 transition 判定。
// 已发布的状态只看 Offline/SignalPoor/AngleAbnormal 三个字段；其余为采样辅助。
//
// Initialized 必须**逐字段**独立——三个字段在同一 health_check 周期里依次调用
// publishHealthIfChanged，若用单一 bool 共享，第一次调用置 true 后第二/三次会
// 把"默认零值 vs 当前 0"误判成"无跳变"，跳过 publish，导致 *Recover 漏发。
type DeviceHealthStatus struct {
	WifiRSSI                 int    // wifi信号强度(dBm)
	AcceleraRaw              string // 原始格式: X:Y:Z:V[:extra...]
	InstallStyle             int    // 安装方式: 0=顶装, 1=侧装, 2=角装
	Offline                  int    // 0=在线, 1=离线（read 失败兜底）
	SignalPoor               int    // 0=正常, 1=信号差
	AngleAbnormal            int    // 0=正常, 1=倾角异常
	OfflineInitialized       bool   // FieldOffline 是否已记录过
	SignalPoorInitialized    bool   // FieldSignalPoor 是否已记录过
	AngleAbnormalInitialized bool   // FieldAngleAbnormal 是否已记录过
	LastCheckTime            time.Time
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
		m.logger.Warn("Failed to get device properties, marking offline",
			zap.String("device_uid", deviceUID),
			zap.Error(err),
		)
		m.publishHealthIfChanged(ctx, tid, did, deviceUID, observation.FieldOffline, 1)
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

	// 按 iot:alarm:stream 标准格式：每个 status 一条 alarm，category/eventName 用 dataCategoryFromFieldAndValue 的准确值。
	// transition-only：值未变 / 首次观测=0（健康）都不再下发，避免每 10min 重复刷 alarm_events。
	m.publishHealthIfChanged(ctx, tid, did, deviceUID, observation.FieldOffline, 0) // OfflineRecover（read 成功）
	m.publishHealthIfChanged(ctx, tid, did, deviceUID, observation.FieldSignalPoor, health.SignalPoor)
	m.publishHealthIfChanged(ctx, tid, did, deviceUID, observation.FieldAngleAbnormal, health.AngleAbnormal)
}

// publishHealthIfChanged 仅在值发生跳变时下发设备健康类 alarm/recover。
//
// 旧逻辑每次健康检查（10min）无条件 publish offline=0/signal_poor/angle_abnormal 三条，
// 即便状态从未变化也会每 10min 灌一条 alarm_events（cardagg 端没有 onset dedup），导致：
//   - alarm_events 表周期性灌行
//   - cards.unhandled_alarm_X 计数失真
//   - pop_alarm 反复刷新 / push 重复轰炸
//
// 现在按 (device_uid, fieldKey) 缓存上次发布值，仅 0→1 / 1→0 跳变时下发。
//
// 首次观测：无条件 publish 当前值，让 cardagg 把外部观察到的状态同步到 alarm_events 表。
// 历史"value=0 时静默初始化"会让 qinglan 重启后 stale active alarm 永远清不掉
// （设备实际健康但 DB 还挂着重启前的 active 行）；改 publish 后：
//   - cardagg.HandleRecoveryWithTypes 对无 active 同类 alarm 是 noop（幂等）→ 无副作用
//   - 有 active 同类 alarm → 正确 auto-resolve
// onset (value=1) 路径走 InsertAlarmAndUpdateCard，cardagg Phase C 的 DedupWhileActive
// 防御重复 onset。
//
// 这是 qinglan 侧的源头去重；cardagg Phase C 的 DedupWhileActive 是 defense-in-depth。
func (m *DeviceSubscriptionManager) publishHealthIfChanged(ctx context.Context, tenantID, deviceID, deviceUID, fieldKey string, value int) {
	m.mu.Lock()
	prev, ok := m.prevHealth[deviceUID]
	if !ok {
		prev = DeviceHealthStatus{}
	}
	var lastValue int
	var initialized bool
	switch fieldKey {
	case observation.FieldOffline:
		lastValue, initialized = prev.Offline, prev.OfflineInitialized
	case observation.FieldSignalPoor:
		lastValue, initialized = prev.SignalPoor, prev.SignalPoorInitialized
	case observation.FieldAngleAbnormal:
		lastValue, initialized = prev.AngleAbnormal, prev.AngleAbnormalInitialized
	default:
		m.mu.Unlock()
		// 未知字段直接透传（保留极端兜底，不应到这里）
		m.publishDeviceAlarm(ctx, tenantID, deviceID, deviceUID, fieldKey, value)
		return
	}

	// 跳变去重：仅当**该字段**已初始化且值未变才 skip。
	// 首次观测（!initialized）一律 publish，让 cardagg 把当前实际状态同步进 DB——
	// 健康 (value=0) 触发 *Recover 清掉 stale active；异常 (value=1) 触发 onset。
	if initialized && lastValue == value {
		m.mu.Unlock()
		return
	}

	// 首次观测 / 真跳变：先更新缓存再放锁 publish
	switch fieldKey {
	case observation.FieldOffline:
		prev.Offline = value
		prev.OfflineInitialized = true
	case observation.FieldSignalPoor:
		prev.SignalPoor = value
		prev.SignalPoorInitialized = true
	case observation.FieldAngleAbnormal:
		prev.AngleAbnormal = value
		prev.AngleAbnormalInitialized = true
	}
	m.prevHealth[deviceUID] = prev
	m.mu.Unlock()

	m.publishDeviceAlarm(ctx, tenantID, deviceID, deviceUID, fieldKey, value)
}

// RefreshHealthFromProps 把任意"成功"读到的设备属性当作健康证据，仅做正向恢复（value=0）：
//   - 读通本身证明在线 → 发 OfflineRecover
//   - props 含 wifi_rssi 且 > -88 → 发 SignalPoorRecover
//   - props 含 accelera + install_style 且角度合法 → 发 AngleAbnormalRecover
//
// 设计约束（与 healthCheckMonitor 互补）：
//   - 不做反向告警：onset (value=1) 路径由 healthCheckMonitor 独占。任何外部查询都可能因
//     瞬时超时/设备 app 抖动 失败，让它发 Offline=1 会放大误报。
//   - 失败语义不变：调用方 GetDeviceProperties 读失败时，不会进到这里。
//   - 幂等：publishHealthIfChanged 是 transition-only，重复调零副作用；与 10min 健康检查
//     的 prevHealth 共享缓存，恢复事件不会重复下发。
//
// 触发场景：install/排错/HTTP 查询命中 wifi_rssi/accelera 后，几秒内 alarm 即恢复，
// 而不必等下一个 10min 健康检查 tick。
func (m *DeviceSubscriptionManager) RefreshHealthFromProps(ctx context.Context, deviceUID string, props map[string]interface{}) {
	if deviceUID == "" || len(props) == 0 {
		return
	}
	tid, did, ok := m.resolveTenantDevice(ctx, deviceUID)
	if !ok {
		return
	}

	// 读通即为在线证据（OfflineRecover 总是发起；transition-only 会自然过滤无 active 的 noop）
	m.publishHealthIfChanged(ctx, tid, did, deviceUID, observation.FieldOffline, 0)

	h := &DeviceHealthStatus{}
	hasWifi := false
	if wifiVal, has := props["wifi_rssi"]; has {
		h.WifiRSSI = toIntFromInterface(wifiVal)
		hasWifi = true
	}
	hasAcc := false
	if accVal, has := props["accelera"]; has {
		switch v := accVal.(type) {
		case string:
			h.AcceleraRaw = v
		default:
			h.AcceleraRaw = fmt.Sprintf("%v", accVal)
		}
		hasAcc = true
	}
	hasStyle := false
	if styleVal, has := props["install_model"]; has {
		h.InstallStyle = toIntFromInterface(styleVal)
		hasStyle = true
	} else if styleVal, has := props["radar_install_style"]; has {
		h.InstallStyle = toIntFromInterface(styleVal)
		hasStyle = true
	}

	m.validateDeviceHealth(h, deviceUID)

	// SignalPoor 恢复：只在观测到好信号时发，且必须有 wifi_rssi（否则 h.WifiRSSI=0 默认值会假阳）
	if hasWifi && h.SignalPoor == 0 {
		m.publishHealthIfChanged(ctx, tid, did, deviceUID, observation.FieldSignalPoor, 0)
	}
	// AngleAbnormal 恢复：accelera + install_style 二者齐备且校验通过才发
	if hasAcc && hasStyle && h.AngleAbnormal == 0 {
		m.publishHealthIfChanged(ctx, tid, did, deviceUID, observation.FieldAngleAbnormal, 0)
	}
}

// resolveTenantDevice 优先从内存订阅表取 (tenant_id, device_id)，未命中则回落到 device_store。
// 与 checkDeviceHealth 里的解析路径保持一致；外部调用方（install/HTTP）可能没在 subscriptionsByUID 里
// （设备未认证或刚断线），所以 deviceRepo fallback 必要。
func (m *DeviceSubscriptionManager) resolveTenantDevice(ctx context.Context, deviceUID string) (string, string, bool) {
	m.mu.RLock()
	sub, ok := m.subscriptionsByUID[deviceUID]
	m.mu.RUnlock()
	if ok {
		sub.mu.RLock()
		tid := strings.TrimSpace(sub.TenantID)
		did := strings.TrimSpace(sub.DeviceID)
		sub.mu.RUnlock()
		if tid != "" && did != "" {
			return tid, did, true
		}
	}
	if m.deviceRepo == nil {
		return "", "", false
	}
	ds, err := m.deviceRepo.GetDeviceStoreInfo(ctx, deviceUID)
	if err != nil || ds == nil {
		return "", "", false
	}
	tid := strings.TrimSpace(ds.TenantID)
	did := strings.TrimSpace(ds.DeviceID)
	if tid == "" || did == "" {
		return "", "", false
	}
	return tid, did, true
}

// validateDeviceHealth 验证设备健康状态，计算 SignalPoor 和 AngleAbnormal
func (m *DeviceSubscriptionManager) validateDeviceHealth(health *DeviceHealthStatus, deviceUID string) {
	health.SignalPoor = 0
	health.AngleAbnormal = 0

	if health.WifiRSSI <= -88 {
		health.SignalPoor = 1
	}

	x, y, z, calibrated := m.parseAccelerator(health.AcceleraRaw)
	if os.Getenv("QINGLAN_VERBOSE_LOG") == "true" {
		log.Printf("[HEALTH_DEBUG] device=%s accelera=%q install_style=%d x=%.2f y=%.2f calibrated=%v",
			deviceUID, health.AcceleraRaw, health.InstallStyle, x, y, calibrated)
	}
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

// parseAccelerator 解析 accelera 字符串格式: X:Y:Z:V[:extra...]
// 返回: x, y, z, calibrated(V==1)
//
// firmware 不同批次会在 V 之后追加字段（如温度/电压/sample count），
// 所以接受「至少 4 段」，多余尾部字段忽略 —— 否则新 firmware 一上线就被误判
// AngleAbnormal=1（实测 2026-05-03 Radar_1797 上报 6 段 -0.29:-5.71:-5.72:1:6553.4:5.6）。
func (m *DeviceSubscriptionManager) parseAccelerator(accRaw string) (float64, float64, float64, bool) {
	accRaw = strings.TrimSpace(accRaw)
	parts := strings.Split(accRaw, ":")
	if len(parts) < 4 {
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

// validateAngle 根据安装方式验证角度是否在正常范围内（与厂家 checkAccelera 一致）
// installStyle: 0=顶装, 1=侧装, 2=墙角
// 返回: true=正常, false=异常
func (m *DeviceSubscriptionManager) validateAngle(x, y float64, installStyle int) bool {
	ax := absFloat(x)
	switch installStyle {
	case 0: // 顶装：水平，两轴 ±10°
		return ax <= 10 && absFloat(y) <= 10
	case 1: // 侧装：近乎垂直，|x|≤10 且 y < -60（y=-60 为异常）
		return ax <= 10 && y < -60
		
	case 2: // 墙角：45°~60° 倾角支架，|x|≤10 且 -60 ≤ y ≤ -45    临时放宽到69
		return ax <= 10 && y >= -69 && y <= -45
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

// toIntFromInterface 将 interface{} 转为 int（兼容 JSON 解析后的 float64 与 firmware 字符串数字）
//
// firmware 用 MQTT 返回的字段经常是字符串数字（如 wifi_rssi="-45"，install_model="2"），
// 必须 ParseInt 取出来。漏 case string 会让 -45 这种好信号被吞成 0，导致 SignalPoor 判定永远 fail-open。
func toIntFromInterface(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case int64:
		return int(n)
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(n)); err == nil {
			return i
		}
	}
	return 0
}

// publishDeviceAlarm 发布设备类：离线/恢复、SensorDetached/恢复 → iot:alarm:stream（直接影响报警）；信号差、倾角 → iot:event:stream（按使能落库）。
// tenantID、deviceID（UUID）须由调用方从认证/订阅或 device_store 解析后传入；本函数不再查库。
// fieldKey：observation 字段（FieldOffline/FieldSignalPoor/FieldAngleAbnormal/FieldDetached）。
func (m *DeviceSubscriptionManager) publishDeviceAlarm(ctx context.Context, tenantID, deviceID, deviceUID, fieldKey string, value int) {
	if m.streamPublisher == nil {
		log.Printf("⚠️ publishDeviceAlarm: streamPublisher is nil, device=%s field=%s value=%d", deviceUID, fieldKey, value)
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
		log.Printf("⚠️ publishDeviceAlarm: empty eventName, device=%s field=%s value=%d", deviceUID, fieldKey, value)
		return
	}
	cid := m.streamPublisher.GetCardID(ctx, deviceUID)
	// 健康检查触发的衍生告警（Offline/SignalPoor/AngleException 类）— 落 iot:alarm:stream
	// 后立刻被 wisefido-iot 写 iot_timeseries，这里再打日志冗余。仅 verbose 时回放。
	if os.Getenv("QINGLAN_VERBOSE_LOG") == "true" {
		log.Printf("📢 publishDeviceAlarm: device=%s field=%s value=%d event=%s cid=%s tid=%s did=%s", deviceUID, fieldKey, value, eventName, cid, tid, did)
	}
	ts := time.Now().UnixMilli()
	eventStatus := "start"
	if value == 0 {
		eventStatus = "end"
	}
	item := observation.EventItem{
		EventSince:  ts,
		EventStatus: eventStatus,
		TrackID:     observation.TrackDevice,
		EventValue:  int64(value),
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
	// 设备类（onset + recover）一律走 iot:alarm:stream；cardagg 端 *Recover 触发 auto-recover
	// 解除 active alarm。原则：设备类报警都是自动恢复，不需要人工标记。
	directAlarm := eventName == alarm.AlarmTypeOffline || eventName == alarm.AlarmTypeOfflineRecover ||
		eventName == alarm.SensorDetached || eventName == alarm.SensorDetachedRecover ||
		eventName == alarm.SignalPoor || eventName == alarm.SingalPoorRecover ||
		eventName == alarm.AngleException || eventName == alarm.AngleExceptionRecover
	topicType := "event"
	if directAlarm {
		topicType = "alarm"
	}
	msg := redis.NewSingleItemMessage(tid, cid, deviceUID, did, "Radar", ts, topicType, eventName, data)
	if directAlarm {
		if err := m.streamPublisher.PublishAlarm(ctx, msg); err != nil {
			m.logger.Warn("Failed to publish device alarm", zap.String("event_name", eventName), zap.String("device_uid", deviceUID), zap.Error(err))
		} else {
			consumer.QinglanHotPathLog(m.logger, "Published device alarm → iot:alarm:stream", zap.String("device_uid", deviceUID), zap.String("event_name", eventName), zap.String("field", fieldKey), zap.Int("value", value))
		}
	} else {
		if err := m.streamPublisher.PublishEvent(ctx, msg); err != nil {
			m.logger.Warn("Failed to publish device event", zap.String("event_name", eventName), zap.String("device_uid", deviceUID), zap.Error(err))
		} else {
			consumer.QinglanHotPathLog(m.logger, "Published device event → iot:event:stream", zap.String("device_uid", deviceUID), zap.String("event_name", eventName), zap.String("field", fieldKey), zap.Int("value", value))
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
