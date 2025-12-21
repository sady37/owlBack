package repository

import (
	"context"

	"wisefido-data/internal/domain"
)

// AlarmCloudRepository 云端告警策略Repository接口
type AlarmCloudRepository interface {
	// GetAlarmCloud 获取租户的告警策略配置
	GetAlarmCloud(ctx context.Context, tenantID string) (*domain.AlarmCloud, error)

	// UpsertAlarmCloud 创建或更新租户的告警策略配置
	// 注意：UNIQUE(tenant_id)，使用UPSERT语义
	UpsertAlarmCloud(ctx context.Context, tenantID string, alarmCloud *domain.AlarmCloud) error

	// GetSystemAlarmCloud 获取系统默认告警策略模板
	// tenant_id = SystemTenantID (00000000-0000-0000-0000-000000000001)
	GetSystemAlarmCloud(ctx context.Context) (*domain.AlarmCloud, error)

	// DeleteAlarmCloud 删除租户的告警策略配置（回退到系统默认）
	DeleteAlarmCloud(ctx context.Context, tenantID string) error
}

