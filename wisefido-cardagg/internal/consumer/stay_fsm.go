package consumer

import (
	"context"
	"time"

	"owl-common/alarm"
	"owl-common/redis"
)

// Stay 状态机：武装 pending（S）与解除 pending（R），与 BathRoomState.StayFSM* 同步展示。
// 窗口 150s；武装需 Enter 后 >2 条 activity（即计数≥3）且窗口内 number>0，且 TotalPeople==1；解除需 Exit 后窗口内 TotalPeople==0。
const (
	stayWindowTTLMs       = int64(150 * 1000)
	stayMinActivityForArm = 3 // >2 条 activity
)

const (
	stayPhaseIdle          = ""
	stayPhaseArmWindow     = "arm_window"
	stayPhaseResolveWindow = "resolve_window"
	stayPhaseArmed         = "armed"
)

// staySession 每 tenant:device 一条；与 Redis Stay pending 配合。
type staySession struct {
	EnterAtMs       int64
	ArmExpireAtMs   int64
	ActivityCount   int
	NumberPositive  bool
	ArmedPending    bool
	ExitAtMs        int64
	ResolveExpireAt int64
}

func (s *staySession) inArmWindow(now int64) bool {
	return s != nil && s.EnterAtMs > 0 && now <= s.ArmExpireAtMs && now >= s.EnterAtMs
}

func (s *staySession) inResolveWindow(now int64) bool {
	return s != nil && s.ExitAtMs > 0 && now <= s.ResolveExpireAt && now >= s.ExitAtMs
}

func staySessionKey(tenantID, deviceID string) string { return tenantID + ":" + deviceID }

func (h *EventHandler) ensureStaySessionLocked(tenantID, deviceID string) *staySession {
	if h.staySessions == nil {
		h.staySessions = make(map[string]*staySession)
	}
	key := staySessionKey(tenantID, deviceID)
	if h.staySessions[key] == nil {
		h.staySessions[key] = &staySession{}
	}
	return h.staySessions[key]
}

func (h *EventHandler) stayPublishPhase(ctx context.Context, m *redis.IoTStreamMessage, deviceID string, ses *staySession, now int64) {
	if h.state == nil || ses == nil || m == nil {
		return
	}
	phase := stayPhaseIdle
	armAt := int64(0)
	resolveAt := int64(0)
	switch {
	case ses.ArmedPending:
		phase = stayPhaseArmed
		armAt = ses.EnterAtMs
		if ses.inResolveWindow(now) {
			resolveAt = ses.ExitAtMs
		}
	case ses.inResolveWindow(now):
		phase = stayPhaseResolveWindow
		resolveAt = ses.ExitAtMs
	case ses.inArmWindow(now) && !ses.ArmedPending:
		phase = stayPhaseArmWindow
		armAt = ses.EnterAtMs
	}
	_ = h.state.PublishBathRoomStayFSM(ctx, m.CardID, deviceID, m.DeviceUID, phase, armAt, resolveAt)
}

// stayOnEnter S1/S4：新 Enter 清 Stay pending，重置武装窗。
func (h *EventHandler) stayOnEnter(ctx context.Context, m *redis.IoTStreamMessage, deviceID string) {
	if m.TenantID == "" || deviceID == "" {
		return
	}
	now := m.Timestamp
	if now == 0 {
		now = time.Now().UnixMilli()
	}
	h.staySesMu.Lock()
	ses := h.ensureStaySessionLocked(m.TenantID, deviceID)
	_ = h.alarms.RemovePendingAlarm(ctx, m.TenantID, m.CardID, deviceID, alarm.Stay)
	ses.EnterAtMs = now
	ses.ArmExpireAtMs = now + stayWindowTTLMs
	ses.ActivityCount = 0
	ses.NumberPositive = false
	ses.ArmedPending = false
	ses.ExitAtMs = 0
	ses.ResolveExpireAt = 0
	publishNow := time.Now().UnixMilli()
	h.stayPublishPhase(ctx, m, deviceID, ses, publishNow)
	h.staySesMu.Unlock()
}

// stayOnExit R1：开 150s 解除窗。
func (h *EventHandler) stayOnExit(ctx context.Context, m *redis.IoTStreamMessage, deviceID string) {
	if m.TenantID == "" || deviceID == "" {
		return
	}
	now := m.Timestamp
	if now == 0 {
		now = time.Now().UnixMilli()
	}
	h.staySesMu.Lock()
	ses := h.ensureStaySessionLocked(m.TenantID, deviceID)
	ses.ExitAtMs = now
	ses.ResolveExpireAt = now + stayWindowTTLMs
	ses.EnterAtMs = 0
	ses.ArmExpireAtMs = 0
	ses.ActivityCount = 0
	ses.NumberPositive = false
	publishNow := time.Now().UnixMilli()
	h.stayPublishPhase(ctx, m, deviceID, ses, publishNow)
	h.staySesMu.Unlock()
}

// stayOnNumberPeople 在 PublishBathRoomStateFromEvent 之后调用：用最新 TotalPeople 推进武装/解除。
func (h *EventHandler) stayOnNumberPeople(ctx context.Context, m *redis.IoTStreamMessage, deviceID string, totalPeople int) {
	if m.TenantID == "" || deviceID == "" || h.state == nil {
		return
	}
	now := time.Now().UnixMilli()
	ts := m.Timestamp
	if ts == 0 {
		ts = now
	}
	h.staySesMu.Lock()
	ses := h.ensureStaySessionLocked(m.TenantID, deviceID)

	if ses.inArmWindow(now) && ts >= ses.EnterAtMs && totalPeople > 0 {
		ses.NumberPositive = true
	}

	if totalPeople >= 2 {
		_ = h.alarms.RemovePendingAlarm(ctx, m.TenantID, m.CardID, deviceID, alarm.Stay)
		ses.ArmedPending = false
	}

	if ses.inResolveWindow(now) && totalPeople == 0 {
		_ = h.alarms.RemovePendingAlarm(ctx, m.TenantID, m.CardID, deviceID, alarm.Stay)
		ses.ArmedPending = false
		ses.ExitAtMs = 0
		ses.ResolveExpireAt = 0
	}

	h.tryArmStayPendingLocked(ctx, m, deviceID, now, ses)
	h.stayPublishPhase(ctx, m, deviceID, ses, now)
	h.staySesMu.Unlock()
}

// stayOnActivity：武装窗内累计 activity，满足 S2+S3 且 TotalPeople==1 则 AddStayPending。
func (h *EventHandler) stayOnActivity(ctx context.Context, m *redis.IoTStreamMessage, deviceID string) {
	if m.TenantID == "" || deviceID == "" || h.state == nil {
		return
	}
	_, stayOn := h.enablement.IsEnabled(ctx, m.TenantID, deviceID, alarm.Stay)
	if !stayOn {
		return
	}
	now := time.Now().UnixMilli()
	h.staySesMu.Lock()
	ses := h.ensureStaySessionLocked(m.TenantID, deviceID)
	if !ses.inArmWindow(now) {
		h.staySesMu.Unlock()
		return
	}
	ses.ActivityCount++
	h.tryArmStayPendingLocked(ctx, m, deviceID, now, ses)
	h.stayPublishPhase(ctx, m, deviceID, ses, now)
	h.staySesMu.Unlock()
}

func (h *EventHandler) tryArmStayPendingLocked(ctx context.Context, m *redis.IoTStreamMessage, deviceID string, now int64, ses *staySession) {
	if ses.ArmedPending || !ses.inArmWindow(now) {
		return
	}
	if ses.ActivityCount < stayMinActivityForArm || !ses.NumberPositive {
		return
	}
	curr, err := h.state.ReadCardStatus(ctx, m.CardID)
	if err != nil || curr == nil || curr.BathRoomState == nil {
		return
	}
	if curr.BathRoomState.TotalPeople != 1 {
		return
	}
	_ = h.alarms.AddStayPendingIfEnabled(ctx, m.TenantID, deviceID, ses.EnterAtMs)
	ses.ArmedPending = true
}
