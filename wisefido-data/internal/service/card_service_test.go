// +build integration

package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"testing"

	"wisefido-data/internal/repository"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTestDBForCard 设置测试数据库
func setupTestDBForCard(t *testing.T) *sql.DB {
	return getTestDBForService(t)
}

// getTestLoggerForCard 获取测试日志记录器
func getTestLoggerForCard() *zap.Logger {
	return getTestLogger()
}

// hashStringForCard 计算字符串的 SHA256 hash
func hashStringForCard(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}

// createTestTenantAndUnitForCard 创建测试租户和unit（cards需要unit_id和bed_id）
func createTestTenantAndUnitForCard(t *testing.T, db *sql.DB) (string, string, string) {
	tenantID := "00000000-0000-0000-0000-000000000985"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Card Tenant", "test-card.local",
	)
	require.NoError(t, err)

	// 创建测试building
	buildingID := "00000000-0000-0000-0000-000000000984"
	_, err = db.Exec(
		`INSERT INTO buildings (building_id, tenant_id, building_name, branch_tag)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (building_id) DO UPDATE SET building_name = EXCLUDED.building_name, branch_tag = EXCLUDED.branch_tag`,
		buildingID, tenantID, "Test Building", "BRANCH-1",
	)
	require.NoError(t, err)

	// 创建测试unit
	unitID := "00000000-0000-0000-0000-000000000983"
	_, err = db.Exec(
		`INSERT INTO units (unit_id, tenant_id, unit_name, building, floor, unit_type, branch_tag, unit_number, timezone, is_public_space, is_multi_person_room)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name, building = EXCLUDED.building, floor = EXCLUDED.floor, unit_type = EXCLUDED.unit_type, branch_tag = EXCLUDED.branch_tag, unit_number = EXCLUDED.unit_number, timezone = EXCLUDED.timezone, is_public_space = EXCLUDED.is_public_space, is_multi_person_room = EXCLUDED.is_multi_person_room`,
		unitID, tenantID, "Test Unit 001", "Test Building", "1F", "Facility", "BRANCH-1", "001", "America/Denver", false, false,
	)
	require.NoError(t, err)

	// 创建测试room
	roomID := "00000000-0000-0000-0000-000000000982"
	_, err = db.Exec(
		`INSERT INTO rooms (room_id, tenant_id, unit_id, room_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_id) DO UPDATE SET room_name = EXCLUDED.room_name`,
		roomID, tenantID, unitID, "Test Room 001",
	)
	require.NoError(t, err)

	// 创建测试bed
	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	bedID := "00000000-0000-0000-0000-000000000981"
	_, err = db.Exec(
		`INSERT INTO beds (bed_id, tenant_id, room_id, bed_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (bed_id) DO UPDATE SET bed_name = EXCLUDED.bed_name`,
		bedID, tenantID, roomID, "Test Bed 001",
	)
	require.NoError(t, err)

	return tenantID, unitID, bedID
}

// createTestResidentForCard 创建测试住户
func createTestResidentForCard(t *testing.T, db *sql.DB, tenantID, unitID, accountSuffix string) string {
	var residentID string
	account := "test_resident_" + accountSuffix
	nickname := "Test Resident " + accountSuffix
	accountHash := hashStringForCard(account)
	err := db.QueryRow(`
		INSERT INTO residents (
			tenant_id, resident_account, resident_account_hash, nickname,
			status, can_view_status, unit_id, admission_date
		) VALUES (
			$1, $2, $3, $4,
			'active', true, $5, CURRENT_DATE
		)
		RETURNING resident_id::text
	`, tenantID, account, accountHash, nickname, unitID).Scan(&residentID)
	require.NoError(t, err)
	return residentID
}

// createTestCardForService 创建测试卡片
func createTestCardForService(t *testing.T, db *sql.DB, tenantID, cardType, cardName, cardAddress string, bedID, unitID, residentID sql.NullString) string {
	var cardID string
	devicesJSON := json.RawMessage("[]")
	residentsJSON := json.RawMessage("[]")
	if residentID.Valid {
		residentsJSON, _ = json.Marshal([]string{residentID.String})
	}

	err := db.QueryRow(`
		INSERT INTO cards (
			tenant_id, card_type, bed_id, unit_id, card_name, card_address,
			resident_id, devices, residents
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING card_id::text
	`, tenantID, cardType, bedID, unitID, cardName, cardAddress, residentID, devicesJSON, residentsJSON).Scan(&cardID)
	require.NoError(t, err)
	return cardID
}

// cleanupTestDataForCard 清理测试数据
func cleanupTestDataForCard(t *testing.T, db *sql.DB, tenantID string) {
	_, _ = db.Exec(`DELETE FROM cards WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM resident_caregivers WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM resident_contacts WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM resident_phi WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM residents WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM beds WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM rooms WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM units WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	_, _ = db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// TestCardService_GetCardOverview_Basic 测试基本的 GetCardOverview 功能
func TestCardService_GetCardOverview_Basic(t *testing.T) {
	db := setupTestDBForCard(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, bedID := createTestTenantAndUnitForCard(t, db)
	defer cleanupTestDataForCard(t, db, tenantID)

	// 创建 Repository 和 Service
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := getTestLoggerForCard()
	cardService := NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	ctx := context.Background()

	// 创建测试数据
	residentID := createTestResidentForCard(t, db, tenantID, unitID, "001")
	cardID := createTestCardForService(t, db, tenantID, "ActiveBed", "Test Card 1", "Test Address 1",
		sql.NullString{String: bedID, Valid: true},
		sql.NullString{String: unitID, Valid: true},
		sql.NullString{String: residentID, Valid: true},
	)

	// 测试查询（Resident 用户）
	req := GetCardOverviewRequest{
		TenantID:        tenantID,
		CurrentUserID:   residentID,
		CurrentUserType: "resident",
	}

	resp, err := cardService.GetCardOverview(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Items, 1)
	require.Equal(t, cardID, resp.Items[0].CardID)
	require.Equal(t, "ActiveBed", resp.Items[0].CardType)
	require.Equal(t, "Test Card 1", resp.Items[0].CardName)
}

// TestCardService_GetCardOverview_FamilyUser 测试 Family 用户类型
func TestCardService_GetCardOverview_FamilyUser(t *testing.T) {
	db := setupTestDBForCard(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, bedID := createTestTenantAndUnitForCard(t, db)
	defer cleanupTestDataForCard(t, db, tenantID)

	// 创建 Repository 和 Service
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := getTestLoggerForCard()
	cardService := NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	ctx := context.Background()

	// 创建测试数据
	residentID := createTestResidentForCard(t, db, tenantID, unitID, "001")
	
	// 创建 contact
	var contactID string
	err := db.QueryRow(`
		INSERT INTO resident_contacts (
			tenant_id, resident_id, contact_id, slot, is_enabled, relationship,
			contact_first_name, contact_last_name, receive_sms, receive_email
		) VALUES (
			$1, $2, gen_random_uuid(), 'A', true, 'Family',
			'Test', 'Contact', false, false
		)
		RETURNING contact_id::text
	`, tenantID, residentID).Scan(&contactID)
	require.NoError(t, err)

	cardID := createTestCardForService(t, db, tenantID, "ActiveBed", "Test Card 1", "Test Address 1",
		sql.NullString{String: bedID, Valid: true},
		sql.NullString{String: unitID, Valid: true},
		sql.NullString{String: residentID, Valid: true},
	)

	// 测试查询（Family 用户）
	req := GetCardOverviewRequest{
		TenantID:        tenantID,
		CurrentUserID:   contactID,
		CurrentUserType: "family",
	}

	resp, err := cardService.GetCardOverview(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Items, 1)
	require.Equal(t, cardID, resp.Items[0].CardID)
}

// TestCardService_GetCardOverview_EmptyResult 测试空结果
func TestCardService_GetCardOverview_EmptyResult(t *testing.T) {
	db := setupTestDBForCard(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, _, _ := createTestTenantAndUnitForCard(t, db)
	defer cleanupTestDataForCard(t, db, tenantID)

	// 创建 Repository 和 Service
	cardsRepo := repository.NewPostgresCardsRepository(db)
	residentsRepo := repository.NewPostgresResidentsRepository(db)
	devicesRepo := repository.NewPostgresDevicesRepository(db)
	usersRepo := repository.NewPostgresUsersRepository(db)
	logger := getTestLoggerForCard()
	cardService := NewCardService(cardsRepo, residentsRepo, devicesRepo, usersRepo, db, logger)

	ctx := context.Background()

	// 测试查询（没有卡片）
	req := GetCardOverviewRequest{
		TenantID:        tenantID,
		CurrentUserID:   "00000000-0000-0000-0000-000000000000",
		CurrentUserType: "resident",
	}

	resp, err := cardService.GetCardOverview(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Items, 0)
	require.Equal(t, 0, resp.Total)
}

