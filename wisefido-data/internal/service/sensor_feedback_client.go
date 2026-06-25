package service

// sensor_feedback_client.go — alarm 反馈事件触发通道。
//
// 护士 handle 告警后，wisefido-data 直调 sensor live engine HTTP 端点即时处理该条反馈
// （单一消费者，不走 Redis；与 veto 通道同机制 [[layout_authority_ai_correction_model]]）。
// best-effort：sensor 不可达仅 warn，不阻塞 handle 成功——无批处理兜底，丢即丢（event_id 去重在 sensor 侧）。

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// notifySensorFeedback POST sensor /roomengine/feedback/ingest {event_id}。非 fall 类 sensor 自行忽略。
func (s *alarmEventService) notifySensorFeedback(ctx context.Context, eventID string) {
	body, _ := json.Marshal(map[string]string{"event_id": eventID})
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, sensorHTTPBaseURL()+"/roomengine/feedback/ingest", bytes.NewReader(body))
	if err != nil {
		s.logger.Warn("notifySensorFeedback: build request", zap.String("event_id", eventID), zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.Warn("notifySensorFeedback: call sensor failed (feedback dropped, no batch fallback)",
			zap.String("event_id", eventID), zap.Error(err))
		return
	}
	defer resp.Body.Close()
	s.logger.Info("feedback_notified_sensor",
		zap.String("event_id", eventID), zap.Int("status", resp.StatusCode))
}
