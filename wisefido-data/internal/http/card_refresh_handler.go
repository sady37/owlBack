package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	rediscommon "owl-common/redis"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"wisefido-data/internal/service"
)

// CardRefreshHandler 处理 POST /admin/api/v1/cards/{cardId}/refresh-devices：
// 把该卡所有设备 publish 到 iot:probe:device:stream，触发 wisefido-qinglan/wisefido-sleepace
// 立刻跑一次 health_check（不等 80s/10min 周期 ticker），缩短前端 refresh→状态变更延迟。
type CardRefreshHandler struct {
	db          *sql.DB
	redisClient *redislib.Client
	logger      *zap.Logger
}

func NewCardRefreshHandler(db *sql.DB, redisClient *redislib.Client, logger *zap.Logger) *CardRefreshHandler {
	return &CardRefreshHandler{db: db, redisClient: redisClient, logger: logger}
}

func (h *CardRefreshHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Fail("method not allowed"))
		return
	}

	ctx := r.Context()
	_, tenantID, _, _, ok := service.MustSession(ctx)
	if !ok || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, Fail("missing or invalid authorization"))
		return
	}

	// URL: /admin/api/v1/cards/{cardId}/refresh-devices
	path := strings.TrimSuffix(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 6 || parts[len(parts)-1] != "refresh-devices" {
		writeJSON(w, http.StatusNotFound, Fail("not found"))
		return
	}
	cardID := strings.TrimSpace(parts[len(parts)-2])
	if cardID == "" {
		writeJSON(w, http.StatusBadRequest, Fail("missing card_id"))
		return
	}

	devices, err := h.loadCardDevices(ctx, tenantID, cardID)
	if err != nil {
		h.logger.Warn("card refresh: load devices failed",
			zap.String("tenant_id", tenantID),
			zap.String("card_id", cardID),
			zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, Fail("load devices failed"))
		return
	}
	if len(devices) == 0 {
		writeJSON(w, http.StatusOK, Ok(map[string]interface{}{"requested": 0, "card_id": cardID}))
		return
	}

	requested := 0
	nowMs := time.Now().UnixMilli()
	for _, d := range devices {
		if d.DeviceUID == "" || d.DeviceAddr == "" {
			continue
		}
		values := map[string]interface{}{
			"device_uid":     d.DeviceUID,
			"device_addr":      d.DeviceAddr,
			"device_type":    d.DeviceType,
			"tenant_id":      tenantID,
			"card_id":        cardID,
			"trigger_source": "manual_refresh",
			"timestamp":      nowMs,
		}
		if _, err := rediscommon.PublishToStream(ctx, h.redisClient,
			rediscommon.StreamProbeDevice.Name, values,
			rediscommon.StreamProbeDevice.MaxLen, rediscommon.StreamProbeDevice.RetentionSeconds,
		); err != nil {
			h.logger.Warn("card refresh: publish probe failed",
				zap.String("device_uid", d.DeviceUID),
				zap.Error(err))
			continue
		}
		requested++
	}

	h.logger.Info("card refresh dispatched",
		zap.String("tenant_id", tenantID),
		zap.String("card_id", cardID),
		zap.Int("requested", requested),
		zap.Int("total_devices", len(devices)),
	)

	writeJSON(w, http.StatusOK, Ok(map[string]interface{}{
		"card_id":   cardID,
		"requested": requested,
	}))
}

type cardDeviceTuple struct {
	DeviceUID  string
	DeviceAddr   string
	DeviceType string
}

func (h *CardRefreshHandler) loadCardDevices(ctx context.Context, tenantID, cardID string) ([]cardDeviceTuple, error) {
	var devicesJSON []byte
	err := h.db.QueryRowContext(ctx,
		`SELECT devices FROM cards WHERE tenant_id = $1 AND card_id = $2`,
		tenantID, cardID,
	).Scan(&devicesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found")
		}
		return nil, err
	}
	if len(devicesJSON) == 0 {
		return nil, nil
	}
	var raw []map[string]interface{}
	if err := json.Unmarshal(devicesJSON, &raw); err != nil {
		return nil, fmt.Errorf("parse devices json: %w", err)
	}
	out := make([]cardDeviceTuple, 0, len(raw))
	for _, d := range raw {
		t := cardDeviceTuple{}
		if v, ok := d["device_uid"].(string); ok {
			t.DeviceUID = strings.TrimSpace(v)
		}
		if v, ok := d["device_addr"].(string); ok {
			t.DeviceAddr = strings.TrimSpace(v)
		}
		if v, ok := d["device_type"].(string); ok {
			t.DeviceType = strings.TrimSpace(v)
		}
		if t.DeviceUID == "" || t.DeviceAddr == "" {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}
