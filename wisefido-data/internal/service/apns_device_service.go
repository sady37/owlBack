package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wisefido-data/internal/notify"

	"go.uber.org/zap"
)

// APNSDeviceService iOS 设备 Token 管理 + 推送发送
// DB 操作（apns_devices 表）+ 调用 notify.APNSSender
//
// v2 schema：apns_devices(user_id UUID, push_token, platform, environment, ...)
//   - 无 tenant_id 列 → SendAlarmPush 通过 JOIN users 用 INET prefix 过滤
//   - 无 user_type 列 → RegisterDevice HTTP gate 已限定 staff，DB 内隐式只剩 staff
//   - notify_mode 过滤走 should_notify_user(user_id, NOW())
type APNSDeviceService struct {
	db         *sql.DB
	sender     *notify.APNSSender // nil = APNs 未配置，静默跳过（不影响启动）
	fcm        *notify.FCMSender  // nil = FCM 未配置，android 分支静默跳过
	logger     *zap.Logger
	popCounter *StaffPopPendingCounter // nil 可选：不查库，badge 恒按告警至少为 1
}

// NewAPNSDeviceService popCounter 可传 nil（仅关闭待 pop 统计注入，不影响编译与其它逻辑）。
func NewAPNSDeviceService(db *sql.DB, sender *notify.APNSSender, fcm *notify.FCMSender, logger *zap.Logger, popCounter *StaffPopPendingCounter) *APNSDeviceService {
	return &APNSDeviceService{db: db, sender: sender, fcm: fcm, logger: logger, popCounter: popCounter}
}

// Register 注册或更新 push token（iOS APNs / Android FCM 共用此入口）
// 同一 (user_id, push_token) 复注册直接 upsert
func (s *APNSDeviceService) Register(
	ctx context.Context,
	userID, pushToken, platform, env string,
) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO apns_devices
		    (user_id, push_token, platform, environment, is_active, last_seen_at, registered_at, updated_at)
		VALUES ($1::UUID, $2, $3, $4, TRUE, NOW(), NOW(), NOW())
		ON CONFLICT (user_id, push_token) DO UPDATE SET
		    platform     = EXCLUDED.platform,
		    environment  = EXCLUDED.environment,
		    is_active    = TRUE,
		    last_seen_at = NOW(),
		    updated_at   = NOW()
	`, userID, pushToken, platform, env)
	return err
}

// Unregister 登出时标记 token 为 inactive
func (s *APNSDeviceService) Unregister(ctx context.Context, userID, pushToken string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE apns_devices
		SET is_active = FALSE, updated_at = NOW()
		WHERE user_id = $1::UUID AND push_token = $2
	`, userID, pushToken)
	return err
}

// deactivateStaleToken APNs 返回 410 时自动失效（goroutine 内调用，不对外暴露）
func (s *APNSDeviceService) deactivateStaleToken(pushToken string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = s.db.ExecContext(ctx, `
		UPDATE apns_devices
		SET is_active = FALSE, updated_at = NOW()
		WHERE push_token = $1
	`, pushToken)
	s.logger.Info("[APNS] stale token deactivated",
		zap.String("token_prefix", tokenPrefix(pushToken)))
}

// AlarmNotification 报警推送参数
// TenantID 是 IPv6 /48 CIDR（如 "fd00:0:5::/48"）；由 cardagg 派生发来
type AlarmNotification struct {
	TenantID   string
	CardID     string
	CardName   string
	EventType  string
	AlarmLevel int // 0=EMERG 1=ALERT 2=CRITICAL 3=ERROR 4=WARNING
}

// SendAlarmPush 向租户下所有符合 notify_mode 的 iOS staff 异步推送报警
// 若 sender 为 nil（未配置 APNs），静默返回
//
// 过滤链：
//   - u.tenant_id = $1::INET    限定租户
//   - u.status = 'active'        见 should_notify_user 内部检查
//   - d.is_active                token 未被 410/UNREGISTERED 失效
//   - d.platform IN (ios,watchos,android) 按行分流：ios/watchos→APNs，android→FCM
//   - should_notify_user(...)    统一 forever/off/login_only/work_time 时间窗 + 节假日
func (s *APNSDeviceService) SendAlarmPush(ctx context.Context, n AlarmNotification) {
	if s.sender == nil && s.fcm == nil {
		return
	}

	type row struct{ token, env, platform, userID string }
	var devices []row

	rows, err := s.db.QueryContext(ctx, `
		SELECT d.push_token, d.environment, d.platform, d.user_id::TEXT
		FROM   apns_devices d
		JOIN   users u ON u.user_id = d.user_id
		WHERE  u.tenant_id = $1::INET
		  AND  d.is_active = TRUE
		  AND  d.platform IN ('ios', 'watchos', 'android')
		  AND  should_notify_user(d.user_id, NOW())
	`, n.TenantID)
	if err != nil {
		s.logger.Warn("[APNS] query devices failed", zap.Error(err))
		return
	}
	defer rows.Close()
	for rows.Next() {
		var d row
		if err := rows.Scan(&d.token, &d.env, &d.platform, &d.userID); err == nil && d.token != "" {
			devices = append(devices, d)
		}
	}
	if len(devices) == 0 {
		return
	}

	for _, d := range devices {
		go func(token, env, platform, staffUserID string) {
			sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			badge := 1
			if s.popCounter != nil && staffUserID != "" {
				nPop, err := s.popCounter.CountForStaff(sendCtx, n.TenantID, staffUserID)
				if err != nil {
					s.logger.Debug("[PUSH] pop pending count failed", zap.Error(err))
				} else if nPop > 0 {
					badge = nPop
				}
			}
			title, body, critical := buildAlarmText(n)

			if platform == "android" {
				if s.fcm == nil {
					return
				}
				msg := s.buildFCMMessage(n, badge, title, body, critical)
				err := s.fcm.Send(sendCtx, token, msg)
				if err == nil {
					s.logger.Info("[FCM] push sent",
						zap.String("token_prefix", tokenPrefix(token)),
						zap.String("event", n.EventType))
					return
				}
				if err == notify.ErrFCMTokenInvalid {
					s.deactivateStaleToken(token)
				} else {
					s.logger.Warn("[FCM] send failed",
						zap.String("token_prefix", tokenPrefix(token)),
						zap.Error(err))
				}
				return
			}

			if s.sender == nil {
				return
			}
			payload := s.buildPayload(n, badge, title, body, critical)
			err := s.sender.Send(sendCtx, token, env, platform, payload)
			if err == nil {
				s.logger.Info("[APNS] push sent",
					zap.String("token_prefix", tokenPrefix(token)),
					zap.String("event", n.EventType))
				return
			}
			if err == notify.ErrDeviceTokenInvalid {
				s.deactivateStaleToken(token)
			} else if strings.Contains(err.Error(), "BadDeviceToken") {
				s.deactivateStaleToken(token)
			} else {
				s.logger.Warn("[APNS] send failed",
					zap.String("token_prefix", tokenPrefix(token)),
					zap.Error(err))
			}
		}(d.token, d.env, d.platform, d.userID)
	}
}

var alarmLevelLabel = [5]string{"EMERG", "ALERT", "CRITICAL", "ERROR", "WARNING"}

// buildAlarmText 平台无关文案：iOS(APNs) 与 Android(FCM) 必须显示一致。
// critical = 是否 EMERG/ALERT 级（走加急声道）。
func buildAlarmText(n AlarmNotification) (title, body string, critical bool) {
	label := "报警"
	if n.AlarmLevel >= 0 && n.AlarmLevel <= 4 {
		label = alarmLevelLabel[n.AlarmLevel]
	}
	return fmt.Sprintf("[%s] %s", label, n.CardName), n.EventType, n.AlarmLevel <= 1
}

// buildPayload aps.badge = 该 staff 可见「待 pop」卡数（与 alarm_state.pop_alarm+event_id 同源）；告警推送至少 1。
// content-available=1 便于后台拉齐角标，不替代 alert。
func (s *APNSDeviceService) buildPayload(n AlarmNotification, badge int, title, body string, critical bool) notify.APNSPayload {
	if badge < 1 {
		badge = 1
	}
	b := badge

	p := notify.APNSPayload{
		APS: notify.APSDict{
			Alert:            notify.APSAlert{Title: title, Body: body},
			Sound:            "default",
			Badge:            &b,
			ContentAvailable: 1,
		},
		CardID:     n.CardID,
		EventType:  n.EventType,
		AlarmLevel: n.AlarmLevel,
	}
	if critical {
		p.APS.Sound = "critical.caf"
		p.APS.InterruptionLevel = "critical"
	}
	return p
}

// buildFCMMessage FCM v1 信封；data key 与 iOS payload 对齐（值全 string），badge 同 iOS 待 pop 卡数口径。
func (s *APNSDeviceService) buildFCMMessage(n AlarmNotification, badge int, title, body string, critical bool) notify.FCMMessage {
	if badge < 1 {
		badge = 1
	}
	b := badge
	sound, np := "default", "PRIORITY_DEFAULT"
	if critical {
		sound, np = "critical", "PRIORITY_MAX"
	}
	return notify.FCMMessage{
		Notification: notify.FCMNotification{Title: title, Body: body},
		Android: notify.FCMAndroidConfig{
			Priority: "high",
			Notification: notify.FCMAndroidNotification{
				Sound:                sound,
				NotificationPriority: np,
				NotificationCount:    &b,
			},
		},
		Data: map[string]string{
			"card_id":     n.CardID,
			"event_type":  n.EventType,
			"alarm_level": strconv.Itoa(n.AlarmLevel),
			"badge_count": strconv.Itoa(badge),
		},
	}
}

func tokenPrefix(token string) string {
	if len(token) >= 8 {
		return token[:8]
	}
	return token
}
