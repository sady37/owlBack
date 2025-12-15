package httpapi

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

func (s *StubHandler) AdminServiceLevels(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin/api/v1/service-levels" || r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// DB-backed read for service levels (global lookup table, no tenant_id needed)
	if s != nil && s.DB != nil {
		rows, err := s.DB.QueryContext(
			r.Context(),
			`SELECT level_code, description, color, color_hex, priority
			 FROM service_levels
			 ORDER BY priority ASC, level_code ASC`,
		)
		if err != nil {
			fmt.Printf("[AdminServiceLevels] Query error: %v\n", err)
			writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to list service levels: %v", err)))
			return
		}
		defer rows.Close()
		items := []any{}
		for rows.Next() {
			var levelCode, desc, color string
			var colorHex sql.NullString
			var priority int
			if err := rows.Scan(&levelCode, &desc, &color, &colorHex, &priority); err != nil {
				fmt.Printf("[AdminServiceLevels] Scan error: %v\n", err)
				writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to scan service level: %v", err)))
				return
			}
			item := map[string]any{
				"level_code":  levelCode,
				"description": desc,
				"color":       color,
				"priority":    priority,
			}
			if colorHex.Valid {
				item["color_hex"] = colorHex.String
			}
			items = append(items, item)
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"items": items, "total": len(items)}))
		return
	}
	writeJSON(w, http.StatusOK, Fail("database not available"))
}

func (s *StubHandler) AdminCardOverview(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin/api/v1/card-overview" || r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 对齐 GetCardOverviewResult
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": []any{},
		"pagination": map[string]any{
			"size":      10,
			"page":      1,
			"count":     0,
			"total":     0,
			"sort":      "",
			"direction": 0,
		},
	}))
}
func (s *StubHandler) SettingsMonitor(w http.ResponseWriter, r *http.Request) {
	// /settings/api/v1/monitor/sleepace/:deviceId
	// /settings/api/v1/monitor/radar/:deviceId
	if strings.HasPrefix(r.URL.Path, "/settings/api/v1/monitor/sleepace/") {
		if r.Method == http.MethodGet || r.Method == http.MethodPut {
			// 返回 SleepaceMonitorSettings（字段较多，给默认值即可）
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"left_bed_start_hour": 0, "left_bed_start_minute": 0,
				"left_bed_end_hour": 0, "left_bed_end_minute": 0,
				"left_bed_duration": 0, "left_bed_alarm_level": "disabled",
				"min_heart_rate": 0, "heart_rate_slow_duration": 0, "heart_rate_slow_alarm_level": "disabled",
				"max_heart_rate": 0, "heart_rate_fast_duration": 0, "heart_rate_fast_alarm_level": "disabled",
				"min_breath_rate": 0, "breath_rate_slow_duration": 0, "breath_rate_slow_alarm_level": "disabled",
				"max_breath_rate": 0, "breath_rate_fast_duration": 0, "breath_rate_fast_alarm_level": "disabled",
				"breath_pause_duration": 0, "breath_pause_alarm_level": "disabled",
				"body_move_duration": 0, "body_move_alarm_level": "disabled",
				"nobody_move_duration": 0, "nobody_move_alarm_level": "disabled",
				"no_turn_over_duration": 0, "no_turn_over_alarm_level": "disabled",
				"situp_alarm_level": "disabled",
				"onbed_duration":    0, "onbed_alarm_level": "disabled",
				"fall_alarm_level": "disabled",
			}))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/settings/api/v1/monitor/radar/") {
		if r.Method == http.MethodGet || r.Method == http.MethodPut {
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"radar_function_mode":     0,
				"suspected_fall_duration": 0, "fall_alarm_level": "disabled",
				"posture_detection_alarm_level": "disabled",
				"sitting_on_ground_duration":    0, "sitting_on_ground_alarm_level": "disabled",
				"stay_detection_duration": 0, "stay_alarm_level": "disabled",
				"leave_detection_start_hour": 0, "leave_detection_start_minute": 0,
				"leave_detection_end_hour": 0, "leave_detection_end_minute": 0,
				"leave_detection_duration": 0, "leave_alarm_level": "disabled",
				"lower_heart_rate": 0, "heart_rate_slow_alarm_level": "disabled",
				"upper_heart_rate": 0, "heart_rate_fast_alarm_level": "disabled",
				"lower_breath_rate": 0, "breath_rate_slow_alarm_level": "disabled",
				"upper_breath_rate": 0, "breath_rate_fast_alarm_level": "disabled",
				"breath_pause_alarm_level": "disabled",
				"weak_vital_duration":      0, "weak_vital_sensitivity": 0, "weak_vital_alarm_level": "disabled",
				"inactivity_alarm_level": "disabled",
			}))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
func (s *StubHandler) SleepaceReports(w http.ResponseWriter, r *http.Request) {
	// GET /sleepace/api/v1/sleepace/reports/:id
	// GET /sleepace/api/v1/sleepace/reports/:id/detail
	// GET /sleepace/api/v1/sleepace/reports/:id/dates
	if !strings.HasPrefix(r.URL.Path, "/sleepace/api/v1/sleepace/reports/") || r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/detail") {
		writeJSON(w, http.StatusOK, Ok(map[string]any{
			"id": 0, "deviceId": "", "deviceCode": "", "recordCount": 0,
			"startTime": 0, "endTime": 0, "date": 0, "stopMode": 0, "timeStep": 0, "timezone": 0,
			"report": "",
		}))
		return
	}
	if strings.HasSuffix(r.URL.Path, "/dates") {
		writeJSON(w, http.StatusOK, Ok([]int{}))
		return
	}
	// list
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"items": []any{},
		"pagination": map[string]any{
			"size": 10, "page": 1, "count": 0, "sort": "", "direction": 0,
		},
	}))
}
func (s *StubHandler) DeviceRelations(w http.ResponseWriter, r *http.Request) {
	// GET /device/api/v1/device/:id/relations
	if !strings.HasPrefix(r.URL.Path, "/device/api/v1/device/") || !strings.HasSuffix(r.URL.Path, "/relations") {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/device/api/v1/device/"), "/relations")
	writeJSON(w, http.StatusOK, Ok(map[string]any{
		"deviceId": id, "deviceName": "stub", "deviceInternalCode": "", "deviceType": 0,
		"addressId": "", "addressName": "", "addressType": 0,
		"residents": []any{},
	}))
}
func (s *StubHandler) Example(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api/v1/example/items" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, Ok(map[string]any{"items": []any{}, "total": 0}))
		return
	case strings.HasPrefix(r.URL.Path, "/api/v1/example/") && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, Ok(map[string]any{}))
		return
	case r.URL.Path == "/api/v1/example/item" && r.Method == http.MethodPost:
		writeJSON(w, http.StatusOK, Fail("database not available"))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
