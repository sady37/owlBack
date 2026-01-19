package http

import (
	"context"
	"database/sql"
	"fmt"
	"owl-common/encode"
	commonredis "owl-common/redis"
	"sync"
	"wisefido-radar/internal/config"
	"wisefido-radar/internal/models"
	"wisefido-radar/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// AuthService 认证服务
// 从 service 包移到此包以避免循环依赖
type AuthService struct {
	config      *config.Config
	db          *sql.DB
	deviceRepo  *repository.DeviceRepository
	redisClient *redis.Client
	logger      *zap.Logger
	deviceCache *sync.Map // 设备缓存（key: uid, value: *DeviceWithLocation）
}

var (
	authServiceDeviceCache = &sync.Map{}
)

// NewAuthService 创建认证服务
func NewAuthService(cfg *config.Config, db *sql.DB, deviceRepo *repository.DeviceRepository, redisClient *redis.Client, logger *zap.Logger) *AuthService {
	return &AuthService{
		config:      cfg,
		db:          db,
		deviceRepo:  deviceRepo,
		redisClient: redisClient,
		logger:      logger,
		deviceCache: authServiceDeviceCache,
	}
}

// AuthenticateDevice 认证设备并返回 MQTT 配置，同时发布认证事件到 Redis Stream
// 参考 radar-ql-v3/simple-https.py 的实现逻辑
func (s *AuthService) AuthenticateDevice(ctx context.Context, req *models.AuthRequest, remoteAddr string) (*models.AuthResponse, error) {
	s.logger.Info("Device authentication request",
		zap.String("uid", req.UID),
		zap.Int("type", req.Type),
		zap.String("mcu_hw", req.MCU.HW),
		zap.String("radar_hw", req.Radar.HW),
		zap.String("remote_addr", remoteAddr),
	)

	// 1. 发布认证请求到 Redis Stream
	s.publishAuthRequest(ctx, req, remoteAddr)

	// 2. 验证设备合法性并查询设备位置和硬件信息
	device, locationInfo, err := s.validateDeviceAndGetLocation(ctx, req.UID)
	if err != nil {
		s.logger.Warn("Device validation failed",
			zap.String("uid", req.UID),
			zap.Error(err),
		)

		// 发布认证失败事件
		s.publishAuthResponseFailure(ctx, req.UID, err.Error())

		// 如果错误是 "deny"，直接返回 deny 响应
		if err.Error() == "deny" {
			return &models.AuthResponse{
				Msg:  "deny",
				Code: 401,
				Data: nil,
			}, nil
		}

		return &models.AuthResponse{
			Msg:  "设备验证失败: " + err.Error(),
			Code: 401,
			Data: nil,
		}, nil // 返回错误响应，但不返回 error（让调用者处理响应）
	}

	// 3. 检查并更新设备硬件信息到 device_store 表
	if err := s.updateDeviceHardwareInfo(ctx, req.UID, req); err != nil {
		// 只记录 debug 级别日志，不影响认证流程
		s.logger.Debug("Hardware info update failed, continuing authentication",
			zap.String("uid", req.UID),
			zap.Error(err),
		)
		// 不返回错误，继续认证流程
	}

	// 4. 缓存设备位置和硬件信息供后续使用
	if locationInfo != nil {
		deviceWithLocation := &DeviceWithLocation{
			Device:       nil, // auth时可能还没有devices记录
			LocationInfo: locationInfo,
		}
		s.deviceCache.Store(req.UID, deviceWithLocation)
		s.logger.Debug("Cached device location info for auth",
			zap.String("uid", req.UID),
			zap.Bool("has_location", locationInfo != nil),
		)
	}

	// 5. 生成 MQTT 连接配置
	mqttConfig := s.generateMQTTConfig(req.UID, device)

	// 6. 构建响应
	response := &models.AuthResponse{
		Msg:  "操作成功",
		Code: 200,
		Data: &models.AuthData{
			UID:  req.UID,
			MQTT: mqttConfig,
		},
	}

	s.logger.Info("Device authenticated successfully",
		zap.String("uid", req.UID),
		zap.String("tenant_id", device.TenantID),
		zap.String("device_type", device.DeviceType),
		zap.String("mqtt_server", mqttConfig.Server),
		zap.Int("mqtt_port", mqttConfig.Port),
	)

	// 7. 发布认证成功事件到 Redis Stream（使用位置信息）
	s.publishAuthResponseSuccess(ctx, req.UID, device, locationInfo, mqttConfig)

	return response, nil
}

// DeviceWithLocation 带位置信息的设备结构
type DeviceWithLocation struct {
	Device       *repository.Device
	LocationInfo *repository.DeviceLocationInfo
}

// validateDeviceAndGetLocation 验证设备合法性并获取位置信息
// 参考 wisefido-radar/internal/repository/device.go 中的 GetOrCreateDeviceFromStore
func (s *AuthService) validateDeviceAndGetLocation(ctx context.Context, uid string) (*DeviceStoreInfo, *repository.DeviceLocationInfo, error) {
	// 1. 从 device_store 表查询设备基本信息
	device, err := s.getDeviceStoreInfo(ctx, uid)
	if err != nil {
		return nil, nil, err
	}

	// 2. 查询设备位置信息（如果devices表中有记录）
	locationInfo, err := s.deviceRepo.GetDeviceLocationInfoByIdentifier(ctx, uid)
	if err != nil {
		// 设备可能还未在devices表中创建，这是正常的，返回空的位置信息
		s.logger.Debug("Device location info not found (device may not be created in devices table yet)",
			zap.String("uid", uid),
			zap.Error(err),
		)
		locationInfo = &repository.DeviceLocationInfo{}
	}

	return device, locationInfo, nil
}

// DeviceStoreInfo 设备库存信息（用于认证）
type DeviceStoreInfo struct {
	UID         string
	DeviceType  string
	TenantID    string
	AllowAccess bool
}

// getDeviceStoreInfo 从 device_store 表获取设备信息
func (s *AuthService) getDeviceStoreInfo(ctx context.Context, uid string) (*DeviceStoreInfo, error) {
	query := `
		SELECT
			device_uid,
			device_type,
			tenant_id::text,
			allow_access
		FROM device_store
		WHERE device_uid = $1
		LIMIT 1
	`

	var ds DeviceStoreInfo

	err := s.db.QueryRowContext(ctx, query, uid).Scan(
		&ds.UID,
		&ds.DeviceType,
		&ds.TenantID,
		&ds.AllowAccess,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found in device_store")
		}
		return nil, fmt.Errorf("failed to query device_store: %w", err)
	}

	// 检查设备访问权限：allow_access 必须为 True，否则拒绝认证
	if !ds.AllowAccess {
		return nil, fmt.Errorf("deny")
	}

	// 检查设备是否已分配给租户
	unallocatedTenantID := "00000000-0000-0000-0000-000000000000"
	if ds.TenantID == unallocatedTenantID {
		return nil, fmt.Errorf("device not allocated to tenant")
	}

	return &ds, nil
}

// generateMQTTConfig 生成 MQTT 连接配置
// 参考 radar-ql-v3/simple-https.py 的响应格式
func (s *AuthService) generateMQTTConfig(uid string, device *DeviceStoreInfo) *models.MQTTConfig {
	cfg := s.config.Radar.DeviceMQTT

	// 生成 clientId（使用 uid 作为标识）
	clientID := fmt.Sprintf("radar-%s", uid)
	if cfg.ClientIDPrefix != "" {
		clientID = fmt.Sprintf("%s-%s", cfg.ClientIDPrefix, uid)
	}

	return &models.MQTTConfig{
		ClientID:  clientID,
		Timeout:   cfg.Timeout,
		Keepalive: cfg.Keepalive,
		Server:    cfg.Server,
		Port:      cfg.Port,
		Account:   cfg.Account,
		Pwd:       cfg.Password,
		Protocol:  cfg.Protocol,
		Prefix:    cfg.Prefix,
		ProductID: cfg.ProductID,
	}
}

// publishAuthRequest 发布设备认证请求到 Redis Stream
func (s *AuthService) publishAuthRequest(ctx context.Context, req *models.AuthRequest, remoteAddr string) {
	// 构建设备信息
	deviceInfo := encode.BuildAuthRequestFromHTTPRequest(
		req.UID,
		req.Type,
		req.MCU.HW,
		req.MCU.SW,
		req.Radar.HW,
		req.Radar.SW,
		remoteAddr,
	)

	// 编码为 Redis Stream 事件
	authRequest := encode.EncodeAuthRequest(req.UID, "Radar", remoteAddr, deviceInfo)

	// 验证事件
	if err := encode.ValidateAuthStreamEvent(authRequest); err != nil {
		s.logger.Warn("Failed to validate auth request event",
			zap.String("uid", req.UID),
			zap.Error(err),
		)
		return
	}

	// 查询 device_id（通过 device_uid 从 devices 表查询，如果设备还未创建则返回 nil）
	var deviceID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT device_id::text
		FROM devices
		WHERE device_uid = $1
		LIMIT 1
	`, req.UID).Scan(&deviceID)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Warn("Failed to query device_id for auth request",
			zap.String("uid", req.UID),
			zap.Error(err),
		)
	}

	// 使用标准格式构建输出
	// 字段顺序：device_id → device_uid → device_type → tenant_id → timestamp → topic_type → category → data_value → 位置信息
	// 提取 category 字段到顶层
	category := ""
	if dataValue, ok := authRequest.DataValue["category"].(string); ok {
		category = dataValue
	}

	// 构建完整的输出对象
	// 字段顺序：device_id → device_uid → device_type → tenant_id → timestamp → topic_type → category → data_value → 位置信息
	encodedData := map[string]interface{}{
		"device_id":   getStringOrNullFromNullString(deviceID),
		"device_uid":  authRequest.DeviceUID,
		"device_type": authRequest.DeviceType,
		"tenant_id":   authRequest.TenantID,
		"timestamp":   authRequest.Timestamp,
		"topic_type":  authRequest.TopicType,
		"category":    category,
		"data_value":  authRequest.DataValue, // 直接使用对象，不要序列化为字符串
	}

	// 添加可选的位置信息字段（都设为 null，auth_request时设备可能还未绑定位置）
	encodedData["branch_id"] = nil
	encodedData["building_id"] = nil
	encodedData["unit_id"] = nil
	encodedData["room_id"] = nil
	encodedData["bed_id"] = nil

	// 发布到 Redis Stream（使用该 stream 的配置）
	streamName := "iot:auth:stream"
	maxLen, retentionSeconds := s.config.GetStreamConfig(streamName)
	streamID, err := commonredis.PublishJSONToStream(ctx, s.redisClient, streamName, encodedData, maxLen, retentionSeconds)
	if err != nil {
		s.logger.Error("Failed to publish auth request to Redis Stream",
			zap.String("uid", req.UID),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("Published auth request to Redis Stream",
		zap.String("uid", req.UID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)
}

// publishAuthResponseSuccess 发布认证成功响应到 Redis Stream
func (s *AuthService) publishAuthResponseSuccess(
	ctx context.Context,
	uid string,
	device *DeviceStoreInfo,
	locationInfo *repository.DeviceLocationInfo,
	mqttConfig *models.MQTTConfig,
) {
	// 编码为 Redis Stream 事件
	authResponse := encode.EncodeAuthResponse(
		uid,
		"Radar",
		device.TenantID,
		"success",
		mqttConfig.Server,
		mqttConfig.Port,
		fmt.Sprintf("Device authenticated successfully, MQTT server: %s:%d", mqttConfig.Server, mqttConfig.Port),
	)

	// 验证事件
	if err := encode.ValidateAuthStreamEvent(authResponse); err != nil {
		s.logger.Warn("Failed to validate auth response event",
			zap.String("uid", uid),
			zap.Error(err),
		)
		return
	}

	// 查询 device_id（通过 device_uid 从 devices 表查询）
	var deviceID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT device_id::text
		FROM devices
		WHERE device_uid = $1
		LIMIT 1
	`, uid).Scan(&deviceID)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Warn("Failed to query device_id for auth response success",
			zap.String("uid", uid),
			zap.Error(err),
		)
	}

	// 使用标准格式构建输出
	// 提取 category 字段到顶层
	category := ""
	if dataValue, ok := authResponse.DataValue["category"].(string); ok {
		category = dataValue
	}

	// 构建完整的输出对象
	// 字段顺序：device_id → device_uid → device_type → tenant_id → timestamp → topic_type → category → data_value → 位置信息
	encodedData := map[string]interface{}{
		"device_id":   getStringOrNullFromNullString(deviceID),
		"device_uid":  authResponse.DeviceUID,
		"device_type": authResponse.DeviceType,
		"tenant_id":   authResponse.TenantID,
		"timestamp":   authResponse.Timestamp,
		"topic_type":  authResponse.TopicType,
		"category":    category,
		"data_value":  authResponse.DataValue, // 直接使用对象，不要序列化为字符串
	}

	// 添加可选的位置信息字段（从 locationInfo 中提取）
	if locationInfo != nil {
		encodedData["branch_id"] = locationInfo.BranchID
		encodedData["building_id"] = locationInfo.BuildingID
		encodedData["unit_id"] = locationInfo.UnitID
		encodedData["room_id"] = locationInfo.RoomID
		encodedData["bed_id"] = locationInfo.BedID
	} else {
		encodedData["branch_id"] = nil
		encodedData["building_id"] = nil
		encodedData["unit_id"] = nil
		encodedData["room_id"] = nil
		encodedData["bed_id"] = nil
	}

	// 发布到 Redis Stream（使用该 stream 的配置）
	streamName := "iot:auth:stream"
	maxLen, retentionSeconds := s.config.GetStreamConfig(streamName)
	streamID, err := commonredis.PublishJSONToStream(ctx, s.redisClient, streamName, encodedData, maxLen, retentionSeconds)
	if err != nil {
		s.logger.Error("Failed to publish auth response to Redis Stream",
			zap.String("uid", uid),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("Published auth response (success) to Redis Stream",
		zap.String("uid", uid),
		zap.String("tenant_id", device.TenantID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)
}

// publishAuthResponseFailure 发布认证失败响应到 Redis Stream
func (s *AuthService) publishAuthResponseFailure(ctx context.Context, uid string, errorMsg string) {
	// 编码为 Redis Stream 事件
	authResponse := encode.EncodeAuthResponse(
		uid,
		"Radar",
		"00000000-0000-0000-0000-000000000000", // 失败时使用默认值
		"failure",
		"", // 失败时无 MQTT 服务器
		0,  // 失败时无 MQTT 端口
		errorMsg,
	)

	// 验证事件
	if err := encode.ValidateAuthStreamEvent(authResponse); err != nil {
		s.logger.Warn("Failed to validate auth response event",
			zap.String("uid", uid),
			zap.Error(err),
		)
		return
	}

	// 使用标准格式构建输出
	// 提取 category 字段到顶层
	category := ""
	if dataValue, ok := authResponse.DataValue["category"].(string); ok {
		category = dataValue
	}

	// 查询 device_id（通过 device_uid 从 devices 表查询）
	var deviceID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT device_id::text
		FROM devices
		WHERE device_uid = $1
		LIMIT 1
	`, uid).Scan(&deviceID)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Warn("Failed to query device_id for auth response failure",
			zap.String("uid", uid),
			zap.Error(err),
		)
	}

	// 构建完整的输出对象
	// 字段顺序：device_id → device_uid → device_type → tenant_id → timestamp → topic_type → category → data_value → 位置信息
	encodedData := map[string]interface{}{
		"device_id":   getStringOrNullFromNullString(deviceID),
		"device_uid":  authResponse.DeviceUID,
		"device_type": authResponse.DeviceType,
		"tenant_id":   authResponse.TenantID,
		"timestamp":   authResponse.Timestamp,
		"topic_type":  authResponse.TopicType,
		"category":    category,
		"data_value":  authResponse.DataValue, // 直接使用对象，不要序列化为字符串
	}

	// 添加可选的位置信息字段（失败时设为 null）
	encodedData["branch_id"] = nil
	encodedData["building_id"] = nil
	encodedData["unit_id"] = nil
	encodedData["room_id"] = nil
	encodedData["bed_id"] = nil

	// 发布到 Redis Stream（使用该 stream 的配置）
	streamName := "iot:auth:stream"
	maxLen, retentionSeconds := s.config.GetStreamConfig(streamName)
	streamID, err := commonredis.PublishJSONToStream(ctx, s.redisClient, streamName, encodedData, maxLen, retentionSeconds)
	if err != nil {
		s.logger.Error("Failed to publish auth response to Redis Stream",
			zap.String("uid", uid),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("Published auth response (failure) to Redis Stream",
		zap.String("uid", uid),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)
}

// updateDeviceHardwareInfo 更新设备硬件信息到 device_store 表
// 只在硬件信息发生变化时更新
func (s *AuthService) updateDeviceHardwareInfo(ctx context.Context, uid string, req *models.AuthRequest) error {
	// 检查是否有完整的硬件信息
	if req.MCU.HW == "" || req.MCU.SW == "" || req.Radar.HW == "" || req.Radar.SW == "" {
		s.logger.Info("Incomplete hardware info, skipping update",
			zap.String("uid", uid),
			zap.String("mcu_hw", req.MCU.HW),
			zap.String("mcu_sw", req.MCU.SW),
			zap.String("radar_hw", req.Radar.HW),
			zap.String("radar_sw", req.Radar.SW),
		)
		return nil
	}

	// 根据 type 字段确定 device_type
	// type=1 -> Radar, 其他类型根据实际需求映射
	deviceType := "Radar" // 默认值
	if req.Type == 1 {
		deviceType = "Radar"
	} else {
		// 可以根据其他type值映射不同的device_type
		s.logger.Warn("Unknown device type, using default 'Radar'",
			zap.String("uid", uid),
			zap.Int("type", req.Type),
		)
	}

	// 根据 MCU.HW 解析设备型号和通讯方式
	deviceModel := ""
	commMode := ""

	// 解析设备型号和通讯方式（根据文档）
	// 4G设备: MCU.hw = "1.0", "1.1"
	// Wi-Fi设备: HC2对应MCU.hw = "2.0", TK2对应MCU.hw = "2.3"
	if req.MCU.HW == "1.0" || req.MCU.HW == "1.1" {
		// 4G设备
		deviceModel = req.MCU.HW
		commMode = "4G"
	} else if req.MCU.HW == "2.0" {
		// HC2 Wi-Fi设备
		deviceModel = "HC2"
		commMode = "wifi"
	} else if req.MCU.HW == "2.3" {
		// TK2 Wi-Fi设备
		deviceModel = "Tk2"
		commMode = "wifi"
	} else {
		// 其他情况，保持原值，默认假设为wifi
		deviceModel = req.MCU.HW
		commMode = "wifi"
		s.logger.Warn("Unknown MCU.HW value, using default",
			zap.String("uid", uid),
			zap.String("mcu_hw", req.MCU.HW),
			zap.String("device_model", deviceModel),
			zap.String("comm_mode", commMode),
		)
	}

	// 拼接硬件信息（与数据库中存储的格式一致）
	// mcu_model: "2.0-Dec 17 2025 10:22:19"
	mcuModel := fmt.Sprintf("%s-%s", req.MCU.HW, req.MCU.SW)
	// firmware_version: "2.3-Jun 25 2025 11:33:44"
	firmwareVersion := fmt.Sprintf("%s-%s", req.Radar.HW, req.Radar.SW)

	// 首先查询当前的所有硬件相关信息（包括mac和imei）
	queryCurrent := `
		SELECT device_type, device_model, comm_mode, mcu_model, firmware_version, imei, mac
		FROM device_store 
		WHERE device_uid = $1
	`

	var currentDeviceType, currentDeviceModel, currentCommMode, currentMcuModel, currentFirmwareVersion sql.NullString
	var currentImei, currentMac sql.NullString
	err := s.db.QueryRowContext(ctx, queryCurrent, uid).Scan(
		&currentDeviceType,
		&currentDeviceModel,
		&currentCommMode,
		&currentMcuModel,
		&currentFirmwareVersion,
		&currentImei,
		&currentMac,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("Device not found in device_store, skipping hardware info update",
				zap.String("uid", uid),
			)
			return nil
		}
		s.logger.Error("Failed to query current hardware info",
			zap.String("uid", uid),
			zap.Error(err),
		)
		return fmt.Errorf("failed to query current hardware info: %w", err)
	}

	// 检查所有硬件信息是否有变化
	deviceTypeChanged := false
	deviceModelChanged := false
	commModeChanged := false
	mcuChanged := false
	firmwareChanged := false
	imeiChanged := false
	macChanged := false

	// 检查device_type变化
	if !currentDeviceType.Valid || currentDeviceType.String != deviceType {
		deviceTypeChanged = true
		s.logger.Debug("Device type changed",
			zap.String("uid", uid),
			zap.String("old", currentDeviceType.String),
			zap.String("new", deviceType),
		)
	}

	// 检查device_model变化
	if !currentDeviceModel.Valid || currentDeviceModel.String != deviceModel {
		deviceModelChanged = true
		s.logger.Debug("Device model changed",
			zap.String("uid", uid),
			zap.String("old", currentDeviceModel.String),
			zap.String("new", deviceModel),
		)
	}

	// 检查comm_mode变化
	if !currentCommMode.Valid || currentCommMode.String != commMode {
		commModeChanged = true
		s.logger.Debug("Comm mode changed",
			zap.String("uid", uid),
			zap.String("old", currentCommMode.String),
			zap.String("new", commMode),
		)
	}

	// 检查mcu_model变化
	if !currentMcuModel.Valid || currentMcuModel.String != mcuModel {
		mcuChanged = true
		s.logger.Debug("MCU model changed",
			zap.String("uid", uid),
			zap.String("old", currentMcuModel.String),
			zap.String("new", mcuModel),
		)
	}

	// 检查firmware_version变化
	if !currentFirmwareVersion.Valid || currentFirmwareVersion.String != firmwareVersion {
		firmwareChanged = true
		s.logger.Debug("Firmware version changed",
			zap.String("uid", uid),
			zap.String("old", currentFirmwareVersion.String),
			zap.String("new", firmwareVersion),
		)
	}

	// 检查imei和mac变化（mac和imei是独立的两个字段，直接取值，不做假设）
	// mac字段：直接使用req.MCU.MAC的值
	if !currentMac.Valid || currentMac.String != req.MCU.MAC {
		macChanged = true
		s.logger.Debug("MAC changed",
			zap.String("uid", uid),
			zap.String("old", currentMac.String),
			zap.String("new", req.MCU.MAC),
		)
	}

	// imei字段：直接使用req.MCU.ICCID的值
	if !currentImei.Valid || currentImei.String != req.MCU.ICCID {
		imeiChanged = true
		s.logger.Debug("IMEI/ICCID changed",
			zap.String("uid", uid),
			zap.String("old", currentImei.String),
			zap.String("new", req.MCU.ICCID),
		)
	}

	// 如果没有变化，跳过更新
	if !deviceTypeChanged && !deviceModelChanged && !commModeChanged && !mcuChanged && !firmwareChanged && !imeiChanged && !macChanged {
		s.logger.Info("Hardware info unchanged, skipping update",
			zap.String("uid", uid),
			zap.String("device_type", deviceType),
			zap.String("device_model", deviceModel),
			zap.String("comm_mode", commMode),
			zap.String("mcu_model", mcuModel),
			zap.String("firmware_version", firmwareVersion),
			zap.String("mac", req.MCU.MAC),
			zap.String("iccid", req.MCU.ICCID),
		)
		return nil
	}

	// 确定mac和imei字段的值（两个独立的字段，直接取值，不做假设）
	macValue := req.MCU.MAC
	imeiValue := req.MCU.ICCID

	// 只要有任何一个字段变化，就更新所有字段
	queryUpdate := `
		UPDATE device_store 
		SET device_type = $1,
		    device_model = $2,
		    comm_mode = $3,
		    mcu_model = $4,
		    firmware_version = $5,
		    imei = $6,
		    mac = $7
		WHERE device_uid = $8
	`
	params := []interface{}{
		deviceType,
		deviceModel,
		commMode,
		mcuModel,
		firmwareVersion,
		imeiValue,
		macValue,
		uid,
	}

	result, err := s.db.ExecContext(ctx, queryUpdate, params...)

	if err != nil {
		s.logger.Error("Failed to update device hardware info",
			zap.String("uid", uid),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update device hardware info: %w", err)
	}

	// 检查是否真的更新了记录
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		changeLog := []string{}
		if deviceTypeChanged {
			changeLog = append(changeLog, "device_type")
		}
		if deviceModelChanged {
			changeLog = append(changeLog, "device_model")
		}
		if commModeChanged {
			changeLog = append(changeLog, "comm_mode")
		}
		if mcuChanged {
			changeLog = append(changeLog, "mcu_model")
		}
		if firmwareChanged {
			changeLog = append(changeLog, "firmware_version")
		}
		if imeiChanged {
			changeLog = append(changeLog, "imei")
		}
		if macChanged {
			changeLog = append(changeLog, "mac")
		}

		s.logger.Info("Updated device hardware info",
			zap.String("uid", uid),
			zap.Strings("changed_fields", changeLog),
			zap.String("device_type", deviceType),
			zap.String("device_model", deviceModel),
			zap.String("comm_mode", commMode),
			zap.String("mcu_model", mcuModel),
			zap.String("firmware_version", firmwareVersion),
			zap.String("mac", req.MCU.MAC),
			zap.String("iccid", req.MCU.ICCID),
			zap.Int64("rows_affected", rowsAffected),
		)
	} else {
		s.logger.Warn("No rows affected when updating device hardware info",
			zap.String("uid", uid),
			zap.String("device_type", deviceType),
			zap.String("device_model", deviceModel),
			zap.String("comm_mode", commMode),
			zap.String("mcu_model", mcuModel),
			zap.String("firmware_version", firmwareVersion),
			zap.String("mac", req.MCU.MAC),
			zap.String("iccid", req.MCU.ICCID),
		)
	}

	return nil
}

// getStringOrNullFromNullString 如果 sql.NullString 有效，返回字符串，否则返回 nil
func getStringOrNullFromNullString(s sql.NullString) interface{} {
	if s.Valid && s.String != "" {
		return s.String
	}
	return nil
}
