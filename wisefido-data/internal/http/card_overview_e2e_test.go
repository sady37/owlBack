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

// setupCardOverviewE2ETestData 设置端到端测试数据
func setupCardOverviewE2ETestData(t *testing.T, db *sql.DB, tenantID string) (unitID1, unitID2, bedID1, bedID2, residentID1, residentID2, contactID, cardID1, cardID2, userID string) {
	ctx := context.Background()

	// 1. 创建 building
	buildingID := "00000000-0000-0000-0000-000000000965"
	_, err := db.ExecContext(ctx,
		`INSERT INTO buildings (building_id, tenant_id, building_name, branch_tag)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (building_id) DO UPDATE SET building_name = EXCLUDED.building_name, branch_tag = EXCLUDED.branch_tag`,
		buildingID, tenantID, "Test Building", "BRANCH-1",
	)
	require.NoError(t, err)

	// 2. 创建两个单元（一个非 share，一个 share）
	unitID1 = "00000000-0000-0000-0000-000000000966"
	_, err = db.ExecContext(ctx,
		`INSERT INTO units (unit_id, tenant_id, unit_name, building, floor, unit_type, branch_tag, unit_number, timezone, is_public_space, is_multi_person_room)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name, building = EXCLUDED.building, floor = EXCLUDED.floor, unit_type = EXCLUDED.unit_type, branch_tag = EXCLUDED.branch_tag, unit_number = EXCLUDED.unit_number, timezone = EXCLUDED.timezone, is_public_space = EXCLUDED.is_public_space, is_multi_person_room = EXCLUDED.is_multi_person_room`,
		unitID1, tenantID, "Test Unit 001", "Test Building", "1F", "Facility", "BRANCH-1", "001", "America/Denver", false, false,
	)
	require.NoError(t, err)

	unitID2 = "00000000-0000-0000-0000-000000000967"
	_, err = db.ExecContext(ctx,
		`INSERT INTO units (unit_id, tenant_id, unit_name, building, floor, unit_type, branch_tag, unit_number, timezone, is_public_space, is_multi_person_room)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name, building = EXCLUDED.building, floor = EXCLUDED.floor, unit_type = EXCLUDED.unit_type, branch_tag = EXCLUDED.branch_tag, unit_number = EXCLUDED.unit_number, timezone = EXCLUDED.timezone, is_public_space = EXCLUDED.is_public_space, is_multi_person_room = EXCLUDED.is_multi_person_room`,
		unitID2, tenantID, "Test Unit 002 (Share)", "Test Building", "1F", "Facility", "BRANCH-1", "002", "America/Denver", true, false,
	)
	require.NoError(t, err)

	// 3. 创建房间
	roomID1 := "00000000-0000-0000-0000-000000000968"
	_, err = db.ExecContext(ctx,
		`INSERT INTO rooms (room_id, tenant_id, unit_id, room_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_id) DO UPDATE SET room_name = EXCLUDED.room_name`,
		roomID1, tenantID, unitID1, "Test Room 001",
	)
	require.NoError(t, err)

	roomID2 := "00000000-0000-0000-0000-000000000969"
	_, err = db.ExecContext(ctx,
		`INSERT INTO rooms (room_id, tenant_id, unit_id, room_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_id) DO UPDATE SET room_name = EXCLUDED.room_name`,
		roomID2, tenantID, unitID2, "Test Room 002",
	)
	require.NoError(t, err)

	// 4. 创建床位
	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	bedID1 = "00000000-0000-0000-0000-000000000970"
	_, err = db.ExecContext(ctx,
		`INSERT INTO beds (bed_id, tenant_id, room_id, bed_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (bed_id) DO UPDATE SET bed_name = EXCLUDED.bed_name`,
		bedID1, tenantID, roomID1, "Test Bed 001",
	)
	require.NoError(t, err)

	bedID2 = "00000000-0000-0000-0000-000000000971"
	_, err = db.ExecContext(ctx,
		`INSERT INTO beds (bed_id, tenant_id, room_id, bed_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (bed_id) DO UPDATE SET bed_name = EXCLUDED.bed_name`,
		bedID2, tenantID, roomID2, "Test Bed 002",
	)
	require.NoError(t, err)

	// 5. 创建住户
	residentID1 = "00000000-0000-0000-0000-000000000972"
	accountHash1 := hashStringForCardOverviewE2E("test_resident_e2e_001")
	_, err = db.ExecContext(ctx,
		`INSERT INTO residents (resident_id, tenant_id, resident_account, resident_account_hash, nickname, status, can_view_status, unit_id, admission_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_DATE)
		 ON CONFLICT (resident_id) DO UPDATE SET nickname = EXCLUDED.nickname`,
		residentID1, tenantID, "test_resident_e2e_001", accountHash1, "Test Resident E2E 001", "active", true, unitID1,
	)
	require.NoError(t, err)

	residentID2 = "00000000-0000-0000-0000-000000000973"
	accountHash2 := hashStringForCardOverviewE2E("test_resident_e2e_002")
	_, err = db.ExecContext(ctx,
		`INSERT INTO residents (resident_id, tenant_id, resident_account, resident_account_hash, nickname, status, can_view_status, unit_id, admission_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_DATE)
		 ON CONFLICT (resident_id) DO UPDATE SET nickname = EXCLUDED.nickname`,
		residentID2, tenantID, "test_resident_e2e_002", accountHash2, "Test Resident E2E 002", "active", true, unitID1,
	)
	require.NoError(t, err)

	// 6. 创建联系人（Family 用户）
	contactID = "00000000-0000-0000-0000-000000000974"
	_, err = db.ExecContext(ctx,
		`INSERT INTO resident_contacts (
			contact_id, tenant_id, resident_id, slot, is_enabled, relationship,
			contact_first_name, contact_last_name, receive_sms, receive_email
		) VALUES (
			$1, $2, $3, 'A', true, 'Family',
			'Test', 'Contact', false, false
		)
		ON CONFLICT (contact_id) DO UPDATE SET relationship = EXCLUDED.relationship`,
		contactID, tenantID, residentID1,
	)
	require.NoError(t, err)

	// 7. 创建 Staff 用户（Caregiver）
	userID = "00000000-0000-0000-0000-000000000975"
	userAccountHash := hashStringForCardOverviewE2E("test_caregiver_e2e")
	_, err = db.ExecContext(ctx,
		`INSERT INTO users (user_id, tenant_id, user_account, user_account_hash, password_hash, nickname, role, status, branch_tag)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'active', $8)
		 ON CONFLICT (tenant_id, user_account) DO UPDATE SET role = EXCLUDED.role, branch_tag = EXCLUDED.branch_tag`,
		userID, tenantID, "test_caregiver_e2e", userAccountHash, []byte("password_hash"), "Test Caregiver E2E", "Caregiver", "BRANCH-1",
	)
	require.NoError(t, err)

	// 8. 创建权限配置（Caregiver assigned_only）
	_, err = db.ExecContext(ctx,
		`INSERT INTO role_permissions (tenant_id, role_code, resource_type, permission_type, assigned_only, branch_only)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT ((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)), role_code, resource_type, permission_type)
		 DO UPDATE SET assigned_only = EXCLUDED.assigned_only, branch_only = EXCLUDED.branch_only`,
		SystemTenantID(), "Caregiver", "cards", "R", true, false,
	)
	require.NoError(t, err)

	// 9. 创建住户分配关系（resident_caregivers）
	userListJSON, _ := json.Marshal([]string{userID})
	_, err = db.ExecContext(ctx,
		`INSERT INTO resident_caregivers (tenant_id, resident_id, userList)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (tenant_id, resident_id) DO UPDATE SET userList = EXCLUDED.userList`,
		tenantID, residentID1, userListJSON,
	)
	require.NoError(t, err)

	// 10. 创建卡片
	cardID1 = "00000000-0000-0000-0000-000000000976"
	devicesJSON1 := json.RawMessage("[]")
	residentsJSON1, _ := json.Marshal([]string{residentID1})
	_, err = db.ExecContext(ctx,
		`INSERT INTO cards (card_id, tenant_id, card_type, bed_id, unit_id, card_name, card_address, resident_id, devices, residents)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (card_id) DO UPDATE SET card_name = EXCLUDED.card_name`,
		cardID1, tenantID, "ActiveBed", bedID1, unitID1, "Test Card 1", "Test Address 1", residentID1, devicesJSON1, residentsJSON1,
	)
	require.NoError(t, err)

	cardID2 = "00000000-0000-0000-0000-000000000977"
	devicesJSON2 := json.RawMessage("[]")
	residentsJSON2, _ := json.Marshal([]string{residentID1, residentID2})
	_, err = db.ExecContext(ctx,
		`INSERT INTO cards (card_id, tenant_id, card_type, bed_id, unit_id, card_name, card_address, resident_id, devices, residents)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (card_id) DO UPDATE SET card_name = EXCLUDED.card_name`,
		cardID2, tenantID, "Location", nil, unitID1, "Test Unit Card 1", "Test Unit Address 1", nil, devicesJSON2, residentsJSON2,
	)
	require.NoError(t, err)

	return unitID1, unitID2, bedID1, bedID2, residentID1, residentID2, contactID, cardID1, cardID2, userID
}

// hashStringForCardOverviewE2E 计算字符串的 SHA256 hash
func hashStringForCardOverviewE2E(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}

// cleanupCardOverviewE2ETestData 清理端到端测试数据
func cleanupCardOverviewE2ETestData(t *testing.T, db *sql.DB, tenantID string) {
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, `DELETE FROM cards WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM resident_caregivers WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM resident_contacts WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM resident_phi WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM residents WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM role_permissions WHERE tenant_id = $1`, SystemTenantID())
	_, _ = db.ExecContext(ctx, `DELETE FROM beds WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM rooms WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM units WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	_, _ = db.ExecContext(ctx, `DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// TestCardOverviewE2E_ResidentUser 测试 Resident 用户的端到端流程
func TestCardOverviewE2E_ResidentUser(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建测试租户
	tenantID := "00000000-0000-0000-0000-000000000964"
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Card Overview E2E Tenant", "test-card-overview-e2e.local",
	)
	require.NoError(t, err)
	defer cleanupCardOverviewE2ETestData(t, db, tenantID)

	// 创建测试数据
	_, _, _, _, residentID1, _, _, cardID1, cardID2, _ := setupCardOverviewE2ETestData(t, db, tenantID)

	// 创建完整的服务栈
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := zap.NewNop()
	cardService := service.NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	// 创建 Handler
	stub := NewStubHandler(nil, nil, db)
	handler := NewCardOverviewHandler(stub, cardService, logger)

	// 创建 HTTP 请求（Resident 用户）
	u, err := url.Parse("/admin/api/v1/card-overview")
	require.NoError(t, err)
	q := u.Query()
	q.Set("tenant_id", tenantID)
	u.RawQuery = q.Encode()
	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Set("X-User-Id", residentID1)
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
	// Resident 应该能看到自己的 ActiveBed 卡片和 Unit 卡片（如果他是第一个住户）
	require.GreaterOrEqual(t, len(items), 1)

	// 验证卡片类型规范化（Location → Unit）
	foundCard1 := false
	foundCard2 := false
	for _, itemAny := range items {
		item, ok := itemAny.(map[string]any)
		require.True(t, ok)
		if item["card_id"] == cardID1 {
			foundCard1 = true
			require.Equal(t, "ActiveBed", item["card_type"])
		}
		if item["card_id"] == cardID2 {
			foundCard2 = true
			require.Equal(t, "Unit", item["card_type"]) // 应该是 'Unit' 而不是 'Location'
		}
	}
	require.True(t, foundCard1, "Should find ActiveBed card")
	require.True(t, foundCard2, "Should find Unit card")
}

// TestCardOverviewE2E_FamilyUser 测试 Family 用户的端到端流程
func TestCardOverviewE2E_FamilyUser(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建测试租户
	tenantID := "00000000-0000-0000-0000-000000000963"
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Card Overview E2E Tenant 2", "test-card-overview-e2e-2.local",
	)
	require.NoError(t, err)
	defer cleanupCardOverviewE2ETestData(t, db, tenantID)

	// 创建测试数据
	_, _, _, _, _, _, contactID, cardID1, _, _ := setupCardOverviewE2ETestData(t, db, tenantID)

	// 创建完整的服务栈
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := zap.NewNop()
	cardService := service.NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	// 创建 Handler
	stub := NewStubHandler(nil, nil, db)
	handler := NewCardOverviewHandler(stub, cardService, logger)

	// 创建 HTTP 请求（Family 用户）
	u, err := url.Parse("/admin/api/v1/card-overview")
	require.NoError(t, err)
	q := u.Query()
	q.Set("tenant_id", tenantID)
	u.RawQuery = q.Encode()
	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Set("X-User-Id", contactID)
	req.Header.Set("X-User-Type", "family")

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
	// Family 用户应该能看到关联住户的卡片
	require.GreaterOrEqual(t, len(items), 1)

	// 验证能看到关联住户的卡片
	foundCard1 := false
	for _, itemAny := range items {
		item, ok := itemAny.(map[string]any)
		require.True(t, ok)
		if item["card_id"] == cardID1 {
			foundCard1 = true
			break
		}
	}
	require.True(t, foundCard1, "Family user should see resident's card")
}

// TestCardOverviewE2E_StaffUser_AssignedOnly 测试 Staff 用户（AssignedOnly）的端到端流程
func TestCardOverviewE2E_StaffUser_AssignedOnly(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建测试租户
	tenantID := "00000000-0000-0000-0000-000000000962"
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Card Overview E2E Tenant 3", "test-card-overview-e2e-3.local",
	)
	require.NoError(t, err)
	defer cleanupCardOverviewE2ETestData(t, db, tenantID)

	// 创建测试数据
	_, _, _, _, _, _, _, cardID1, cardID2, userID := setupCardOverviewE2ETestData(t, db, tenantID)

	// 创建完整的服务栈
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := zap.NewNop()
	cardService := service.NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	// 创建 Handler
	stub := NewStubHandler(nil, nil, db)
	handler := NewCardOverviewHandler(stub, cardService, logger)

	// 创建 HTTP 请求（Staff 用户，AssignedOnly）
	u, err := url.Parse("/admin/api/v1/card-overview")
	require.NoError(t, err)
	q := u.Query()
	q.Set("tenant_id", tenantID)
	u.RawQuery = q.Encode()
	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Set("X-User-Id", userID)
	req.Header.Set("X-User-Type", "staff")
	req.Header.Set("X-User-Role", "Caregiver")

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
	// Caregiver 应该只能看到分配的住户的卡片（residentID1）
	require.GreaterOrEqual(t, len(items), 1)

	// 验证只能看到分配的卡片
	foundCard1 := false
	foundCard2 := false
	for _, itemAny := range items {
		item, ok := itemAny.(map[string]any)
		require.True(t, ok)
		if item["card_id"] == cardID1 {
			foundCard1 = true
		}
		if item["card_id"] == cardID2 {
			foundCard2 = true
		}
	}
	// 至少应该能看到一个卡片（cardID1 或 cardID2）
	require.True(t, foundCard1 || foundCard2, "Should find at least one assigned resident's card")
}

// TestCardOverviewE2E_ShareUnit_AccessDenied 测试 Share Unit 拒绝访问
func TestCardOverviewE2E_ShareUnit_AccessDenied(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建测试租户
	tenantID := "00000000-0000-0000-0000-000000000961"
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Card Overview E2E Tenant 4", "test-card-overview-e2e-4.local",
	)
	require.NoError(t, err)
	defer cleanupCardOverviewE2ETestData(t, db, tenantID)

	// 创建测试数据（包含 Share Unit）
	_, unitID2, _, _, residentID1, _, _, _, _, _ := setupCardOverviewE2ETestData(t, db, tenantID)

	// 创建 Share Unit 的卡片
	cardIDShare := "00000000-0000-0000-0000-000000000978"
	devicesJSON := json.RawMessage("[]")
	residentsJSON, _ := json.Marshal([]string{residentID1})
	_, err = db.ExecContext(ctx,
		`INSERT INTO cards (card_id, tenant_id, card_type, bed_id, unit_id, card_name, card_address, resident_id, devices, residents)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (card_id) DO UPDATE SET card_name = EXCLUDED.card_name`,
		cardIDShare, tenantID, "Location", nil, unitID2, "Test Share Unit Card", "Test Share Address", nil, devicesJSON, residentsJSON,
	)
	require.NoError(t, err)

	// 创建完整的服务栈
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := zap.NewNop()
	cardService := service.NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	// 创建 Handler
	stub := NewStubHandler(nil, nil, db)
	handler := NewCardOverviewHandler(stub, cardService, logger)

	// 创建 HTTP 请求（Resident 用户）
	u, err := url.Parse("/admin/api/v1/card-overview")
	require.NoError(t, err)
	q := u.Query()
	q.Set("tenant_id", tenantID)
	u.RawQuery = q.Encode()
	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Set("X-User-Id", residentID1)
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

	// 验证 Share Unit 卡片不在结果中
	foundShareCard := false
	for _, itemAny := range items {
		item, ok := itemAny.(map[string]any)
		require.True(t, ok)
		if item["card_id"] == cardIDShare {
			foundShareCard = true
			break
		}
	}
	require.False(t, foundShareCard, "Share Unit card should not be visible to resident")
}

// TestCardOverviewE2E_FamilyView_Calculation 测试 family_view 计算
func TestCardOverviewE2E_FamilyView_Calculation(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// 创建测试租户
	tenantID := "00000000-0000-0000-0000-000000000960"
	ctx := context.Background()
	_, err := db.ExecContext(ctx,
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Card Overview E2E Tenant 5", "test-card-overview-e2e-5.local",
	)
	require.NoError(t, err)
	defer cleanupCardOverviewE2ETestData(t, db, tenantID)

	// 创建测试数据
	_, _, _, _, residentID1, _, _, cardID1, _, _ := setupCardOverviewE2ETestData(t, db, tenantID)

	// 创建完整的服务栈
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := zap.NewNop()
	cardService := service.NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	// 创建 Handler
	stub := NewStubHandler(nil, nil, db)
	handler := NewCardOverviewHandler(stub, cardService, logger)

	// 创建 HTTP 请求（Resident 用户）
	u, err := url.Parse("/admin/api/v1/card-overview")
	require.NoError(t, err)
	q := u.Query()
	q.Set("tenant_id", tenantID)
	u.RawQuery = q.Encode()
	req := httptest.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Set("X-User-Id", residentID1)
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
	require.GreaterOrEqual(t, len(items), 1)

	// 验证 family_view 字段存在
	for _, itemAny := range items {
		item, ok := itemAny.(map[string]any)
		require.True(t, ok)
		if item["card_id"] == cardID1 {
			_, hasFamilyView := item["family_view"]
			require.True(t, hasFamilyView, "Card should have family_view field")
			// residentID1 的 can_view_status = true，所以 family_view 应该是 true
			require.True(t, item["family_view"].(bool), "family_view should be true for resident with can_view_status = true")
			break
		}
	}
}

