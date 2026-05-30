// Package alarmpush cardagg → wisefido-data 的 alarm 推送 trigger。
//
// 职责：cardagg 写入 alarm_events 成功后异步通知 wisefido-data 触发 APNs。
//   - 单源原则：alarm 由 cardagg 持久化 → 推送 trigger 也由 cardagg 发；sensor 不再持有此责
//   - wisefido-data 只做 APNs HTTPS gateway + apns_devices 寿命管理
//   - 过滤链（notify_mode / work_time / 节假日 / login_only）在 wisefido-data 端
//     用 SQL should_notify_user() 统一执行
//
// 触发条件（由 caller 控制）：
//   - alarm_events INSERT 成功（非 dedup / 非 SkippedNotify）
//   - tenant_id 和 card_id 均存在
//
// 配置：WISEFIDO_DATA_ALARM_PUSH_URL / INTERNAL_ALARM_PUSH_SECRET 任一未设直接 no-op，
// 便于 dev/test 环境无 APNs 凭证时正常跑业务流程。
package alarmpush

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// NotifyWisefidoData 异步触发 APNs；env 缺失时静默 no-op。
func NotifyWisefidoData(logger *zap.Logger, tenantID, cardID, deviceAddr, eventID, eventType, alarmLevel string) {
	base := strings.TrimSpace(os.Getenv("WISEFIDO_DATA_ALARM_PUSH_URL"))
	sec := strings.TrimSpace(os.Getenv("INTERNAL_ALARM_PUSH_SECRET"))
	if base == "" || sec == "" || tenantID == "" || cardID == "" {
		return
	}
	base = strings.TrimRight(base, "/")
	payload := map[string]string{
		"tenant_id":   tenantID,
		"card_id":     cardID,
		"device_addr": deviceAddr,
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
