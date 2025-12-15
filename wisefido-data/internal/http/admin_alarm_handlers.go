package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (s *StubHandler) AdminAlarm(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/admin/api/v1/alarm-cloud":
		if r.Method == http.MethodGet {
			if s != nil && s.DB != nil {
				// Get tenant_id from query or header
				tenantID := r.URL.Query().Get("tenant_id")
				if tenantID == "" || tenantID == "null" {
					tenantID = r.Header.Get("X-Tenant-Id")
				}
				// Normalize: empty string or "null" means use SystemTenantID
				if tenantID == "" || tenantID == "null" {
					tenantID = SystemTenantID()
				}
				// Query alarm_cloud table
				var offlineAlarm, lowBattery, deviceFailure sql.NullString
				var deviceAlarmsRaw, conditionsRaw, notificationRulesRaw []byte
				var found bool
				err := s.DB.QueryRowContext(
					r.Context(),
					`SELECT OfflineAlarm, LowBattery, DeviceFailure, device_alarms, conditions, notification_rules
					 FROM alarm_cloud WHERE tenant_id = $1`,
					tenantID,
				).Scan(&offlineAlarm, &lowBattery, &deviceFailure, &deviceAlarmsRaw, &conditionsRaw, &notificationRulesRaw)
				if err == nil {
					found = true
				} else if err == sql.ErrNoRows {
					// If not found, try SystemTenantID as fallback
					if tenantID != SystemTenantID() {
						err = s.DB.QueryRowContext(
							r.Context(),
							`SELECT OfflineAlarm, LowBattery, DeviceFailure, device_alarms, conditions, notification_rules
							 FROM alarm_cloud WHERE tenant_id = $1`,
							SystemTenantID(),
						).Scan(&offlineAlarm, &lowBattery, &deviceFailure, &deviceAlarmsRaw, &conditionsRaw, &notificationRulesRaw)
						if err == nil {
							found = true
							tenantID = SystemTenantID() // Update tenant_id to reflect the actual source
						}
					}
				}
				if !found {
					// Return empty config if not found (should not happen if init script was run)
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
				var deviceAlarms map[string]any
				if len(deviceAlarmsRaw) > 0 {
					_ = json.Unmarshal(deviceAlarmsRaw, &deviceAlarms)
				} else {
					deviceAlarms = map[string]any{}
				}
				var conditions any
				if len(conditionsRaw) > 0 {
					_ = json.Unmarshal(conditionsRaw, &conditions)
				}
				var notificationRules any
				if len(notificationRulesRaw) > 0 {
					_ = json.Unmarshal(notificationRulesRaw, &notificationRules)
				}
				result := map[string]any{
					"tenant_id":     tenantID,
					"device_alarms": deviceAlarms,
				}
				if offlineAlarm.Valid {
					result["OfflineAlarm"] = offlineAlarm.String
				}
				if lowBattery.Valid {
					result["LowBattery"] = lowBattery.String
				}
				if deviceFailure.Valid {
					result["DeviceFailure"] = deviceFailure.String
				}
				if conditions != nil {
					result["conditions"] = conditions
				}
				if notificationRules != nil {
					result["notification_rules"] = notificationRules
				}
				writeJSON(w, http.StatusOK, Ok(result))
				return
			}
			// No DB: return stub
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
				// Get tenant_id from payload or header
				tenantID, _ := payload["tenant_id"].(string)
				if tenantID == "" {
					tenantID = r.Header.Get("X-Tenant-Id")
				}
				if tenantID == "" {
					tenantID = SystemTenantID()
				}
				// Parse fields
				var offlineAlarm, lowBattery, deviceFailure sql.NullString
				if val, ok := payload["OfflineAlarm"].(string); ok && val != "" {
					offlineAlarm = sql.NullString{String: val, Valid: true}
				}
				if val, ok := payload["LowBattery"].(string); ok && val != "" {
					lowBattery = sql.NullString{String: val, Valid: true}
				}
				if val, ok := payload["DeviceFailure"].(string); ok && val != "" {
					deviceFailure = sql.NullString{String: val, Valid: true}
				}
				var deviceAlarmsJSON []byte
				if val, ok := payload["device_alarms"].(map[string]any); ok && val != nil {
					deviceAlarmsJSON, _ = json.Marshal(val)
				} else {
					deviceAlarmsJSON = []byte("{}")
				}
				var conditionsJSON []byte
				if val, ok := payload["conditions"]; ok && val != nil {
					conditionsJSON, _ = json.Marshal(val)
				}
				var notificationRulesJSON []byte
				if val, ok := payload["notification_rules"]; ok && val != nil {
					notificationRulesJSON, _ = json.Marshal(val)
				}
				// Upsert: INSERT ... ON CONFLICT DO UPDATE
				_, err := s.DB.ExecContext(
					r.Context(),
					`INSERT INTO alarm_cloud (tenant_id, OfflineAlarm, LowBattery, DeviceFailure, device_alarms, conditions, notification_rules)
					 VALUES ($1, $2, $3, $4, $5, $6, $7)
					 ON CONFLICT (tenant_id) DO UPDATE SET
					   OfflineAlarm = EXCLUDED.OfflineAlarm,
					   LowBattery = EXCLUDED.LowBattery,
					   DeviceFailure = EXCLUDED.DeviceFailure,
					   device_alarms = EXCLUDED.device_alarms,
					   conditions = EXCLUDED.conditions,
					   notification_rules = EXCLUDED.notification_rules`,
					tenantID, offlineAlarm, lowBattery, deviceFailure, deviceAlarmsJSON, conditionsJSON, notificationRulesJSON,
				)
				if err != nil {
					writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to update alarm_cloud: %v", err)))
					return
				}
				// Return updated config
				var resultOfflineAlarm, resultLowBattery, resultDeviceFailure sql.NullString
				var resultDeviceAlarmsRaw, resultConditionsRaw, resultNotificationRulesRaw []byte
				_ = s.DB.QueryRowContext(
					r.Context(),
					`SELECT OfflineAlarm, LowBattery, DeviceFailure, device_alarms, conditions, notification_rules
					 FROM alarm_cloud WHERE tenant_id = $1`,
					tenantID,
				).Scan(&resultOfflineAlarm, &resultLowBattery, &resultDeviceFailure, &resultDeviceAlarmsRaw, &resultConditionsRaw, &resultNotificationRulesRaw)
				var resultDeviceAlarms map[string]any
				if len(resultDeviceAlarmsRaw) > 0 {
					_ = json.Unmarshal(resultDeviceAlarmsRaw, &resultDeviceAlarms)
				} else {
					resultDeviceAlarms = map[string]any{}
				}
				var resultConditions any
				if len(resultConditionsRaw) > 0 {
					_ = json.Unmarshal(resultConditionsRaw, &resultConditions)
				}
				var resultNotificationRules any
				if len(resultNotificationRulesRaw) > 0 {
					_ = json.Unmarshal(resultNotificationRulesRaw, &resultNotificationRules)
				}
				result := map[string]any{
					"tenant_id":     tenantID,
					"device_alarms": resultDeviceAlarms,
				}
				if resultOfflineAlarm.Valid {
					result["OfflineAlarm"] = resultOfflineAlarm.String
				}
				if resultLowBattery.Valid {
					result["LowBattery"] = resultLowBattery.String
				}
				if resultDeviceFailure.Valid {
					result["DeviceFailure"] = resultDeviceFailure.String
				}
				if resultConditions != nil {
					result["conditions"] = resultConditions
				}
				if resultNotificationRules != nil {
					result["notification_rules"] = resultNotificationRules
				}
				writeJSON(w, http.StatusOK, Ok(result))
				return
			}
			// No DB: return stub
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
