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
	deviceAddr := extractDeviceAddrFromPath(path)

	if deviceAddr == "" {
		writeJSON(w, http.StatusOK, Fail("device_addr is required"))
		return
	}

	// 根据路径后缀路由到不同的处理函数
	if strings.HasSuffix(path, "/download") {
		if r.Method == http.MethodPost {
			h.DownloadReport(w, r, deviceAddr)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else if strings.HasSuffix(path, "/detail") {
		h.GetSleepaceReportDetail(w, r, deviceAddr)
	} else if strings.HasSuffix(path, "/dates") {
		h.GetSleepaceReportDates(w, r, deviceAddr)
	} else {
		h.GetSleepaceReports(w, r, deviceAddr)
	}
}

// GetSleepaceReports 获取睡眠报告列表
// GET /sleepace/api/v1/sleepace/reports/:id?startDate=20240820&endDate=20240830&page=1&size=10
func (h *SleepaceReportHandler) GetSleepaceReports(w http.ResponseWriter, r *http.Request, deviceAddr string) {
	ctx := r.Context()
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 权限检查
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")
	if err := h.checkReportPermission(ctx, tenantID, deviceAddr, currentUserID, currentUserType, currentUserRole, "read"); err != nil {
		h.logger.Warn("GetSleepaceReports permission denied",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", deviceAddr),
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
		DeviceAddr:  deviceAddr,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		PageSize:  size,
	}

	resp, err := h.sleepaceReportService.GetSleepaceReports(ctx, req)
	if err != nil {
		h.logger.Error("GetSleepaceReports failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", deviceAddr),
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
func (h *SleepaceReportHandler) GetSleepaceReportDetail(w http.ResponseWriter, r *http.Request, deviceAddr string) {
	ctx := r.Context()
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 权限检查
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")
	if err := h.checkReportPermission(ctx, tenantID, deviceAddr, currentUserID, currentUserType, currentUserRole, "read"); err != nil {
		h.logger.Warn("GetSleepaceReportDetail permission denied",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", deviceAddr),
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
		DeviceAddr: deviceAddr,
		Date:     date,
	}

	resp, err := h.sleepaceReportService.GetSleepaceReportDetail(ctx, req)
	if err != nil {
		h.logger.Error("GetSleepaceReportDetail failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", deviceAddr),
			zap.Int("date", date),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	// 转换为前端格式（兼容 v1.0）
	result := map[string]any{
		"id":                       resp.ID,
		"deviceAddr":               resp.DeviceAddr,
		"deviceUid":                resp.DeviceUID,
		"recordCount":              resp.RecordCount,
		"startTime":                resp.StartTime,
		"endTime":                  resp.EndTime,
		"date":                     resp.Date,
		"stopMode":                 resp.StopMode,
		"timeStep":                 resp.TimeStep,
		"timezone":                 resp.Timezone,
		"report":                   resp.Report,
		"weeklySleepEfficiency": resp.WeeklySleepEfficiency,
	}

	writeJSON(w, http.StatusOK, Ok(result))
}

// GetSleepaceReportDates 获取有效日期列表
// GET /sleepace/api/v1/sleepace/reports/:id/dates
func (h *SleepaceReportHandler) GetSleepaceReportDates(w http.ResponseWriter, r *http.Request, deviceAddr string) {
	ctx := r.Context()
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 权限检查
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")
	if err := h.checkReportPermission(ctx, tenantID, deviceAddr, currentUserID, currentUserType, currentUserRole, "read"); err != nil {
		h.logger.Warn("GetSleepaceReportDates permission denied",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", deviceAddr),
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
		DeviceAddr: deviceAddr,
	}

	resp, err := h.sleepaceReportService.GetSleepaceReportDates(ctx, req)
	if err != nil {
		h.logger.Error("GetSleepaceReportDates failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", deviceAddr),
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

// DownloadReport 从厂家 get24HourDailyWithMaxReport 拉取并写入 sleepace_report（读权限即可）
// POST /sleepace/api/v1/sleepace/reports/:id/download?startTime=1234567890&endTime=1234567890
func (h *SleepaceReportHandler) DownloadReport(w http.ResponseWriter, r *http.Request, deviceAddr string) {
	ctx := r.Context()
	tenantID, ok := h.base.tenantIDFromReq(w, r)
	if !ok {
		return
	}

	// 权限检查（下载需要 manage 权限）
	currentUserID := r.Header.Get("X-User-Id")
	currentUserType := r.Header.Get("X-User-Type")
	currentUserRole := r.Header.Get("X-User-Role")
	// 与列表/详情一致：能看报告即可触发从厂家拉取并入库（get24HourDailyWithMaxReport）
	if err := h.checkReportPermission(ctx, tenantID, deviceAddr, currentUserID, currentUserType, currentUserRole, "read"); err != nil {
		h.logger.Warn("DownloadReport permission denied",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", deviceAddr),
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

	sleepaceDeviceID, deviceUID, err := h.getSleepaceDownloadIdentity(ctx, tenantID, deviceAddr)
	if err != nil {
		h.logger.Error("Failed to resolve sleepace report identity",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", deviceAddr),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to resolve device: %v", err)))
		return
	}

	req := service.DownloadReportRequest{
		TenantID:         tenantID,
		DeviceAddr:       deviceAddr,
		SleepaceDeviceID: sleepaceDeviceID,
		DeviceUID:        deviceUID,
		StartTime:        startTime,
		EndTime:          endTime,
	}

	err = h.sleepaceReportService.DownloadReport(ctx, req)
	if err != nil {
		h.logger.Error("DownloadReport failed",
			zap.String("tenant_id", tenantID),
			zap.String("device_addr", deviceAddr),
			zap.String("sleepace_device_id", sleepaceDeviceID),
			zap.String("device_uid", deviceUID),
			zap.Int64("start_time", startTime),
			zap.Int64("end_time", endTime),
			zap.Error(err),
		)
		writeJSON(w, http.StatusOK, Fail(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}

// getSleepaceDownloadIdentity 解析厂家 deviceId（device_factory_meta.device_code 优先）与 deviceName（device_factory_meta.device_uid）。
// v2: tenant scope 用 INET <<= /48 prefix；devices 表已无 status 列，缺该过滤（软删通过 device_ipv6 退回 trash 池）。
func (h *SleepaceReportHandler) getSleepaceDownloadIdentity(ctx context.Context, tenantID, deviceAddr string) (sleepaceDeviceID, deviceUID string, err error) {
	if h.db == nil {
		return "", "", fmt.Errorf("database connection not available")
	}
	q := `
		SELECT COALESCE(NULLIF(TRIM(dfm.device_code), ''), dfm.device_uid), dfm.device_uid
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		WHERE d.device_addr <<= $1::INET
		  AND d.device_addr = $2::INET
		LIMIT 1
	`
	err = h.db.QueryRowContext(ctx, q, tenantID, deviceAddr).Scan(&sleepaceDeviceID, &deviceUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("device not found")
		}
		return "", "", fmt.Errorf("lookup device: %w", err)
	}
	if sleepaceDeviceID == "" || deviceUID == "" {
		return "", "", fmt.Errorf("incomplete device identifiers")
	}
	return sleepaceDeviceID, deviceUID, nil
}

// extractDeviceAddrFromPath 从路径中提取 device_addr (URL :id 占位为 vendor route 锁定保留)。
// 路径格式：/sleepace/api/v1/sleepace/reports/:id[/detail|/dates|/download]
func extractDeviceAddrFromPath(path string) string {
	prefix := "/sleepace/api/v1/sleepace/reports/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	deviceAddr := strings.TrimPrefix(path, prefix)
	deviceAddr = strings.TrimSuffix(deviceAddr, "/detail")
	deviceAddr = strings.TrimSuffix(deviceAddr, "/dates")
	deviceAddr = strings.TrimSuffix(deviceAddr, "/download")
	deviceAddr = strings.TrimSuffix(deviceAddr, "/")
	return deviceAddr
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
func (h *SleepaceReportHandler) checkReportPermission(ctx context.Context, tenantID, deviceAddr, userID, userType, userRole, permissionType string) error {
	if h.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// 1. 通过 device_addr 获取关联的住户信息
	residentInfo, err := h.getResidentByDeviceAddr(ctx, tenantID, deviceAddr)
	if err != nil {
		// 如果设备没有关联住户，允许访问（fallback）
		return nil
	}

	// 2. Staff 角色权限检查
	if userType == "staff" && userRole != "" {
		// 2.1 Caregiver/Nurse：检查 assign-only
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

		// 2.2 Manager：Phase 3 按 Current Branch (user_branches.is_primary) 严格过滤
		if userRole == "Manager" {
			perm, err := GetResourcePermission(h.db, ctx, userRole, "residents", "R")
			if err == nil && perm.BranchOnly {
				// 1) 拿 user 的 Current Branch /56
				var userBranch sql.NullString
				_ = h.db.QueryRowContext(ctx, `
					SELECT host(branch_prefix) || '/56' FROM user_branches
					 WHERE user_id = $1::UUID AND is_primary = TRUE AND valid_to IS NULL LIMIT 1`,
					userID,
				).Scan(&userBranch)
				if !userBranch.Valid || userBranch.String == "" {
					// 没挂任何 branch = tenant-wide（admin 走前面 SystemAdmin 分支，到这里基本表 unassigned）
					return fmt.Errorf("access denied: manager has no current branch")
				}
				// 2) 拿 resident 的 Current Branch /56 (resident_unit 反推)
				var residentBranch sql.NullString
				_ = h.db.QueryRowContext(ctx, `
					SELECT host(network(set_masklen(spatial_prefix, 56))) || '/56'
					  FROM resident_unit
					 WHERE resident_id = $1::INET AND valid_to IS NULL
					 ORDER BY valid_from DESC LIMIT 1`,
					residentInfo.ResidentID,
				).Scan(&residentBranch)
				if !residentBranch.Valid || residentBranch.String != userBranch.String {
					return fmt.Errorf("access denied: resident belongs to different branch")
				}
			}
			return nil
		}
	}

	// 3. 其他角色：默认允许（SystemAdmin 等）
	return nil
}

// residentInfo 住户信息（用于权限检查）
type residentInfo struct {
	ResidentID string
	BranchTag  sql.NullString
	UnitID     sql.NullString
}

// getResidentByDeviceAddr 查询路径：device_addr → resident_unit.spatial_prefix LPM → residents。
func (h *SleepaceReportHandler) getResidentByDeviceAddr(ctx context.Context, tenantID, deviceAddr string) (*residentInfo, error) {
	query := `
		SELECT host(r.resident_id),
		       NULL::text AS branch_tag,
		       host(network(set_masklen(d.device_addr, 80))) || '/80' AS unit_id
		FROM devices d
		JOIN resident_unit ru
		  ON ru.valid_to IS NULL
		 AND d.device_addr <<= ru.spatial_prefix
		JOIN residents r ON r.resident_id = ru.resident_id
		WHERE d.device_addr <<= $1::INET
		  AND d.device_addr = $2::INET
		ORDER BY masklen(ru.spatial_prefix) DESC
		LIMIT 1
	`

	var info residentInfo
	err := h.db.QueryRowContext(ctx, query, tenantID, deviceAddr).Scan(
		&info.ResidentID,
		&info.BranchTag,
		&info.UnitID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no resident found for device")
		}
		return nil, fmt.Errorf("failed to get resident by device_addr: %w", err)
	}

	return &info, nil
}

// isResidentAssignedToUser 检查住户是否分配给该用户
// resident_caregivers 表通过 user_list (JSONB) 存储用户ID列表
func (h *SleepaceReportHandler) isResidentAssignedToUser(ctx context.Context, tenantID, residentID, userID string) bool {
	// 查询 resident_caregivers 表的 user_list 字段（JSONB 数组）
	query := `
		SELECT user_list
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

