package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	commonredis "owl-common/redis"
	"sync"
	"wisefido-qinglan/internal/config"
	"wisefido-qinglan/internal/domain"
	"wisefido-qinglan/internal/models"
	"wisefido-qinglan/internal/repository"
	"wisefido-qinglan/internal/service"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	platformTrashTenantID       = "00000000-0000-0000-0000-000000000000"
	platformSystemTenantID      = "00000000-0000-0000-0000-000000000001"
	platformUnallocatedTenantID = "00000000-0000-0000-0000-000000000002"
	// authFailureStreamTenantID 认证失败写 Auth Stream 的 tenant_id：设备可能在库，仅占位，不表示 device_store 已改为 Unallocated
	authFailureStreamTenantID = platformUnallocatedTenantID
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
	// 输出到标准输出，确保总是可见
	log.Printf("=== Auth Request === Device UID: %s, Type: %d, MCU_HW: %s, Radar_HW: %s, RemoteAddr: %s",
		req.UID, req.Type, req.MCU.HW, req.Radar.HW, remoteAddr)

	s.logger.Info("Device authentication request",
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
			log.Printf("❌ Device %s not found in device_store, creating new record (pending)", req.UID)
			s.logger.Info("Device not found in device_store, creating new record (pending)",
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

			// 新创建的设备已分配给系统租户（000...001），allow_access = FALSE，需要管理员处理
			s.logger.Info("Device created in device_store (pending approval, assigned to system tenant)",
				zap.String("uid", req.UID),
				zap.String("device_id", device.DeviceID),
				zap.String("tenant_id", device.TenantID),
			)
			// 新设备需要管理员审批，直接返回 pending 响应
			// 传递刚创建的 device.DeviceID，避免重新查询
			s.publishAuthResponseFailure(ctx, req.UID, "device access pending approval (assigned to system tenant)", device.DeviceID)
			return &models.AuthResponse{
				Msg:  "Device pending administrator approval",
				Code: 401,
				Data: nil,
			}, nil
		} else {
			log.Printf("❌ Device validation failed for %s: %v", req.UID, err)
			s.logger.Warn("Device validation failed",
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

			// 如果是 pending 状态（分配给系统租户），返回 pending 响应
			if err.Error() == "device pending approval (assigned to system tenant)" {
				return &models.AuthResponse{
					Msg:  "Device pending administrator approval",
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
	// 如果执行到这里，说明设备已验证通过（allow_access = true 且已分配给租户）

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

	s.logger.Info("Device authenticated successfully",
		zap.String("uid", req.UID),
		zap.String("tenant_id", device.TenantID),
		zap.String("device_type", device.DeviceType),
		zap.String("mqtt_server", mqttConfig.Server),
		zap.Int("mqtt_port", mqttConfig.Port),
		zap.String("auth_url", authURL),
	)
	log.Printf("[AUTH_FLOW] uid=%s step=device_validated tenant_id=%s device_id=%s device_type=%s mqtt_server=%s mqtt_port=%d",
		req.UID, device.TenantID, device.DeviceID, device.DeviceType, mqttConfig.Server, mqttConfig.Port)

	if s.cardMapping != nil {
		s.cardMapping.RefreshBaseline(ctx, req.UID)
	}

	// 6. 发布认证成功事件到 Redis Stream
	log.Printf("✅ Auth success for device %s, publishing success response", req.UID)
	s.publishAuthResponseSuccess(ctx, req.UID, device, mqttConfig)

	// 7. 认证成功后仅开启周期性订阅（不立即订阅MQTT主题，因为启动时已订阅）
	// 创建订阅记录，让周期性订阅机制来处理monitor订阅命令
	if s.subscriptionManager != nil {
		if err := s.subscriptionManager.EnablePeriodicSubscription(ctx, req.UID, device.DeviceID); err != nil {
			s.logger.Warn("Failed to enable periodic subscription after authentication",
				zap.String("uid", req.UID),
				zap.String("device_id", device.DeviceID),
				zap.Error(err),
			)
			log.Printf("⚠️ Failed to enable periodic subscription for device %s after authentication: %v", req.UID, err)
			// 不返回错误，认证仍然成功，订阅失败不影响认证结果
		} else {
			s.logger.Info("Enabled periodic subscription after authentication",
				zap.String("uid", req.UID),
				zap.String("device_id", device.DeviceID),
			)
			log.Printf("✅ Enabled periodic subscription for device %s after authentication", req.UID)
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

	// 1. 系统级接入：allow_access=FALSE 则拒绝认证（不发 MQTT 凭证）
	if !ds.AllowAccess {
		log.Printf("auth denied: allow_access=false device_uid=%s", deviceUID)
		return nil, fmt.Errorf("deny")
	}

	if ds.TenantID == platformTrashTenantID {
		return nil, fmt.Errorf("device in trash tenant")
	}
	if ds.TenantID == platformSystemTenantID {
		return nil, fmt.Errorf("device pending approval (assigned to system tenant)")
	}
	if ds.TenantID == platformUnallocatedTenantID {
		return nil, fmt.Errorf("device not allocated to tenant")
	}

	return ds, nil
}

// DeviceStoreInfo 设备库存信息（用于认证）
// 对应 device_store 表的完整结构
// 注意：此类型定义在 repository 包中，此处使用类型别名以保持兼容性
type DeviceStoreInfo = repository.DeviceStoreInfo

// createDeviceStoreRecord 在 device_store 表中创建新设备记录（当设备不存在时）
// 状态设为 pending（allow_access = FALSE），让客户在前端处理
func (s *AuthService) createDeviceStoreRecord(ctx context.Context, uid string, req *models.AuthRequest) (*DeviceStoreInfo, error) {
	// 根据 type 字段确定 device_type
	deviceType := "Radar" // 默认值
	if req.Type == 1 {
		deviceType = "Radar"
	}

	// 生成 device_id (UUID)
	deviceID := uuid.New().String()

	// 插入新记录，分配给系统租户（000...001），allow_access = FALSE（pending 状态）
	// 系统管理员可以在前端看到这些设备并处理
	query := `
		INSERT INTO device_store (
			device_id, device_uid, device_type, tenant_id, allow_access
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING device_id, device_uid, device_type, tenant_id, allow_access
	`

	var ds DeviceStoreInfo

	err := s.db.QueryRowContext(ctx, query,
		deviceID, uid, deviceType, platformSystemTenantID, false,
	).Scan(
		&ds.DeviceID,
		&ds.DeviceUID,
		&ds.DeviceType,
		&ds.TenantID,
		&ds.AllowAccess,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create device_store record: %w", err)
	}

	s.logger.Info("Created new device_store record (pending, assigned to system tenant)",
		zap.String("uid", uid),
		zap.String("device_id", ds.DeviceID),
		zap.String("device_type", ds.DeviceType),
		zap.String("tenant_id", ds.TenantID),
		zap.Bool("allow_access", ds.AllowAccess),
	)

	return &ds, nil
}

// generateMQTTConfig 生成 MQTT 连接配置
// 参考 radar-ql-v3/simple-https.py 的响应格式
func (s *AuthService) generateMQTTConfig(uid string) *models.MQTTConfig {
	cfg := s.config.MQTT.RadarDeviceMQTT

	// 检查服务器地址是否为本地回环地址（设备无法连接）
	if cfg.Server == "127.0.0.1" || cfg.Server == "localhost" || cfg.Server == "::1" {
		s.logger.Warn("⚠️ MQTT server address is localhost, devices cannot connect! Please set RADAR_MQTT_SERVER environment variable",
			zap.String("current_server", cfg.Server),
			zap.String("device_uid", uid),
		)
		log.Printf("⚠️ WARNING: MQTT server address is %s, devices cannot connect! Please set RADAR_MQTT_SERVER environment variable", cfg.Server)
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

	// 先查 device_store：在库用真实 tenant_id，未入库则 auth 流用 trash
	resolvedTenantID := platformTrashTenantID
	var deviceID sql.NullString
	var storeTenantID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT device_id::text, tenant_id::text
		FROM device_store
		WHERE device_uid = $1
		LIMIT 1
	`, req.UID).Scan(&deviceID, &storeTenantID)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Warn("Failed to query device_store for auth request",
			zap.String("uid", req.UID),
			zap.Error(err),
		)
	}
	if err == nil && storeTenantID.Valid && storeTenantID.String != "" {
		resolvedTenantID = storeTenantID.String
	}

	authRequest := commonredis.BuildAuthRequestMessage(req.UID, "Radar", resolvedTenantID, remoteAddr, deviceInfo)
	if deviceID.Valid {
		authRequest.DeviceID = deviceID.String
	}

	// 将 AuthMessage 转换为 map（用于发布到 Redis Stream）
	// 使用 owl-common/redis 中的 AuthMessage 类型定义
	encodedData := authMessageToMap(authRequest, deviceID)

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

	s.logger.Info("Published auth request to Redis Stream",
		zap.String("uid", req.UID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)
	// 输出到标准输出，便于在终端查看
	log.Printf("Published auth request to stream %s (stream_id: %s) for device %s", streamName, streamID, req.UID)
	log.Printf("[AUTH_FLOW] uid=%s step=auth_request_stream_published stream=%s stream_id=%s tenant_id=%s device_id=%s",
		req.UID, streamName, streamID, resolvedTenantID, deviceID.String)
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

	// 查询 device_id（优先从 devices 表，否则使用 device_store）
	var deviceID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT device_id::text
		FROM devices
		WHERE device_uid = $1
		LIMIT 1
	`, uid).Scan(&deviceID)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Warn("Failed to query device_id from devices table for auth response success",
			zap.String("uid", uid),
			zap.Error(err),
		)
	}
	if !deviceID.Valid {
		deviceID = sql.NullString{String: device.DeviceID, Valid: true}
	}
	if deviceID.Valid {
		authResponse.DeviceID = deviceID.String
	}

	// 将 AuthMessage 转换为 map（用于发布到 Redis Stream）
	// 使用 owl-common/redis 中的 AuthMessage 类型定义
	encodedData := authMessageToMap(authResponse, deviceID)

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

	s.logger.Info("Published auth response (success) to Redis Stream",
		zap.String("uid", uid),
		zap.String("tenant_id", device.TenantID),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)
	// 输出到标准输出，便于在终端查看
	log.Printf("Published auth response (success) to stream %s (stream_id: %s) for device %s", streamName, streamID, uid)
	log.Printf("[AUTH_FLOW] uid=%s step=auth_response_success_stream_published stream=%s stream_id=%s tenant_id=%s device_id=%s",
		uid, streamName, streamID, device.TenantID, authResponse.DeviceID)
}

// publishAuthResponseFailure 发布认证失败响应到 Redis Stream
func (s *AuthService) publishAuthResponseFailure(ctx context.Context, uid string, errorMsg string, deviceID ...string) {
	// 认证失败：设备仍可能在 device_store；Stream 上 tenant_id 用 Unallocated 占位，不表示库内 tenant 已改为 002。
	authResponse := commonredis.BuildAuthResponseMessage(
		uid,
		"Radar",
		authFailureStreamTenantID,
		"failure",
		"", // 失败时无 MQTT 服务器
		0,  // 失败时无 MQTT 端口
		errorMsg,
	)

	// 如果提供了 deviceID 参数，直接使用；否则查询数据库
	var deviceIDStr sql.NullString
	if len(deviceID) > 0 && deviceID[0] != "" {
		deviceIDStr = sql.NullString{String: deviceID[0], Valid: true}
		authResponse.DeviceID = deviceID[0]
	} else {
		// 查询 device_id（通过 device_uid 从 device_store 表查询）
		err := s.db.QueryRowContext(ctx, `
			SELECT device_id::text
			FROM device_store
			WHERE device_uid = $1
			LIMIT 1
		`, uid).Scan(&deviceIDStr)
		if err != nil && err != sql.ErrNoRows {
			s.logger.Warn("Failed to query device_id from device_store for auth response failure",
				zap.String("uid", uid),
				zap.Error(err),
			)
		}
		if deviceIDStr.Valid {
			authResponse.DeviceID = deviceIDStr.String
		}
	}

	// 将 AuthMessage 转换为 map（用于发布到 Redis Stream）
	// 使用 owl-common/redis 中的 AuthMessage 类型定义
	encodedData := authMessageToMap(authResponse, deviceIDStr)

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

	s.logger.Info("Published auth response (failure) to Redis Stream",
		zap.String("uid", uid),
		zap.String("stream", streamName),
		zap.String("stream_id", streamID),
	)
	// 输出到标准输出，便于在终端查看
	log.Printf("Published auth response (failure) to stream %s (stream_id: %s) for device %s", streamName, streamID, uid)
	log.Printf("[AUTH_FLOW] uid=%s step=auth_response_failure_stream_published stream=%s stream_id=%s device_id=%s error=%s",
		uid, streamName, streamID, authResponse.DeviceID, errorMsg)
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

// authMessageToMap 将 AuthMessage 转换为 map（用于发布到 Redis Stream）
// 使用 IoTStreamMessage 格式（与 iot:monitor:stream 等保持一致）
// 字段顺序：device_id, device_uid, device_type, tenant_id, timestamp, topic_type, category, data_value
// 注意：Go 的 map 是无序的，所以 Redis Stream 中字段顺序可能随机，但不影响功能（通过字段名访问）
func authMessageToMap(msg commonredis.AuthMessage, deviceID sql.NullString) map[string]interface{} {
	// 确定 device_id
	actualDeviceID := ""
	if deviceID.Valid && deviceID.String != "" {
		actualDeviceID = deviceID.String
	} else if msg.DeviceID != "" {
		actualDeviceID = msg.DeviceID
	}

	// 序列化 dataValue 为 JSON 字符串（PublishToStream 需要字符串值）
	dataValueJSON, _ := json.Marshal(msg.DataValue)

	result := make(map[string]interface{})

	// 按 IoTStreamMessage 定义顺序添加字段（与 iot:monitor:stream 等格式一致）
	// 注意：虽然代码中按顺序添加，但 Go map 是无序的，Redis Stream 中显示顺序可能随机
	if actualDeviceID != "" {
		result["device_id"] = actualDeviceID
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
