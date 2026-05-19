package ipam_test

// e2e_test 在 owl_v2 上跑 IPAM + DDNS 端到端流程：
//   1. 在 demo tenant (slot=3) 下创建 临时 branch (Test-Phase-B)
//   2. 在该 branch 下加 site (building 0, floor 9)
//   3. 在 site 下加 unit
//   4. 在 unit 下加 room
//   5. 在 room 下加 bed
//   6. 用任一未绑定的 device_factory_meta record 派生 device /128，注册
//   7. 调 DDNS client 写 BIND zone
//   8. dig 查询 forward AAAA 验证
//   9. cleanup 删除测试数据
//
// 跑法：
//   cd owl-common
//   OWL_PG_DSN="host=localhost port=5432 user=postgres password=postgres dbname=owl_v2 sslmode=disable" \
//   DDNS_TSIG_SECRET="qDcMwcwqjVfR7gKkyfA3j413160w7SXhGbfI1yNb8v8=" \
//   go test -v -run TestE2E_IPAM_DDNS ./ipam/
//
// 不跑：
//   - 跳过条件：env OWL_PG_DSN 或 DDNS_TSIG_SECRET 缺失 → t.Skip

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"owl-common/card"
	"owl-common/ddns"
	"owl-common/ipam"

	_ "github.com/lib/pq"
)

func TestE2E_IPAM_DDNS(t *testing.T) {
	dsn := os.Getenv("OWL_PG_DSN")
	tsigSecret := os.Getenv("DDNS_TSIG_SECRET")
	if dsn == "" || tsigSecret == "" {
		t.Skip("OWL_PG_DSN or DDNS_TSIG_SECRET not set; skipping integration test")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 同时启用 kea sync — Allocate 成功后 lease 写到 kea
	keaURL := os.Getenv("KEA_CTRL_URL")
	if keaURL == "" {
		keaURL = "http://localhost:8000"
	}
	keaClient := ipam.NewKeaClient(keaURL, "owl", "owl-kea-dev-2026")
	if v, err := keaClient.VersionGet(ctx); err != nil {
		t.Fatalf("kea VersionGet failed (kea not up?): %v", err)
	} else {
		t.Logf("✓ kea connected: %s", v)
	}
	pg := ipam.NewPGBackendWithKea(db, keaClient)

	// demo tenant /48 = fd00:0:3::/48 (used implicitly via the test branch fd00:0:3:6300::/56)
	_ = netip.MustParsePrefix("fd00:0:3::/48")

	// 清理上一次失败留下的 test branch（branch_slot=99 留作专用测试 slot）
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `
			DELETE FROM devices WHERE spatial_addr << 'fd00:0:3:6300::/56'::INET;
			DELETE FROM beds WHERE bed_id << 'fd00:0:3:6300::/56'::INET;
			DELETE FROM rooms WHERE room_id << 'fd00:0:3:6300::/56'::INET;
			DELETE FROM units WHERE unit_id << 'fd00:0:3:6300::/56'::INET;
			DELETE FROM sites WHERE site_id << 'fd00:0:3:6300::/56'::INET;
			DELETE FROM branches WHERE branch_id = 'fd00:0:3:6300::/56'::INET;
		`)
	})

	// Step 0: clean up first to avoid prev-run pollution
	_, _ = db.ExecContext(ctx, `
		DELETE FROM devices WHERE spatial_addr << 'fd00:0:3:6300::/56'::INET;
		DELETE FROM beds WHERE bed_id << 'fd00:0:3:6300::/56'::INET;
		DELETE FROM rooms WHERE room_id << 'fd00:0:3:6300::/56'::INET;
		DELETE FROM units WHERE unit_id << 'fd00:0:3:6300::/56'::INET;
		DELETE FROM sites WHERE site_id << 'fd00:0:3:6300::/56'::INET;
		DELETE FROM branches WHERE branch_id = 'fd00:0:3:6300::/56'::INET;
	`)

	// 由于 AllocateBranch 是 MAX+1 自动分配，无法精确控制 slot=99；
	// 直接 INSERT 一个 branch_slot=99 (0x63) 测试 branch 用 (避免与 demo 现有 1..4 冲突)
	_, err = db.ExecContext(ctx, `
		INSERT INTO branches (branch_id, branch_slot, branch_name, timezone)
		VALUES ('fd00:0:3:6300::/56', 99, 'Test-Phase-B', 'America/Denver')
	`)
	if err != nil {
		t.Fatalf("insert test branch: %v", err)
	}
	branch := netip.MustParsePrefix("fd00:0:3:6300::/56")
	// 直接 INSERT 时手动补 kea lease (绕过 Allocate path 的 hook)
	if err := keaClient.RecordPrefixLease(ctx, branch, ipam.EncodeDUIDForPrefix(branch), 56, "branch:Test-Phase-B (manual)", 0); err != nil {
		t.Fatalf("kea record branch lease: %v", err)
	}
	t.Logf("✓ test branch fd00:0:3:6300::/56")

	// Step 2: AllocateSite
	site, err := pg.AllocateSite(ctx, branch, 0, 9, ipam.SiteAttrs{
		Name: "TestBldg0 Floor9",
	})
	if err != nil {
		t.Fatalf("AllocateSite: %v", err)
	}
	t.Logf("✓ site %s", site)

	// Step 3: AllocateUnit
	unit, err := pg.AllocateUnit(ctx, site, ipam.UnitAttrs{
		Name:     "Test101",
		UnitType: 1, // Private
	})
	if err != nil {
		t.Fatalf("AllocateUnit: %v", err)
	}
	t.Logf("✓ unit %s", unit)

	// Step 4: AllocateRoom
	room, err := pg.AllocateRoom(ctx, unit, ipam.RoomAttrs{
		Name:      "Bedroom",
		RoomType:  card.RoomTypeDefault, // v2 简化：bedroom 合并到 Default
		IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("AllocateRoom: %v", err)
	}
	t.Logf("✓ room %s", room)

	// Step 5: AllocateBed
	bed, err := pg.AllocateBed(ctx, room, ipam.BedAttrs{Name: "BedA"})
	if err != nil {
		t.Fatalf("AllocateBed: %v", err)
	}
	t.Logf("✓ bed %s", bed)

	// Step 6: 找一个未绑定的 device_factory_meta record (HC2) 派生 device
	var deviceID, deviceUID string
	err = db.QueryRowContext(ctx, `
		SELECT dfm.device_id::text, dfm.device_uid
		FROM device_factory_meta dfm
		LEFT JOIN devices d ON d.device_id = dfm.device_id
		WHERE dfm.device_model = 'HC2' AND d.device_id IS NULL
		LIMIT 1
	`).Scan(&deviceID, &deviceUID)
	if err != nil {
		t.Fatalf("find unbound device: %v", err)
	}
	t.Logf("  using device_uid=%s", deviceUID)

	deviceAddr, err := pg.RegisterDevice(ctx, bed, deviceID, deviceUID)
	if err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}
	t.Logf("✓ device %s registered", deviceAddr)

	// Step 7: DDNS register
	dc, err := ddns.New(ddns.Config{
		Server: "127.0.0.1", Port: 5353,
		KeyName: "ddns-update", Algorithm: "hmac-sha256",
		Secret: tsigSecret, OwlDomain: "owl.",
	})
	if err != nil {
		t.Fatalf("ddns.New: %v", err)
	}
	zone := ddns.ZoneForTenant(3, "owl.") // tenant3.owl.
	// tenant3.owl. zone 不在 BIND 配置里（默认仅 tenant1）；只能写已配置 zone
	zone = "tenant1.owl." // demo 用 tenant1 zone 测试 DDNS 链路
	shortName := "test-phase-b-bed"
	if err := dc.RegisterDevice(ctx, deviceAddr, shortName, zone); err != nil {
		t.Fatalf("DDNS RegisterDevice: %v", err)
	}
	t.Logf("✓ DDNS register %s.%s -> %s", shortName, zone, deviceAddr)

	// Step 8: dig 查 forward AAAA 验证
	dig, err := exec.CommandContext(ctx, "dig", "@127.0.0.1", "-p", "5353",
		shortName+"."+zone, "AAAA", "+short").Output()
	if err != nil {
		t.Fatalf("dig: %v", err)
	}
	got := strings.TrimSpace(string(dig))
	want := deviceAddr.String()
	if got != want {
		t.Errorf("dig AAAA mismatch:\n  got:  %s\n  want: %s", got, want)
	} else {
		t.Logf("✓ dig forward AAAA matches: %s", got)
	}

	// Cleanup DDNS
	if err := dc.UnregisterDevice(ctx, deviceAddr, shortName, zone); err != nil {
		t.Logf("cleanup DDNS unregister: %v (non-fatal)", err)
	}

	// Step 9: 验证 kea lease database 真的有这些 prefix lease
	verifyKeaLease := func(t *testing.T, prefix netip.Prefix, label string) {
		duid := ipam.EncodeDUIDForPrefix(prefix)
		body, err := postKea(ctx, keaURL, "owl", "owl-kea-dev-2026",
			`{"command":"lease6-get-by-duid","service":["dhcp6"],"arguments":{"duid":"`+duid+`"}}`)
		if err != nil {
			t.Errorf("kea verify %s: %v", label, err)
			return
		}
		if !strings.Contains(body, prefix.Masked().Addr().String()) {
			t.Errorf("kea lease for %s (%s) not found in DB", label, prefix)
			return
		}
		t.Logf("✓ kea lease audit verified: %s = %s", label, prefix)
	}
	verifyKeaLease(t, netip.MustParsePrefix("fd00:0:3:6300::/56"), "branch")
	verifyKeaLease(t, site, "site")
	verifyKeaLease(t, unit, "unit")
	verifyKeaLease(t, room, "room")
	verifyKeaLease(t, bed, "bed")
}

// postKea is a small helper for the verification step (avoids exposing KeaClient internals).
func postKea(ctx context.Context, url, user, pass, jsonBody string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url+"/", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(user, pass)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return string(b), nil
}
