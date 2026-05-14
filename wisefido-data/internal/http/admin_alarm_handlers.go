package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AdminAlarm — admin/api/v1/alarm-cloud GET/PUT
//
// v2 schema：tenant 级 alarm 配置存 spatial_config 行 (spatial_prefix=tenant_prefix /48, config_key='alarm.cloud_config')，
// config_value JSONB 内含 v1 形字段（lowercase）：offlinealarm / lowbattery / devicefailure /
// device_alarms / conditions / notification_rules / metadata。
//
// 与 FE 兼容做大小写兜底：读时回填首字母大写 (OfflineAlarm 等) 让 FE 直接消费；
// 写时统一存 lowercase（v2 owl_v2 init script 落地的 alarm.cloud_config 已是 lowercase）。
const alarmCloudConfigKey = "alarm.cloud_config"

func (s *StubHandler) AdminAlarm(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/admin/api/v1/alarm-cloud":
		if r.Method == http.MethodGet {
			if s != nil && s.DB != nil {
				tenantID := r.URL.Query().Get("tenant_id")
				if tenantID == "" || tenantID == "null" {
					tenantID = r.Header.Get("X-Tenant-Id")
				}
				if tenantID == "" || tenantID == "null" {
					tenantID = SystemTenantID()
				}
				cfg, found, err := readAlarmCloudConfig(r, s.DB, tenantID)
				if err == nil && !found && tenantID != SystemTenantID() {
					// fallback to system tenant
					cfg, found, err = readAlarmCloudConfig(r, s.DB, SystemTenantID())
					if found {
						tenantID = SystemTenantID()
					}
				}
				if err != nil {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to read alarm_cloud: %v", err)))
					return
				}
				if !found {
					writeJSON(w, http.StatusOK, Ok(map[string]any{
						"tenant_id":          tenantID,
						"OfflineAlarm":       nil,
						"LowBattery":         nil,
						"DeviceFailure":      nil,
						"device_alarms":      map[string]any{},
						"conditions":         nil,
						"notification_rules": nil,
					}))
					return
				}
				writeJSON(w, http.StatusOK, Ok(alarmCloudConfigToResponse(tenantID, cfg)))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"tenant_id":     nil,
				"device_alarms": map[string]any{},
			}))
			return
		}
		if r.Method == http.MethodPut {
			if s != nil && s.DB != nil {
				var payload map[string]any
				if err := readBodyJSON(r, 1<<20, &payload); err != nil {
					writeJSON(w, http.StatusOK, Fail("invalid body"))
					return
				}
				tenantID, _ := payload["tenant_id"].(string)
				if tenantID == "" {
					tenantID = r.Header.Get("X-Tenant-Id")
				}
				if tenantID == "" {
					tenantID = SystemTenantID()
				}
				cfg := alarmCloudPayloadToConfig(payload)
				cfgJSON, _ := json.Marshal(cfg)
				_, err := s.DB.ExecContext(
					r.Context(),
					`INSERT INTO spatial_config (spatial_prefix, config_key, config_value, source)
					 VALUES ($1::INET, $2, $3::JSONB, 'manual_ui')
					 ON CONFLICT (spatial_prefix, config_key) DO UPDATE SET
					   config_value = EXCLUDED.config_value,
					   source = 'manual_ui',
					   updated_at = NOW()`,
					tenantID, alarmCloudConfigKey, cfgJSON,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update alarm_cloud: %v", err)))
					return
				}
				resultCfg, _, err := readAlarmCloudConfig(r, s.DB, tenantID)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to read alarm_cloud after update: %v", err)))
					return
				}
				writeJSON(w, http.StatusOK, Ok(alarmCloudConfigToResponse(tenantID, resultCfg)))
				return
			}
			writeJSON(w, http.StatusOK, Ok(map[string]any{
				"tenant_id":     nil,
				"device_alarms": map[string]any{},
			}))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	case r.URL.Path == "/admin/api/v1/alarm-events":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// 对齐 GetAlarmEventsResult
		writeJSON(w, http.StatusOK, Ok(map[string]any{
			"items": []any{},
			"pagination": map[string]any{
				"size":  10,
				"page":  1,
				"count": 0,
				"total": 0,
			},
		}))
		return
	case strings.HasPrefix(r.URL.Path, "/admin/api/v1/alarm-events/") && strings.HasSuffix(r.URL.Path, "/handle"):
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

// readAlarmCloudConfig 读取 tenant 级 alarm 配置 JSONB；返回 nil + found=false 表示无该 tenant 配置。
func readAlarmCloudConfig(r *http.Request, db *sql.DB, tenantID string) (map[string]any, bool, error) {
	var raw []byte
	err := db.QueryRowContext(r.Context(),
		`SELECT config_value FROM spatial_config WHERE spatial_prefix = $1::INET AND config_key = $2`,
		tenantID, alarmCloudConfigKey,
	).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	cfg := map[string]any{}
	if len(raw) > 0 {
		if e := json.Unmarshal(raw, &cfg); e != nil {
			return nil, false, e
		}
	}
	return cfg, true, nil
}

// alarmCloudConfigToResponse 把 spatial_config JSONB 翻成 v1 alarm_cloud 形（FE 期望的首字母大写键）。
func alarmCloudConfigToResponse(tenantID string, cfg map[string]any) map[string]any {
	out := map[string]any{
		"tenant_id": tenantID,
	}
	// JSONB 内字段为 lowercase；FE 期望 OfflineAlarm/LowBattery/DeviceFailure 三键首字母大写
	if v, ok := cfg["offlinealarm"].(string); ok && v != "" {
		out["OfflineAlarm"] = v
	}
	if v, ok := cfg["lowbattery"].(string); ok && v != "" {
		out["LowBattery"] = v
	}
	if v, ok := cfg["devicefailure"].(string); ok && v != "" {
		out["DeviceFailure"] = v
	}
	if v, ok := cfg["device_alarms"].(map[string]any); ok {
		out["device_alarms"] = v
	} else {
		out["device_alarms"] = map[string]any{}
	}
	if v, ok := cfg["conditions"]; ok && v != nil {
		out["conditions"] = v
	}
	if v, ok := cfg["notification_rules"]; ok && v != nil {
		out["notification_rules"] = v
	}
	return out
}

// alarmCloudPayloadToConfig FE PUT 入参（v1 形）→ spatial_config JSONB（lowercase）。
func alarmCloudPayloadToConfig(payload map[string]any) map[string]any {
	cfg := map[string]any{}
	if v, ok := payload["OfflineAlarm"].(string); ok && v != "" {
		cfg["offlinealarm"] = v
	}
	if v, ok := payload["LowBattery"].(string); ok && v != "" {
		cfg["lowbattery"] = v
	}
	if v, ok := payload["DeviceFailure"].(string); ok && v != "" {
		cfg["devicefailure"] = v
	}
	if v, ok := payload["device_alarms"].(map[string]any); ok && v != nil {
		cfg["device_alarms"] = v
	} else {
		cfg["device_alarms"] = map[string]any{}
	}
	if v, ok := payload["conditions"]; ok && v != nil {
		cfg["conditions"] = v
	}
	if v, ok := payload["notification_rules"]; ok && v != nil {
		cfg["notification_rules"] = v
	}
	return cfg
}
