package http

import (
	"context"
	"database/sql"
	"fmt"
	"wisefido-radar/internal/config"
	"wisefido-radar/internal/models"
	"wisefido-radar/internal/repository"
	
	"go.uber.org/zap"
)

// AuthService 认证服务
// 从 service 包移到此包以避免循环依赖
type AuthService struct {
	config     *config.Config
	db         *sql.DB
	deviceRepo *repository.DeviceRepository
	logger     *zap.Logger
}

// NewAuthService 创建认证服务
func NewAuthService(cfg *config.Config, db *sql.DB, deviceRepo *repository.DeviceRepository, logger *zap.Logger) *AuthService {
	return &AuthService{
		config:     cfg,
		db:         db,
		deviceRepo: deviceRepo,
		logger:     logger,
	}
}

// AuthenticateDevice 认证设备并返回 MQTT 配置
// 参考 radar-ql-v3/simple-https.py 的实现逻辑
func (s *AuthService) AuthenticateDevice(ctx context.Context, req *models.AuthRequest) (*models.AuthResponse, error) {
	s.logger.Info("Device authentication request",
		zap.String("uid", req.UID),
		zap.Int("type", req.Type),
		zap.String("mcu_hw", req.MCU.HW),
		zap.String("radar_hw", req.Radar.HW),
	)
	
	// 1. 验证设备合法性
	device, err := s.validateDevice(ctx, req.UID)
	if err != nil {
		s.logger.Warn("Device validation failed",
			zap.String("uid", req.UID),
			zap.Error(err),
		)
		return &models.AuthResponse{
			Msg:  "设备验证失败: " + err.Error(),
			Code: 401,
			Data: nil,
		}, nil // 返回错误响应，但不返回 error（让调用者处理响应）
	}
	
	// 2. 生成 MQTT 连接配置
	mqttConfig := s.generateMQTTConfig(req.UID, device)
	
	// 3. 构建响应
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
	
	return response, nil
}

// validateDevice 验证设备合法性
// 参考 wisefido-radar/internal/repository/device.go 中的 GetOrCreateDeviceFromStore
func (s *AuthService) validateDevice(ctx context.Context, uid string) (*DeviceStoreInfo, error) {
	// 直接从 device_store 表查询（设备认证时，设备可能还未在 devices 表中创建）
	// 这与 MQTT 连接时的逻辑不同：MQTT 连接时会自动创建 devices 记录，但认证时不需要
	return s.getDeviceStoreInfo(ctx, uid)
}

// DeviceStoreInfo 设备库存信息（用于认证）
type DeviceStoreInfo struct {
	DeviceStoreID string
	DeviceType    string
	SerialNumber  string
	UID           string
	TenantID      string
	AllowAccess   bool
}

// getDeviceStoreInfo 从 device_store 表获取设备信息
func (s *AuthService) getDeviceStoreInfo(ctx context.Context, uid string) (*DeviceStoreInfo, error) {
	query := `
		SELECT
			device_store_id::text,
			device_type,
			serial_number,
			uid,
			tenant_id::text,
			allow_access
		FROM device_store
		WHERE uid = $1 OR serial_number = $1
		LIMIT 1
	`
	
	var ds DeviceStoreInfo
	var serialNumber, uidValue sql.NullString
	
	err := s.db.QueryRowContext(ctx, query, uid).Scan(
		&ds.DeviceStoreID,
		&ds.DeviceType,
		&serialNumber,
		&uidValue,
		&ds.TenantID,
		&ds.AllowAccess,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found in device_store")
		}
		return nil, fmt.Errorf("failed to query device_store: %w", err)
	}
	
	if serialNumber.Valid {
		ds.SerialNumber = serialNumber.String
	}
	if uidValue.Valid {
		ds.UID = uidValue.String
	}
	
	// 检查设备访问权限
	if !ds.AllowAccess {
		return nil, fmt.Errorf("device access not allowed")
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

