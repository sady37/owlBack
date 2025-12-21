package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// SleepaceReportHandler Sleepace 睡眠报告 Handler
type SleepaceReportHandler struct {
	sleepaceReportService service.SleepaceReportService
	base                  *StubHandler // 用于 tenantIDFromReq
	db                    *sql.DB      // 用于查询设备信息
	logger                *zap.Logger
}

// NewSleepaceReportHandler 创建 SleepaceReportHandler
func NewSleepaceReportHandler(sleepaceReportService service.SleepaceReportService, db interface{}, logger *zap.Logger) *SleepaceReportHandler {
	var dbConn *sql.DB
	if db != nil {
		if d, ok := db.(*sql.DB); ok {
			dbConn = d
		}
	}
	return &SleepaceReportHandler{
		sleepaceReportService: sleepaceReportService,
		base:                  &StubHandler{},
		db:                    dbConn,
		logger:                logger,
	}
}

// ServeHTTP 处理 HTTP 请求
func (h *SleepaceReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路由：/sleepace/api/v1/sleepace/reports/:id
	// 支持：
	//   - GET /sleepace/api/v1/sleepace/reports/:id - 获取报告列表
	//   - GET /sleepace/api/v1/sleepace/reports/:id/detail - 获取报告详情
	//   - GET /sleepace/api/v1/sleepace/reports/:id/dates - 获取有效日期列表

	path := r.URL.Path
	deviceID := extractDeviceIDFromPath(path)

	if deviceID == "" {
		writeJSON(w, http.StatusOK, Fail("device_id is required"))
		return
	}

	// 根据路径后缀路由到不同的处理函数
	if strings.HasSuffix(path, "/download") {
		if r.Method == http.MethodPost {
			h.DownloadReport(w, r, deviceID)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else if strings.HasSuffix(path, "/detail") {
		h.GetSleepaceReportDetail(w, r, deviceID)
	} else if strings.HasSuffix(path, "/dates") {
		h.GetSleepaceReportDates(w, r, deviceID)
	} else {
		h.GetSleepaceReports(w, r, deviceID)
	}
}

// GetSleepaceReports 获取睡眠报告列表
// GET /sleepace/api/v1/sleepace/reports/:id?startDate=20240820&endDate=20240830&page=1&size=10
func (h *SleepaceReportHandler) GetSleepaceReports(w http.ResponseWriter, r *http.Request, deviceID string) {
	ctx := r.Context()
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 权限检查
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")
	if err := h.checkReportPermission(ctx, tenantID, deviceID, currentUserID, currentUserType, currentUserRole, "read"); err != nil {
		h.logger.Warn("GetSleepaceReports permission denied",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("user_id", currentUserID),
			zap.String("user_type", currentUserType),
			zap.String("user_role", currentUserRole),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 解析查询参数
	startDate, _ := parseIntQuery(r, "startDate", 0)
	endDate, _ := parseIntQuery(r, "endDate", 0)
	page, _ := parseIntQuery(r, "page", 1)
	size, _ := parseIntQuery(r, "size", 10)

	req := service.GetSleepaceReportsRequest{
		TenantID:  tenantID,
		DeviceID:  deviceID,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		PageSize:  size,
	}

	resp, err := h.sleepaceReportService.GetSleepaceReports(ctx, req)
	if err != nil {
		h.logger.Error("GetSleepaceReports failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换为前端格式（兼容 v1.0）
	result := map[string]any{
		"items": resp.Items,
		"pagination": map[string]any{
			"size":      resp.Size,
			"page":      resp.Page,
			"count":     resp.Total,
			"total":     resp.Total,
			"sort":      "",
			"direction": 0,
		},
	}

	writeJSON(w, http.StatusOK, Ok(result))
}

// GetSleepaceReportDetail 获取睡眠报告详情
// GET /sleepace/api/v1/sleepace/reports/:id/detail?date=20240820
func (h *SleepaceReportHandler) GetSleepaceReportDetail(w http.ResponseWriter, r *http.Request, deviceID string) {
	ctx := r.Context()
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 权限检查
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")
	if err := h.checkReportPermission(ctx, tenantID, deviceID, currentUserID, currentUserType, currentUserRole, "read"); err != nil {
		h.logger.Warn("GetSleepaceReportDetail permission denied",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("user_id", currentUserID),
			zap.String("user_type", currentUserType),
			zap.String("user_role", currentUserRole),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 解析查询参数
	date, err := parseIntQuery(r, "date", 0)
	if err != nil || date == 0 {
		writeJSON(w, http.StatusOK, Fail("date parameter is required (YYYYMMDD format)"))
		return
	}

	req := service.GetSleepaceReportDetailRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
		Date:     date,
	}

	resp, err := h.sleepaceReportService.GetSleepaceReportDetail(ctx, req)
	if err != nil {
		h.logger.Error("GetSleepaceReportDetail failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Int("date", date),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换为前端格式（兼容 v1.0）
	result := map[string]any{
		"id":          resp.ID,
		"deviceId":    resp.DeviceID,
		"deviceCode":  resp.DeviceCode,
		"recordCount": resp.RecordCount,
		"startTime":   resp.StartTime,
		"endTime":     resp.EndTime,
		"date":        resp.Date,
		"stopMode":    resp.StopMode,
		"timeStep":    resp.TimeStep,
		"timezone":    resp.Timezone,
		"report":      resp.Report,
	}

	writeJSON(w, http.StatusOK, Ok(result))
}

// GetSleepaceReportDates 获取有效日期列表
// GET /sleepace/api/v1/sleepace/reports/:id/dates
func (h *SleepaceReportHandler) GetSleepaceReportDates(w http.ResponseWriter, r *http.Request, deviceID string) {
	ctx := r.Context()
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 权限检查
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")
	if err := h.checkReportPermission(ctx, tenantID, deviceID, currentUserID, currentUserType, currentUserRole, "read"); err != nil {
		h.logger.Warn("GetSleepaceReportDates permission denied",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("user_id", currentUserID),
			zap.String("user_type", currentUserType),
			zap.String("user_role", currentUserRole),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	req := service.GetSleepaceReportDatesRequest{
		TenantID: tenantID,
		DeviceID: deviceID,
	}

	resp, err := h.sleepaceReportService.GetSleepaceReportDates(ctx, req)
	if err != nil {
		h.logger.Error("GetSleepaceReportDates failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 直接返回日期数组（兼容 v1.0）
	writeJSON(w, http.StatusOK, Ok(resp.Dates))
}

// ============================================
// 辅助方法
// ============================================

// DownloadReport 手动触发下载报告
// POST /sleepace/api/v1/sleepace/reports/:id/download?startTime=1234567890&endTime=1234567890
func (h *SleepaceReportHandler) DownloadReport(w http.ResponseWriter, r *http.Request, deviceID string) {
	ctx := r.Context()
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 权限检查（下载需要 manage 权限）
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")
	if err := h.checkReportPermission(ctx, tenantID, deviceID, currentUserID, currentUserType, currentUserRole, "manage"); err != nil {
		h.logger.Warn("DownloadReport permission denied",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("user_id", currentUserID),
			zap.String("user_type", currentUserType),
			zap.String("user_role", currentUserRole),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 解析查询参数
	startTime, err := parseInt64Query(r, "startTime", 0)
	if err != nil || startTime == 0 {
		writeJSON(w, http.StatusOK, Fail("startTime parameter is required (Unix timestamp in seconds)"))
		return
	}

	endTime, err := parseInt64Query(r, "endTime", 0)
	if err != nil || endTime == 0 {
		writeJSON(w, http.StatusOK, Fail("endTime parameter is required (Unix timestamp in seconds)"))
		return
	}

	// 获取设备信息（需要 device_code）
	// 通过 device_id 查询 devices 表获取 device_code（serial_number 或 uid）
	deviceCode, err := h.getDeviceCode(ctx, tenantID, deviceID)
	if err != nil {
		h.logger.Error("Failed to get device code",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to get device code: %v", err)))
		return
	}

	req := service.DownloadReportRequest{
		TenantID:   tenantID,
		DeviceID:   deviceID,
		DeviceCode: deviceCode,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	err = h.sleepaceReportService.DownloadReport(ctx, req)
	if err != nil {
		h.logger.Error("DownloadReport failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_id", deviceID),
			zap.String("device_code", deviceCode),
			zap.Int64("start_time", startTime),
			zap.Int64("end_time", endTime),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// getDeviceCode 通过 device_id 获取 device_code（serial_number 或 uid）
func (h *SleepaceReportHandler) getDeviceCode(ctx context.Context, tenantID, deviceID string) (string, error) {
	if h.db == nil {
		return "", fmt.Errorf("database connection not available")
	}

	query := `
		SELECT COALESCE(serial_number, uid, '') as device_code
		FROM devices
		WHERE tenant_id = $1::uuid
		  AND device_id = $2::uuid
		  AND status <> 'disabled'
		LIMIT 1
	`
	var deviceCode string
	err := h.db.QueryRowContext(ctx, query, tenantID, deviceID).Scan(&deviceCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("device not found")
		}
		return "", fmt.Errorf("failed to get device code: %w", err)
	}
	if deviceCode == "" {
		return "", fmt.Errorf("device code not found (serial_number and uid are both empty)")
	}
	return deviceCode, nil
}

// extractDeviceIDFromPath 从路径中提取 device_id
// 路径格式：/sleepace/api/v1/sleepace/reports/:id 或 /sleepace/api/v1/sleepace/reports/:id/detail
func extractDeviceIDFromPath(path string) string {
	// 移除前缀
	prefix := "/sleepace/api/v1/sleepace/reports/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	// 提取 device_id（移除后缀如 /detail, /dates, /download）
	deviceID := strings.TrimPrefix(path, prefix)
	deviceID = strings.TrimSuffix(deviceID, "/detail")
	deviceID = strings.TrimSuffix(deviceID, "/dates")
	deviceID = strings.TrimSuffix(deviceID, "/download")
	deviceID = strings.TrimSuffix(deviceID, "/")

	return deviceID
}

// parseIntQuery 解析整数查询参数
func parseIntQuery(r *http.Request, key string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue, err
	}
	return parsed, nil
}

// parseInt64Query 解析 int64 查询参数
func parseInt64Query(r *http.Request, key string, defaultValue int64) (int64, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue, err
	}
	return parsed, nil
}

// ============================================
// 权限检查
// ============================================

// checkReportPermission 检查睡眠报告权限
// 权限规则：
// 1. 住户及相关联系人：可查看住户自己的睡眠报告
// 2. Caregiver/Nurse：可查看、处理 assign-only 住户的睡眠报告
// 3. Manager：可查看、处理 branch 住户的睡眠报告，如果 branch=null，处理 branch=null 的 unit 的住户
func (h *SleepaceReportHandler) checkReportPermission(ctx context.Context, tenantID, deviceID, userID, userType, userRole, permissionType string) error {
	if h.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// 1. 通过 device_id 获取关联的住户信息
	residentInfo, err := h.getResidentByDeviceID(ctx, tenantID, deviceID)
	if err != nil {
		// 如果设备没有关联住户，允许访问（fallback）
		return nil
	}

	// 2. 住户及相关联系人：只能查看自己的
	if userType == "resident" || userType == "family" {
		if residentInfo.ResidentID != userID {
			return fmt.Errorf("access denied: can only view own reports")
		}
		return nil
	}

	// 3. Staff 角色权限检查
	if userType == "staff" && userRole != "" {
		// 3.1 Caregiver/Nurse：检查 assign-only
		if userRole == "Caregiver" || userRole == "Nurse" {
			// 检查权限配置
			perm, err := GetResourcePermission(h.db, ctx, userRole, "residents", "R")
			if err == nil && perm.AssignedOnly {
				// 检查是否分配给该用户
				if !h.isResidentAssignedToUser(ctx, tenantID, residentInfo.ResidentID, userID) {
					return fmt.Errorf("access denied: resident not assigned to you")
				}
			}
			return nil
		}

		// 3.2 Manager：检查 branch-only
		if userRole == "Manager" {
			// 检查权限配置
			perm, err := GetResourcePermission(h.db, ctx, userRole, "residents", "R")
			if err == nil && perm.BranchOnly {
				// 获取用户的 branch_tag
				var userBranchTag sql.NullString
				err := h.db.QueryRowContext(ctx,
					`SELECT branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
					tenantID, userID,
				).Scan(&userBranchTag)
				if err == nil {
					// 检查住户的 branch_tag
					if !userBranchTag.Valid || userBranchTag.String == "" {
						// 用户 branch_tag 为 NULL：只能访问 branch_tag 为 NULL 或 '-' 的住户
						if residentInfo.BranchTag.Valid && residentInfo.BranchTag.String != "" && residentInfo.BranchTag.String != "-" {
							return fmt.Errorf("access denied: can only access residents with branch_tag NULL or '-'")
						}
					} else {
						// 用户 branch_tag 有值：只能访问匹配的 branch
						if !residentInfo.BranchTag.Valid || residentInfo.BranchTag.String != userBranchTag.String {
							return fmt.Errorf("access denied: resident belongs to different branch")
						}
					}
				}
			}
			return nil
		}
	}

	// 4. 其他角色：默认允许（SystemAdmin 等）
	return nil
}

// residentInfo 住户信息（用于权限检查）
type residentInfo struct {
	ResidentID string
	BranchTag  sql.NullString
	UnitID     sql.NullString
}

// getResidentByDeviceID 通过 device_id 获取关联的住户信息
// 查询路径：devices → beds → residents 或 devices → rooms → units → residents
func (h *SleepaceReportHandler) getResidentByDeviceID(ctx context.Context, tenantID, deviceID string) (*residentInfo, error) {
	// 查询设备关联的住户（优先通过 bed，其次通过 room）
	query := `
		SELECT DISTINCT
			r.resident_id::text,
			u.branch_tag,
			u.unit_id::text
		FROM devices d
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id
		LEFT JOIN rooms rm ON (d.bound_room_id = rm.room_id OR b.room_id = rm.room_id)
		LEFT JOIN units u ON rm.unit_id = u.unit_id
		LEFT JOIN residents r ON (r.bed_id = b.bed_id OR r.room_id = rm.room_id OR r.unit_id = u.unit_id)
		WHERE d.tenant_id = $1::uuid
		  AND d.device_id = $2::uuid
		  AND r.resident_id IS NOT NULL
		LIMIT 1
	`

	var info residentInfo
	err := h.db.QueryRowContext(ctx, query, tenantID, deviceID).Scan(
		&info.ResidentID,
		&info.BranchTag,
		&info.UnitID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no resident found for device")
		}
		return nil, fmt.Errorf("failed to get resident by device_id: %w", err)
	}

	return &info, nil
}

// isResidentAssignedToUser 检查住户是否分配给该用户
// resident_caregivers 表通过 userList (JSONB) 存储用户ID列表
func (h *SleepaceReportHandler) isResidentAssignedToUser(ctx context.Context, tenantID, residentID, userID string) bool {
	// 查询 resident_caregivers 表的 userList 字段（JSONB 数组）
	query := `
		SELECT userList
		FROM resident_caregivers
		WHERE tenant_id = $1::uuid
		  AND resident_id = $2::uuid
		LIMIT 1
	`
	var userListJSON []byte
	err := h.db.QueryRowContext(ctx, query, tenantID, residentID).Scan(&userListJSON)
	if err != nil {
		// 如果查询失败或记录不存在，返回 false
		return false
	}

	// 解析 JSONB 数组
	var userList []string
	if err := json.Unmarshal(userListJSON, &userList); err != nil {
		// 如果解析失败，返回 false
		return false
	}

	// 检查 userID 是否在列表中
	for _, id := range userList {
		if id == userID {
			return true
		}
	}

	return false
}

