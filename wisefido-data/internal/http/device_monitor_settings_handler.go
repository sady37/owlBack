package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"owl-common/alarm"
	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// DeviceMonitorSettingsHandler 设备监控配置 Handler
type DeviceMonitorSettingsHandler struct {
	deviceMonitorSettingsService service.DeviceMonitorSettingsService
	logger                       *zap.Logger
	base                         *StubHandler // 用于 tenantIDFromReq
	usersRepo                    repository.UsersRepository
	userBranchesRepo             repository.UserBranchesRepository
	devicesRepo                  repository.DevicesRepository
	unitsRepo                    repository.UnitsRepository
	redisClient                  *redis.Client
	db                           *sql.DB
}

// NewDeviceMonitorSettingsHandler 创建设备监控配置 Handler
func NewDeviceMonitorSettingsHandler(
	deviceMonitorSettingsService service.DeviceMonitorSettingsService,
	usersRepo repository.UsersRepository,
	userBranchesRepo repository.UserBranchesRepository,
	devicesRepo repository.DevicesRepository,
	unitsRepo repository.UnitsRepository,
	redisClient *redis.Client,
	db *sql.DB,
	logger *zap.Logger,
) *DeviceMonitorSettingsHandler {
	return &DeviceMonitorSettingsHandler{
		deviceMonitorSettingsService: deviceMonitorSettingsService,
		logger:                       logger,
		base:                         &StubHandler{}, // 用于 tenantIDFromReq
		usersRepo:                    usersRepo,
		userBranchesRepo:             userBranchesRepo,
		devicesRepo:                  devicesRepo,
		unitsRepo:                    unitsRepo,
		redisClient:                  redisClient,
		db:                           db,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *DeviceMonitorSettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 检查是否是获取默认设置的请求
	if strings.HasPrefix(path, "/settings/api/v1/monitor/default/sleepad") {
		if r.Method == http.MethodGet {
			h.GetDefaultDeviceMonitorSettings(w, r, "sleepad")
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if strings.HasPrefix(path, "/settings/api/v1/monitor/default/radar") {
		if r.Method == http.MethodGet {
			h.GetDefaultDeviceMonitorSettings(w, r, "radar")
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 解析设备类型和设备ID
	var deviceType string
	var deviceID string

	if strings.HasPrefix(path, "/settings/api/v1/monitor/sleepad/") {
		deviceType = "sleepad"
		deviceID = strings.TrimPrefix(path, "/settings/api/v1/monitor/sleepad/")
	} else if strings.HasPrefix(path, "/settings/api/v1/monitor/radar/") {
		deviceType = "radar"
		deviceID = strings.TrimPrefix(path, "/settings/api/v1/monitor/radar/")
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 验证 deviceID 不为空且不包含 "/"，且不能是 "undefined" 或 "null"
	if deviceID == "" || strings.Contains(deviceID, "/") || deviceID == "undefined" || deviceID == "null" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 根据 HTTP 方法分发
	switch r.Method {
	case http.MethodGet:
		h.GetDeviceMonitorSettings(w, r, deviceType, deviceID)
	case http.MethodPut:
		h.UpdateDeviceMonitorSettings(w, r, deviceType, deviceID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GetDeviceMonitorSettings 获取设备监控配置
func (h *DeviceMonitorSettingsHandler) GetDeviceMonitorSettings(w http.ResponseWriter, r *http.Request, deviceType, deviceID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	alarmItems, err := h.deviceMonitorSettingsService.GetDeviceMonitorSettings(ctx, tenantID, deviceID, deviceType)
	if err != nil {
		h.logger.Error("GetDeviceMonitorSettings failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(alarmItems))
}

// UpdateDeviceMonitorSettings 更新设备监控配置
func (h *DeviceMonitorSettingsHandler) UpdateDeviceMonitorSettings(w http.ResponseWriter, r *http.Request, deviceType, deviceID string) {
	ctx := r.Context()

	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 1. 安全验证：从 session 获取真实的 user_id（不能信任 header）
	// TODO: 如果系统有 session 管理，应该从 session/context 获取，而不是 header
	// 当前实现：使用 header，但会进行严格验证
	sessionUserID := r.Header.Get("X-User-Id")
	if sessionUserID == "" {
		writeJSON(w, http.StatusOK, Fail("user_id is required"))
		return
	}

	// 2. 安全验证：验证 tenant_id 与 user_id 一致性
	if err := h.verifyUserTenantConsistency(ctx, sessionUserID, tenantID); err != nil {
		h.logger.Warn("Security check failed: user-tenant mismatch",
			zap.String("user_id", sessionUserID),
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail("unauthorized: user does not belong to this tenant"))
		return
	}

	// 3. 安全验证：验证 user role (Manager/admin)
	if err := h.verifyUserRole(ctx, sessionUserID, tenantID, []string{"Manager", "admin"}); err != nil {
		h.logger.Warn("Security check failed: insufficient role",
			zap.String("user_id", sessionUserID),
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail("unauthorized: insufficient role (requires Manager or admin)"))
		return
	}

	// 4. 安全验证：验证 device_id 与 user_id/tenant_id 一致性
	if err := h.verifyDeviceAccess(ctx, sessionUserID, tenantID, deviceID); err != nil {
		h.logger.Warn("Security check failed: device access denied",
			zap.String("user_id", sessionUserID),
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail("unauthorized: device access denied"))
		return
	}

	// 4.1. 设备验证增强：仅允许设置在线的设备
	if err := h.verifyDeviceOnline(ctx, tenantID, deviceID); err != nil {
		h.logger.Warn("Security check failed: device is offline",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail("device must be online to update settings"))
		return
	}

	// 5. 安全验证：验证 branch_id 与 user_id/device_id 一致性
	if err := h.verifyBranchAccess(ctx, sessionUserID, tenantID, deviceID); err != nil {
		h.logger.Warn("Security check failed: branch access denied",
			zap.String("user_id", sessionUserID),
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail("unauthorized: branch access denied"))
		return
	}

	var payload map[string]interface{}
	if err := readBodyJSON(r, 1<<20, &payload); err != nil {
		writeJSON(w, http.StatusOK, Fail("invalid body"))
		return
	}

	// 从 payload 解析 alarm_items 数组
	alarmItemsRaw, ok := payload["alarm_items"]
	if !ok {
		writeJSON(w, http.StatusOK, Fail("alarm_items is required"))
		return
	}

	// 转换为 JSON 再解析为 []alarm.AlarmItem
	alarmItemsJSON, err := json.Marshal(alarmItemsRaw)
	if err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to marshal alarm_items: "+err.Error()))
		return
	}

	var alarmItems []alarm.AlarmItem
	if err := json.Unmarshal(alarmItemsJSON, &alarmItems); err != nil {
		writeJSON(w, http.StatusOK, Fail("failed to unmarshal alarm_items: "+err.Error()))
		return
	}

	// 6. 验证 AlarmItems 中的 AlarmType 是否属于该设备类型
	if err := h.verifyAlarmTypesForDevice(deviceType, alarmItems); err != nil {
		h.logger.Warn("Validation failed: invalid alarm types for device",
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail("invalid alarm types for device: "+err.Error()))
		return
	}

	// 7. 验证参数值范围（阈值合理性）
	if err := h.validateAlarmItemParams(deviceType, alarmItems); err != nil {
		h.logger.Warn("Validation failed: invalid parameter values",
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail("invalid parameter values: "+err.Error()))
		return
	}

	// 使用已验证的 sessionUserID（不能使用 payload 中的 user_id）
	result, err := h.deviceMonitorSettingsService.UpdateDeviceMonitorSettings(ctx, tenantID, deviceID, deviceType, sessionUserID, alarmItems, nil)
	if err != nil {
		h.logger.Error("UpdateDeviceMonitorSettings failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 根据设备类型处理返回结果
	if deviceType == "radar" {
		// Radar 设备返回 UpdateRadarResult
		if radarResult, ok := result.(*service.UpdateRadarResult); ok {
			// 如果配置无变化
			if radarResult.NoChange {
				writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
					"success":        true,
					"no_change":      true,
					"message":        "no Change",
					"device_write":   false,
					"database_write": false,
				}))
				return
			}

			// 构建失败项列表（转换为前端格式）
			failedItems := make([]map[string]interface{}, len(radarResult.FailedAlarmTypes))
			for i, item := range radarResult.FailedAlarmTypes {
				failedItems[i] = map[string]interface{}{
					"alarm_type":   item.AlarmType,
					"is_enabled":   item.IsEnabled,
					"alarm_level":  item.AlarmLevel,
					"alarm_params": item.AlarmParams,
				}
			}

			writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
				"success":        radarResult.DeviceWriteSuccess && radarResult.DBWriteSuccess,
				"device_write":   radarResult.DeviceWriteSuccess,
				"database_write": radarResult.DBWriteSuccess,
				"failed_items":   failedItems, // 失败的项，vue 需要更新为旧值
			}))
			return
		}
	} else {
		// Sleepace 设备返回 map
		if sleepaceResult, ok := result.(map[string]interface{}); ok {
			writeJSON(w, http.StatusOK, Ok(sleepaceResult))
			return
		}
	}

	// 未知的返回类型
	writeJSON(w, http.StatusOK, Fail("unknown result type"))
}

// GetDefaultDeviceMonitorSettings 获取默认设备监控配置
// 阈值：硬编码（与 System 租户模板设备的值相同）
// Alarm Level：优先从当前租户的 alarm_cloud 读取，如果没有则使用硬编码值
func (h *DeviceMonitorSettingsHandler) GetDefaultDeviceMonitorSettings(w http.ResponseWriter, r *http.Request, deviceType string) {
	ctx := r.Context()

	// 从请求中获取 tenantID
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	alarmItems, err := h.deviceMonitorSettingsService.GetDefaultDeviceMonitorSettings(ctx, tenantID, deviceType)
	if err != nil {
		h.logger.Error("GetDefaultDeviceMonitorSettings failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_type", deviceType),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(alarmItems))
}

// ============================================
// 安全验证函数
// ============================================

// verifyUserTenantConsistency 验证 tenant_id 与 user_id 一致性
// 1. tenant_id: user_id 一致
func (h *DeviceMonitorSettingsHandler) verifyUserTenantConsistency(ctx context.Context, userID, tenantID string) error {
	if userID == "" || tenantID == "" {
		return fmt.Errorf("user_id and tenant_id are required")
	}

	user, err := h.usersRepo.GetUser(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("user not found or access denied: %w", err)
	}

	if user.TenantID != tenantID {
		return fmt.Errorf("user does not belong to tenant: user.tenant_id=%s, expected=%s", user.TenantID, tenantID)
	}

	return nil
}

// verifyUserRole 验证 user role (Manager/admin)
// 3. user_id Role: Manager, admin
func (h *DeviceMonitorSettingsHandler) verifyUserRole(ctx context.Context, userID, tenantID string, allowedRoles []string) error {
	if userID == "" || tenantID == "" {
		return fmt.Errorf("user_id and tenant_id are required")
	}

	user, err := h.usersRepo.GetUser(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 检查 role 是否在允许列表中（不区分大小写）
	userRoleLower := strings.ToLower(user.Role)
	for _, allowedRole := range allowedRoles {
		if strings.ToLower(allowedRole) == userRoleLower {
			return nil
		}
	}

	// SystemAdmin 和 SystemOperator 也允许
	if strings.EqualFold(user.Role, "SystemAdmin") || strings.EqualFold(user.Role, "SystemOperator") {
		return nil
	}

	return fmt.Errorf("insufficient role: user.role=%s, allowed=%v", user.Role, allowedRoles)
}

// verifyDeviceAccess 验证 device_id 与 user_id/tenant_id 一致性
// 2. branch_id: user_id device_id 一致（通过设备所属的 branch_id 验证）
func (h *DeviceMonitorSettingsHandler) verifyDeviceAccess(ctx context.Context, userID, tenantID, deviceID string) error {
	if userID == "" || tenantID == "" || deviceID == "" {
		return fmt.Errorf("user_id, tenant_id and device_id are required")
	}

	// 验证设备存在且属于该租户
	device, err := h.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("device not found or access denied: %w", err)
	}

	if device.TenantID != tenantID {
		return fmt.Errorf("device does not belong to tenant: device.tenant_id=%s, expected=%s", device.TenantID, tenantID)
	}

	return nil
}

// verifyBranchAccess 验证 branch_id 与 user_id/device_id 一致性
// 2. branch_id: user_id device_id 一致
func (h *DeviceMonitorSettingsHandler) verifyBranchAccess(ctx context.Context, userID, tenantID, deviceID string) error {
	if userID == "" || tenantID == "" || deviceID == "" {
		return fmt.Errorf("user_id, tenant_id and device_id are required")
	}

	// 获取设备信息
	device, err := h.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	// 获取设备的 branch_id（通过 unit_id）
	deviceBranchID := ""
	if device.UnitID.Valid && device.UnitID.String != "" {
		unit, err := h.unitsRepo.GetUnit(ctx, tenantID, device.UnitID.String)
		if err == nil && unit.BranchID.Valid {
			deviceBranchID = unit.BranchID.String
		}
	}

	// 如果设备没有关联 branch_id，跳过验证（允许访问）
	if deviceBranchID == "" {
		return nil
	}

	// 获取用户的所有 branch_id 列表
	userBranches, err := h.userBranchesRepo.GetUserBranches(ctx, tenantID, userID)
	if err != nil {
		// 如果查询失败，记录警告但允许访问（向后兼容）
		h.logger.Warn("Failed to get user branches, allowing access",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return nil
	}

	// 检查用户的 branch_id 列表中是否包含设备的 branch_id
	for _, userBranch := range userBranches {
		if userBranch.BranchID == deviceBranchID {
			return nil
		}
	}

	// 如果用户没有关联任何 branch，也允许访问（向后兼容）
	if len(userBranches) == 0 {
		h.logger.Warn("User has no branch associations, allowing access",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.String("device_branch_id", deviceBranchID),
		)
		return nil
	}

	return fmt.Errorf("user does not have access to device branch: user_branches=%v, device_branch_id=%s",
		getBranchIDs(userBranches), deviceBranchID)
}

// getBranchIDs 从 UserBranch 列表提取 branch_id 列表
func getBranchIDs(branches []*domain.UserBranch) []string {
	ids := make([]string, len(branches))
	for i, b := range branches {
		ids[i] = b.BranchID
	}
	return ids
}

// verifyDeviceOnline 验证设备是否在线（通过 wisefido-qinglan HTTP API 实时查询）
// 仅允许设置在线的设备
func (h *DeviceMonitorSettingsHandler) verifyDeviceOnline(ctx context.Context, tenantID, deviceID string) error {
	// 获取设备信息
	device, err := h.devicesRepo.GetDevice(ctx, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	if device.DeviceUID == "" {
		// 如果设备没有 device_uid，跳过在线状态检查
		h.logger.Warn("Device has no device_uid, skipping online status check",
			zap.String("device_id", deviceID),
		)
		return nil
	}

	// 通过 wisefido-qinglan HTTP API 实时查询设备在线状态
	if err := h.deviceMonitorSettingsService.CheckDeviceOnlineStatus(ctx, device.DeviceUID); err != nil {
		h.logger.Warn("Device is not online",
			zap.String("device_id", deviceID),
			zap.String("device_uid", device.DeviceUID),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// verifyAlarmTypesForDevice 验证 AlarmItems 中的 AlarmType 是否属于该设备类型
func (h *DeviceMonitorSettingsHandler) verifyAlarmTypesForDevice(deviceType string, alarmItems []alarm.AlarmItem) error {
	normalizedType := strings.ToLower(deviceType)

	// 获取该设备类型支持的 AlarmType 列表
	var allowedTypes []string
	switch normalizedType {
	case "radar":
		allowedTypes = alarm.GetRadarAlarmTypes()
	case "sleepad":
		allowedTypes = alarm.GetSleepPadAlarmTypes()
	default:
		return fmt.Errorf("unsupported device type: %s (normalized: %s)", deviceType, normalizedType)
	}

	// 构建允许的 AlarmType 集合（用于快速查找）
	allowedSet := make(map[string]bool)
	for _, t := range allowedTypes {
		allowedSet[t] = true
	}

	// 验证每个 AlarmItem 的 AlarmType
	for _, item := range alarmItems {
		if item.AlarmType == "" {
			continue // 跳过空的 AlarmType
		}
		if !allowedSet[item.AlarmType] {
			return fmt.Errorf("alarm type '%s' is not supported for device type '%s'", item.AlarmType, deviceType)
		}
	}

	return nil
}

// validateAlarmItemParams 验证参数值范围（阈值合理性）
// 注意：此函数仅验证参数范围，不设置默认值。默认值由 AlarmItem 设置。
// 雷达：呼吸心率参数范围验证
// Radar_AbnormalRespiratoryRate:
//   - max (upper breath): [10, 100]
//   - min (lower breath): [1, 20]
//   - 验证 min < max
// Radar_AbnormalHeartRate:
//   - max (upper heart): [10, 255]
//   - min (lower heart): [1, 100]
//   - 验证 min < max
func (h *DeviceMonitorSettingsHandler) validateAlarmItemParams(deviceType string, alarmItems []alarm.AlarmItem) error {
	if deviceType != "radar" {
		// 目前只验证雷达设备的参数
		return nil
	}

	// 构建 AlarmItem 索引
	itemMap := make(map[string]alarm.AlarmItem)
	for _, item := range alarmItems {
		if item.AlarmType != "" {
			itemMap[item.AlarmType] = item
		}
	}

	// 验证雷达呼吸心率参数
	// Radar_AbnormalRespiratoryRate: max (upper breath), min (lower breath)
	if item, ok := itemMap["Radar_AbnormalRespiratoryRate"]; ok && item.AlarmParams != nil {
		if max, ok := item.AlarmParams["max"]; ok {
			upperBreath := getIntValue(max)
			if upperBreath < 10 || upperBreath > 100 {
				return fmt.Errorf("Radar_AbnormalRespiratoryRate.max must be in range [10, 100], got %d", upperBreath)
			}
		}
		if min, ok := item.AlarmParams["min"]; ok {
			lowerBreath := getIntValue(min)
			if lowerBreath < 1 || lowerBreath > 20 {
				return fmt.Errorf("Radar_AbnormalRespiratoryRate.min must be in range [1, 20], got %d", lowerBreath)
			}
		}
		// 验证 min < max
		if max, ok := item.AlarmParams["max"]; ok {
			if min, ok := item.AlarmParams["min"]; ok {
				upperBreath := getIntValue(max)
				lowerBreath := getIntValue(min)
				if lowerBreath >= upperBreath {
					return fmt.Errorf("Radar_AbnormalRespiratoryRate: min (%d) must be less than max (%d)", lowerBreath, upperBreath)
				}
			}
		}
	}

	// Radar_AbnormalHeartRate: max (upper heart), min (lower heart)
	if item, ok := itemMap["Radar_AbnormalHeartRate"]; ok && item.AlarmParams != nil {
		if max, ok := item.AlarmParams["max"]; ok {
			upperHeart := getIntValue(max)
			if upperHeart < 10 || upperHeart > 255 {
				return fmt.Errorf("Radar_AbnormalHeartRate.max must be in range [10, 255], got %d", upperHeart)
			}
		}
		if min, ok := item.AlarmParams["min"]; ok {
			lowerHeart := getIntValue(min)
			if lowerHeart < 1 || lowerHeart > 100 {
				return fmt.Errorf("Radar_AbnormalHeartRate.min must be in range [1, 100], got %d", lowerHeart)
			}
		}
		// 验证 min < max
		if max, ok := item.AlarmParams["max"]; ok {
			if min, ok := item.AlarmParams["min"]; ok {
				upperHeart := getIntValue(max)
				lowerHeart := getIntValue(min)
				if lowerHeart >= upperHeart {
					return fmt.Errorf("Radar_AbnormalHeartRate: min (%d) must be less than max (%d)", lowerHeart, upperHeart)
				}
			}
		}
	}

	return nil
}

// getIntValue 从 interface{} 获取 int 值
func getIntValue(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		// 尝试解析字符串
		var result int
		if _, err := fmt.Sscanf(x, "%d", &result); err == nil {
			return result
		}
	}
	return 0
}

