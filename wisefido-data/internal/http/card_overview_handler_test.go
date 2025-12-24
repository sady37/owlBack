// +build integration

package httpapi

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"wisefido-data/internal/repository"
	"wisefido-data/internal/service"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupCardOverviewTestData 设置 Card Overview 测试数据
func setupCardOverviewTestData(t *testing.T, db *sql.DB, tenantID string) (unitID, bedID, residentID, cardID string) {
	ctx := context.Background()

	// 1. 创建 building
	buildingID := "00000000-0000-0000-0000-000000000975"
	_, err := db.ExecContext(ctx,
		`INSERT INTO buildings (building_id, tenant_id, building_name, branch_tag)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (building_id) DO UPDATE SET building_name = EXCLUDED.building_name, branch_tag = EXCLUDED.branch_tag`,
		buildingID, tenantID, "Test Building", "BRANCH-1",
	)
	require.NoError(t, err)

	// 2. 创建单元（unit）
	unitID = "00000000-0000-0000-0000-000000000976"
	_, err = db.ExecContext(ctx,
		`INSERT INTO units (unit_id, tenant_id, unit_name, building, floor, unit_type, branch_tag, unit_number, timezone, is_public_space, is_multi_person_room)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name, building = EXCLUDED.building, floor = EXCLUDED.floor, unit_type = EXCLUDED.unit_type, branch_tag = EXCLUDED.branch_tag, unit_number = EXCLUDED.unit_number, timezone = EXCLUDED.timezone, is_public_space = EXCLUDED.is_public_space, is_multi_person_room = EXCLUDED.is_multi_person_room`,
		unitID, tenantID, "Test Unit 001", "Test Building", "1F", "Facility", "BRANCH-1", "001", "America/Denver", false, false,
	)
	require.NoError(t, err)

	// 3. 创建房间（room）
	roomID := "00000000-0000-0000-0000-000000000977"
	_, err = db.ExecContext(ctx,
		`INSERT INTO rooms (room_id, tenant_id, unit_id, room_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_id) DO UPDATE SET room_name = EXCLUDED.room_name`,
		roomID, tenantID, unitID, "Test Room 001",
	)
	require.NoError(t, err)

	// 4. 创建床位（bed）
	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	bedID = "00000000-0000-0000-0000-000000000978"
	_, err = db.ExecContext(ctx,
		`INSERT INTO beds (bed_id, tenant_id, room_id, bed_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (bed_id) DO UPDATE SET bed_name = EXCLUDED.bed_name`,
		bedID, tenantID, roomID, "Test Bed 001",
	)
	require.NoError(t, err)

	// 5. 创建住户（resident）
	residentID = "00000000-0000-0000-0000-000000000979"
	accountHash := hashStringForCardOverview("test_resident_001")
	_, err = db.ExecContext(ctx,
		`INSERT INTO residents (resident_id, tenant_id, resident_account, resident_account_hash, nickname, status, can_view_status, unit_id, admission_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_DATE)
		 ON CONFLICT (resident_id) DO UPDATE SET nickname = EXCLUDED.nickname`,
		residentID, tenantID, "test_resident_001", accountHash, "Test Resident 001", "active", true, unitID,
	)
	require.NoError(t, err)

	// 6. 创建卡片（card）
	cardID = "00000000-0000-0000-0000-000000000980"
	devicesJSON := json.RawMessage("[]")
	residentsJSON, _ := json.Marshal([]string{residentID})
	_, err = db.ExecContext(ctx,
		`INSERT INTO cards (card_id, tenant_id, card_type, bed_id, unit_id, card_name, card_address, resident_id, devices, residents)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (card_id) DO UPDATE SET card_name = EXCLUDED.card_name`,
		cardID, tenantID, "ActiveBed", bedID, unitID, "Test Card 1", "Test Address 1", residentID, devicesJSON, residentsJSON,
	)
	require.NoError(t, err)

	return unitID, bedID, residentID, cardID
}

// hashStringForCardOverview 计算字符串的 SHA256 hash
func hashStringForCardOverview(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}

// cleanupCardOverviewTestData 清理 Card Overview 测试数据
func cleanupCardOverviewTestData(t *testing.T, db *sql.DB, tenantID string) {
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, `DELETE FROM cards WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM resident_caregivers WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM resident_contacts WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM resident_phi WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM residents WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM beds WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM rooms WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM units WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// getTestDBForCardOverview 获取测试数据库连接
func getTestDBForCardOverview(t *testing.T) *sql.DB {
	return setupTestDB(t)
}

// getTestLoggerForCardOverview 获取测试日志记录器
func getTestLoggerForCardOverview() *zap.Logger {
	return zap.NewNop()
}

// TestCardOverviewHandler_GetCardOverview_Basic 测试基本的 GetCardOverview 功能
func TestCardOverviewHandler_GetCardOverview_Basic(t *testing.T) {
	db := getTestDBForCardOverview(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建测试租户
	tenantID := "00000000-0000-0000-0000-000000000974"
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Card Overview Tenant", "test-card-overview.local",
	)
	require.NoError(t, err)
	defer cleanupCardOverviewTestData(t, db, tenantID)

	// 创建测试数据
	_, _, residentID, cardID := setupCardOverviewTestData(t, db, tenantID)

	// 创建 Repository 和 Service
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := getTestLoggerForCardOverview()
	cardService := service.NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	// 创建 Handler
	stub := NewStubHandler(nil, nil, db)
	handler := NewCardOverviewHandler(stub, cardService, logger)

	// 创建 HTTP 请求
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/card-overview?tenant_id="+tenantID, nil)
	req.Header.Set("X-User-Id", residentID)
	req.Header.Set("X-User-Type", "resident")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(w, req)

	// 验证响应
	require.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	require.Equal(t, "success", result["type"])

	data, ok := result["result"].(map[string]any)
	require.True(t, ok, "result should be a map")
	require.NotNil(t, data["items"])

	items, ok := data["items"].([]any)
	require.True(t, ok, "items should be an array")
	require.Len(t, items, 1)

	item, ok := items[0].(map[string]any)
	require.True(t, ok, "item should be a map")
	require.Equal(t, cardID, item["card_id"])
	require.Equal(t, "ActiveBed", item["card_type"]) // 测试数据创建的是 ActiveBed 卡片
	require.Equal(t, "Test Card 1", item["card_name"])
}

// TestCardOverviewHandler_GetCardOverview_WithSearch 测试搜索功能
func TestCardOverviewHandler_GetCardOverview_WithSearch(t *testing.T) {
	db := getTestDBForCardOverview(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建测试租户
	tenantID := "00000000-0000-0000-0000-000000000973"
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Card Overview Tenant 2", "test-card-overview-2.local",
	)
	require.NoError(t, err)
	defer cleanupCardOverviewTestData(t, db, tenantID)

	// 创建测试数据
	_, _, residentID, cardID := setupCardOverviewTestData(t, db, tenantID)

	// 创建 Repository 和 Service
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := getTestLoggerForCardOverview()
	cardService := service.NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	// 创建 Handler
	stub := NewStubHandler(nil, nil, db)
	handler := NewCardOverviewHandler(stub, cardService, logger)

	// 创建 HTTP 请求（带搜索参数）
	u, err := url.Parse("/admin/api/v1/card-overview")
	require.NoError(t, err)
	q := u.Query()
	q.Set("tenant_id", tenantID)
	q.Set("search", "Test Card")
	u.RawQuery = q.Encode()
	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Set("X-User-Id", residentID)
	req.Header.Set("X-User-Type", "resident")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(w, req)

	// 验证响应
	require.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	require.Equal(t, "success", result["type"])

	data, ok := result["result"].(map[string]any)
	require.True(t, ok, "result should be a map")
	require.NotNil(t, data["items"])

	items, ok := data["items"].([]any)
	require.True(t, ok, "items should be an array")
	require.Len(t, items, 1)

	item, ok := items[0].(map[string]any)
	require.True(t, ok, "item should be a map")
	require.Equal(t, cardID, item["card_id"])
}

// TestCardOverviewHandler_GetCardOverview_EmptyResult 测试空结果
func TestCardOverviewHandler_GetCardOverview_EmptyResult(t *testing.T) {
	db := getTestDBForCardOverview(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建测试租户
	tenantID := "00000000-0000-0000-0000-000000000972"
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Card Overview Tenant 3", "test-card-overview-3.local",
	)
	require.NoError(t, err)
	defer cleanupCardOverviewTestData(t, db, tenantID)

	// 创建 Repository 和 Service
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := getTestLoggerForCardOverview()
	cardService := service.NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	// 创建 Handler
	stub := NewStubHandler(nil, nil, db)
	handler := NewCardOverviewHandler(stub, cardService, logger)

	// 创建 HTTP 请求（没有卡片）
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/card-overview?tenant_id="+tenantID, nil)
	req.Header.Set("X-User-Id", "00000000-0000-0000-0000-000000000000")
	req.Header.Set("X-User-Type", "resident")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(w, req)

	// 验证响应
	require.Equal(t, http.StatusOK, w.Code)

	var result map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	require.Equal(t, "success", result["type"])

	data, ok := result["result"].(map[string]any)
	require.True(t, ok)
	require.NotNil(t, data["items"])

	items, ok := data["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 0)
}

// TestCardOverviewHandler_GetCardOverview_MethodNotAllowed 测试方法不允许
func TestCardOverviewHandler_GetCardOverview_MethodNotAllowed(t *testing.T) {
	db := getTestDBForCardOverview(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建 Repository 和 Service
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := getTestLoggerForCardOverview()
	cardService := service.NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	// 创建 Handler
	stub := NewStubHandler(nil, nil, db)
	handler := NewCardOverviewHandler(stub, cardService, logger)

	// 创建 POST 请求（应该返回 405）
	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/card-overview", nil)
	w := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(w, req)

	// 验证响应
	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestCardOverviewHandler_GetCardOverview_NotFound 测试路径不存在
func TestCardOverviewHandler_GetCardOverview_NotFound(t *testing.T) {
	db := getTestDBForCardOverview(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建 Repository 和 Service
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := getTestLoggerForCardOverview()
	cardService := service.NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	// 创建 Handler
	stub := NewStubHandler(nil, nil, db)
	handler := NewCardOverviewHandler(stub, cardService, logger)

	// 创建请求（错误的路径）
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/card-overview-wrong", nil)
	w := httptest.NewRecorder()

	// 执行请求
	handler.ServeHTTP(w, req)

	// 验证响应
	require.Equal(t, http.StatusNotFound, w.Code)
}

