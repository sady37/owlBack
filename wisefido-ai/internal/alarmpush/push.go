// Package alarmpush 与 wisefido-cardagg alarm_service.notifyAlarmPushAsync 同一条 HTTP 链路：
// POST wisefido-data /internal/v1/push/alarm（APNs 仅 staff，见 wisefido-data APNSDeviceService）。
// cardagg 无独立 HTTP 服务，故 AI 与 cardagg 均直接调 data；责任与推送范围一致。
package alarmpush

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// NotifyWisefidoData 异步通知 wisefido-data 触发 APNs（env：WISEFIDO_DATA_ALARM_PUSH_URL、INTERNAL_ALARM_PUSH_SECRET）
func NotifyWisefidoData(logger *zap.Logger, tenantID, cardID, deviceID, eventID, eventType, alarmLevel string) {
	base := strings.TrimSpace(os.Getenv("WISEFIDO_DATA_ALARM_PUSH_URL"))
	sec := strings.TrimSpace(os.Getenv("INTERNAL_ALARM_PUSH_SECRET"))
	if base == "" || sec == "" || tenantID == "" || cardID == "" {
		return
	}
	base = strings.TrimRight(base, "/")
	payload := map[string]string{
		"tenant_id":   tenantID,
		"card_id":     cardID,
		"device_id":   deviceID,
		"event_id":    eventID,
		"event_type":  eventType,
		"alarm_level": alarmLevel,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	go func() {
		c, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(c, http.MethodPost, base+"/internal/v1/push/alarm", bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Internal-Secret", sec)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if logger != nil {
				logger.Warn("alarm push to wisefido-data failed", zap.Error(err))
			}
			return
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= 300 && logger != nil {
			logger.Warn("alarm push to wisefido-data bad status", zap.Int("status", resp.StatusCode))
		}
	}()
}

// LookupCardIDByDevice 与 owl-common/card findCardIDByDevice 一致（无事务版）
func LookupCardIDByDevice(ctx context.Context, db *sql.DB, deviceID string) (string, error) {
	if db == nil || deviceID == "" {
		return "", sql.ErrNoRows
	}
	var cardID string
	err := db.QueryRowContext(ctx, `
		SELECT card_id FROM cards
		WHERE devices @> $1::jsonb
		LIMIT 1
	`, fmt.Sprintf(`[{"device_id":"%s"}]`, deviceID)).Scan(&cardID)
	return cardID, err
}
