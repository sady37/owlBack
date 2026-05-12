package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// AlarmResource 标识 alarm 类配置的两个层级。
type AlarmResource string

const (
	AlarmResourceCloud  AlarmResource = "alarm_cloud"  // tenant 层（AlarmCloud.vue）
	AlarmResourceDevice AlarmResource = "alarm_device" // device 层（radar/sleepace-monitor-settings.vue）
)

// AlarmAction 标识 read / write 两种操作。
//
// 这里把 v1 的 C/R/U/D 简化为 R/W：alarm 这两个资源没有"独立 create/delete"语义（spatial_config
// 行 upsert 一次性表达，没有真删；删 = 改成默认）。
type AlarmAction string

const (
	AlarmActionRead  AlarmAction = "read"
	AlarmActionWrite AlarmAction = "write"
)

// IsAlarmAccessAllowed 按用户产品规则判断是否允许 alarm 配置访问：
//
//   alarm_cloud (tenant level)
//     READ:  所有角色（B2B + B2C 都允许）
//     WRITE: 仅 tenant_admin / platform_admin
//
//   alarm_device (per-device)
//     B2B:  family 完全禁；tenant_admin / manager / nurse 可读+可写；其它角色（caregiver/viewer 等）禁
//     B2C:  所有角色都可读+可写
//
// 角色名约定：传入 v1 PascalCase（"Admin" / "Manager" / "Family" ...），内部 normalize 到 v2 snake_case。
// tenantID 是 v2 INET CIDR（如 'fd00:0:3::/48'）。
func IsAlarmAccessAllowed(ctx context.Context, db *sql.DB,
	tenantID, role string, resource AlarmResource, action AlarmAction) (bool, error) {

	v2Role := normalizeAlarmRole(role)

	// platform_admin 是全平台超级角色，绕过所有矩阵
	if v2Role == "platform_admin" {
		return true, nil
	}

	switch resource {
	case AlarmResourceCloud:
		if action == AlarmActionRead {
			return true, nil
		}
		// WRITE: 只允许 tenant_admin
		return v2Role == "tenant_admin", nil

	case AlarmResourceDevice:
		kind, err := lookupTenantKind(ctx, db, tenantID)
		if err != nil {
			// 查不到 tenant kind 时保守拒绝（避免误授权）
			return false, fmt.Errorf("failed to lookup tenant kind: %w", err)
		}
		if kind == "B2C" {
			// B2C 家庭场景：所有人都可读+可写（家属各自管自家）
			return true, nil
		}
		// B2B（机构场景）
		if v2Role == "family" {
			return false, nil
		}
		return v2Role == "tenant_admin" || v2Role == "manager" || v2Role == "nurse", nil
	}

	return false, fmt.Errorf("unknown alarm resource: %s", resource)
}

// lookupTenantKind 读 tenants.kind ('B2B' / 'B2C')。返回值 uppercase。
func lookupTenantKind(ctx context.Context, db *sql.DB, tenantID string) (string, error) {
	if db == nil {
		return "B2B", nil // dev fallback
	}
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}
	var kind string
	err := db.QueryRowContext(ctx,
		`SELECT kind FROM tenants WHERE tenant_id = $1::inet`,
		tenantID,
	).Scan(&kind)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(kind), nil
}

// normalizeAlarmRole 把 v1 PascalCase / 大小写不一致的 role string 映射到 v2 snake_case。
// 与 service.mapRoleToV2 一致（不直接调以避免跨文件耦合）。
func normalizeAlarmRole(role string) string {
	switch role {
	case "SystemAdmin", "SystemOperator", "PlatformAdmin", "platform_admin":
		return "platform_admin"
	case "Admin", "TenantAdmin", "tenant_admin":
		return "tenant_admin"
	case "Manager", "manager":
		return "manager"
	case "Nurse", "nurse":
		return "nurse"
	case "Caregiver", "caregiver":
		return "caregiver"
	case "Family", "family":
		return "family"
	case "Viewer", "viewer":
		return "viewer"
	}
	return strings.ToLower(strings.TrimSpace(role))
}
