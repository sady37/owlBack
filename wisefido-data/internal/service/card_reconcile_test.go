// card_reconcile_test.go — ReconcileCards 集成测试套
//
// 跑法：
//   PGPASSWORD=postgres go test -v ./internal/service/ -run TestReconcile -count=1
//
// 测试隔离：用 fd00:0:99::/48 测试 tenant（demo 数据在 fd00:0:3，互不干扰）。
// 每个 testcase 自己 cleanTestTenant 重置；不依赖跑顺序。
//
// 覆盖度：A (规则) / B (split↔merge 切换) / C (LPM) / D (Repository hook) /
//        E (CloudEvent emit) / F (幂等) / G (edge cases)。共 ~38 cases。

package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"testing"

	"wisefido-data/internal/domain"
	"wisefido-data/internal/repository"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// =========================================================================
// 测试 tenant 拓扑（一次性建好骨架，所有测试共用）
//
//   Tenant   fd00:0:99::/48
//   Branch   fd00:0:99:100::/56
//   Site     fd00:0:99:101::/64    site_name='TestBldg'
//   Units:
//     Private  fd00:0:99:101:65::/80   unit_type=1
//       Bedroom    fd00:0:99:101:65:100::/88
//         BedA     fd00:0:99:101:65:101::/96
//         BedB     fd00:0:99:101:65:102::/96
//       Bathroom   fd00:0:99:101:65:200::/88
//     Share    fd00:0:99:101:c9::/80   unit_type=2
//       Bedroom    fd00:0:99:101:c9:100::/88
//         BedA     fd00:0:99:101:c9:101::/96
//         BedB     fd00:0:99:101:c9:102::/96
//     Public   fd00:0:99:101:1::/80    unit_type=3
//       LivingRoom fd00:0:99:101:1:100::/88
// =========================================================================

const (
	tTenant = "fd00:0:99::/48"
	tBranch = "fd00:0:99:100::/56"
	tSite   = "fd00:0:99:101::/64"

	tUnitPrivate  = "fd00:0:99:101:65::/80"
	tUnitShare    = "fd00:0:99:101:c9::/80"
	tUnitPublic   = "fd00:0:99:101:1::/80"

	tRoomPrivBedroom  = "fd00:0:99:101:65:100::/88"
	tRoomPrivBathroom = "fd00:0:99:101:65:200::/88"
	tBedPrivA         = "fd00:0:99:101:65:101::/96"
	tBedPrivB         = "fd00:0:99:101:65:102::/96"

	tRoomShareBedroom = "fd00:0:99:101:c9:100::/88"
	tBedShareA        = "fd00:0:99:101:c9:101::/96"
	tBedShareB        = "fd00:0:99:101:c9:102::/96"

	tRoomPublicLiv = "fd00:0:99:101:1:100::/88"

	// resident HoAs (slot 编 0xff01:N)
	tRes1 = "fd00:0:99:ff01:1::/128"
	tRes2 = "fd00:0:99:ff01:2::/128"
	tRes3 = "fd00:0:99:ff01:3::/128"
)

var (
	tDB       *sql.DB
	tCardSync *CardSyncService
)

// =========================================================================
// TestMain — 建测试 tenant 骨架，跑所有测试，最后清理
// =========================================================================

func TestMain(m *testing.M) {
	var err error
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/owl_v2?sslmode=disable"
	}
	tDB, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(1)
	}
	defer tDB.Close()

	if err := tDB.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "ping db: %v — skip card reconcile tests\n", err)
		os.Exit(0)
	}

	if err := setupTestTopology(); err != nil {
		fmt.Fprintf(os.Stderr, "setup topology: %v\n", err)
		os.Exit(1)
	}

	cardRepo := repository.NewPostgresCardRepository(tDB, zap.NewNop())
	tCardSync = NewCardSyncService(cardRepo, nil /* publisher=nil → 跳过 CloudEvent；E 组用 observer */, nil, zap.NewNop())
	tCardSync.SetReconcileDeps(tDB, nil, "owl.")

	code := m.Run()

	teardownTestTenant()
	os.Exit(code)
}

// setupTestTopology — 建一次测试 tenant + branch + units/rooms/beds 骨架；幂等
func setupTestTopology() error {
	exec := func(q string, args ...interface{}) error {
		if _, err := tDB.Exec(q, args...); err != nil {
			return fmt.Errorf("setup: %w (q=%q)", err, q)
		}
		return nil
	}
	if err := exec(`INSERT INTO tenants (tenant_id, tenant_slot, tenant_name, status, kind) VALUES ($1::INET, 99, 'TestTenant', 'active', 'B2B') ON CONFLICT (tenant_id) DO NOTHING`, tTenant); err != nil {
		return err
	}
	if err := exec(`INSERT INTO branches (branch_id, branch_slot, branch_name, timezone) VALUES ($1::INET, 1, 'TestBranch', 'UTC') ON CONFLICT (branch_id) DO NOTHING`, tBranch); err != nil {
		return err
	}
	// sites_slot_consistent CHECK: site_slot = (building<<4) | floor → 1<<4|1=17
	if err := exec(`INSERT INTO sites (site_id, site_slot, building, floor, site_name) VALUES ($1::INET, 17, 1, 1, 'TestBldg') ON CONFLICT (site_id) DO NOTHING`, tSite); err != nil {
		return err
	}
	if err := exec(`INSERT INTO units (unit_id, unit_slot, unit_name, unit_type, unit_property) VALUES
		 ($1::INET, 101, 'TPrivate', 1, 1),
		 ($2::INET, 201, 'TShare',   2, 1),
		 ($3::INET, 1,   'TPublic',  3, 1) ON CONFLICT (unit_id) DO NOTHING`, tUnitPrivate, tUnitShare, tUnitPublic); err != nil {
		return err
	}
	if err := exec(`INSERT INTO rooms (room_id, room_slot, room_name, room_type, is_primary) VALUES
		 ($1::INET, 1, 'PrivBedroom',  'bedroom',    TRUE),
		 ($2::INET, 2, 'PrivBathroom', 'bathroom',   FALSE),
		 ($3::INET, 1, 'ShareBedroom', 'bedroom',    TRUE),
		 ($4::INET, 1, 'PubLiv',       'livingroom', FALSE) ON CONFLICT (room_id) DO NOTHING`, tRoomPrivBedroom, tRoomPrivBathroom, tRoomShareBedroom, tRoomPublicLiv); err != nil {
		return err
	}
	if err := exec(`INSERT INTO beds (bed_id, bed_slot, bed_name) VALUES
		 ($1::INET, 1, 'PrivBedA'),
		 ($2::INET, 2, 'PrivBedB'),
		 ($3::INET, 1, 'ShareBedA'),
		 ($4::INET, 2, 'ShareBedB') ON CONFLICT (bed_id) DO NOTHING`, tBedPrivA, tBedPrivB, tBedShareA, tBedShareB); err != nil {
		return err
	}
	if err := exec(`INSERT INTO residents (resident_id, resident_slot, nickname, resident_account, status) VALUES
		 ($1::INET, 1, 'TR1', 'TR001', 'active'),
		 ($2::INET, 2, 'TR2', 'TR002', 'active'),
		 ($3::INET, 3, 'TR3', 'TR003', 'active') ON CONFLICT (resident_id) DO NOTHING`, tRes1, tRes2, tRes3); err != nil {
		return err
	}
	return nil
}

// teardownTestTenant — 测试结束清理（保留 schema 骨架以便 -count=N 多次跑）
// 只清 mutating 表：cards / devices / device_factory_meta / resident_unit；residents/units/rooms/beds 骨架保留
func teardownTestTenant() {
	_, _ = tDB.Exec(`DELETE FROM cards WHERE card_id <<= $1::INET`, tTenant)
	_, _ = tDB.Exec(`DELETE FROM resident_unit WHERE resident_id <<= $1::INET`, tTenant)
	// devices.device_id 是 FK 到 device_factory_meta；删 factory_meta 触发 CASCADE 删 devices
	_, _ = tDB.Exec(`DELETE FROM device_factory_meta WHERE device_id IN (SELECT device_id FROM devices WHERE device_ipv6 <<= $1::INET)`, tTenant)
}

// resetTestTenantState — 每个 test case 开始前调，清 mutating state
func resetTestTenantState(t *testing.T) {
	t.Helper()
	teardownTestTenant()
}

// =========================================================================
// Helpers — DB 写入 / 查询
// =========================================================================

// seedDevice — 插一台 device 到指定 spatial_prefix（取 prefix 网络部分 + 随机 host 32 bit）
// 先插 device_factory_meta（FK 要求），再插 devices
func seedDevice(t *testing.T, spatialPrefix string, monitoringEnabled bool) string {
	t.Helper()
	var addr string
	if err := tDB.QueryRow(`
		SELECT host(
			set_masklen(network(set_masklen($1::INET, 96)), 128)
			| ('::' || lpad(to_hex((random()*16777215)::int), 4, '0') || ':' || lpad(to_hex((random()*65535)::int), 4, '0'))::INET
		)
	`, spatialPrefix).Scan(&addr); err != nil {
		t.Fatalf("seedDevice: derive addr from %s: %v", spatialPrefix, err)
	}
	deviceIPv6 := addr + "/128"
	// device_factory_meta 先建（FK ON DELETE CASCADE 保证清理时一起删）
	var deviceID string
	if err := tDB.QueryRow(`
		INSERT INTO device_factory_meta (device_id, device_uid, device_type, import_date)
		VALUES (gen_random_uuid(), 'TEST-' || substr(md5(random()::text), 1, 10), 'Radar', NOW())
		RETURNING device_id::text
	`).Scan(&deviceID); err != nil {
		t.Fatalf("seedDevice: factory_meta: %v", err)
	}
	if _, err := tDB.Exec(`
		INSERT INTO devices (device_id, device_ipv6, monitoring_enabled, access)
		VALUES ($1::UUID, $2::INET, $3, TRUE)
	`, deviceID, deviceIPv6, monitoringEnabled); err != nil {
		t.Fatalf("seedDevice: insert devices: %v", err)
	}
	return deviceID
}

// seedResidentUnit — 给指定 resident 插一行 active resident_unit
func seedResidentUnit(t *testing.T, residentID, spatialPrefix string) {
	t.Helper()
	if _, err := tDB.Exec(`
		INSERT INTO resident_unit (resident_id, spatial_prefix, move_reason)
		VALUES ($1::INET, $2::INET, 'test_seed')
	`, residentID, spatialPrefix); err != nil {
		t.Fatalf("seedResidentUnit: %v", err)
	}
}

// cardSnapshot — 查 scope 内 cards 表，按 spatial_prefix 排序返回
type cardRow struct {
	Prefix     string
	CardType   string
	CardName   string
	ResidentID string // host 格式（无 /128 后缀），空表示 NULL
}

func cardsInScope(t *testing.T, scope string) []cardRow {
	t.Helper()
	rows, err := tDB.Query(`
		SELECT card_id::text,
		       CASE masklen(card_id)
		           WHEN  48 THEN 'tenant'
		           WHEN  56 THEN 'branch'
		           WHEN  64 THEN 'site'
		           WHEN  80 THEN CASE WHEN card_name = 'public' THEN 'public' ELSE 'unit' END
		           WHEN  88 THEN 'room'
		           WHEN  96 THEN 'bed'
		           WHEN 128 THEN 'device'
		       END AS card_type,
		       COALESCE(card_name, ''),
		       COALESCE(host(resident_id), '')
		  FROM cards
		 WHERE card_id <<= $1::INET
		 ORDER BY card_id
	`, scope)
	if err != nil {
		t.Fatalf("cardsInScope: %v", err)
	}
	defer rows.Close()
	out := []cardRow{}
	for rows.Next() {
		var r cardRow
		if err := rows.Scan(&r.Prefix, &r.CardType, &r.CardName, &r.ResidentID); err != nil {
			t.Fatalf("cardsInScope scan: %v", err)
		}
		out = append(out, r)
	}
	return out
}

// assertCards — 简洁断言：scope 内 cards 应等于 expectedPrefixes（顺序无关，仅看 prefix 集合）
func assertCards(t *testing.T, scope string, expectedPrefixes ...string) []cardRow {
	t.Helper()
	cards := cardsInScope(t, scope)
	got := make([]string, len(cards))
	for i, c := range cards {
		got[i] = c.Prefix
	}
	sort.Strings(got)
	want := make([]string, len(expectedPrefixes))
	copy(want, expectedPrefixes)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Errorf("cards count mismatch in %s: got %d %v, want %d %v", scope, len(got), got, len(want), want)
		return cards
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("cards mismatch in %s: got %v, want %v", scope, got, want)
			return cards
		}
	}
	return cards
}

// findCard — 在 snapshot 里找指定 prefix 的 cardRow（fail if 找不到）
func findCard(t *testing.T, cards []cardRow, prefix string) cardRow {
	t.Helper()
	for _, c := range cards {
		if c.Prefix == prefix {
			return c
		}
	}
	t.Fatalf("findCard: prefix %s not in snapshot %v", prefix, cards)
	return cardRow{}
}

// reconcile — 调 ReconcileCards 并 fail-on-error 包装
func reconcile(t *testing.T, scope string) {
	t.Helper()
	if err := tCardSync.ReconcileCards(context.Background(), scope); err != nil {
		t.Fatalf("ReconcileCards(%s): %v", scope, err)
	}
}

// =========================================================================
// 组 A：ReconcileCards 核心规则
// =========================================================================

func TestReconcile_A1_EmptyScope(t *testing.T) {
	resetTestTenantState(t)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate) // 期望 0 张
}

func TestReconcile_A2_ShareUnitOneBedMerge(t *testing.T) {
	resetTestTenantState(t)
	// v3 统一规则: Share unit 1 bed device → merge mode → /80 unit card
	// （v2 中 Share 永远 split 的特殊行为已取消，Share / Private 统一走 activeBed>1 才 split）
	seedDevice(t, tBedShareA, true)
	reconcile(t, tUnitShare)
	assertCards(t, tUnitShare, tUnitShare)
}

func TestReconcile_A3_PrivateOneBedMerge(t *testing.T) {
	resetTestTenantState(t)
	// Private unit + cards 空 + 1 个 device 装在 bed → merge: 只 1 张 /80
	seedDevice(t, tBedPrivA, true)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tUnitPrivate)
}

func TestReconcile_A4_PrivateTwoBedSplit(t *testing.T) {
	resetTestTenantState(t)
	// v3: 2 bed device 同 room (PrivBedroom 内 active_bed=2) + 无 room/unit device
	// → split mode 下 L1 unit card 不出（无 unit-level device），L2 /88 也不出（room device 也没）
	// → 仅 2 张 /96
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedPrivB, true)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tBedPrivA, tBedPrivB)
}

func TestReconcile_A5_PrivateRoomDeviceOnly(t *testing.T) {
	resetTestTenantState(t)
	// Private + 1 device 装在 room (/88 没 bed 子级) → 1 张 /80（升到 unit）
	seedDevice(t, tRoomPrivBathroom, true)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tUnitPrivate)
}

func TestReconcile_A6_PublicRoomDeviceOnly(t *testing.T) {
	resetTestTenantState(t)
	// v3: bedCount=0 + 1 room-level device → single mode → /80 unit card 装该 device
	// (套娃只在 raw_active_bed ≥ 1 时启用；raw=0 一律 /80 兜底)
	seedDevice(t, tRoomPublicLiv, true)
	reconcile(t, tUnitPublic)
	assertCards(t, tUnitPublic, tUnitPublic)
}

func TestReconcile_A7_MonitoringDisabled(t *testing.T) {
	resetTestTenantState(t)
	// device 装在 bed 但 monitoring=false → 不进 expected，无卡
	seedDevice(t, tBedPrivA, false)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate) // 0 张
}

func TestReconcile_A8_PrivateOneBedPlusBathroom(t *testing.T) {
	resetTestTenantState(t)
	// Private + 1 bed device + 1 bathroom radar (room 级)
	// activeBed=1 → merge mode → 1 张 /80（bed + bathroom 都隐含其中）
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tRoomPrivBathroom, true)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tUnitPrivate)
}

// =========================================================================
// 组 B：split↔merge 双向切换（stateless rule：activeBed>1 split / ≤1 merge，无 latching）
// =========================================================================

func TestReconcile_B1_MergeToSplit(t *testing.T) {
	resetTestTenantState(t)
	// 1 bed device + 无 room/unit device → merge: 1 张 /80
	seedDevice(t, tBedPrivA, true)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tUnitPrivate)

	// 加第 2 bed device → 触发 split; 无 room/unit device → 仅 2 张 /96（/80 消失因无 unit-level device）
	seedDevice(t, tBedPrivB, true)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tBedPrivA, tBedPrivB)
}

func TestReconcile_B2_SplitMinusOneFallsBackToMerge(t *testing.T) {
	resetTestTenantState(t)
	// 2 bed → split (2 张 /96)
	d1 := seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedPrivB, true)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tBedPrivA, tBedPrivB)

	// 删 1 bed device → 剩 activeBed=1 → 无单向阀 latching ([[card_split_rule_v3]])
	// → 回退到 MERGE 模式 → 仅 /80 unit card
	if _, err := tDB.Exec(`DELETE FROM devices WHERE device_id=$1::UUID`, d1); err != nil {
		t.Fatal(err)
	}
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tUnitPrivate)
}

func TestReconcile_B3_SplitMinusAllBeds(t *testing.T) {
	resetTestTenantState(t)
	d1 := seedDevice(t, tBedPrivA, true)
	d2 := seedDevice(t, tBedPrivB, true)
	reconcile(t, tUnitPrivate)
	// 删两个 bed device → cards 表当前还有 1 张 /96 还在（reconcile 前），先删 device 再 reconcile
	_, _ = tDB.Exec(`DELETE FROM devices WHERE device_id IN ($1::UUID, $2::UUID)`, d1, d2)
	reconcile(t, tUnitPrivate)
	// 0 active bed → 无卡
	assertCards(t, tUnitPrivate)
}

func TestReconcile_B4_SplitAddRoomDevice(t *testing.T) {
	resetTestTenantState(t)
	// 2 bed → split (2 张 /96)
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedPrivB, true)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tBedPrivA, tBedPrivB)

	// 加 bathroom radar (PrivBathroom 0 active bed) → /88 room card 出现
	// PrivBedroom 仍 active_bed=2 → 不出 /88（无 room device）
	// 无 unit-level device → 不出 /80
	seedDevice(t, tRoomPrivBathroom, true)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tBedPrivA, tBedPrivB, tRoomPrivBathroom)
}

// =========================================================================
// 组 C：LPM resident_id 双向 overlap
// =========================================================================

func TestReconcile_C1_ResidentInUnit80(t *testing.T) {
	resetTestTenantState(t)
	// 2 bed device + bathroom radar → split mode 下:
	//   2 张 /96 (bed) + 1 张 /88 (bathroom, 0 active bed in PrivBathroom)
	//   PrivBedroom 内 active_bed=2 → 无 room device 不出 /88
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedPrivB, true)
	seedDevice(t, tRoomPrivBathroom, true)
	seedResidentUnit(t, tRes1, tUnitPrivate) // resident 在 /80 整 unit
	reconcile(t, tUnitPrivate)
	cards := assertCards(t, tUnitPrivate, tBedPrivA, tBedPrivB, tRoomPrivBathroom)
	// 所有 cards 应 overlay resident = TR1（resident 在 unit /80，LPM 双向 overlap 找到所有 cards）
	for _, c := range cards {
		if c.ResidentID != "fd00:0:99:ff01:1::" {
			t.Errorf("card %s expected resident TR1, got rid=%q name=%q", c.Prefix, c.ResidentID, c.CardName)
		}
	}
}

func TestReconcile_C2_ResidentInBed96(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedPrivB, true)
	seedResidentUnit(t, tRes1, tBedPrivA) // resident 绑具体 bed，但 unit 内唯一
	reconcile(t, tUnitPrivate)
	cards := assertCards(t, tUnitPrivate, tBedPrivA, tBedPrivB)
	// v3 unit fallback: unit 内唯一 resident → 所有该 unit 内 cards overlay 该 resident
	// 体现 "Private unit 整 unit 归一人" 语义（不需 unit_type 分支）
	for _, c := range cards {
		if c.ResidentID != "fd00:0:99:ff01:1::" {
			t.Errorf("%s expected TR1 (unit fallback), got %q", c.Prefix, c.ResidentID)
		}
	}
}

func TestReconcile_C2b_ShareUnitNoUnitFallback(t *testing.T) {
	resetTestTenantState(t)
	// Share unit 多 resident → fallback 不触发（cnt != 1），各 bed card 各自命中
	seedDevice(t, tBedShareA, true)
	seedDevice(t, tBedShareB, true)
	seedResidentUnit(t, tRes1, tBedShareA)
	seedResidentUnit(t, tRes2, tBedShareB)
	reconcile(t, tUnitShare)
	cards := assertCards(t, tUnitShare, tBedShareA, tBedShareB)
	if c := findCard(t, cards, tBedShareA); c.ResidentID != "fd00:0:99:ff01:1::" {
		t.Errorf("ShareBedA expected TR1, got %q", c.ResidentID)
	}
	if c := findCard(t, cards, tBedShareB); c.ResidentID != "fd00:0:99:ff01:2::" {
		t.Errorf("ShareBedB expected TR2, got %q", c.ResidentID)
	}
}

func TestReconcile_C3_ResidentInRoom88(t *testing.T) {
	resetTestTenantState(t)
	// v3: room device + 0 bed → single mode /80 unit card（不再出 /88）
	seedDevice(t, tRoomPublicLiv, true)
	seedResidentUnit(t, tRes1, tRoomPublicLiv)
	reconcile(t, tUnitPublic)
	cards := assertCards(t, tUnitPublic, tUnitPublic)
	if c := findCard(t, cards, tUnitPublic); c.ResidentID != "fd00:0:99:ff01:1::" {
		t.Errorf("unit card expected TR1 (LPM 双向找到 /88 子级 resident), got %q", c.ResidentID)
	}
}

func TestReconcile_C4_ResidentInBranch56(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedPrivA, true) // 1 bed device → merge mode → /80 unit card
	seedResidentUnit(t, tRes1, tBranch)
	reconcile(t, tUnitPrivate)
	cards := assertCards(t, tUnitPrivate, tUnitPrivate)
	if c := findCard(t, cards, tUnitPrivate); c.ResidentID != "" {
		t.Errorf("expected NoOne (branch /56 不算 unit 占用), got %q", c.ResidentID)
	}
}

func TestReconcile_C5_ShareUnitMultipleResidents(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedShareA, true)
	seedDevice(t, tBedShareB, true)
	seedResidentUnit(t, tRes1, tBedShareA)
	seedResidentUnit(t, tRes2, tBedShareB)
	reconcile(t, tUnitShare)
	cards := assertCards(t, tUnitShare, tBedShareA, tBedShareB)
	if c := findCard(t, cards, tBedShareA); c.ResidentID != "fd00:0:99:ff01:1::" {
		t.Errorf("ShareBedA expected TR1, got %q", c.ResidentID)
	}
	if c := findCard(t, cards, tBedShareB); c.ResidentID != "fd00:0:99:ff01:2::" {
		t.Errorf("ShareBedB expected TR2, got %q", c.ResidentID)
	}
}

// =========================================================================
// 组 D：Repository hook 自动触发
// =========================================================================

func TestReconcile_D1_ResidentsRepoCreate(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedPrivB, true)

	// 装 ResidentsRepo + hook
	residentsRepo := repository.NewPostgresResidentsRepository(tDB)
	hookCalls := []string{}
	residentsRepo.SetOnResidentUnitChange(func(_ context.Context, scope string) {
		hookCalls = append(hookCalls, scope)
		_ = tCardSync.ReconcileCards(context.Background(), scope)
	})

	// 用 repo 直接调 CreateResident（绕过 service，因为 ResidentService.Create 有 branch 校验需要 scope ctx）
	unitID := tUnitPrivate
	in := &domain.ResidentCreateInput{
		Nickname: "D1Resident",
		UnitID:   &unitID,
	}
	_, err := residentsRepo.CreateResident(context.Background(), tTenant, in, "00000000-0000-0000-0000-000000000000", "Admin")
	if err != nil {
		t.Fatalf("CreateResident: %v", err)
	}
	if len(hookCalls) != 1 {
		t.Errorf("hook should fire once for create, got %d times", len(hookCalls))
	}
	if len(hookCalls) > 0 && hookCalls[0] != tUnitPrivate {
		t.Errorf("hook scope expected %s, got %s", tUnitPrivate, hookCalls[0])
	}
	// v3: 2 bed device + 无 room/unit device → 仅 2 张 /96（不出 /80）
	cards := assertCards(t, tUnitPrivate, tBedPrivA, tBedPrivB)
	// 两张 /96 都应有 resident_id（resident 在 /80 unit，LPM 双向 overlap 找到子级 bed cards）
	if c := findCard(t, cards, tBedPrivA); c.ResidentID == "" {
		t.Errorf("BedA should have resident_id after create, got NoOne")
	}
	_, _ = tDB.Exec(`DELETE FROM residents WHERE nickname='D1Resident'`)
}

func TestReconcile_D2_ResidentsRepoUpdateMoveUnit(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedShareA, true)
	// 老挂 Private
	seedResidentUnit(t, tRes1, tBedPrivA)
	reconcile(t, tUnitPrivate)
	reconcile(t, tUnitShare)
	assertCards(t, tUnitPrivate, tUnitPrivate) // Private + 1 bed = merge
	if c := findCard(t, cardsInScope(t, tUnitPrivate), tUnitPrivate); c.ResidentID != "fd00:0:99:ff01:1::" {
		t.Errorf("before move: Private unit should have TR1")
	}

	// 装 hook
	residentsRepo := repository.NewPostgresResidentsRepository(tDB)
	hookCalls := []string{}
	residentsRepo.SetOnResidentUnitChange(func(_ context.Context, scope string) {
		hookCalls = append(hookCalls, scope)
		_ = tCardSync.ReconcileCards(context.Background(), scope)
	})

	// Move resident 1 from Private → Share BedA
	newBed := tBedShareA
	in := &domain.ResidentUpdateInput{BedID: &newBed}
	if err := residentsRepo.UpdateResident(context.Background(), tTenant, "fd00:0:99:ff01:1::", in, "00000000-0000-0000-0000-000000000000", "Admin"); err != nil {
		t.Fatalf("UpdateResident: %v", err)
	}
	if len(hookCalls) < 2 {
		t.Errorf("UpdateResident hook should fire 2x (old+new scope), got %d: %v", len(hookCalls), hookCalls)
	}
	// v3: Private unit activeBed=1（bed device 还在）→ merge mode → /80 NoOne（resident 已搬走）
	if c := findCard(t, cardsInScope(t, tUnitPrivate), tUnitPrivate); c.ResidentID != "" {
		t.Errorf("after move: Private /80 should be NoOne, got %q", c.ResidentID)
	}
	// v3: Share unit 也走统一规则 — 1 bed device + 无其他 → /80 unit card (merge)
	if c := findCard(t, cardsInScope(t, tUnitShare), tUnitShare); c.ResidentID != "fd00:0:99:ff01:1::" {
		t.Errorf("after move: Share /80 should overlay TR1 (LPM 找子级 /96), got %q", c.ResidentID)
	}
}

func TestReconcile_D3_ResidentsRepoSoftDelete(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedPrivA, true)
	seedResidentUnit(t, tRes1, tBedPrivA)
	reconcile(t, tUnitPrivate)

	residentsRepo := repository.NewPostgresResidentsRepository(tDB)
	residentsRepo.SetOnResidentUnitChange(func(_ context.Context, scope string) {
		_ = tCardSync.ReconcileCards(context.Background(), scope)
	})

	if err := residentsRepo.SoftDelete(context.Background(), "fd00:0:99:ff01:1::"); err != nil {
		t.Fatal(err)
	}
	if c := findCard(t, cardsInScope(t, tUnitPrivate), tUnitPrivate); c.ResidentID != "" {
		t.Errorf("after soft delete: /80 should be NoOne, got %q", c.ResidentID)
	}
	// 还原（避免污染后续）
	_, _ = tDB.Exec(`UPDATE residents SET status='active' WHERE resident_id=$1::INET`, tRes1)
}

func TestReconcile_D5_DevicesRepoMonitoringToggle(t *testing.T) {
	resetTestTenantState(t)
	devID := seedDevice(t, tBedPrivA, true)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tUnitPrivate)

	devicesRepo := repository.NewPostgresDevicesRepository(tDB)
	hookCalls := []string{}
	devicesRepo.SetOnDeviceChange(func(_ context.Context, scope string) {
		hookCalls = append(hookCalls, scope)
		_ = tCardSync.ReconcileCards(context.Background(), scope)
	})

	// 关 monitoring → cards 应消失（device 不算 active）
	dev := &domain.Device{MonitoringEnabled: false, Access: true}
	if err := devicesRepo.UpdateDeviceWithFlags(context.Background(), tTenant, devID, dev,
		false, false, false, true, false); err != nil {
		t.Fatal(err)
	}
	if len(hookCalls) == 0 {
		t.Errorf("UpdateDevice hook should fire, got 0 calls")
	}
	assertCards(t, tUnitPrivate) // 0 张
}

// =========================================================================
// 组 E：CloudEvent emit（用 observer 拦截 diffs，不走真实 publisher）
// =========================================================================

// captureDiffs — 设置 observer，返回拦截到的 diffs 列表（按调用顺序累加）
func captureDiffs(t *testing.T) *[]cardDiff {
	t.Helper()
	captured := []cardDiff{}
	tCardSync.SetReconcileObserver(func(_ string, diffs []cardDiff) {
		captured = append(captured, diffs...)
	})
	t.Cleanup(func() { tCardSync.SetReconcileObserver(nil) })
	return &captured
}

func countDiffOp(diffs []cardDiff, op string) int {
	n := 0
	for _, d := range diffs {
		if d.op == op {
			n++
		}
	}
	return n
}

func TestReconcile_E1_CreateEmit(t *testing.T) {
	resetTestTenantState(t)
	got := captureDiffs(t)
	seedDevice(t, tBedShareA, true)
	reconcile(t, tUnitShare) // 新建 /96 卡 (NoOne)
	if countDiffOp(*got, "create") != 1 {
		t.Errorf("expect 1 create, got %d (all=%v)", countDiffOp(*got, "create"), *got)
	}
}

func TestReconcile_E2_DeleteEmit(t *testing.T) {
	resetTestTenantState(t)
	d := seedDevice(t, tBedShareA, true)
	reconcile(t, tUnitShare) // 先建卡
	got := captureDiffs(t)
	_, _ = tDB.Exec(`DELETE FROM devices WHERE device_id=$1::UUID`, d)
	reconcile(t, tUnitShare) // 删卡
	if countDiffOp(*got, "delete") != 1 {
		t.Errorf("expect 1 delete, got %d (all=%v)", countDiffOp(*got, "delete"), *got)
	}
}

func TestReconcile_E3_AdmissionEmit(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedShareA, true)
	reconcile(t, tUnitShare) // 先建空卡 NoOne
	got := captureDiffs(t)
	seedResidentUnit(t, tRes1, tBedShareA)
	reconcile(t, tUnitShare) // resident 入住
	if countDiffOp(*got, "admission") != 1 {
		t.Errorf("expect 1 admission, got %d (all=%v)", countDiffOp(*got, "admission"), *got)
	}
}

func TestReconcile_E4_DischargeEmit(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedShareA, true)
	seedResidentUnit(t, tRes1, tBedShareA)
	reconcile(t, tUnitShare) // 先 resident 占着
	got := captureDiffs(t)
	_, _ = tDB.Exec(`UPDATE resident_unit SET valid_to=NOW() WHERE resident_id=$1::INET AND valid_to IS NULL`, tRes1)
	reconcile(t, tUnitShare) // resident 离开
	if countDiffOp(*got, "discharge") != 1 {
		t.Errorf("expect 1 discharge, got %d (all=%v)", countDiffOp(*got, "discharge"), *got)
	}
}

func TestReconcile_E5_TransferEmit(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedShareA, true)
	seedResidentUnit(t, tRes1, tBedShareA)
	reconcile(t, tUnitShare) // BedA: TR1
	got := captureDiffs(t)
	// TR1 走 → TR2 顶上同一张床
	_, _ = tDB.Exec(`UPDATE resident_unit SET valid_to=NOW() WHERE resident_id=$1::INET AND valid_to IS NULL`, tRes1)
	seedResidentUnit(t, tRes2, tBedShareA)
	reconcile(t, tUnitShare)
	if countDiffOp(*got, "transfer") != 1 {
		t.Errorf("expect 1 transfer, got %d (all=%v)", countDiffOp(*got, "transfer"), *got)
	}
}

func TestReconcile_E6_NameChangedEmit(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedShareA, true)
	seedResidentUnit(t, tRes1, tBedShareA)
	reconcile(t, tUnitShare) // BedA: TR1
	got := captureDiffs(t)
	// 改 nickname；resident_id 不变
	_, _ = tDB.Exec(`UPDATE residents SET nickname='TR1Renamed' WHERE resident_id=$1::INET`, tRes1)
	reconcile(t, tUnitShare)
	if countDiffOp(*got, "name_changed") != 1 {
		t.Errorf("expect 1 name_changed, got %d (all=%v)", countDiffOp(*got, "name_changed"), *got)
	}
	// 还原
	_, _ = tDB.Exec(`UPDATE residents SET nickname='TR1' WHERE resident_id=$1::INET`, tRes1)
}

func TestReconcile_E7_NoEmitOnIdempotent(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedShareA, true)
	reconcile(t, tUnitShare)
	got := captureDiffs(t)
	reconcile(t, tUnitShare) // 完全幂等
	if len(*got) != 0 {
		t.Errorf("expect 0 diffs on idempotent reconcile, got %v", *got)
	}
}

func TestReconcile_E8_MultipleDiffsOneReconcile(t *testing.T) {
	resetTestTenantState(t)
	got := captureDiffs(t)
	// 一次 reconcile 产生多个 diff：2 张新 card 各带 resident → 2 admission（computeCardDiff
	// 对"新卡 + 有 resident"直接 emit admission 单 op，不分别 emit create+admission）
	seedDevice(t, tBedShareA, true)
	seedDevice(t, tBedShareB, true)
	seedResidentUnit(t, tRes1, tBedShareA)
	seedResidentUnit(t, tRes2, tBedShareB)
	reconcile(t, tUnitShare)
	if countDiffOp(*got, "admission") != 2 {
		t.Errorf("expect 2 admission, got %d (all=%v)", countDiffOp(*got, "admission"), *got)
	}
	// create 不应出现（每张新卡只 emit 一种 op；有 resident 走 admission 分支）
	if countDiffOp(*got, "create") != 0 {
		t.Errorf("expect 0 create (admission absorbs create for new-card-with-resident), got %d", countDiffOp(*got, "create"))
	}
}

// =========================================================================
// 组 F：幂等
// =========================================================================

func TestReconcile_F1_IdempotentTwice(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedPrivB, true)
	seedResidentUnit(t, tRes1, tUnitPrivate)
	reconcile(t, tUnitPrivate)
	cards1 := cardsInScope(t, tUnitPrivate)

	// 再跑一次 — 内容不应变化
	reconcile(t, tUnitPrivate)
	cards2 := cardsInScope(t, tUnitPrivate)
	if len(cards1) != len(cards2) {
		t.Errorf("idempotent: count changed %d → %d", len(cards1), len(cards2))
	}
}

func TestReconcile_F2_TenantWideEqualsPerUnit(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedShareA, true)
	seedDevice(t, tRoomPublicLiv, true)

	// 一次性 tenant /48 reconcile
	reconcile(t, tTenant)
	gotAll := cardsInScope(t, tTenant)
	cardCount := len(gotAll)

	// 单独跑 3 个 unit reconcile（应该等价）
	reconcile(t, tUnitPrivate)
	reconcile(t, tUnitShare)
	reconcile(t, tUnitPublic)
	gotPer := cardsInScope(t, tTenant)
	if len(gotPer) != cardCount {
		t.Errorf("tenant-wide vs per-unit mismatch: %d vs %d", cardCount, len(gotPer))
	}
}

// =========================================================================
// 组 G：Edge cases
// =========================================================================

func TestReconcile_G1_DeviceNotInAnyUnit(t *testing.T) {
	resetTestTenantState(t)
	// device_ipv6 在 tenant /48 内但不属于任何 unit (host-only)，应不进 expected
	var devID string
	if err := tDB.QueryRow(`
		INSERT INTO device_factory_meta (device_id, device_uid, device_type, import_date)
		VALUES (gen_random_uuid(), 'TEST-G1', 'Radar', NOW())
		RETURNING device_id::text
	`).Scan(&devID); err != nil {
		t.Fatal(err)
	}
	if _, err := tDB.Exec(`
		INSERT INTO devices (device_id, device_ipv6, monitoring_enabled, access)
		VALUES ($1::UUID, 'fd00:0:99::abcd:1234/128', TRUE, TRUE)
	`, devID); err != nil {
		t.Fatal(err)
	}
	reconcile(t, tTenant)
	assertCards(t, tTenant) // 整 tenant 无卡
}

func TestReconcile_G2_NoActiveResidentUnit(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedPrivA, true)
	// 不插 resident_unit
	reconcile(t, tUnitPrivate)
	cards := assertCards(t, tUnitPrivate, tUnitPrivate)
	if c := findCard(t, cards, tUnitPrivate); c.ResidentID != "" {
		t.Errorf("no resident → expect NoOne, got %q", c.ResidentID)
	}
	if c := findCard(t, cards, tUnitPrivate); c.CardName != CardNameNoResident {
		t.Errorf("card_name should be NoOne, got %q", c.CardName)
	}
}

func TestReconcile_G4_ScopeTenant(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedShareA, true)
	reconcile(t, tTenant) // 一次扫整 tenant
	cards := cardsInScope(t, tTenant)
	if len(cards) < 2 {
		t.Errorf("tenant scope reconcile should cover both units, got %d cards", len(cards))
	}
}

func TestReconcile_G5_ScopeBranch(t *testing.T) {
	resetTestTenantState(t)
	seedDevice(t, tBedPrivA, true)
	reconcile(t, tBranch) // /56 scope
	assertCards(t, tBranch, tUnitPrivate)
}

// =========================================================================
// 组 H：v3 套娃规则（room-level 嵌套层）
// =========================================================================

func TestReconcile_H1_RoomMultipleBedKeepsRoomCard(t *testing.T) {
	resetTestTenantState(t)
	// PrivBedroom 内 2 bed (101+102) + room-level radar
	// → split mode；room active_bed=2 → /88 room card 装 radar
	// → 期望 3 张：bed 101 + bed 102 + room PrivBedroom
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedPrivB, true)
	seedDevice(t, tRoomPrivBedroom, true) // room-level radar (bed slot=0，不在任何 bed)
	reconcile(t, tUnitPrivate)
	assertCards(t, tUnitPrivate, tBedPrivA, tBedPrivB, tRoomPrivBedroom)
}

func TestReconcile_H2_SplitMinusOneAddRoomDeviceMerges(t *testing.T) {
	resetTestTenantState(t)
	// 起点：2 bed → split (2 张 /96)
	seedDevice(t, tBedPrivA, true)
	bedB := seedDevice(t, tBedPrivB, true)
	reconcile(t, tUnitPrivate)
	// 删 bedB + 加 PrivBedroom room-level radar
	_, _ = tDB.Exec(`DELETE FROM devices WHERE device_id=$1::UUID`, bedB)
	seedDevice(t, tRoomPrivBedroom, true)
	reconcile(t, tUnitPrivate)
	// 无单向阀 latching：unit activeBed=1 → MERGE → 仅 /80 unit card
	// （bed device 与 room device 都隐含其中；split 状态消失）
	assertCards(t, tUnitPrivate, tUnitPrivate)
}

func TestReconcile_H3_RoomZeroBedKeepsRoomCard(t *testing.T) {
	resetTestTenantState(t)
	// 2 bed → unit activeBed=2 → split mode
	seedDevice(t, tBedPrivA, true)
	seedDevice(t, tBedPrivB, true)
	reconcile(t, tUnitPrivate)
	// 加 PrivBathroom radar（不同 room，无 bed）
	seedDevice(t, tRoomPrivBathroom, true)
	reconcile(t, tUnitPrivate)
	// split 模式下 PrivBathroom activeBed=0 + hasDev → 独立 /88 room card（不上推 /80）
	assertCards(t, tUnitPrivate, tBedPrivA, tBedPrivB, tRoomPrivBathroom)
}

func TestReconcile_H4_UnitTypeIndependence(t *testing.T) {
	resetTestTenantState(t)
	// v3 关键性质：相同 device 拓扑，无论 unit_type=Private/Share/Public，cards 集合一致
	// Share unit + 2 bed → 也走套娃 (与 Private 一致)
	seedDevice(t, tBedShareA, true)
	seedDevice(t, tBedShareB, true)
	reconcile(t, tUnitShare)
	// Share unit (unit_type=2) 不再走"自然 anchor"，与 Private 用同一规则:
	// 2 bed 同 room (ShareBedroom) + 无 room device → 仅 2 张 /96，不出 /80 / /88
	assertCards(t, tUnitShare, tBedShareA, tBedShareB)
}
