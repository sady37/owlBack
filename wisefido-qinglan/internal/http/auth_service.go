package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	commonredis "owl-common/redis"
	"sync"
	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/domain"
	"wisefido-qinglan/internal/models"
	"wisefido-qinglan/internal/repository"
	"wisefido-qinglan/internal/service"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	// v2 tenant /48 host 表示（对齐 owlRD/dbv2/90_seed_tenants.sql 种子）。
	// 仓库层 SELECT host(network(set_masklen(device_ipv6, 48))) 返回的就是这个 host repr。
	platformSystemTenantID = "fd00:0:1::" // slot=1, system 默认归属
	platformTrashTenantID  = "fd00:0:2::" // slot=2, drop 设备 / 待分配回收站
	// authFailureStreamTenantID 认证失败 Stream 上的 tenant_id 占位（v2: 用 trash 池）
	authFailureStreamTenantID = platformTrashTenantID
)

// AuthService 认证服务
// 从 service 包移到此包以避免循环依赖
type AuthService struct {
	config              *config.Config
	db                  *sql.DB
	deviceRepo          repository.DeviceRepository
	redisClient         *redis.Client
	logger              *zap.Logger
	deviceCache         *sync.Map                 // 设备缓存（key: uid, value: *domain.DeviceWithLocation），使用 domain.DeviceCache
	subscriptionManager DeviceSubscriptionManager // 设备订阅管理器接口
	cardMapping         *service.CardMappingService
}

// NewAuthService 创建认证服务
func NewAuthService(cfg *config.Config, db *sql.DB, deviceRepo repository.DeviceRepository, redisClient *redis.Client, logger *zap.Logger, subscriptionManager DeviceSubscriptionManager, cardMapping *service.CardMappingService) *AuthService {
	return &AuthService{
		config:              cfg,
		db:                  db,
		deviceRepo:          deviceRepo,
		redisClient:         redisClient,
		logger:              logger,
		deviceCache:         domain.DeviceCache,
		subscriptionManager: subscriptionManager,
		cardMapping:         cardMapping,
	}
}

func (s *AuthService) AuthenticateDevice(ctx context.Context, req *models.AuthRequest, remoteAddr string) (*models.AuthResponse, error) {
	s.logger.Info("device auth request",
		zap.String("uid", req.UID),
		zap.Int("type", req.Type),
		zap.String("mcu_hw", req.MCU.HW),
		zap.String("radar_hw", req.Radar.HW),
		zap.String("remote_addr", remoteAddr),
	)

	// 1. 发布认证请求到 Redis Stream（总是执行，即使用户不存在）
	s.publishAuthRequest(ctx, req, remoteAddr)

	// 2. 验证设备合法性
	device, err := s.validateDeviceAndGetLocation(ctx, req.UID)
	if err != nil {
		// 如果 device_store 中没有记录，创建新记录（pending 状态）
		if err.Error() == "device not found in device_store" {
			s.logger.Info("device not in device_store, creating pending record",
				zap.String("uid", req.UID),
			)

			device, err = s.createDeviceStoreRecord(ctx, req.UID, req)
			if err != nil {
				s.logger.Error("Failed to create device_store record",
					zap.String("uid", req.UID),
					zap.Error(err),
				)
				s.publishAuthResponseFailure(ctx, req.UID, "failed to create device record: "+err.Error())
				return &models.AuthResponse{
					Msg:  "Device registration failed: " + err.Error(),
					Code: 401,
					Data: nil,
				}, nil
			}

			// 新创建的设备已分配给 Trash 租户（fd00:0:2::/48），access = FALSE，需要管理员审核
			s.logger.Info("Device created in device_store (pending approval, assigned to system tenant)",
				zap.String("uid", req.UID),
				zap.String("device_id", device.DeviceUID),
				zap.String("tenant_id", device.TenantID),
			)
			// 新设备需要管理员审批，直接返回 pending 响应
			// 传递刚创建的 device.DeviceUID，避免重新查询
			s.publishAuthResponseFailure(ctx, req.UID, "device access pending approval (assigned to system tenant)", device.DeviceUID)
			return &models.AuthResponse{
				Msg:  "Device pending administrator approval",
				Code: 401,
				Data: nil,
			}, nil
		} else {
			s.logger.Warn("device validation failed",
				zap.String("uid", req.UID),
				zap.Error(err),
			)
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
				Msg:  "Device validation failed: " + err.Error(),
				Code: 401,
				Data: nil,
			}, nil
		}
	}

	// 验证已在 validateDeviceAndGetLocation() 中完成
	// 如果执行到这里，说明设备已验证通过（access = true 且已分配给租户）

	// 3. 检查并更新设备硬件信息到 device_store 表
	if err := s.updateDeviceHardwareInfo(ctx, req.UID, req); err != nil {
		// 只记录 debug 级别日志，不影响认证流程
		s.logger.Debug("Hardware info update failed, continuing authentication",
			zap.String("uid", req.UID),
			zap.Error(err),
		)
		// 不返回错误，继续认证流程
	}

	// 4. 生成 MQTT 连接配置
	mqttConfig := s.generateMQTTConfig(req.UID)

	// 5. 构建响应（authUrl 明确 8443 认证入口，供支持字段的固件使用）
	authURL := s.config.DeviceAuthURL()
	response := &models.AuthResponse{
		Msg:  "Operation success",
		Code: 200,
		Data: &models.AuthData{
			UID:     req.UID,
			MQTT:    mqttConfig,
			AuthURL: authURL,
		},
	}

	s.logger.Info("device authenticated",
		zap.String("uid", req.UID),
		zap.String("tenant_id", device.TenantID),
		zap.String("device_id", device.DeviceUID),
		zap.String("device_type", device.DeviceType),
		zap.String("mqtt_server", mqttConfig.Server),
		zap.Int("mqtt_port", mqttConfig.Port),
		zap.String("auth_url", authURL),
	)

	if s.cardMapping != nil {
		s.cardMapping.RefreshBaseline(ctx, req.UID)
	}

	// 6. 发布认证成功事件到 Redis Stream
	s.publishAuthResponseSuccess(ctx, req.UID, device, mqttConfig)

	// 7. 认证成功后仅开启周期性订阅（不立即订阅MQTT主题，因为启动时已订阅）
	// 创建订阅记录，让周期性订阅机制来处理monitor订阅命令
	if s.subscriptionManager != nil {
		if err := s.subscriptionManager.EnablePeriodicSubscription(ctx, req.UID, device.DeviceUID); err != nil {
			s.logger.Warn("enable periodic subscription failed",
				zap.String("uid", req.UID),
				zap.String("device_id", device.DeviceUID),
				zap.Error(err),
			)
		} else {
			s.logger.Info("enabled periodic subscription",
				zap.String("uid", req.UID),
				zap.String("device_id", device.DeviceUID),
			)
		}
	}

	return response, nil
}

// validateDeviceAndGetLocation 验证设备合法性
// 通过 repository 层获取设备信息，符合分层架构
func (s *AuthService) validateDeviceAndGetLocation(ctx context.Context, deviceUID string) (*DeviceStoreInfo, error) {
	// 通过 repository 层获取 device_store 信息
	ds, err := s.deviceRepo.GetDeviceStoreInfo(ctx, deviceUID)
	if err != nil {
		return nil, err
	}

	// 1. 系统级接入：access=FALSE 则拒绝认证（不发 MQTT 凭证）
	if !ds.Access {
		s.logger.Info("auth denied: access=false", zap.String("device_uid", deviceUID))
		return nil, fmt.Errorf("deny")
	}

	if ds.TenantID == platformTrashTenantID {
		return nil, fmt.Errorf("device in trash tenant")
	}
	// system (000...001) 和 unallocated (000...002) 均允许通过：
	// access=TRUE 已视为审批通过，租户分配可在认证后由管理员调整。

	return ds, nil
}

// DeviceStoreInfo 设备库存信息（用于认证）
// 对应 device_store 表的完整结构
// 注意：此类型定义在 repository 包中，此处使用类型别名以保持兼容性
type DeviceStoreInfo = repository.DeviceStoreInfo

// createDeviceStoreRecord 创建新设备记录（Phase 2: 写 device_factory_meta + devices）
//
// 状态：access=FALSE（pending），分配到 trash /48 池 fd00:0:2::/48，
// 等 platform_admin 决定接受（迁 System / 真 tenant）或保持丢弃。
// 业务约定：未知 UID MQTT 来源 → Trash；已知 UID 经 wisefido-data 导入 → System。
func (s *AuthService) createDeviceStoreRecord(ctx context.Context, uid string, req *models.AuthRequest) (*DeviceStoreInfo, error) {
	deviceType := "Radar"
	if req.Type == 1 {
		deviceType = "Radar"
	}

	// 委托给 repo（Phase 2: INSERT dfm + devices，identity = device_uid）
	stub := &domain.Device{
		DeviceUID:  uid,
		DeviceType: sql.NullString{String: deviceType, Valid: true},
	}
	if err := s.deviceRepo.(interface {
		CreateDevice(ctx context.Context, device *domain.Device) error
	}).CreateDevice(ctx, stub); err != nil {
		return nil, fmt.Errorf("failed to create device factory + devices: %w", err)
	}

	// 回读完整 DeviceStoreInfo
	ds, err := s.deviceRepo.GetDeviceStoreInfo(ctx, uid)
	if err != nil {
		return &DeviceStoreInfo{
			DeviceUID:  uid,
			DeviceType: deviceType,
			TenantID:   platformTrashTenantID,
			Access:     false,
		}, nil
	}

	s.logger.Info("Created new device (pending, trash tenant pool fd00:0:2::/48)",
		zap.String("uid", uid),
		zap.String("device_uid", ds.DeviceUID),
		zap.String("device_type", ds.DeviceType),
		zap.String("tenant_id", ds.TenantID),
		zap.Bool("access", ds.Access),
	)
	return ds, nil
}

// generateMQTTConfig 生成 MQTT 连接配置
// 参考 radar-ql-v3/simple-https.py 的响应格式
func (s *AuthService) generateMQTTConfig(uid string) *models.MQTTConfig {
	cfg := s.config.MQTT.RadarDeviceMQTT

	// 检查服务器地址是否为本地回环地址（设备无法连接）
	if cfg.Server == "127.0.0.1" || cfg.Server == "localhost" || cfg.Server == "::1" {
		s.logger.Warn("mqtt server is localhost — devices cannot connect; set RADAR_MQTT_SERVER",
			zap.String("current_server", cfg.Server),
			zap.String("device_uid", uid),
		)
	}

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

// buildAuthRequestFromHTTPRequest 从 HTTP 请求构建认证请求信息
func buildAuthRequestFromHTTPRequest(
	deviceUID string,
	deviceType int,
	mcuHW string,
	mcuSW string,
	radarHW string,
	radarSW string,
	remoteAddr string,
) map[string]interface{} {
	deviceInfo := make(map[string]interface{})

	// 添加版本信息（放在 log 对象中）
	logInfo := make(map[string]interface{})
	logInfo["device_type"] = deviceType
	if mcuHW != "" {
		logInfo["mcu_hw"] = mcuHW
	}
	if mcuSW != "" {
		logInfo["mcu_sw"] = mcuSW
	}
	if radarHW != "" {
		logInfo["radar_hw"] = radarHW
	}
	if radarSW != "" {
		logInfo["radar_sw"] = radarSW
	}

	deviceInfo["log"] = logInfo

	return deviceInfo
}

// publishAuthRequest 发布设备认证请求到 Redis Stream
func (s *AuthService) publishAuthRequest(ctx context.Context, req *models.AuthRequest, remoteAddr string) {
	// 构建设备信息
	deviceInfo := buildAuthRequestFromHTTPRequest(
		req.UID,
		req.Type,
		req.MCU.HW,
		req.MCU.SW,
		req.Radar.HW,
		req.Radar.SW,
		remoteAddr,
	)

	// Phase 2 一刀切：identity 收口到 device_uid；从 dfm + devices 派生 tenant_id 和 device_addr
	resolvedTenantID := platformTrashTenantID
	var deviceAddr sql.NullString
	var storeTenantID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(host(d.device_addr), '') AS device_addr,
		       COALESCE(host(network(set_masklen(d.device_addr, 48))), '') AS tenant_id
		FROM device_factory_meta dfm
		LEFT JOIN devices d ON d.device_uid = dfm.device_uid
		WHERE dfm.device_uid = $1
		LIMIT 1
	`, req.UID).Scan(&deviceAddr, &storeTenantID)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Warn("Failed to query device_factory_meta for auth request",
			zap.String("uid", req.UID),
			zap.Error(err),
		)
	}
	if err == nil && storeTenantID.Valid && storeTenantID.String != "" {
		resolvedTenantID = storeTenantID.String
	}

	authRequest := commonredis.BuildAuthRequestMessage(req.UID, "Radar", resolvedTenantID, remoteAddr, deviceInfo)
	if deviceAddr.Valid {
		authRequest.DeviceAddr = deviceAddr.String
	}

	// 将 AuthMessage 转换为 map（用于发布到 Redis Stream）
	encodedData := authMessageToMap(authRequest, deviceAddr)

	// 发布到 Redis Stream（使用该 stream 的配置）
	streamName := commonredis.StreamAuth.Name
	maxLen, retentionSeconds := s.config.GetStreamConfig(streamName)
	streamID, err := commonredis.PublishToStream(ctx, s.redisClient, streamName, encodedData, maxLen, retentionSeconds)
	if err != nil {
		s.logger.Error("Failed to publish auth request to Redis Stream",
			zap.String("uid", req.UID),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("auth request published",
		zap.String("uid", req.UID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
		zap.String("tenant_id", resolvedTenantID),
		zap.String("device_addr", deviceAddr.String),
	)
}

// publishAuthResponseSuccess 发布认证成功响应到 Redis Stream
func (s *AuthService) publishAuthResponseSuccess(
	ctx context.Context,
	uid string,
	device *DeviceStoreInfo,
	mqttConfig *models.MQTTConfig,
) {
	// 构建认证响应消息（使用 owl-common/redis 中的统一格式）
	// 注意：auth 消息不包含位置信息（只在 event/alarm 中包含）
	authResponse := commonredis.BuildAuthResponseMessage(
		uid,
		"Radar",
		device.TenantID,
		"success",
		mqttConfig.Server,
		mqttConfig.Port,
		fmt.Sprintf("Device authenticated successfully, MQTT server: %s:%d", mqttConfig.Server, mqttConfig.Port),
	)

	// Phase 2 一刀切：identity = device_uid；查 device_addr 用作 handshake response
	var deviceAddr sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(host(d.device_addr), '')
		FROM device_factory_meta dfm
		LEFT JOIN devices d ON d.device_uid = dfm.device_uid
		WHERE dfm.device_uid = $1
		LIMIT 1
	`, uid).Scan(&deviceAddr)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Warn("Failed to query device_addr for auth response success",
			zap.String("uid", uid),
			zap.Error(err),
		)
	}
	if !deviceAddr.Valid {
		if addrText := device.DeviceAddrText(); addrText != "" {
			deviceAddr = sql.NullString{String: addrText, Valid: true}
		}
	}
	if deviceAddr.Valid {
		authResponse.DeviceAddr = deviceAddr.String
	}

	// 将 AuthMessage 转换为 map（用于发布到 Redis Stream）
	encodedData := authMessageToMap(authResponse, deviceAddr)

	// 发布到 Redis Stream（使用该 stream 的配置）
	streamName := commonredis.StreamAuth.Name
	maxLen, retentionSeconds := s.config.GetStreamConfig(streamName)
	streamID, err := commonredis.PublishToStream(ctx, s.redisClient, streamName, encodedData, maxLen, retentionSeconds)
	if err != nil {
		s.logger.Error("Failed to publish auth response to Redis Stream",
			zap.String("uid", uid),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("auth success response published",
		zap.String("uid", uid),
		zap.String("tenant_id", device.TenantID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
		zap.String("device_addr", authResponse.DeviceAddr),
	)
}

// publishAuthResponseFailure 发布认证失败响应到 Redis Stream
// Phase 2 一刀切：deviceAddr 替代旧 deviceID 参数（device_id UUID 退役）。
func (s *AuthService) publishAuthResponseFailure(ctx context.Context, uid string, errorMsg string, deviceAddrHint ...string) {
	authResponse := commonredis.BuildAuthResponseMessage(
		uid,
		"Radar",
		authFailureStreamTenantID,
		"failure",
		"",
		0,
		errorMsg,
	)

	var deviceAddrStr sql.NullString
	if len(deviceAddrHint) > 0 && deviceAddrHint[0] != "" {
		deviceAddrStr = sql.NullString{String: deviceAddrHint[0], Valid: true}
		authResponse.DeviceAddr = deviceAddrHint[0]
	} else {
		err := s.db.QueryRowContext(ctx, `
			SELECT COALESCE(host(d.device_addr), '')
			FROM device_factory_meta dfm
			LEFT JOIN devices d ON d.device_uid = dfm.device_uid
			WHERE dfm.device_uid = $1
			LIMIT 1
		`, uid).Scan(&deviceAddrStr)
		if err != nil && err != sql.ErrNoRows {
			s.logger.Warn("Failed to query device_addr for auth response failure",
				zap.String("uid", uid),
				zap.Error(err),
			)
		}
		if deviceAddrStr.Valid {
			authResponse.DeviceAddr = deviceAddrStr.String
		}
	}

	encodedData := authMessageToMap(authResponse, deviceAddrStr)

	// 发布到 Redis Stream（使用该 stream 的配置）
	streamName := commonredis.StreamAuth.Name
	maxLen, retentionSeconds := s.config.GetStreamConfig(streamName)
	streamID, err := commonredis.PublishToStream(ctx, s.redisClient, streamName, encodedData, maxLen, retentionSeconds)
	if err != nil {
		s.logger.Error("Failed to publish auth response to Redis Stream",
			zap.String("uid", uid),
			zap.String("stream", streamName),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("auth failure response published",
		zap.String("uid", uid),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
		zap.String("device_addr", authResponse.DeviceAddr),
		zap.String("error", errorMsg),
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

	// v2.5: dfm 出厂字段 + firmware_version（drs 已退役）
	queryCurrent := `
		SELECT dfm.device_type::text,
		       dfm.device_model,
		       dfm.comm_mode,
		       dfm.mcu_model,
		       dfm.firmware_version,
		       dfm.imei,
		       dfm.mac_wifi
		FROM device_factory_meta dfm
		WHERE dfm.device_uid = $1
		LIMIT 1
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
			s.logger.Warn("Device not found in device_factory_meta, skipping hardware info update",
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

	// 检查imei和mac变化（只有当请求中的值不为空且与数据库中的值不同时，才更新）
	// mac字段：只有当请求中的MAC不为空时才检查变化
	if req.MCU.MAC != "" {
		if !currentMac.Valid || currentMac.String != req.MCU.MAC {
			macChanged = true
			s.logger.Debug("MAC changed",
				zap.String("uid", uid),
				zap.String("old", currentMac.String),
				zap.String("new", req.MCU.MAC),
			)
		}
	}

	// imei字段：只有当请求中的ICCID不为空时才检查变化
	if req.MCU.ICCID != "" {
		if !currentImei.Valid || currentImei.String != req.MCU.ICCID {
			imeiChanged = true
			s.logger.Debug("IMEI/ICCID changed",
				zap.String("uid", uid),
				zap.String("old", currentImei.String),
				zap.String("new", req.MCU.ICCID),
			)
		}
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

	// 确定mac和imei字段的值
	// 如果请求中的值为空，保持数据库中的现有值不变
	macValue := req.MCU.MAC
	if macValue == "" && currentMac.Valid {
		macValue = currentMac.String
	}

	imeiValue := req.MCU.ICCID
	if imeiValue == "" && currentImei.Valid {
		imeiValue = currentImei.String
	}

	// v2 split write：
	//   - dfm 出厂字段（device_type/device_model/comm_mode/mcu_model/imei/mac_wifi/firmware_version）
	//     注：dfm 设计为"导入即不变"，但实际设备首次连接 + 重新 flash 都会上报，需允许覆盖（含 firmware）。
	updDfm := `
		UPDATE device_factory_meta
		SET device_type = $1::device_type_enum,
		    device_model = $2,
		    comm_mode = $3,
		    mcu_model = $4,
		    imei = $5,
		    mac_wifi = $6,
		    firmware_version = $7
		WHERE device_uid = $8
	`
	result, err := s.db.ExecContext(ctx, updDfm, deviceType, deviceModel, commMode, mcuModel, imeiValue, macValue, firmwareVersion, uid)
	if err != nil {
		return fmt.Errorf("failed to update device_factory_meta: %w", err)
	}

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

// OTAPropertyPublisher publishes OTA reconfig properties to devices via MQTT.
// Currently disabled -- OTA reconfig is handled by the scheduler.
// type OTAPropertyPublisher interface {
// 	SetOTAReconfig(ctx context.Context, uid string, props map[string]interface{}) error
// }

// authMessageToMap 将 AuthMessage 转换为 map（Phase 2: device_addr 替代旧 device_id）。
func authMessageToMap(msg commonredis.AuthMessage, deviceAddr sql.NullString) map[string]interface{} {
	actualDeviceAddr := ""
	if deviceAddr.Valid && deviceAddr.String != "" {
		actualDeviceAddr = deviceAddr.String
	} else if msg.DeviceAddr != "" {
		actualDeviceAddr = msg.DeviceAddr
	}

	dataValueJSON, _ := json.Marshal(msg.DataValue)

	result := make(map[string]interface{})
	if actualDeviceAddr != "" {
		result["device_addr"] = actualDeviceAddr
	}
	result["device_uid"] = msg.DeviceUID
	result["device_type"] = msg.DeviceType
	result["tenant_id"] = msg.TenantID
	result["timestamp"] = fmt.Sprintf("%d", msg.Timestamp)
	result["topic_type"] = msg.TopicType
	result["category"] = msg.Category
	result[commonredis.DataValueKey] = string(dataValueJSON)

	return result
}
