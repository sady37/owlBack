package service

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type DeviceStatusEntry struct {
	Online        bool   `json:"online"`
	Heart         int    `json:"heart"`
	Breath        int    `json:"breath"`
	BedStatus     int    `json:"bed_status"`
	TurnOver      int    `json:"turn_over"`
	BodyMove      int    `json:"body_move"`
	SitUp         int    `json:"sit_up"`
	InitStatus    int    `json:"init_status"`
	SignalQuality int    `json:"signal_quality"` // 0-100%
	SleepStage    int    `json:"sleep_stage"`
	InbedStatus   int    `json:"inbed_status"`
	DeviceFailure bool   `json:"device_failure"`
	Detached      bool   `json:"detached"`
	ActiveAlarm   string `json:"active_alarm,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DeviceStatusTracker struct {
	mu      sync.RWMutex
	devices map[string]*DeviceStatusEntry // key: device_uid（内部统一）
	mapping *CardMappingService
	logger  *zap.Logger
}

func NewDeviceStatusTracker(mapping *CardMappingService, logger *zap.Logger) *DeviceStatusTracker {
	return &DeviceStatusTracker{
		devices: make(map[string]*DeviceStatusEntry),
		mapping: mapping,
		logger:  logger,
	}
}

func (t *DeviceStatusTracker) getOrCreate(deviceCode string) *DeviceStatusEntry {
	e, ok := t.devices[deviceCode]
	if !ok {
		e = &DeviceStatusEntry{}
		t.devices[deviceCode] = e
	}
	e.UpdatedAt = time.Now()
	return e
}

func (t *DeviceStatusTracker) UpdateConnection(deviceCode string, online bool) {
	t.mu.Lock()
	e := t.getOrCreate(deviceCode)
	e.Online = online
	t.mu.Unlock()
}

func (t *DeviceStatusTracker) UpdateRealtime(deviceCode string, heart, breath, bedStatus, turnOver, bodyMove, sitUp, initStatus, signalQuality int) {
	t.mu.Lock()
	e := t.getOrCreate(deviceCode)
	e.Heart = heart
	e.Breath = breath
	e.BedStatus = bedStatus
	e.TurnOver = turnOver
	e.BodyMove = bodyMove
	e.SitUp = sitUp
	e.InitStatus = initStatus
	e.SignalQuality = signalQuality
	t.mu.Unlock()
}

func (t *DeviceStatusTracker) UpdateSleepStage(deviceCode string, stage int) {
	t.mu.Lock()
	e := t.getOrCreate(deviceCode)
	e.SleepStage = stage
	t.mu.Unlock()
}

func (t *DeviceStatusTracker) UpdateInbedStatus(deviceCode string, status int) {
	t.mu.Lock()
	e := t.getOrCreate(deviceCode)
	e.InbedStatus = status
	t.mu.Unlock()
}

func (t *DeviceStatusTracker) UpdateSensor(deviceCode string, failure, detached bool) {
	t.mu.Lock()
	e := t.getOrCreate(deviceCode)
	e.DeviceFailure = failure
	e.Detached = detached
	t.mu.Unlock()
}

func (t *DeviceStatusTracker) UpdateAlarm(deviceCode string, activeAlarm string) {
	t.mu.Lock()
	e := t.getOrCreate(deviceCode)
	e.ActiveAlarm = activeAlarm
	t.mu.Unlock()
}

// --- query methods ---

type DeviceStatusItem struct {
	DeviceUID  string             `json:"device_uid"`
	DeviceCode string             `json:"device_code,omitempty"`
	Status     string             `json:"status"`
	Detail     *DeviceStatusEntry `json:"detail,omitempty"`
}

func (t *DeviceStatusTracker) GetByUID(uid string) *DeviceStatusItem {
	code := ""
	if t.mapping != nil {
		code = t.mapping.UIDToCode(uid)
	}
	return t.buildItem(uid, code)
}

func (t *DeviceStatusTracker) GetByCode(code string) *DeviceStatusItem {
	uid := ""
	if t.mapping != nil {
		uid = t.mapping.CodeToUID(code)
	}
	if uid == "" {
		return &DeviceStatusItem{DeviceCode: code, Status: "unknown"}
	}
	return t.buildItem(uid, code)
}

func (t *DeviceStatusTracker) buildItem(uid, code string) *DeviceStatusItem {
	t.mu.RLock()
	entry, ok := t.devices[uid]
	t.mu.RUnlock()
	if !ok {
		return &DeviceStatusItem{DeviceUID: uid, DeviceCode: code, Status: "unknown"}
	}
	status := "offline"
	if entry.Online {
		status = "online"
	}
	snapshot := *entry
	return &DeviceStatusItem{
		DeviceUID:  uid,
		DeviceCode: code,
		Status:     status,
		Detail:     &snapshot,
	}
}

func (t *DeviceStatusTracker) GetAll() []DeviceStatusItem {
	t.mu.RLock()
	defer t.mu.RUnlock()
	items := make([]DeviceStatusItem, 0, len(t.devices))
	for uid, entry := range t.devices {
		code := ""
		if t.mapping != nil {
			code = t.mapping.UIDToCode(uid)
		}
		status := "offline"
		if entry.Online {
			status = "online"
		}
		snapshot := *entry
		items = append(items, DeviceStatusItem{
			DeviceUID:  uid,
			DeviceCode: code,
			Status:     status,
			Detail:     &snapshot,
		})
	}
	return items
}
