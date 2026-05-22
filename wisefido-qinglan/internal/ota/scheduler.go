package ota

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// OTAPushFn is the function signature for pushing OTA via MQTT
type OTAPushFn func(uid string, data map[string]interface{}) error

// TCPPushFn pushes OTA via TCP (for old MCU devices)
type TCPPushFn func(req PushRequest) PushResult

// Scheduler periodically scans for devices that need OTA updates
type Scheduler struct {
	db        *sql.DB
	otaPushFn OTAPushFn
	tcpPushFn TCPPushFn
	interval  time.Duration
	fwDir     string
	fwURL     string
	logger    *zap.Logger

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewScheduler creates a new OTA scheduler
func NewScheduler(db *sql.DB, otaPushFn OTAPushFn, tcpPushFn TCPPushFn, interval time.Duration, fwDir, fwURL string, logger *zap.Logger) *Scheduler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Scheduler{
		db:        db,
		otaPushFn: otaPushFn,
		tcpPushFn: tcpPushFn,
		interval:  interval,
		fwDir:     fwDir,
		fwURL:     fwURL,
		logger:    logger,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.wg.Add(1)
	go s.run(ctx)
	s.logger.Info("ota scheduler started",
		zap.Duration("interval", s.interval),
		zap.String("fw_dir", s.fwDir),
	)
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	s.logger.Info("ota scheduler stopped")
}

func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run once immediately
	s.scan(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.scan(ctx)
		}
	}
}

// scan v2：从 device_ota + devices + dfm 派生可升级设备并推送固件
//
// v2 schema：
//   - device_ota（PK device_ipv6）: target_firmware_version / target_mcu_model / approve_way / schedule / status
//   - devices.access=TRUE 才允许接入
//   - 路由 device_uid 用 dfm.device_uid
func (s *Scheduler) scan(ctx context.Context) {
	if s.otaPushFn == nil {
		return
	}

	// Phase 2: device_ota PK = device_uid (logMAC)；直接 JOIN devices(device_uid) 查 access
	// approve_way 值集：system_schedule / tenant_schedule / system_manual / tenant_manual
	query := `
		SELECT o.device_uid,
		       COALESCE(o.approve_way, '')                   AS approve_way,
		       COALESCE(o.schedule::text, '')                AS schedule_ts,
		       COALESCE(o.target_firmware_version, '')       AS ota_target_fw,
		       COALESCE(o.target_mcu_model, '')              AS ota_target_mcu,
		       COALESCE(o.updated_at, NOW())                 AS ota_updated_at
		FROM device_ota o
		JOIN devices d ON d.device_uid = o.device_uid
		WHERE d.access = TRUE
		  AND o.approve_way IN ('tenant_schedule', 'tenant_manual')
		  AND o.approved_at IS NOT NULL
		  AND COALESCE(o.status, 'idle') IN ('idle', 'scheduled')
		  AND ((o.target_firmware_version IS NOT NULL AND o.target_firmware_version != '')
		    OR (o.target_mcu_model IS NOT NULL AND o.target_mcu_model != ''))
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		s.logger.Warn("ota scheduler query", zap.Error(err))
		return
	}
	defer rows.Close()

	count := 0
	pushed := 0
	for rows.Next() {
		var uid, way, schedule, targetFW, targetMCU string
		var updatedAt time.Time
		if err := rows.Scan(&uid, &way, &schedule, &targetFW, &targetMCU, &updatedAt); err != nil {
			s.logger.Warn("ota scheduler scan row", zap.Error(err))
			continue
		}
		count++

		if strings.HasSuffix(way, "_schedule") && schedule != "" {
			if !IsInScheduleWindow(schedule, time.Now(), updatedAt) {
				s.logger.Debug("schedule not in window",
					zap.String("uid", uid),
					zap.String("schedule", schedule),
				)
				continue
			}
		}

		if targetFW == "" && targetMCU == "" {
			s.logger.Debug("no target firmware set", zap.String("uid", uid))
			continue
		}

		verInfo := &FwVersionInfo{}
		data := map[string]interface{}{}
		pushReq := PushRequest{UID: uid}

		if targetMCU != "" {
			localPath := filepath.Join(s.fwDir, targetMCU)
			sz, err := getFileSize(localPath)
			if err != nil {
				s.logger.Warn("mcu firmware not found",
					zap.String("uid", uid),
					zap.String("path", localPath),
					zap.Error(err),
				)
				continue
			}
			sha, _ := getFileSHA256(localPath)
			if v := ParseUpdateINI(s.fwDir, targetMCU); v != nil {
				verInfo.EspVer = v.EspVer
			}
			url := fmt.Sprintf("%s/%s", s.fwURL, targetMCU)
			data["espfileUrl"] = url
			data["espfilesha256"] = sha
			data["espfilesize"] = sz
			data["espver"] = verInfo.EspVer
			pushReq.EspVersion = verInfo.EspVer
			pushReq.EspFileURL = url
			pushReq.EspFileSize = uint32(sz)
			pushReq.EspSHA256 = sha
		}

		if targetFW != "" {
			localPath := filepath.Join(s.fwDir, targetFW)
			sz, err := getFileSize(localPath)
			if err != nil {
				s.logger.Warn("radar firmware not found",
					zap.String("uid", uid),
					zap.String("path", localPath),
					zap.Error(err),
				)
				continue
			}
			sha, _ := getFileSHA256(localPath)
			if v := ParseUpdateINI(s.fwDir, targetFW); v != nil {
				verInfo.RadarVer = v.RadarVer
			}
			url := fmt.Sprintf("%s/%s", s.fwURL, targetFW)
			data["radarfileUrl"] = url
			data["radarfilesha256"] = sha
			data["radarfilesize"] = sz
			data["radarver"] = verInfo.RadarVer
			pushReq.RadarFirmware = targetFW
			pushReq.RadarVersion = verInfo.RadarVer
			pushReq.RadarSHA256 = sha
		}

		s.logger.Info("pushing ota",
			zap.String("uid", uid),
			zap.String("mcu", targetMCU),
			zap.String("mcu_ver", verInfo.EspVer),
			zap.String("fw", targetFW),
			zap.String("fw_ver", verInfo.RadarVer),
		)

		pushed_ok := false
		if s.tcpPushFn != nil {
			result := s.tcpPushFn(pushReq)
			if result.Success {
				s.logger.Info("tcp push ok", zap.String("uid", uid))
				pushed_ok = true
			} else {
				s.logger.Info("tcp push failed, trying mqtt",
					zap.String("uid", uid),
					zap.String("msg", result.Message),
				)
				if s.otaPushFn != nil {
					if err := s.otaPushFn(uid, data); err != nil {
						s.logger.Warn("mqtt fallback failed", zap.String("uid", uid), zap.Error(err))
					} else {
						pushed_ok = true
					}
				}
			}
		} else if s.otaPushFn != nil {
			if err := s.otaPushFn(uid, data); err != nil {
				s.logger.Warn("mqtt push failed", zap.String("uid", uid), zap.Error(err))
				continue
			}
			pushed_ok = true
		}

		if pushed_ok {
			_, err = s.db.ExecContext(ctx, `
				UPDATE device_ota
				SET status = 'downloading', updated_at = NOW(), last_attempted_at = NOW()
				WHERE device_uid = $1
			`, uid)
			if err != nil {
				s.logger.Warn("device_ota status update", zap.String("uid", uid), zap.Error(err))
			}
		}

		pushed++
	}

	if count > 0 {
		s.logger.Info("ota scan complete",
			zap.Int("eligible", count),
			zap.Int("pushed", pushed),
		)
	}
}

// findFirmware searches for a firmware file matching the device model
func (s *Scheduler) findFirmware(model, targetVer string) (string, error) {
	if model == "" {
		return "", fmt.Errorf("empty device model")
	}

	// Search in subdirectories matching the model
	entries, err := os.ReadDir(s.fwDir)
	if err != nil {
		return "", fmt.Errorf("read firmware dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := strings.ToLower(entry.Name())
		modelLower := strings.ToLower(model)
		if !strings.Contains(dirName, modelLower) {
			continue
		}

		// Search for firmware files in this subdirectory
		subDir := filepath.Join(s.fwDir, entry.Name())
		files, err := os.ReadDir(subDir)
		if err != nil {
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			if !strings.HasSuffix(name, ".bin") && !strings.HasSuffix(name, ".img") {
				continue
			}
			// If target version specified, match it
			if targetVer != "" && !strings.Contains(name, targetVer) {
				continue
			}
			return filepath.Join(entry.Name(), name), nil
		}
	}

	return "", fmt.Errorf("no firmware found for model=%s targetVer=%s", model, targetVer)
}

// FwVersionInfo holds version info for a firmware file from update.ini
type FwVersionInfo struct {
	EspVer   string
	RadarVer string
}

// parseConfigINI reads ota/update.ini and returns version info for a specific filename
// ParseUpdateINI reads ota/update.ini and returns version info for a specific filename
func ParseUpdateINI(fwDir, targetFile string) *FwVersionInfo {
	iniPath := filepath.Join(fwDir, "update.ini")
	data, err := os.ReadFile(iniPath)
	if err != nil {
		return nil
	}
	// Find the section for targetFile
	var info *FwVersionInfo
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "file:") {
			// 已解析完目标 section → 遇到下一个 file: 就返回；否则继续寻找
			if info != nil {
				return info
			}
			fname := strings.TrimSpace(strings.TrimPrefix(line, "file:"))
			if fname == targetFile {
				info = &FwVersionInfo{}
			}
			continue
		}
		if info == nil {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		switch parts[0] {
		case "espsfver":
			info.EspVer = strings.TrimSpace(parts[1])
		case "radarsfver":
			info.RadarVer = strings.TrimSpace(parts[1])
		}
	}
	// 目标 section 是文件最后一段
	return info
}

// getFileSize returns the size of a file
func getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// getFileSHA256 returns the SHA256 hash of a file
func getFileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ParseSchedule parses a schedule string in MMDDHH+xH format
// e.g. "01150200+8H" means January 15, 02:00, 8 hour window
func ParseSchedule(s string) (month, day, hour, windowHours int, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0, 0, 0, fmt.Errorf("empty schedule")
	}

	// Split on '+'
	parts := strings.SplitN(s, "+", 2)
	if len(parts) != 2 {
		return 0, 0, 0, 0, fmt.Errorf("invalid schedule format: %s (expected MMDDHH+xH)", s)
	}

	datePart := parts[0]
	windowPart := strings.TrimSuffix(strings.ToUpper(parts[1]), "H")

	if len(datePart) != 6 {
		return 0, 0, 0, 0, fmt.Errorf("invalid date part: %s (expected MMDDHH)", datePart)
	}

	month, err = strconv.Atoi(datePart[0:2])
	if err != nil || month < 1 || month > 12 {
		return 0, 0, 0, 0, fmt.Errorf("invalid month: %s", datePart[0:2])
	}
	day, err = strconv.Atoi(datePart[2:4])
	if err != nil || day < 1 || day > 31 {
		return 0, 0, 0, 0, fmt.Errorf("invalid day: %s", datePart[2:4])
	}
	hour, err = strconv.Atoi(datePart[4:6])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, 0, 0, fmt.Errorf("invalid hour: %s", datePart[4:6])
	}
	windowHours, err = strconv.Atoi(windowPart)
	if err != nil || windowHours < 1 {
		return 0, 0, 0, 0, fmt.Errorf("invalid window: %s", windowPart)
	}

	return month, day, hour, windowHours, nil
}

// IsInScheduleWindow checks if the given time is within the schedule window.
// 支持三种格式：
//   1. "auto"        → 永远在窗口内（持续自动升级，直到手动改或 status=success）
//   2. "now+Nd"/"Nh" → 从 updatedAt 起 N 天/小时内有效（有明确截止）
//   3. "MMDDHH+xH"   → 绝对时间窗（原格式）
func IsInScheduleWindow(schedule string, now time.Time, updatedAt time.Time) bool {
	s := strings.ToLower(strings.TrimSpace(schedule))
	if s == "" {
		return false
	}

	// 识别码 1: auto → 永远在窗口内
	if s == "auto" {
		return true
	}

	// 识别码 2: now+Nd 或 now+Nh → 从 updatedAt 起的窗口
	if strings.HasPrefix(s, "now+") {
		dur := strings.TrimPrefix(s, "now+")
		if len(dur) < 2 {
			return false
		}
		unit := dur[len(dur)-1]
		nStr := dur[:len(dur)-1]
		n, err := strconv.Atoi(nStr)
		if err != nil || n < 1 {
			return false
		}
		var window time.Duration
		switch unit {
		case 'd':
			window = time.Duration(n) * 24 * time.Hour
		case 'h':
			window = time.Duration(n) * time.Hour
		default:
			return false
		}
		return !now.Before(updatedAt) && now.Before(updatedAt.Add(window))
	}

	// 格式 3: 原 MMDDHH+xH 绝对时间窗
	month, day, hour, windowHours, err := ParseSchedule(schedule)
	if err != nil {
		return false
	}
	start := time.Date(now.Year(), time.Month(month), day, hour, 0, 0, 0, now.Location())
	end := start.Add(time.Duration(windowHours) * time.Hour)
	return now.After(start) && now.Before(end)
}
