// +build integration

package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"testing"
)

// hashString 计算字符串的 SHA256 hash
func hashStringForCards(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}

// createTestTenantAndUnitForCards 创建测试租户和unit（cards需要unit_id）
func createTestTenantAndUnitForCards(t *testing.T, db *sql.DB) (string, string, string) {
	tenantID := "00000000-0000-0000-0000-000000000995"
	_, err := db.Exec(
		`INSERT INTO tenants (tenant_id, tenant_name, domain, status)
		 VALUES ($1, $2, $3, 'active')
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_name = EXCLUDED.tenant_name, domain = EXCLUDED.domain, status = EXCLUDED.status`,
		tenantID, "Test Tenant Cards", "test-cards.local",
	)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	// 创建测试building
	buildingID := "00000000-0000-0000-0000-000000000994"
	_, err = db.Exec(
		`INSERT INTO buildings (building_id, tenant_id, building_name, branch_tag)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (building_id) DO UPDATE SET building_name = EXCLUDED.building_name, branch_tag = EXCLUDED.branch_tag`,
		buildingID, tenantID, "Test Building", "BRANCH-1",
	)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}

	// 创建测试unit（需要branch_tag、unit_number和timezone）
	unitID := "00000000-0000-0000-0000-000000000993"
	_, err = db.Exec(
		`INSERT INTO units (unit_id, tenant_id, unit_name, building, floor, unit_type, branch_tag, unit_number, timezone, is_public_space, is_multi_person_room)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (unit_id) DO UPDATE SET unit_name = EXCLUDED.unit_name, building = EXCLUDED.building, floor = EXCLUDED.floor, unit_type = EXCLUDED.unit_type, branch_tag = EXCLUDED.branch_tag, unit_number = EXCLUDED.unit_number, timezone = EXCLUDED.timezone, is_public_space = EXCLUDED.is_public_space, is_multi_person_room = EXCLUDED.is_multi_person_room`,
		unitID, tenantID, "Test Unit 001", "Test Building", "1F", "Facility", "BRANCH-1", "001", "America/Denver", false, false,
	)
	if err != nil {
		t.Fatalf("Failed to create test unit: %v", err)
	}

	// 创建测试room（ActiveBed 卡片需要 bed_id，bed 需要 room_id）
	roomID := "00000000-0000-0000-0000-000000000992"
	_, err = db.Exec(
		`INSERT INTO rooms (room_id, tenant_id, unit_id, room_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_id) DO UPDATE SET room_name = EXCLUDED.room_name`,
		roomID, tenantID, unitID, "Test Room 001",
	)
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	// 创建测试bed（ActiveBed 卡片需要 bed_id）
	// 注意：bed_type 字段已删除，ActiveBed 判断由应用层动态计算
	bedID := "00000000-0000-0000-0000-000000000991"
	_, err = db.Exec(
		`INSERT INTO beds (bed_id, tenant_id, room_id, bed_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (bed_id) DO UPDATE SET bed_name = EXCLUDED.bed_name`,
		bedID, tenantID, roomID, "Test Bed 001",
	)
	if err != nil {
		t.Fatalf("Failed to create test bed: %v", err)
	}

	return tenantID, unitID, bedID
}

// createTestResidentForCards 创建测试住户（使用随机 account 避免冲突）
func createTestResidentForCards(t *testing.T, db *sql.DB, tenantID, unitID string) string {
	return createTestResidentForCardsWithAccount(t, db, tenantID, unitID, "")
}

// createTestResidentForCardsWithAccount 创建测试住户（指定 account）
func createTestResidentForCardsWithAccount(t *testing.T, db *sql.DB, tenantID, unitID, accountSuffix string) string {
	var residentID string
	account := "test_resident_001"
	if accountSuffix != "" {
		account = "test_resident_" + accountSuffix
	}
	nickname := "Test Resident 001"
	if accountSuffix != "" {
		nickname = "Test Resident " + accountSuffix
	}
	accountHash := hashStringForCards(account)
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
	if err != nil {
		t.Fatalf("Failed to create test resident: %v", err)
	}
	return residentID
}

// createTestCard 创建测试卡片
func createTestCard(t *testing.T, db *sql.DB, tenantID, cardType, cardName, cardAddress string, bedID, unitID, residentID sql.NullString) string {
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
	if err != nil {
		t.Fatalf("Failed to create test card: %v", err)
	}
	return cardID
}

// cleanupTestDataForCards 清理测试数据
func cleanupTestDataForCards(t *testing.T, db *sql.DB, tenantID string) {
	db.Exec(`DELETE FROM cards WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM resident_caregivers WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM resident_contacts WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM resident_phi WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM residents WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM beds WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM rooms WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM units WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM buildings WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tags_catalog WHERE tenant_id = $1`, tenantID)
	db.Exec(`DELETE FROM tenants WHERE tenant_id = $1`, tenantID)
}

// TestPostgresCardsRepository_ListCards_Basic 测试基本的 ListCards 功能
func TestPostgresCardsRepository_ListCards_Basic(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, bedID := createTestTenantAndUnitForCards(t, db)
	defer cleanupTestDataForCards(t, db, tenantID)

	repo := NewPostgresCardsRepository(db)
	ctx := context.Background()

	// 创建测试卡片
	residentID := createTestResidentForCards(t, db, tenantID, unitID)
	cardID1 := createTestCard(t, db, tenantID, "ActiveBed", "Test Card 1", "Test Address 1",
		sql.NullString{String: bedID, Valid: true}, // bed_id (ActiveBed 卡片需要)
		sql.NullString{String: unitID, Valid: true},
		sql.NullString{String: residentID, Valid: true},
	)

	// 测试查询所有卡片
	req := ListCardsRequest{
		TenantID: tenantID,
	}
	cards, err := repo.ListCards(ctx, req)
	if err != nil {
		t.Fatalf("ListCards failed: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("Expected 1 card, got %d", len(cards))
	}
	if cards[0].Card.CardID != cardID1 {
		t.Errorf("Expected card_id %s, got %s", cardID1, cards[0].Card.CardID)
	}
	if cards[0].Card.CardType != "ActiveBed" {
		t.Errorf("Expected card_type 'ActiveBed', got '%s'", cards[0].Card.CardType)
	}
	if cards[0].Card.CardName != "Test Card 1" {
		t.Errorf("Expected card_name 'Test Card 1', got '%s'", cards[0].Card.CardName)
	}
}

// TestPostgresCardsRepository_ListCards_ByCardType 测试按卡片类型过滤
func TestPostgresCardsRepository_ListCards_ByCardType(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, bedID := createTestTenantAndUnitForCards(t, db)
	defer cleanupTestDataForCards(t, db, tenantID)

	repo := NewPostgresCardsRepository(db)
	ctx := context.Background()

	// 创建测试卡片
	residentID := createTestResidentForCards(t, db, tenantID, unitID)
	cardID1 := createTestCard(t, db, tenantID, "ActiveBed", "Test Card 1", "Test Address 1",
		sql.NullString{String: bedID, Valid: true}, // bed_id (ActiveBed 卡片需要)
		sql.NullString{String: unitID, Valid: true},
		sql.NullString{String: residentID, Valid: true},
	)
	// 注意：数据库 schema 中 card_type 可能还是 'Location'，需要根据实际情况调整
	cardID2 := createTestCard(t, db, tenantID, "Location", "Test Card 2", "Test Address 2",
		sql.NullString{String: "", Valid: false}, // bed_id (Location 卡片必须为 NULL)
		sql.NullString{String: unitID, Valid: true},
		sql.NullString{String: "", Valid: false},
	)

	// 测试查询 ActiveBed 卡片
	req := ListCardsRequest{
		TenantID: tenantID,
		CardType: "ActiveBed",
	}
	cards, err := repo.ListCards(ctx, req)
	if err != nil {
		t.Fatalf("ListCards failed: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("Expected 1 ActiveBed card, got %d", len(cards))
	}
	if cards[0].Card.CardID != cardID1 {
		t.Errorf("Expected card_id %s, got %s", cardID1, cards[0].Card.CardID)
	}

	// 测试查询 Location 卡片（数据库 schema 中使用 'Location'）
	req.CardType = "Location"
	cards, err = repo.ListCards(ctx, req)
	if err != nil {
		t.Fatalf("ListCards failed: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("Expected 1 Location card, got %d", len(cards))
	}
	if cards[0].Card.CardID != cardID2 {
		t.Errorf("Expected card_id %s, got %s", cardID2, cards[0].Card.CardID)
	}
}

// TestPostgresCardsRepository_ListCards_ByResidentPermission 测试 Resident 权限过滤
func TestPostgresCardsRepository_ListCards_ByResidentPermission(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, bedID := createTestTenantAndUnitForCards(t, db)
	defer cleanupTestDataForCards(t, db, tenantID)

	repo := NewPostgresCardsRepository(db)
	ctx := context.Background()

	// 创建测试住户和卡片（使用不同的 account 避免冲突）
	residentID1 := createTestResidentForCardsWithAccount(t, db, tenantID, unitID, "001")
	residentID2 := createTestResidentForCardsWithAccount(t, db, tenantID, unitID, "002")

	cardID1 := createTestCard(t, db, tenantID, "ActiveBed", "Test Card 1", "Test Address 1",
		sql.NullString{String: bedID, Valid: true}, // bed_id (ActiveBed 卡片需要)
		sql.NullString{String: unitID, Valid: true},
		sql.NullString{String: residentID1, Valid: true},
	)
	// 创建第二个 bed 用于第二个卡片
	bedID2 := "00000000-0000-0000-0000-000000000990"
	roomID := "00000000-0000-0000-0000-000000000992"
	_, _ = db.Exec(
		`INSERT INTO beds (bed_id, tenant_id, room_id, bed_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (bed_id) DO UPDATE SET bed_name = EXCLUDED.bed_name`,
		bedID2, tenantID, roomID, "Test Bed 002",
	)
	_ = createTestCard(t, db, tenantID, "ActiveBed", "Test Card 2", "Test Address 2",
		sql.NullString{String: bedID2, Valid: true}, // bed_id (ActiveBed 卡片需要)
		sql.NullString{String: unitID, Valid: true},
		sql.NullString{String: residentID2, Valid: true},
	)

	// 测试 Resident 权限过滤（只能看到自己的卡片）
	req := ListCardsRequest{
		TenantID: tenantID,
		PermissionFilter: &PermissionFilter{
			UserID:   residentID1,
			UserType: "resident",
		},
	}
	cards, err := repo.ListCards(ctx, req)
	if err != nil {
		t.Fatalf("ListCards failed: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("Expected 1 card, got %d", len(cards))
	}
	if cards[0].Card.CardID != cardID1 {
		t.Errorf("Expected card_id %s, got %s", cardID1, cards[0].Card.CardID)
	}
}

// TestPostgresCardsRepository_ListCards_BySearch 测试搜索功能
func TestPostgresCardsRepository_ListCards_BySearch(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, bedID := createTestTenantAndUnitForCards(t, db)
	defer cleanupTestDataForCards(t, db, tenantID)

	repo := NewPostgresCardsRepository(db)
	ctx := context.Background()

	// 创建测试卡片
	residentID := createTestResidentForCards(t, db, tenantID, unitID)
	cardID1 := createTestCard(t, db, tenantID, "ActiveBed", "Test Card 1", "Test Address 1",
		sql.NullString{String: bedID, Valid: true}, // bed_id (ActiveBed 卡片需要)
		sql.NullString{String: unitID, Valid: true},
		sql.NullString{String: residentID, Valid: true},
	)
	// 创建第二个 bed 用于第二个卡片
	bedID2 := "00000000-0000-0000-0000-000000000990"
	roomID := "00000000-0000-0000-0000-000000000992"
	_, _ = db.Exec(
		`INSERT INTO beds (bed_id, tenant_id, room_id, bed_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (bed_id) DO UPDATE SET bed_name = EXCLUDED.bed_name`,
		bedID2, tenantID, roomID, "Test Bed 002",
	)
	_ = createTestCard(t, db, tenantID, "ActiveBed", "Other Card", "Other Address",
		sql.NullString{String: bedID2, Valid: true}, // bed_id (ActiveBed 卡片需要)
		sql.NullString{String: unitID, Valid: true},
		sql.NullString{String: residentID, Valid: true},
	)

	// 测试搜索
	req := ListCardsRequest{
		TenantID: tenantID,
		Search:   "Test Card",
	}
	cards, err := repo.ListCards(ctx, req)
	if err != nil {
		t.Fatalf("ListCards failed: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("Expected 1 card, got %d", len(cards))
	}
	if cards[0].Card.CardID != cardID1 {
		t.Errorf("Expected card_id %s, got %s", cardID1, cards[0].Card.CardID)
	}
}

// TestPostgresCardsRepository_ListCards_ByCardID 测试按卡片ID查询
func TestPostgresCardsRepository_ListCards_ByCardID(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	tenantID, unitID, bedID := createTestTenantAndUnitForCards(t, db)
	defer cleanupTestDataForCards(t, db, tenantID)

	repo := NewPostgresCardsRepository(db)
	ctx := context.Background()

	// 创建测试卡片
	residentID := createTestResidentForCards(t, db, tenantID, unitID)
	cardID1 := createTestCard(t, db, tenantID, "ActiveBed", "Test Card 1", "Test Address 1",
		sql.NullString{String: bedID, Valid: true}, // bed_id (ActiveBed 卡片需要)
		sql.NullString{String: unitID, Valid: true},
		sql.NullString{String: residentID, Valid: true},
	)

	// 测试按卡片ID查询
	req := ListCardsRequest{
		TenantID: tenantID,
		CardID:   cardID1,
	}
	cards, err := repo.ListCards(ctx, req)
	if err != nil {
		t.Fatalf("ListCards failed: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("Expected 1 card, got %d", len(cards))
	}
	if cards[0].Card.CardID != cardID1 {
		t.Errorf("Expected card_id %s, got %s", cardID1, cards[0].Card.CardID)
	}
}

