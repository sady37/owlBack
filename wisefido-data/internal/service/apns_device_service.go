package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"wisefido-data/internal/notify"

	"go.uber.org/zap"
)

// APNSDeviceService iOS 设备 Token 管理 + 推送发送
// DB 操作（apns_devices 表）+ 调用 notify.APNSSender
type APNSDeviceService struct {
	db         *sql.DB
	sender     *notify.APNSSender // nil = APNs 未配置，静默跳过（不影响启动）
	logger     *zap.Logger
	popCounter *StaffPopPendingCounter // nil 可选：不查库，aps.badge 恒按告警至少为 1
}

// NewAPNSDeviceService popCounter 可传 nil（仅关闭待 pop 统计注入，不影响编译与其它逻辑）。
func NewAPNSDeviceService(db *sql.DB, sender *notify.APNSSender, logger *zap.Logger, popCounter *StaffPopPendingCounter) *APNSDeviceService {
	return &APNSDeviceService{db: db, sender: sender, logger: logger, popCounter: popCounter}
}

// Register 注册或更新 iOS 设备 token
// ON CONFLICT(device_token) 直接 upsert，同一设备重复注册无副作用
func (s *APNSDeviceService) Register(
	ctx context.Context,
	tenantID, userID, userType, deviceToken, env string,
) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO apns_devices
		    (tenant_id, user_id, user_type, device_token, environment, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, TRUE, NOW())
		ON CONFLICT (device_token) DO UPDATE SET
		    tenant_id   = EXCLUDED.tenant_id,
		    user_id     = EXCLUDED.user_id,
		    user_type   = EXCLUDED.user_type,
		    environment = EXCLUDED.environment,
		    is_active   = TRUE,
		    updated_at  = NOW()
	`, tenantID, userID, userType, deviceToken, env)
	return err
}

// Unregister 登出时标记 token 为 inactive
func (s *APNSDeviceService) Unregister(ctx context.Context, tenantID, deviceToken string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE apns_devices
		SET is_active = FALSE, updated_at = NOW()
		WHERE tenant_id = $1 AND device_token = $2
	`, tenantID, deviceToken)
	return err
}

// deactivateStaleToken APNs 返回 410 时自动失效（goroutine 内调用，不对外暴露）
func (s *APNSDeviceService) deactivateStaleToken(deviceToken string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = s.db.ExecContext(ctx, `
		UPDATE apns_devices
		SET is_active = FALSE, updated_at = NOW()
		WHERE device_token = $1
	`, deviceToken)
	s.logger.Info("[APNS] stale token deactivated",
		zap.String("token_prefix", tokenPrefix(deviceToken)))
}

// AlarmNotification 报警推送参数
type AlarmNotification struct {
	TenantID   string
	CardID     string
	CardName   string
	EventType  string
	AlarmLevel int // 0=EMERG 1=ALERT 2=CRITICAL 3=ERROR 4=WARNING
}

// SendAlarmPush 向租户下所有活跃 staff 异步推送报警（仅 user_type=staff；不向住户/联系人推送）
// 若 sender 为 nil（未配置 APNs），静默返回
func (s *APNSDeviceService) SendAlarmPush(ctx context.Context, n AlarmNotification) {
	if s.sender == nil {
		return
	}

	type row struct{ token, env, userID string }
	var devices []row

	rows, err := s.db.QueryContext(ctx, `
		SELECT device_token, environment, user_id
		FROM   apns_devices
		WHERE  tenant_id = $1
		  AND  user_type = 'staff'
		  AND  is_active = TRUE
	`, n.TenantID)
	if err != nil {
		s.logger.Warn("[APNS] query devices failed", zap.Error(err))
		return
	}
	defer rows.Close()
	for rows.Next() {
		var d row
		if err := rows.Scan(&d.token, &d.env, &d.userID); err == nil && d.token != "" {
			devices = append(devices, d)
		}
	}
	if len(devices) == 0 {
		return
	}

	for _, d := range devices {
		go func(token, env, staffUserID string) {
			sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			badge := 1
			if s.popCounter != nil && staffUserID != "" {
				nPop, err := s.popCounter.CountForStaff(sendCtx, n.TenantID, staffUserID)
				if err != nil {
					s.logger.Debug("[APNS] pop pending count failed", zap.Error(err))
				} else if nPop > 0 {
					badge = nPop
				}
			}
			payload := s.buildPayload(n, badge)
			err := s.sender.Send(sendCtx, token, env, payload)
			if err == nil {
				s.logger.Info("[APNS] push sent",
					zap.String("token_prefix", tokenPrefix(token)),
					zap.String("event", n.EventType))
				return
			}
			if err == notify.ErrDeviceTokenInvalid {
				s.deactivateStaleToken(token)
			} else if strings.Contains(err.Error(), "BadDeviceToken") {
				// Treat APNs 400 BadDeviceToken as stale and deactivate to avoid DB pollution
				s.deactivateStaleToken(token)
			} else {
				s.logger.Warn("[APNS] send failed",
					zap.String("token_prefix", tokenPrefix(token)),
					zap.Error(err))
			}
		}(d.token, d.env, d.userID)
	}
}

var alarmLevelLabel = [5]string{"EMERG", "ALERT", "CRITICAL", "ERROR", "WARNING"}

// buildPayload aps.badge = 该 staff 可见「待 pop」卡数（与 alarm_state.pop_alarm+event_id 同源）；告警推送至少 1。
// content-available=1 便于后台拉齐角标，不替代 alert。
func (s *APNSDeviceService) buildPayload(n AlarmNotification, badge int) notify.APNSPayload {
	label := "报警"
	if n.AlarmLevel >= 0 && n.AlarmLevel <= 4 {
		label = alarmLevelLabel[n.AlarmLevel]
	}
	if badge < 1 {
		badge = 1
	}
	b := badge

	p := notify.APNSPayload{
		APS: notify.APSDict{
			Alert: notify.APSAlert{
				Title: fmt.Sprintf("[%s] %s", label, n.CardName),
				Body:  n.EventType,
			},
			Sound:            "default",
			Badge:            &b,
			ContentAvailable: 1,
		},
		CardID:     n.CardID,
		EventType:  n.EventType,
		AlarmLevel: n.AlarmLevel,
	}
	if n.AlarmLevel <= 1 {
		p.APS.Sound = "critical.caf"
		p.APS.InterruptionLevel = "critical"
	}
	return p
}

func tokenPrefix(token string) string {
	if len(token) >= 8 {
		return token[:8]
	}
	return token
}
