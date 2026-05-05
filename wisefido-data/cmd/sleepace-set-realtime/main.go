// sleepace-set-realtime: 把 sleepad 设备的实时上报配置改为 interval=2s + realtimeMode=1。
//
// 三种模式：
//
//	# 单台
//	go run ./cmd/sleepace-set-realtime -uid BM87224700978
//	go run ./cmd/sleepace-set-realtime -code 8amzqonkfyfyy -dry-run
//
//	# 全量扫描（只查不改，用于了解现状）
//	go run ./cmd/sleepace-set-realtime -scan
//
//	# 全量下发（仅 BM8701-2 ≥6.67 支持 mode；其它型号只下 interval）
//	go run ./cmd/sleepace-set-realtime -apply-all -yes
//
// 仅 BM8701-2 + firmware ≥ 6.67 支持 realtimeMode/* 接口；其它型号/老固件 mode 步骤跳过。
// SetRealtimeInterval 老接口所有型号都支持。
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"wisefido-data/internal/config"
	"wisefido-data/internal/service"

	"owl-common/database"

	"go.uber.org/zap"
)

const (
	targetInterval = 2 // seconds
	targetMode     = 1 // 1 = 离床后不上报；0 = 离床后仍上报
	apiTimeout     = 15 * time.Second
)

func main() {
	uid := flag.String("uid", "", "device_uid (e.g. BM87224700978)；与 -code 二选一；与 -scan/-apply-all 互斥")
	code := flag.String("code", "", "device_code (厂家加密 ID, e.g. 8amzqonkfyfyy)")
	dryRun := flag.Bool("dry-run", false, "只 Get 当前 mode，不下发 Set")
	scan := flag.Bool("scan", false, "全量扫描所有 sleepad 设备并 Get 当前 mode（不改任何东西）")
	applyAll := flag.Bool("apply-all", false, "对所有 sleepad 下发 interval=2 + mode=1（需要 -yes 确认）")
	yes := flag.Bool("yes", false, "跳过 -apply-all 的二次确认")
	flag.Parse()

	if *scan && *applyAll {
		log.Fatal("-scan 和 -apply-all 互斥")
	}
	singleMode := *uid != "" || *code != ""
	if singleMode && (*scan || *applyAll) {
		log.Fatal("-uid/-code 不能与 -scan/-apply-all 同时使用")
	}
	if !singleMode && !*scan && !*applyAll {
		log.Fatal("必须指定 -uid/-code 或 -scan 或 -apply-all 之一")
	}
	if *applyAll && !*yes {
		log.Fatal("-apply-all 是批量下发，请加 -yes 确认")
	}

	cfg := config.Load()
	if !cfg.DBEnabled {
		log.Fatal("config.db_enabled = false")
	}
	if cfg.SleepaceGateway.APIBaseURL == "" {
		log.Fatal("SLEEPACE_GATEWAY_API_BASE_URL 未配置")
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("DB connect: %v", err)
	}
	defer db.Close()

	gw := service.NewSleepaceGatewayClient(cfg.SleepaceGateway.APIBaseURL, logger)

	switch {
	case singleMode:
		runSingle(db, gw, *uid, *code, *dryRun)
	case *scan:
		runScan(db, gw)
	case *applyAll:
		runApplyAll(db, gw)
	}
}

// ---------------- 单台 ----------------

func runSingle(db *sql.DB, gw *service.SleepaceGatewayClient, uid, code string, dryRun bool) {
	dev, err := lookupDevice(context.Background(), db, uid, code)
	if err != nil {
		log.Fatalf("lookup device: %v", err)
	}
	printDeviceHeader(dev)

	if dryRun {
		getOnly(gw, dev)
		fmt.Println("[dry-run] 跳过所有 Set")
		return
	}

	getOnly(gw, dev)
	applyOne(gw, dev)
}

// ---------------- 扫描 ----------------

func runScan(db *sql.DB, gw *service.SleepaceGatewayClient) {
	devs, err := listSleepadDevices(context.Background(), db)
	if err != nil {
		log.Fatalf("list sleepad: %v", err)
	}
	fmt.Printf("found %d sleepad device(s)\n\n", len(devs))
	fmt.Printf("%-37s %-15s %-15s %-10s %-10s %-25s\n",
		"device_id", "device_uid", "device_code", "model", "firmware", "current_mode")
	fmt.Println(strings.Repeat("-", 130))
	for _, d := range devs {
		mode := "n/a (unsupported)"
		if isRTSupported(d) {
			ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
			m, err := gw.GetRealtimeModeAfterLeave(ctx, d.DeviceCode)
			cancel()
			if err != nil {
				mode = fmt.Sprintf("ERR: %v", err)
			} else {
				mode = fmt.Sprintf("%d (%s)", m, modeDesc(m))
			}
		}
		fmt.Printf("%-37s %-15s %-15s %-10s %-10s %-25s\n",
			d.DeviceID, d.DeviceUID, d.DeviceCode, d.DeviceModel, dispVer(d.FirmwareVer), mode)
	}
}

// ---------------- 全量下发 ----------------

func runApplyAll(db *sql.DB, gw *service.SleepaceGatewayClient) {
	devs, err := listSleepadDevices(context.Background(), db)
	if err != nil {
		log.Fatalf("list sleepad: %v", err)
	}
	fmt.Printf("applying interval=%d + mode=%d to %d sleepad device(s)\n\n", targetInterval, targetMode, len(devs))

	var ok, failInterval, failMode, mismatch, skipMode int
	for _, d := range devs {
		fmt.Printf("→ %s [%s] %s fw=%s\n", d.DeviceUID, d.DeviceCode, d.DeviceModel, dispVer(d.FirmwareVer))

		// SetRealtimeInterval（所有 sleepad 支持）
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		err := gw.SetRealtimeInterval(ctx, d.DeviceID, d.DeviceCode, targetInterval)
		cancel()
		if err != nil {
			fmt.Printf("    SetRealtimeInterval FAILED: %v\n", err)
			failInterval++
			continue
		}
		fmt.Printf("    SetRealtimeInterval(2) OK\n")

		// SetRealtimeModeAfterLeave（仅 BM8701-2 ≥6.67）
		if !isRTSupported(d) {
			fmt.Printf("    skip mode (model/firmware unsupported)\n")
			skipMode++
			ok++
			continue
		}
		ctx, cancel = context.WithTimeout(context.Background(), apiTimeout)
		err = gw.SetRealtimeModeAfterLeave(ctx, d.DeviceID, d.DeviceCode, targetMode)
		cancel()
		if err != nil {
			fmt.Printf("    SetRealtimeModeAfterLeave FAILED: %v\n", err)
			failMode++
			continue
		}
		// Verify
		ctx, cancel = context.WithTimeout(context.Background(), apiTimeout)
		got, gerr := gw.GetRealtimeModeAfterLeave(ctx, d.DeviceCode)
		cancel()
		if gerr != nil {
			fmt.Printf("    SetRealtimeModeAfterLeave OK; verify Get FAILED: %v\n", gerr)
			failMode++
			continue
		}
		if got != targetMode {
			fmt.Printf("    MISMATCH after Set: got mode=%d expect=%d\n", got, targetMode)
			mismatch++
			continue
		}
		fmt.Printf("    SetRealtimeModeAfterLeave(1) OK + verified\n")
		ok++
	}

	fmt.Printf("\nsummary: total=%d  ok=%d  fail_interval=%d  fail_mode=%d  mismatch=%d  skip_mode=%d\n",
		len(devs), ok, failInterval, failMode, mismatch, skipMode)
	if failInterval+failMode+mismatch > 0 {
		log.Fatal("有失败，请回看上面输出")
	}
}

// ---------------- helpers ----------------

func getOnly(gw *service.SleepaceGatewayClient, d *deviceRecord) {
	if !isRTSupported(*d) {
		fmt.Println("[get] skip (model/firmware 不支持 realtimeMode/get)")
		return
	}
	fmt.Println("[get] GetRealtimeModeAfterLeave …")
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()
	mode, err := gw.GetRealtimeModeAfterLeave(ctx, d.DeviceCode)
	if err != nil {
		fmt.Printf("    FAILED: %v\n", err)
		return
	}
	fmt.Printf("    OK current mode = %d (%s)\n", mode, modeDesc(mode))
}

func applyOne(gw *service.SleepaceGatewayClient, d *deviceRecord) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	fmt.Printf("[set] SetRealtimeInterval(interval=%d) …\n", targetInterval)
	if err := gw.SetRealtimeInterval(ctx, d.DeviceID, d.DeviceCode, targetInterval); err != nil {
		fmt.Printf("    FAILED: %v\n", err)
	} else {
		fmt.Printf("    OK\n")
	}

	if !isRTSupported(*d) {
		fmt.Println("[set] skip mode (model/firmware 不支持)")
		return
	}
	fmt.Printf("[set] SetRealtimeModeAfterLeave(mode=%d) …\n", targetMode)
	if err := gw.SetRealtimeModeAfterLeave(ctx, d.DeviceID, d.DeviceCode, targetMode); err != nil {
		fmt.Printf("    FAILED: %v\n", err)
		return
	}
	fmt.Printf("    OK\n")
	fmt.Println("[verify] GetRealtimeModeAfterLeave …")
	mode, err := gw.GetRealtimeModeAfterLeave(ctx, d.DeviceCode)
	if err != nil {
		fmt.Printf("    FAILED: %v\n", err)
	} else if mode != targetMode {
		fmt.Printf("    MISMATCH: expected %d, got %d (%s)\n", targetMode, mode, modeDesc(mode))
	} else {
		fmt.Printf("    OK mode = %d (%s) ✓\n", mode, modeDesc(mode))
	}
}

func printDeviceHeader(d *deviceRecord) {
	fmt.Printf("[device]\n  device_id    : %s\n  device_uid   : %s\n  device_code  : %s\n  device_model : %s\n  firmware     : %s\n",
		d.DeviceID, d.DeviceUID, d.DeviceCode, d.DeviceModel, d.FirmwareVer)
	fmt.Printf("  realtimeMode supported: %v (model=BM8701-2 && firmware>=6.67)\n\n", isRTSupported(*d))
}

func isRTSupported(d deviceRecord) bool {
	return strings.EqualFold(d.DeviceModel, "BM8701-2") && firmwareGE(d.FirmwareVer, "6.67")
}

type deviceRecord struct {
	DeviceID    string
	DeviceUID   string
	DeviceCode  string
	DeviceModel string
	FirmwareVer string
}

func lookupDevice(ctx context.Context, db *sql.DB, uid, code string) (*deviceRecord, error) {
	q := `
		SELECT d.device_id::text,
		       COALESCE(d.device_uid, '') AS device_uid,
		       COALESCE(NULLIF(TRIM(ds.device_code), ''), '') AS device_code,
		       COALESCE(ds.device_model, '') AS device_model,
		       COALESCE(ds.firmware_version, '') AS firmware_version
		FROM devices d
		JOIN device_store ds ON ds.device_id = d.device_id
		WHERE LOWER(COALESCE(ds.device_type, '')) = 'sleepad'
		  AND ($1 = '' OR d.device_uid = $1)
		  AND ($2 = '' OR ds.device_code = $2)
		LIMIT 2`
	rows, err := db.QueryContext(ctx, q, uid, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var matches []deviceRecord
	for rows.Next() {
		var r deviceRecord
		if err := rows.Scan(&r.DeviceID, &r.DeviceUID, &r.DeviceCode, &r.DeviceModel, &r.FirmwareVer); err != nil {
			return nil, err
		}
		matches = append(matches, r)
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no sleepad device matched uid=%q code=%q", uid, code)
	case 1:
		if matches[0].DeviceCode == "" {
			return nil, fmt.Errorf("device has empty device_code; cannot call sleepace API")
		}
		return &matches[0], nil
	default:
		return nil, fmt.Errorf("more than one device matched; refine -uid/-code")
	}
}

// listSleepadDevices 列出所有需要纳入采样调度的 sleepad 设备（device_code 非空、未禁用）。
func listSleepadDevices(ctx context.Context, db *sql.DB) ([]deviceRecord, error) {
	q := `
		SELECT d.device_id::text,
		       COALESCE(d.device_uid, '') AS device_uid,
		       COALESCE(NULLIF(TRIM(ds.device_code), ''), '') AS device_code,
		       COALESCE(ds.device_model, '') AS device_model,
		       COALESCE(ds.firmware_version, '') AS firmware_version
		FROM devices d
		JOIN device_store ds ON ds.device_id = d.device_id
		WHERE LOWER(COALESCE(ds.device_type, '')) = 'sleepad'
		  AND COALESCE(d.status, '') NOT IN ('disabled', 'removed')
		  AND COALESCE(NULLIF(TRIM(ds.device_code), ''), '') <> ''
		ORDER BY d.device_uid`
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []deviceRecord
	for rows.Next() {
		var r deviceRecord
		if err := rows.Scan(&r.DeviceID, &r.DeviceUID, &r.DeviceCode, &r.DeviceModel, &r.FirmwareVer); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func modeDesc(m int) string {
	switch m {
	case 0:
		return "离床后仍上报"
	case 1:
		return "离床后不上报"
	default:
		return "?"
	}
}

func dispVer(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "(empty)"
	}
	return v
}

// firmwareGE 比较 "6.89" 这种 major.minor 字符串；非数字版本返回 false（保守，不假设支持）。
func firmwareGE(ver, min string) bool {
	a1, a2 := splitVer(ver)
	b1, b2 := splitVer(min)
	if a1 < 0 || b1 < 0 {
		return false
	}
	if a1 != b1 {
		return a1 > b1
	}
	return a2 >= b2
}

func splitVer(v string) (int, int) {
	v = strings.TrimSpace(v)
	if v == "" {
		return -1, -1
	}
	parts := strings.SplitN(v, ".", 2)
	a, err := atoiSafe(parts[0])
	if err != nil {
		return -1, -1
	}
	b := 0
	if len(parts) == 2 {
		b, _ = atoiSafe(parts[1])
	}
	return a, b
}

func atoiSafe(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not int: %q", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
