package ota

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// OTAPushFn is the function signature for pushing OTA via MQTT
type OTAPushFn func(uid string, data map[string]interface{}) error

// Scheduler periodically scans for devices that need OTA updates
type Scheduler struct {
	db        *sql.DB
	otaPushFn OTAPushFn
	interval  time.Duration
	fwDir     string
	fwURL     string

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewScheduler creates a new OTA scheduler
func NewScheduler(db *sql.DB, otaPushFn OTAPushFn, interval time.Duration, fwDir, fwURL string) *Scheduler {
	return &Scheduler{
		db:        db,
		otaPushFn: otaPushFn,
		interval:  interval,
		fwDir:     fwDir,
		fwURL:     fwURL,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.wg.Add(1)
	go s.run(ctx)
	log.Printf("[OTA-Scheduler] started, interval=%v fwDir=%s", s.interval, s.fwDir)
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	log.Printf("[OTA-Scheduler] stopped")
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

// scan queries device_store for devices eligible for OTA and pushes firmware via MQTT
func (s *Scheduler) scan(ctx context.Context) {
	if s.otaPushFn == nil {
		return
	}

	// Query devices with ota_permit='true' and ota_way in (schedule, tenant)
	// that have been approved and are idle/pending
	query := `
		SELECT device_uid, COALESCE(device_model, '') as device_model,
		       COALESCE(firmware_version, '') as firmware_version,
		       COALESCE(ota_way, '') as ota_way,
		       COALESCE(ota_schedule, '') as ota_schedule,
		       COALESCE(ota_target_firmware_version, '') as ota_target_fw,
		       COALESCE(ota_status, 'idle') as ota_status
		FROM device_store
		WHERE ota_permit = 'true'
		  AND ota_way IN ('schedule', 'tenant')
		  AND ota_tenant_approved = true
		  AND COALESCE(ota_status, 'idle') IN ('idle', 'pending')
		  AND allow_access = true
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[OTA-Scheduler] query failed: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	pushed := 0
	for rows.Next() {
		var uid, model, fwVer, way, schedule, targetVer, status string
		if err := rows.Scan(&uid, &model, &fwVer, &way, &schedule, &targetVer, &status); err != nil {
			log.Printf("[OTA-Scheduler] scan row failed: %v", err)
			continue
		}
		count++

		// For schedule mode, check if current time is within the schedule window
		if way == "schedule" && schedule != "" {
			if !IsInScheduleWindow(schedule, time.Now()) {
				log.Printf("[OTA-Scheduler] uid=%s schedule=%s not in window, skipping", uid, schedule)
				continue
			}
		}

		// Find firmware file for this device model
		fwFile, err := s.findFirmware(model, targetVer)
		if err != nil {
			log.Printf("[OTA-Scheduler] uid=%s model=%s no firmware found: %v", uid, model, err)
			continue
		}

		// Get file info
		fwPath := filepath.Join(s.fwDir, fwFile)
		size, err := getFileSize(fwPath)
		if err != nil {
			log.Printf("[OTA-Scheduler] uid=%s firmware size error: %v", uid, err)
			continue
		}
		sha, err := getFileSHA256(fwPath)
		if err != nil {
			log.Printf("[OTA-Scheduler] uid=%s firmware hash error: %v", uid, err)
			continue
		}

		// Build MQTT OTA data payload
		downloadURL := s.fwURL + "/" + fwFile
		version := targetVer
		if version == "" {
			version = filepath.Base(fwFile)
		}

		data := map[string]interface{}{
			"espfileUrl":    downloadURL,
			"espfilesha256": sha,
			"espfilesize":   size,
			"espver":        version,
		}

		log.Printf("[OTA-Scheduler] pushing OTA to uid=%s model=%s firmware=%s version=%s",
			uid, model, fwFile, version)

		if err := s.otaPushFn(uid, data); err != nil {
			log.Printf("[OTA-Scheduler] push failed uid=%s: %v", uid, err)
			continue
		}

		// Update ota_status to 'pushing'
		_, err = s.db.ExecContext(ctx, `
			UPDATE device_store SET ota_status = 'pushing', ota_updated_at = NOW()
			WHERE device_uid = $1
		`, uid)
		if err != nil {
			log.Printf("[OTA-Scheduler] update status failed uid=%s: %v", uid, err)
		}

		pushed++
	}

	if count > 0 {
		log.Printf("[OTA-Scheduler] scan complete: %d eligible, %d pushed", count, pushed)
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

// IsInScheduleWindow checks if the given time is within the schedule window
func IsInScheduleWindow(schedule string, now time.Time) bool {
	month, day, hour, windowHours, err := ParseSchedule(schedule)
	if err != nil {
		return false
	}

	// Build start time for this year
	start := time.Date(now.Year(), time.Month(month), day, hour, 0, 0, 0, now.Location())
	end := start.Add(time.Duration(windowHours) * time.Hour)

	return now.After(start) && now.Before(end)
}
