package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	"go.uber.org/zap"
)

// AlarmCloudService 告警配置服务接口
type AlarmCloudService interface {
	GetAlarmCloudConfig(ctx context.Context, req GetAlarmCloudConfigRequest) (*AlarmCloudConfigResponse, error)
	UpdateAlarmCloudConfig(ctx context.Context, req UpdateAlarmCloudConfigRequest) (*AlarmCloudConfigResponse, error)
}

// alarmCloudService 实现
type alarmCloudService struct {
	alarmCloudRepo repository.AlarmCloudRepository
	logger         *zap.Logger
}

// NewAlarmCloudService 创建 AlarmCloudService 实例
func NewAlarmCloudService(alarmCloudRepo repository.AlarmCloudRepository, logger *zap.Logger) AlarmCloudService {
	return &alarmCloudService{
		alarmCloudRepo: alarmCloudRepo,
		logger:         logger,
	}
}

// GetAlarmCloudConfigRequest 查询告警配置请求
type GetAlarmCloudConfigRequest struct {
	TenantID string
	UserID   string // 当前用户ID（用于权限检查）
	UserRole string // 当前用户角色（用于权限检查）
}

// UpdateAlarmCloudConfigRequest 更新告警配置请求
type UpdateAlarmCloudConfigRequest struct {
	TenantID          string
	UserID            string          // 当前用户ID（用于权限检查）
	UserRole          string          // 当前用户角色（用于权限检查）
	OfflineAlarm      *string         // 可选
	LowBattery        *string         // 可选
	DeviceFailure     *string         // 可选
	DeviceAlarms      json.RawMessage // 可选
	Conditions        json.RawMessage // 可选
	NotificationRules json.RawMessage // 可选
	Metadata          json.RawMessage // 可选
}

// AlarmCloudConfigResponse 告警配置响应
type AlarmCloudConfigResponse struct {
	TenantID          string          `json:"tenant_id"`
	OfflineAlarm      *string         `json:"OfflineAlarm,omitempty"`
	LowBattery        *string         `json:"LowBattery,omitempty"`
	DeviceFailure     *string         `json:"DeviceFailure,omitempty"`
	DeviceAlarms      json.RawMessage `json:"device_alarms"`
	Conditions        json.RawMessage `json:"conditions,omitempty"`
	NotificationRules json.RawMessage `json:"notification_rules,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

// GetAlarmCloudConfig 查询告警配置
func (s *alarmCloudService) GetAlarmCloudConfig(ctx context.Context, req GetAlarmCloudConfigRequest) (*AlarmCloudConfigResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// TODO: 权限检查（需要 role_permissions 表支持）
	// 当前实现：暂时跳过权限检查，后续可以添加

	// 1. 优先查询租户特定配置
	alarmCloud, err := s.alarmCloudRepo.GetAlarmCloud(ctx, req.TenantID)
	if err != nil {
		// 检查是否是 "not found" 错误
		isNotFound := err == sql.ErrNoRows || 
			(err != nil && (fmt.Sprintf("%v", err) == "alarm cloud not found: sql: no rows in result set" || 
				strings.Contains(fmt.Sprintf("%v", err), "alarm cloud not found")))
		
		if isNotFound {
			// 2. 如果租户没有配置，查询系统默认配置
			systemTenantID := "00000000-0000-0000-0000-000000000001"
			systemAlarmCloud, err2 := s.alarmCloudRepo.GetAlarmCloud(ctx, systemTenantID)
			if err2 != nil {
				// 检查是否是 "not found" 错误
				isNotFound2 := err2 == sql.ErrNoRows || 
					(err2 != nil && (fmt.Sprintf("%v", err2) == "alarm cloud not found: sql: no rows in result set" || 
						strings.Contains(fmt.Sprintf("%v", err2), "alarm cloud not found")))
				
				if isNotFound2 {
					// 如果系统默认配置也不存在，返回空配置
					// 注意：返回的 tenant_id 是请求的 tenant_id（与旧 Handler 一致）
					return &AlarmCloudConfigResponse{
						TenantID:     req.TenantID,
						DeviceAlarms: json.RawMessage("{}"),
					}, nil
				}
				return nil, fmt.Errorf("failed to get system alarm cloud: %w", err2)
			}
			// 使用系统默认配置，tenant_id 保持为 SystemTenantID（反映实际来源，与旧 Handler 一致）
			alarmCloud = systemAlarmCloud
		} else {
			return nil, fmt.Errorf("failed to get alarm cloud: %w", err)
		}
	}

	// 3. 转换为响应格式
	resp := &AlarmCloudConfigResponse{
		TenantID:     alarmCloud.TenantID,
		DeviceAlarms: alarmCloud.DeviceAlarms,
	}

	// 处理可选字段
	if alarmCloud.OfflineAlarm != "" {
		resp.OfflineAlarm = &alarmCloud.OfflineAlarm
	}
	if alarmCloud.LowBattery != "" {
		resp.LowBattery = &alarmCloud.LowBattery
	}
	if alarmCloud.DeviceFailure != "" {
		resp.DeviceFailure = &alarmCloud.DeviceFailure
	}
	if len(alarmCloud.Conditions) > 0 {
		resp.Conditions = alarmCloud.Conditions
	}
	if len(alarmCloud.NotificationRules) > 0 {
		resp.NotificationRules = alarmCloud.NotificationRules
	}
	if len(alarmCloud.Metadata) > 0 {
		resp.Metadata = alarmCloud.Metadata
	}

	// 确保 device_alarms 不为空
	if len(resp.DeviceAlarms) == 0 {
		resp.DeviceAlarms = json.RawMessage("{}")
	}

	return resp, nil
}

// UpdateAlarmCloudConfig 更新告警配置
func (s *alarmCloudService) UpdateAlarmCloudConfig(ctx context.Context, req UpdateAlarmCloudConfigRequest) (*AlarmCloudConfigResponse, error) {
	// 参数验证
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// TODO: 权限检查（需要 role_permissions 表支持）
	// 业务规则：只有 SystemAdmin 或 Admin 可以更新告警配置
	// 当前实现：暂时跳过权限检查，后续可以添加

	// 业务规则验证：不能更新系统默认配置
	if req.TenantID == "00000000-0000-0000-0000-000000000001" {
		return nil, fmt.Errorf("cannot update system alarm cloud config")
	}

	// 1. 获取现有配置（如果存在），用于合并未提供的字段
	existingAlarmCloud, err := s.alarmCloudRepo.GetAlarmCloud(ctx, req.TenantID)
	if err != nil {
		// 检查是否是 "not found" 错误
		isNotFound := err == sql.ErrNoRows || 
			(err != nil && (fmt.Sprintf("%v", err) == "alarm cloud not found: sql: no rows in result set" || 
				strings.Contains(fmt.Sprintf("%v", err), "alarm cloud not found")))
		if !isNotFound {
			return nil, fmt.Errorf("failed to get existing alarm cloud: %w", err)
		}
		// 如果是 not found，existingAlarmCloud 为 nil，继续执行
	}

	// 2. 构建更新后的配置（合并现有配置和请求的字段）
	var alarmCloud domain.AlarmCloud
	if existingAlarmCloud != nil {
		// 使用现有配置作为基础
		alarmCloud = *existingAlarmCloud
	} else {
		// 新建配置，尝试使用系统默认配置作为基础
		systemTenantID := "00000000-0000-0000-0000-000000000001"
		systemAlarmCloud, err := s.alarmCloudRepo.GetAlarmCloud(ctx, systemTenantID)
		if err != nil {
			// 检查是否是 "not found" 错误
			isNotFound := err == sql.ErrNoRows || 
				(err != nil && (fmt.Sprintf("%v", err) == "alarm cloud not found: sql: no rows in result set" || 
					strings.Contains(fmt.Sprintf("%v", err), "alarm cloud not found")))
			if isNotFound {
				// 如果系统默认配置不存在，使用空配置
				alarmCloud = domain.AlarmCloud{
					TenantID:     req.TenantID,
					DeviceAlarms: json.RawMessage("{}"),
				}
			} else {
				// 其他错误，返回错误
				return nil, fmt.Errorf("failed to get system alarm cloud: %w", err)
			}
		} else {
			// 使用系统默认配置作为基础
			alarmCloud = *systemAlarmCloud
			alarmCloud.TenantID = req.TenantID // 更新为当前租户ID
		}
	}

	// 3. 更新字段（只更新提供的字段，与旧 Handler 逻辑一致）
	// 注意：旧 Handler 中，如果字段为空字符串，不更新（使用 sql.NullString）
	// 新 Service 中，如果字段为 nil，不更新（使用指针）
	if req.OfflineAlarm != nil {
		alarmCloud.OfflineAlarm = *req.OfflineAlarm
	}
	if req.LowBattery != nil {
		alarmCloud.LowBattery = *req.LowBattery
	}
	if req.DeviceFailure != nil {
		alarmCloud.DeviceFailure = *req.DeviceFailure
	}
	if len(req.DeviceAlarms) > 0 {
		// 验证 JSON 格式
		var test map[string]interface{}
		if err := json.Unmarshal(req.DeviceAlarms, &test); err != nil {
			return nil, fmt.Errorf("invalid device_alarms JSON format: %w", err)
		}
		alarmCloud.DeviceAlarms = req.DeviceAlarms
	} else if existingAlarmCloud == nil {
		// 如果是新建配置且没有提供 device_alarms，使用空对象
		alarmCloud.DeviceAlarms = json.RawMessage("{}")
	}
	// conditions, notification_rules, metadata: 如果提供了就更新，否则保持现有值（或 nil）
	if len(req.Conditions) > 0 {
		// 验证 JSON 格式
		var test interface{}
		if err := json.Unmarshal(req.Conditions, &test); err != nil {
			return nil, fmt.Errorf("invalid conditions JSON format: %w", err)
		}
		alarmCloud.Conditions = req.Conditions
	}
	if len(req.NotificationRules) > 0 {
		// 验证 JSON 格式
		var test interface{}
		if err := json.Unmarshal(req.NotificationRules, &test); err != nil {
			return nil, fmt.Errorf("invalid notification_rules JSON format: %w", err)
		}
		alarmCloud.NotificationRules = req.NotificationRules
	}
	if len(req.Metadata) > 0 {
		// 验证 JSON 格式
		var test interface{}
		if err := json.Unmarshal(req.Metadata, &test); err != nil {
			return nil, fmt.Errorf("invalid metadata JSON format: %w", err)
		}
		alarmCloud.Metadata = req.Metadata
	}

	// 4. 确保 device_alarms 不为空
	if len(alarmCloud.DeviceAlarms) == 0 {
		alarmCloud.DeviceAlarms = json.RawMessage("{}")
	}

	// 5. 调用 Repository 更新（UPSERT 语义）
	err = s.alarmCloudRepo.UpsertAlarmCloud(ctx, req.TenantID, &alarmCloud)
	if err != nil {
		return nil, fmt.Errorf("failed to update alarm cloud: %w", err)
	}

	// 6. 返回更新后的配置
	return s.GetAlarmCloudConfig(ctx, GetAlarmCloudConfigRequest{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		UserRole: req.UserRole,
	})
}

