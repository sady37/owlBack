package httpapi

// 端到端 HTTP test：用 httptest.Server 起 v2 spatial 路由，跑完整 curl-style POST 流程。
//
// 跑法：
//   cd wisefido-data
//   OWL_PG_DSN="host=localhost port=5432 user=postgres password=postgres dbname=owl_v2 sslmode=disable" \
//   DDNS_TSIG_SECRET="qDcMwcwqjVfR7gKkyfA3j413160w7SXhGbfI1yNb8v8=" \
//   go test -v -run TestSpatialV2_HTTP_E2E ./internal/http/

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"owl-common/ddns"
	"owl-common/ipam"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func TestSpatialV2_HTTP_E2E(t *testing.T) {
	dsn := os.Getenv("OWL_PG_DSN")
	tsigSecret := os.Getenv("DDNS_TSIG_SECRET")
	if dsn == "" || tsigSecret == "" {
		t.Skip("OWL_PG_DSN / DDNS_TSIG_SECRET not set; integration test skipped")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	// 清理上一次测试残留 (test branch slot=98 = 0x62 专用 v2 HTTP test)
	cleanup := func() {
		_, _ = db.ExecContext(context.Background(), `
			DELETE FROM devices WHERE spatial_addr << 'fd00:0:3:6200::/56'::INET;
			DELETE FROM beds WHERE spatial_prefix << 'fd00:0:3:6200::/56'::INET;
			DELETE FROM rooms WHERE spatial_prefix << 'fd00:0:3:6200::/56'::INET;
			DELETE FROM units WHERE spatial_prefix << 'fd00:0:3:6200::/56'::INET;
			DELETE FROM sites WHERE spatial_prefix << 'fd00:0:3:6200::/56'::INET;
			DELETE FROM branches WHERE spatial_prefix = 'fd00:0:3:6200::/56'::INET;
		`)
	}
	cleanup()
	t.Cleanup(cleanup)

	// 直接 INSERT test branch (不能用 AllocateBranch 因为它取 MAX+1 不让我控制 slot=98)
	if _, err := db.Exec(`
		INSERT INTO branches (spatial_prefix, branch_slot, branch_name, timezone)
		VALUES ('fd00:0:3:6200::/56', 98, 'V2-HTTP-Test', 'America/Denver')
	`); err != nil {
		t.Fatalf("seed test branch: %v", err)
	}

	// 构造 backend (不接 kea，避免依赖 kea audit；本测试只验 HTTP 层)
	backend := ipam.NewPGBackend(db)

	// 构造 ddns client (验 DDNS 推 BIND)
	ddnsClient, err := ddns.New(ddns.Config{
		Server: "127.0.0.1", Port: 5353,
		KeyName: "ddns-update", Algorithm: "hmac-sha256",
		Secret: tsigSecret, OwlDomain: "owl.",
	})
	if err != nil {
		t.Fatalf("ddns new: %v", err)
	}

	logger := zap.NewNop()
	handler := NewSpatialV2Handler(backend, db, ddnsClient, logger)
	router := NewRouter(logger)
	router.RegisterSpatialV2Routes(handler)
	srv := httptest.NewServer(router)
	defer srv.Close()

	post := func(path string, body any) (status int, parsed map[string]any) {
		buf, _ := json.Marshal(body)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL+path, bytes.NewReader(buf))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("POST %s: %v", path, err)
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatalf("decode response from %s: %v (raw: %s)", path, err, string(raw))
		}
		return resp.StatusCode, parsed
	}

	// Helper: extract result.prefix from response (envelope: {type, code, message, result})
	getResultPrefix := func(t *testing.T, resp map[string]any, label string) string {
		if resp["type"] != "success" {
			t.Fatalf("%s: not success: %+v", label, resp)
		}
		result, ok := resp["result"].(map[string]any)
		if !ok {
			t.Fatalf("%s: response missing result: %+v", label, resp)
		}
		prefix, _ := result["prefix"].(string)
		if prefix == "" {
			t.Fatalf("%s: response.result.prefix empty: %+v", label, resp)
		}
		return prefix
	}

	// === Step 1: AllocateSite (building 0 floor 8) ===
	status, resp := post("/admin/api/v2/spatial/sites", map[string]any{
		"parent": "fd00:0:3:6200::/56", "building": 0, "floor": 8,
		"attrs": map[string]string{"name": "V2HTTP Bldg0 Floor8"},
	})
	if status != 200 || resp["type"] != "success" {
		t.Fatalf("site: %+v", resp)
	}
	site := getResultPrefix(t, resp, "site")
	t.Logf("✓ POST /sites → %s", site)

	// === Step 2: AllocateUnit ===
	status, resp = post("/admin/api/v2/spatial/units", map[string]any{
		"parent": site,
		"attrs":  map[string]any{"name": "V2HTTPUnit", "unit_type": 1},
	})
	if status != 200 || resp["type"] != "success" {
		t.Fatalf("unit: %+v", resp)
	}
	unit := getResultPrefix(t, resp, "unit")
	t.Logf("✓ POST /units → %s", unit)

	// === Step 3: AllocateRoom ===
	status, resp = post("/admin/api/v2/spatial/rooms", map[string]any{
		"parent": unit,
		"attrs":  map[string]any{"name": "Bedroom", "room_type": "bedroom", "is_primary": true},
	})
	if status != 200 || resp["type"] != "success" {
		t.Fatalf("room: %+v", resp)
	}
	room := getResultPrefix(t, resp, "room")
	t.Logf("✓ POST /rooms → %s", room)

	// === Step 4: AllocateBed ===
	status, resp = post("/admin/api/v2/spatial/beds", map[string]any{
		"parent": room,
		"attrs":  map[string]string{"name": "BedA"},
	})
	if status != 200 || resp["type"] != "success" {
		t.Fatalf("bed: %+v", resp)
	}
	bed := getResultPrefix(t, resp, "bed")
	t.Logf("✓ POST /beds → %s", bed)

	// === Step 5: 找空闲 device + RegisterDevice + DDNS ===
	var deviceID, deviceUID string
	if err := db.QueryRow(`
		SELECT dfm.device_id::text, dfm.device_uid
		FROM device_factory_meta dfm
		LEFT JOIN devices d ON d.device_id = dfm.device_id
		WHERE dfm.device_model = 'HC2' AND d.device_id IS NULL LIMIT 1
	`).Scan(&deviceID, &deviceUID); err != nil {
		t.Fatalf("find unbound device: %v", err)
	}
	status, resp = post("/admin/api/v2/spatial/devices", map[string]any{
		"base":       bed,
		"device_id":  deviceID,
		"device_uid": deviceUID,
		"dns_name":   "v2-http-test-bed",
		"dns_zone":   "tenant1.owl.",
	})
	if status != 200 || resp["type"] != "success" {
		t.Fatalf("device: %+v", resp)
	}
	data := resp["result"].(map[string]any)
	devAddr := data["address"].(string)
	dnsFQDN := data["dns_fqdn"].(string)
	t.Logf("✓ POST /devices → %s (DNS %s)", devAddr, dnsFQDN)

	// === Step 6: dig forward AAAA 验证 DDNS 写到 BIND ===
	cmdOut, err := digQuery(t, "v2-http-test-bed.tenant1.owl", "AAAA")
	if err != nil {
		t.Fatalf("dig: %v", err)
	}
	if !strings.Contains(cmdOut, devAddr) {
		t.Errorf("dig AAAA mismatch: got %q want contains %q", cmdOut, devAddr)
	} else {
		t.Logf("✓ dig forward AAAA: %s", strings.TrimSpace(cmdOut))
	}

	// === Step 7: GET /tenants/<encoded> 反查 ===
	encoded := url.PathEscape("fd00:0:3::/48") // ":" 在 path 段需 URL-encode
	getURL := srv.URL + "/admin/api/v2/spatial/tenants/" + encoded + "?prefix=fd00:0:3::/48"
	getResp, err := http.Get(getURL)
	if err != nil {
		t.Fatalf("GET tenants: %v", err)
	}
	defer getResp.Body.Close()
	tBody, _ := io.ReadAll(getResp.Body)
	var tenant map[string]any
	if err := json.Unmarshal(tBody, &tenant); err != nil {
		t.Fatalf("decode tenant: %v (raw: %s)", err, string(tBody))
	}
	if tenant["type"] != "success" {
		t.Fatalf("tenant lookup not success: %+v", tenant)
	}
	tData := tenant["result"].(map[string]any)
	if tData["name"] != "demo" {
		t.Errorf("tenant name expected 'demo' got %v", tData["name"])
	}
	t.Logf("✓ GET /tenants → name=%v slot=%v tz=%v", tData["name"], tData["slot"], tData["timezone"])

	// 清理 DDNS
	if a, err := netip.ParseAddr(devAddr); err == nil {
		_ = ddnsClient.UnregisterDevice(context.Background(), a, "v2-http-test-bed", "tenant1.owl.")
	}
}

func digQuery(t *testing.T, name, qtype string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "dig", "@127.0.0.1", "-p", "5353", name, qtype, "+short").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
