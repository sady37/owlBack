package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"wisefido-data/internal/config"
	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
	"wisefido-qinglan/encode"

	"go.uber.org/zap"
)

// RadarInstall Radar 安装配置服务
// 雷达设备的查询与设置仅通过 qinglan_client（PUT/GET wisefido-qinglan /api/v1/radar/devices/{uid}/properties 等），无其他设备路径。
// 画布 layout 走 room_visual_layout 表（PK=spatial_prefix，canvas JSONB）；
// 2026-05-16 cutover：废弃 v1 的 rooms.layout_config 列 + config_versions 表（v2 都不存在）。
type RadarInstall struct {
	config            *config.Config
	db                *sql.DB
	devicesRepo       repository.DevicesRepository
	cardsRepo         repository.CardsRepository
	configVersionsRepo repository.ConfigVersionsRepository // 保留参数兼容；layout 路径已不再用
	unitsRepo         repository.UnitsRepository
	qinglanClient     *QinglanClient
	logger            *zap.Logger
	// 订阅管理器：记录哪些设备需要订阅（device_id -> bool）
	// 当设备被 bind 时，标记为需要订阅；unbind 时，移除订阅标记
	subscribedDevices map[string]bool
	subscribedMutex   sync.RWMutex // 保护 subscribedDevices 的并发访问
}

// NewRadarInstall 创建 Radar 安装配置服务
func NewRadarInstall(cfg *config.Config, db *sql.DB, devicesRepo repository.DevicesRepository, cardsRepo repository.CardsRepository, configVersionsRepo repository.ConfigVersionsRepository, unitsRepo repository.UnitsRepository, qinglanClient *QinglanClient, logger *zap.Logger) *RadarInstall {
	return &RadarInstall{
		config:            cfg,
		db:                db,
		devicesRepo:       devicesRepo,
		cardsRepo:         cardsRepo,
		configVersionsRepo: configVersionsRepo,
		unitsRepo:         unitsRepo,
		qinglanClient:     qinglanClient,
		logger:            logger,
		subscribedDevices: make(map[string]bool),
	}
}

// ListCardDevices 按 card_id 取卡片上设备，供 GET /card/:cardId/devices；有 cardId 时前端用。
func (s *RadarInstall) ListCardDevices(ctx context.Context, tenantID, cardID string) ([]repository.CardDeviceItem, error) {
	if s.cardsRepo == nil {
		return nil, fmt.Errorf("cards repository not available")
	}
	return s.cardsRepo.GetCardDevices(ctx, tenantID, cardID)
}

// ListCardDevicesByDeviceID 通过 device_id 查找所属卡片，返回 card_id、room_id (/88 fallback 锚)、
// spatialPrefix (/128 device CIDR，本 canvas 的 save target)、该卡设备列表及 layout 配置（三层 fallback 取到的最具体一份）。
//
// scope 来源 = URL 切入点：当前 URL 是 device 切入，故 spatialPrefix 即该 device 的 /128 CIDR。
// 未来若新增 room/unit 切入 URL，新 handler 负责返回相应 /88 或 /80 prefix。
func (s *RadarInstall) ListCardDevicesByDeviceID(ctx context.Context, tenantID, deviceID string) (cardID, roomID, spatialPrefix string, devices []repository.CardDeviceItem, layoutConfig json.RawMessage, err error) {
	if s.cardsRepo == nil {
		return "", "", "", nil, nil, fmt.Errorf("cards repository not available")
	}
	cardID, err = s.cardsRepo.GetCardIDByDeviceID(ctx, tenantID, deviceID)
	if err != nil {
		return "", "", "", nil, nil, err
	}
	devices, err = s.cardsRepo.GetCardDevices(ctx, tenantID, cardID)
	if err != nil {
		return "", "", "", nil, nil, err
	}
	// 入参 deviceIPv6 (host /128) 直接喂给 layout 查询；三档 fallback 走
	//   /128 device 局部 → /88 room → /80 unit
	// 都没命中再 fallback 到 repo 老 stub。
	roomID = ""
	spatialPrefix = ""
	devLookup := ""
	if s.devicesRepo != nil {
		dev, e := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
		if e == nil && dev != nil {
			if dev.DeviceIPv6 != "" {
				devLookup = dev.DeviceIPv6 // 完整 /128，让 getCurrentLayoutByRoomID 三层 fallback
				// API 返给 FE：
				//   - spatial_prefix = /128 device CIDR（save target，URL 切入点决定 scope）
				//   - room_id        = /88 room CIDR（layout 加载 fallback 锚 + 老 FE 兼容）
				spatialPrefix = dev.DeviceIPv6 + "/128"
				if rc, perr := normalizeRoomPrefix88(dev.DeviceIPv6); perr == nil {
					roomID = rc
				}
			}
			if roomID == "" && dev.RoomID.Valid && dev.RoomID.String != "" {
				roomID = dev.RoomID.String
				devLookup = roomID
			}
		}
	}
	if devLookup != "" {
		layoutConfig = s.getCurrentLayoutByRoomID(ctx, tenantID, devLookup)
	}

	// Inject device bindings into layout for unit-level templates that contain
	// temporary iot.deviceId like "Radar01" but lack real device_id values.
	if len(layoutConfig) > 0 && len(devices) > 0 {
		layoutConfig = injectDeviceBindingsIntoLayout(layoutConfig, devices, s.logger)
	}

	return cardID, roomID, spatialPrefix, devices, layoutConfig, nil
}

// injectDeviceBindingsIntoLayout 对 layout 中 device_id 为空的 Radar/Sleepad/Sensor 对象，
// 按 iot.deviceId 序号（Radar01→0, Radar02→1）与 devices 列表对应项建立绑定，
// 并补全顶层 radar/sleepad map供前端 loadFromLayoutConfig 使用。
func injectDeviceBindingsIntoLayout(layout json.RawMessage, devices []repository.CardDeviceItem, logger *zap.Logger) json.RawMessage {
	if len(layout) == 0 || len(devices) == 0 {
		return layout
	}

	var doc map[string]any
	if err := json.Unmarshal(layout, &doc); err != nil {
		return layout
	}

	rawObjects, _ := doc["objects"].([]any)
	if len(rawObjects) == 0 {
		return layout
	}

	type devEntry struct{ id, uid string }
	byType := map[string][]devEntry{}
	for _, d := range devices {
		t := strings.ToLower(d.DeviceType)
		if t == "" {
			t = "radar"
		}
		byType[t] = append(byType[t], devEntry{id: d.DeviceID, uid: d.DeviceUID})
	}

	topMap := map[string]map[string]string{}
	changed := false

	// iterate objects
	for _, raw := range rawObjects {
		obj, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		// skip if already bound
		if did, _ := obj["device_id"].(string); did != "" {
			continue
		}
		if bid, _ := obj["bindedDeviceId"].(string); bid != "" {
			continue
		}

		typeName, _ := obj["typeName"].(string)
		t := strings.ToLower(typeName)
		candidates := byType[t]
		if len(candidates) == 0 {
			continue
		}

		// parse iot.deviceId suffix number
		idx := 0
		if iot, ok2 := nestedString(obj, "device", "iot", "deviceId"); ok2 {
			re := regexp.MustCompile(`(\d+)$`)
			m := re.FindStringSubmatch(iot)
			if len(m) > 1 {
				if n, e := strconv.Atoi(m[1]); e == nil && n > 0 {
					idx = n - 1
				}
			}
		}
		if idx >= len(candidates) {
			idx = len(candidates) - 1
		}

		entry := candidates[idx]
		obj["device_id"] = entry.id
		obj["bindedDeviceId"] = entry.id

		objID, _ := obj["id"].(string)
		if objID != "" && entry.id != "" {
			if topMap[t] == nil {
				topMap[t] = map[string]string{}
			}
			topMap[t][objID] = entry.id
		}
		changed = true
		if logger != nil {
			logger.Info("injectDeviceBindings: bound radar object",
				zap.String("typeName", typeName),
				zap.String("objectId", objID),
				zap.String("deviceId", entry.id),
			)
		}
	}

	if !changed {
		return layout
	}

	for t, m := range topMap {
		doc[t] = m
	}

	out, err := json.Marshal(doc)
	if err != nil {
		return layout
	}
	return out
}

// nestedString 从 map[string]any 深层取字符串字段
func nestedString(m map[string]any, keys ...string) (string, bool) {
	cur := m
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return "", false
		}
		if i == len(keys)-1 {
			s, ok := v.(string)
			return s, ok
		}
		nxt, ok := v.(map[string]any)
		if !ok {
			return "", false
		}
		cur = nxt
	}
	return "", false
}

// getCurrentLayoutByRoomID 取房间当前布局：直接读 room_visual_layout.canvas（PK=spatial_prefix）。
//
// 2026-05-16 cutover：废弃 v1 的 rooms.layout_config / units.layout_config / config_versions 链
// （v2 schema 这三处都不存在；旧实现写入静默失败、读永远 nil）。详 owlBack/doc/AI_fall_detect 与
// schema 27_room_visual_layout.sql。
//
// roomID 应是 /88 CIDR 文本（如 "fd00:0:3:111:3:300::/88"）。room /88 找不到时回退 unit /80
// （unit-level 共享布局；schema rvl_prefix_scope CHECK 允许 80/88 两种 masklen）。
func (s *RadarInstall) getCurrentLayoutByRoomID(ctx context.Context, tenantID, roomID string) json.RawMessage {
	if s.db == nil || roomID == "" {
		return nil
	}
	startCIDR, err := normalizeLayoutPrefix(roomID)
	if err != nil {
		return nil
	}
	// 三档 fallback：longest-prefix-wins (device /128 → room /88 → unit /80)
	// 起点视入参 masklen 而定；广 scope 入参（/80）不会回看 device 局部，避免越权。
	startBits, _ := masklenOfCIDR(startCIDR)
	for _, m := range []int{128, 88, 80} {
		if m > startBits {
			continue // 不向更精确层试探
		}
		var lookup string
		if m == startBits {
			lookup = startCIDR
		} else {
			lookup = truncateCIDRToMask(startCIDR, m)
			if lookup == "" {
				continue
			}
		}
		if canvas := s.queryRoomVisualLayoutCanvas(ctx, lookup); len(canvas) > 0 {
			return canvas
		}
	}
	return nil
}

// masklenOfCIDR 取 CIDR 的 masklen；失败返回 0。
func masklenOfCIDR(cidr string) (int, error) {
	p, err := netip.ParsePrefix(cidr)
	if err != nil {
		return 0, err
	}
	return p.Bits(), nil
}

// normalizeRoomPrefix88 任意 IPv6 (host 或 CIDR) → /88 room CIDR，用于 API 返给 FE 显示的 room_id。
// 与 normalizeLayoutPrefix 不同：本函数永远输出 /88，忽略 host bits（用于 FE 上下文识别 room）。
func normalizeRoomPrefix88(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("empty")
	}
	if idx := strings.IndexByte(s, '/'); idx >= 0 {
		s = s[:idx]
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return "", err
	}
	return netip.PrefixFrom(addr, 88).Masked().String(), nil
}

// queryRoomVisualLayoutCanvas 一次 SELECT canvas FROM room_visual_layout WHERE spatial_prefix。
func (s *RadarInstall) queryRoomVisualLayoutCanvas(ctx context.Context, spatialPrefix string) json.RawMessage {
	var canvas []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT canvas FROM room_visual_layout WHERE spatial_prefix = $1::INET`,
		spatialPrefix,
	).Scan(&canvas)
	if err != nil {
		if err != sql.ErrNoRows && s.logger != nil {
			s.logger.Debug("room_visual_layout SELECT failed",
				zap.String("spatial_prefix", spatialPrefix), zap.Error(err))
		}
		return nil
	}
	return canvas
}

// sha256HexString sha256 hex 64 字符；用于 room_visual_layout.canvas_hash 检测改动。
func sha256HexString(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// truncateCIDRToMask 把 CIDR 文本截短到 newMask 位并 mask 出 network 地址。
//   "fd00:0:3:111:3:300::/88" + 80 → "fd00:0:3:111:3::/80"
// 输入非法 / 已等于或更短 mask → 返回空。
func truncateCIDRToMask(cidr string, newMask int) string {
	p, err := netip.ParsePrefix(strings.TrimSpace(cidr))
	if err != nil {
		return ""
	}
	if p.Bits() <= newMask {
		return ""
	}
	return netip.PrefixFrom(p.Addr(), newMask).Masked().String()
}

// GetLayoutConfigByRoomID 公有版本 layout query；与 getCurrentLayoutByRoomID 等价。
//
// 2026-05-16 cutover：废弃 v1 config_versions 历史时间点 query；当前 schema
// (room_visual_layout) 不保留版本历史，每次 SaveRoomLayout UPDATE 同一行（version++）。
// 时间点 query 语义不再支持；caller 想要历史需另设 room_visual_layout_history。
func (s *RadarInstall) GetLayoutConfigByRoomID(ctx context.Context, tenantID, roomID string) (json.RawMessage, error) {
	_ = tenantID // tenantID 隐含在 roomID 的 /48 prefix；保留参数兼容
	if roomID == "" {
		return nil, nil
	}
	return s.getCurrentLayoutByRoomID(ctx, tenantID, roomID), nil
}

// SaveRoomLayout UPSERT 房间布局到 room_visual_layout 表（PK=spatial_prefix）。
//
// 2026-05-16 cutover：废弃 v1 的 rooms.layout_config 列 + config_versions 表（schema 都不存在；
// 旧实现写入静默失败）。new schema 27_room_visual_layout：spatial_prefix + canvas JSONB + canvas_hash
// + version。canvas 与既有 SHA256 一致时跳过（不增 version），否则 UPSERT，version 自增。
//
// roomID 接受：
//   - host-only IPv6 (FE 常用，URL path 不能含 "/")：默认追加 /88 (room scope)
//   - 已带 /80 或 /88 mask 的 CIDR
// schema CHECK rvl_prefix_scope 仅允许 /80 (unit) 或 /88 (room) masklen。
func (s *RadarInstall) SaveRoomLayout(ctx context.Context, tenantID, roomID string, configData []byte) error {
	_ = tenantID // tenantID 隐含在 roomID 的 /48 prefix，schema CHECK 校验
	if roomID == "" {
		return fmt.Errorf("room_id is required")
	}
	if len(configData) == 0 {
		return fmt.Errorf("config_data is required")
	}
	if s.db == nil {
		return fmt.Errorf("db not available")
	}
	spatialPrefix, err := normalizeLayoutPrefix(roomID)
	if err != nil {
		return fmt.Errorf("room_id invalid: %w", err)
	}

	canvasHash := sha256HexString(configData)

	// 比对 hash：与已存一致则跳过（不增 version、不更新 updated_at）
	var existingHash sql.NullString
	if err := s.db.QueryRowContext(ctx,
		`SELECT canvas_hash FROM room_visual_layout WHERE spatial_prefix = $1::INET`,
		spatialPrefix,
	).Scan(&existingHash); err == nil && existingHash.Valid && existingHash.String == canvasHash {
		s.logger.Debug("SaveRoomLayout: canvas unchanged, skip", zap.String("spatial_prefix", spatialPrefix))
		return nil
	}

	// UPSERT：INSERT ... ON CONFLICT DO UPDATE（version + 1, updated_at = NOW()）
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO room_visual_layout (spatial_prefix, canvas, canvas_hash, version, updated_at)
		VALUES ($1::INET, $2::JSONB, $3, 1, NOW())
		ON CONFLICT (spatial_prefix) DO UPDATE
		SET canvas = EXCLUDED.canvas,
		    canvas_hash = EXCLUDED.canvas_hash,
		    version = room_visual_layout.version + 1,
		    updated_at = NOW()
	`, spatialPrefix, string(configData), canvasHash)
	if err != nil {
		s.logger.Warn("SaveRoomLayout: upsert room_visual_layout failed",
			zap.String("spatial_prefix", spatialPrefix), zap.Error(err))
		return fmt.Errorf("save room layout: %w", err)
	}
	s.logger.Info("SaveRoomLayout: upserted",
		zap.String("spatial_prefix", spatialPrefix), zap.String("canvas_hash", canvasHash))
	return nil
}

// normalizeLayoutPrefix 把入参 roomID/deviceID 规范成 spatial_prefix CIDR（/80, /88 或 /128）。
//
// 三档作用域：
//   - /80  unit 公共布局：caller 显式传 "fd00:0:3:111:3::/80"
//   - /88  room 默认布局：host-only 不带 mask 时默认补 /88
//   - /128 device 局部布局：caller 显式传 "fd00:0:3:111:3:300:7cd4:570c/128" 或带任何 mask 但 host bits 非零
//
// host-only 入参（不含 "/")时的默认推断：host bits 全 0 → /88 room；host bits 非零 → /128 device
//
//   - "fd00:0:3:111:3::"                 → "fd00:0:3:111:3::/88"  (room)
//   - "fd00:0:3:111:3:300::"             → "fd00:0:3:111:3:300::/88" (room)
//   - "fd00:0:3:111:3:300:7cd4:570c"     → "fd00:0:3:111:3:300:7cd4:570c/128" (device)
//   - "fd00:0:3:111:3:300::/88"          → 原样
//   - "fd00:0:3:111:3::/80"              → 原样 (unit 公共)
//   - 其他 masklen / 非法 IPv6            → error
func normalizeLayoutPrefix(roomID string) (string, error) {
	s := strings.TrimSpace(roomID)
	if s == "" {
		return "", fmt.Errorf("empty")
	}
	if !strings.Contains(s, "/") {
		// host-only：host bits 非零认作 device (/128)，否则 /88 room
		addr, err := netip.ParseAddr(s)
		if err != nil {
			return "", fmt.Errorf("parse addr %q: %w", roomID, err)
		}
		// 判断 /88 mask 之下还有非零 bits → device 局部
		room88 := netip.PrefixFrom(addr, 88).Masked()
		if room88.Addr() != addr {
			s = s + "/128"
		} else {
			s = s + "/88"
		}
	}
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return "", fmt.Errorf("parse cidr %q: %w", roomID, err)
	}
	if p.Bits() != 80 && p.Bits() != 88 && p.Bits() != 128 {
		return "", fmt.Errorf("masklen %d not in {80,88,128} (room_visual_layout schema CHECK)", p.Bits())
	}
	return p.Masked().String(), nil
}

// GetDeviceUID 根据 device_id 和 tenant_id 获取设备 UID；设备不存在、未绑定或非雷达类型时返回明确错误
func (s *RadarInstall) GetDeviceUID(ctx context.Context, tenantID, deviceID string) (string, error) {
	if s.devicesRepo == nil {
		return "", fmt.Errorf("devices repository not available")
	}
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return "", fmt.Errorf("设备不存在或未绑定(device_id=%s)，请先在画布中 Bind 真实雷达设备: %w", deviceID, err)
	}
	if !device.DeviceType.Valid || !strings.EqualFold(device.DeviceType.String, "radar") {
		return "", fmt.Errorf("该设备不是雷达类型，无法操作")
	}
	if device.DeviceUID == "" {
		return "", fmt.Errorf("device UID not found for device_id: %s", deviceID)
	}
	return device.DeviceUID, nil
}

// GetDeviceByUID 根据 device_uid 和 tenant_id 获取设备信息
// 用于从 device_uid 转换为 device_id（用于订阅检查等场景）
func (s *RadarInstall) GetDeviceByUID(ctx context.Context, tenantID, deviceUID string) (*domain.Device, error) {
	if s.devicesRepo == nil {
		return nil, fmt.Errorf("devices repository not available")
	}
	return s.devicesRepo.GetDeviceByUID(ctx, tenantID, deviceUID)
}

// GetDeviceStatus 获取设备在线状态
// 通过 wisefido-qinglan: GET /api/v1/radar/devices/{uid}/status
func (s *RadarInstall) GetDeviceStatus(ctx context.Context, deviceUID string) (string, error) {
	if s.qinglanClient == nil {
		return "", fmt.Errorf("qinglan client not available")
	}
	return s.qinglanClient.GetDeviceStatus(ctx, deviceUID)
}

// GetDeviceProperties 读取设备属性
// 通过 wisefido-qinglan: GET /api/v1/radar/devices/{uid}/properties
func (s *RadarInstall) GetDeviceProperties(ctx context.Context, uid string, keys []string) (map[string]interface{}, error) {
	return s.qinglanClient.GetDeviceProperties(ctx, uid, keys)
}

// SetDeviceProperties 设置设备属性，返回设备响应码（200=成功）和错误
// 通过 wisefido-qinglan: PUT /api/v1/radar/devices/{uid}/properties
func (s *RadarInstall) SetDeviceProperties(ctx context.Context, uid string, properties map[string]interface{}) (deviceCode int, err error) {
	return s.qinglanClient.SetDeviceProperties(ctx, uid, properties)
}

// SubscribeRealtimeData 订阅实时数据
// 通过 wisefido-qinglan: POST /api/v1/radar/devices/{uid}/subscribe
func (s *RadarInstall) SubscribeRealtimeData(ctx context.Context, uid string, content interface{}, duration int) error {
	return s.qinglanClient.SubscribeRealtimeData(ctx, uid, content, duration)
}

// CallDeviceFunction 调用设备功能（重启等）
// 通过 wisefido-qinglan: POST /api/v1/radar/devices/{uid}/function
func (s *RadarInstall) CallDeviceFunction(ctx context.Context, uid string, dev int) error {
	return s.qinglanClient.CallDeviceFunction(ctx, uid, dev)
}

// GetOriginalProperties 获取设备原始属性（v1.0 API 格式）
// keys 为空时读全部；有则按序只查这些 key，wisefido-qinglan 内分笔 MQTT、指令间 50ms。
// 返回 JSON 字符串，包含雷达配置参数
func (s *RadarInstall) GetOriginalProperties(ctx context.Context, tenantID, deviceID string, keys []string) (string, error) {
	// 1. 获取设备 UID
	uid, err := s.GetDeviceUID(ctx, tenantID, deviceID)
	if err != nil {
		return "", err
	}
	// 2. 读取属性：keys 为空读全部，否则按序查指定 key
	properties, err := s.GetDeviceProperties(ctx, uid, keys)
	if err != nil {
		return "", err
	}
	// 返回前对 ssid_password 脱敏：冒号后部分改为 *******（SSID:password 中的 password）
	if v, ok := properties["ssid_password"]; ok && v != nil {
		properties["ssid_password"] = encode.MaskSsidPassword(encode.ToStr(v))
	}

	// 3. 转换为 JSON 字符串
	jsonBytes, err := json.Marshal(properties)
	if err != nil {
		return "", fmt.Errorf("failed to marshal properties to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// GetOriginalPropertiesFromDB 从 DB 读取雷达安装配置（设备查不到时的回退）
// 当前无持久化表，返回空 JSON；后续可接 config_versions 等
func (s *RadarInstall) GetOriginalPropertiesFromDB(ctx context.Context, tenantID, deviceID string) (string, error) {
	return "{}", nil
}

// UpdateConfig 更新设备配置。
// config 与 radarMqttConfig 输出对齐：dm（画布 cm/10），install_model 统一 0/1/2，
// boundary_left/right/front/rear，可选 area_{i}_id/type/x1..y4；经 encode.EncodeV1ConfigToDeviceProps 转成
// radar_install_style、rectangle、declare_area 等后通过 qinglan 写入（安装/边界/区域可能触发重启）。
// 返回设备响应码（200=成功）和错误，供 HTTP 层透传 device_code 给前端。
func (s *RadarInstall) UpdateConfig(ctx context.Context, tenantID, deviceID string, config map[string]interface{}) (deviceCode int, err error) {
	// 1. 获取设备 UID
	uid, err := s.GetDeviceUID(ctx, tenantID, deviceID)
	if err != nil {
		return 0, err
	}

	// 2. 前端已完成格式转换（cm→dm、boundary→rectangle 等），此处透传
	properties := config
	encodeLogFields := []zap.Field{zap.String("device_id", deviceID), zap.Int("config_keys", len(config)), zap.Int("properties_keys", len(properties))}
	if v := properties["declare_area"]; v != nil {
		if s, ok := v.(string); ok {
			encodeLogFields = append(encodeLogFields, zap.String("declare_area", s))
		} else {
			encodeLogFields = append(encodeLogFields, zap.Any("declare_area", v))
		}
	}
	s.logger.Info("UpdateConfig: encode result", encodeLogFields...)
	if len(properties) == 0 {
		return 0, fmt.Errorf("config produced no device properties (check request body has install_model/height/rectangle/declare_area)")
	}

	// 便于 MQTT 排查：下发 rectangle 时打日志，可与 /prop/productId/uid/get 及设备 /post 回包对照
	if v, ok := properties["rectangle"]; ok {
		s.logger.Info("UpdateConfig: rectangle to device", zap.String("uid", uid), zap.Any("rectangle", v))
	}

	// 3. 设置设备属性，透传设备响应码给发起端
	return s.SetDeviceProperties(ctx, uid, properties)
}

// BindDevice 绑定设备（通知需要订阅该设备的数据）
// 当 vue-radar 画布中 bind 设备时调用
func (s *RadarInstall) BindDevice(ctx context.Context, tenantID, deviceID string) error {
	// 验证设备是否存在且为雷达类型
	device, err := s.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("设备不存在: %w", err)
	}
	if !device.DeviceType.Valid || !strings.EqualFold(device.DeviceType.String, "radar") {
		return fmt.Errorf("该设备不是雷达类型，无法订阅")
	}
	
	// 标记为需要订阅（加锁保护）
	s.subscribedMutex.Lock()
	s.subscribedDevices[deviceID] = true
	s.subscribedMutex.Unlock()
	
	s.logger.Info("[RADAR_BIND] device bound, subscription enabled",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID))
	return nil
}

// UnbindDevice 解绑设备（取消订阅该设备的数据）
// 当 vue-radar 画布中 unbind 设备时调用
func (s *RadarInstall) UnbindDevice(ctx context.Context, tenantID, deviceID string) error {
	// 移除订阅标记（加锁保护）
	s.subscribedMutex.Lock()
	delete(s.subscribedDevices, deviceID)
	s.subscribedMutex.Unlock()
	
	s.logger.Info("Device unbound, subscription disabled", zap.String("device_id", deviceID), zap.String("tenant_id", tenantID))
	return nil
}

// IsDeviceSubscribed 检查设备是否需要订阅
func (s *RadarInstall) IsDeviceSubscribed(deviceID string) bool {
	s.subscribedMutex.RLock()
	defer s.subscribedMutex.RUnlock()
	return s.subscribedDevices[deviceID]
}

// InitializeSubscriptionsFromLayout 从 layout 配置初始化订阅
// 解析 layout 中的 objects，找出所有已绑定的雷达设备，自动订阅
func (s *RadarInstall) InitializeSubscriptionsFromLayout(ctx context.Context, tenantID string, layoutConfig json.RawMessage) error {
	if len(layoutConfig) == 0 {
		return nil
	}
	
	var layout struct {
		Objects []struct {
			TypeName      string `json:"typeName"`
			BindedDeviceID string `json:"bindedDeviceId,omitempty"`
			DeviceID      string `json:"device_id,omitempty"`
		} `json:"objects"`
	}
	
	if err := json.Unmarshal(layoutConfig, &layout); err != nil {
		s.logger.Warn("Failed to parse layout config for subscription initialization", zap.Error(err))
		return nil // 不阻塞初始化流程
	}
	
	// 找出所有已绑定的雷达设备
	for _, obj := range layout.Objects {
		if obj.TypeName == "Radar" {
			deviceID := obj.BindedDeviceID
			if deviceID == "" {
				deviceID = obj.DeviceID
			}
			if deviceID != "" {
				// 自动订阅
				if err := s.BindDevice(ctx, tenantID, deviceID); err != nil {
					s.logger.Warn("Failed to subscribe device from layout", 
						zap.String("device_id", deviceID), 
						zap.Error(err))
					// 继续处理其他设备，不中断
				}
			}
		}
	}
	
	s.subscribedMutex.RLock()
	count := len(s.subscribedDevices)
	s.subscribedMutex.RUnlock()
	
	s.logger.Info("[RADAR_INIT] initialized subscriptions from layout", 
		zap.Int("subscribed_count", count))
	return nil
}
