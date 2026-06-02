package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"wisefido-data/internal/service"

	"go.uber.org/zap"
)

// AlarmPushHandler cardagg → wisefido-data，触发 iOS APNs。
// 网页告警列表仍为 AlarmEventHandler：/admin/api/v1/alarm-events。
// POST /internal/v1/push/alarm，Header: X-Internal-Secret（/internal/ 跳过 JWT）
type AlarmPushHandler struct {
	apnsSvc *service.APNSDeviceService
	db      *sql.DB
	secret  string
	logger  *zap.Logger
}

func NewAlarmPushHandler(db *sql.DB, apnsSvc *service.APNSDeviceService, logger *zap.Logger) *AlarmPushHandler {
	return &AlarmPushHandler{
		db:      db,
		apnsSvc: apnsSvc,
		secret:  strings.TrimSpace(os.Getenv("INTERNAL_ALARM_PUSH_SECRET")),
		logger:  logger,
	}
}

func (h *AlarmPushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.secret == "" || r.Header.Get("X-Internal-Secret") != h.secret {
		if h.logger != nil {
			h.logger.Warn("[AlarmPush] invalid or missing secret", zap.String("remote", r.RemoteAddr))
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if h.apnsSvc == nil || h.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"skipped":true}`))
		return
	}

	var req struct {
		TenantID   string `json:"tenant_id"`
		CardID     string `json:"card_id"`
		DeviceAddr string `json:"device_addr"`
		EventID    string `json:"event_id"`
		EventType  string `json:"event_type"`
		AlarmLevel string `json:"alarm_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.TenantID == "" || req.CardID == "" || req.EventType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	n := service.AlarmNotification{
		TenantID:   req.TenantID,
		CardID:     req.CardID,
		CardName:   h.resolveDisplayName(r.Context(), req.EventID),
		EventType:  req.EventType,
		AlarmLevel: service.AlarmLevelStringToPushIndex(req.AlarmLevel),
	}
	go h.apnsSvc.SendAlarmPush(context.Background(), n)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

// resolveDisplayName 组装锁屏可读的推送显示名：Unit/Room/DeviceType（alarm_events 快照），
// 缺哪段跳哪段，绝不暴露 card_id(INET) / hex uid。住户姓名是 PHI，刻意不进锁屏推送；
// 护理员 app 内鉴权后再看。event_name 在 push body 单独给。
func (h *AlarmPushHandler) resolveDisplayName(ctx context.Context, eventID string) string {
	var unit, room, devType string
	if eventID != "" {
		_ = h.db.QueryRowContext(ctx, `
			SELECT COALESCE(ae.unit_name, ''), COALESCE(ae.room_name, ''),
			       COALESCE(dfm.device_type::text, '')
			FROM alarm_events ae
			LEFT JOIN device_factory_meta dfm ON dfm.device_uid = ae.device_uid
			WHERE ae.event_id = $1::uuid
		`, eventID).Scan(&unit, &room, &devType)
	}
	if name := joinNonEmpty("/", unit, room, devType); name != "" {
		return name
	}
	return "Alarm"
}

func joinNonEmpty(sep string, parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}
