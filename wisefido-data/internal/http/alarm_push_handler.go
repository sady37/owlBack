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
		DeviceID   string `json:"device_id"`
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

	// v2: cards.card_id INET PK；tenant 由 card_id /48 派生（无独立 tenant_id 列）
	cardName := req.CardID
	var cn sql.NullString
	if err := h.db.QueryRowContext(r.Context(), `
		SELECT card_name FROM cards
		 WHERE card_id = $1::INET
		   AND set_masklen(card_id, 48) = $2::INET
	`, req.CardID, req.TenantID).Scan(&cn); err == nil && cn.Valid && cn.String != "" {
		cardName = cn.String
	}

	n := service.AlarmNotification{
		TenantID:   req.TenantID,
		CardID:     req.CardID,
		CardName:   cardName,
		EventType:  req.EventType,
		AlarmLevel: service.AlarmLevelStringToPushIndex(req.AlarmLevel),
	}
	go h.apnsSvc.SendAlarmPush(context.Background(), n)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}
