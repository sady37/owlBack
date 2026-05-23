package service

import (
	"context"
	"sync"

	"owl-common/card"

	"go.uber.org/zap"
)

var sleepadDeviceTypes = []string{"Sleepad", "SleepPad", "sleepad"}

// CardMappingService — sleepace MQTT 报文 deviceId 字段 → DeviceBaseline 解析层。
//
// MQTT deviceId 实际值：
//   - 正常路径（bind 完成后）= device_code（sleepace 平台密文 ID，如 "r0nqo00b34vtf"）
//   - 兼容路径（pre-bind 早期）= device_uid（logMAC，如 "BM87225200672"）
//
// 两层 in-memory cache：
//   - codeToUID:     device_code → device_uid（启动 bulk 加载 + invalidate 重载）
//   - baselineCache: device_uid → *DeviceBaseline（懒查 + invalidate 清空）
//
// 上游事实源：wisefido-data /internal/device-baseline HTTP API (经 CardAPIClient)。
// 失效触发：config:card:stream 事件 → InvalidateCache / InvalidateByDeviceUID / InvalidateByCardID。
type CardMappingService struct {
	api    *card.CardAPIClient
	logger *zap.Logger

	mu            sync.RWMutex
	codeToUID     map[string]string               // device_code → device_uid
	baselineCache map[string]*card.DeviceBaseline // device_uid → full identity (DeviceBaseline)

	readyCh chan struct{} // closed when cache is ready; re-created on invalidate
}

func NewCardMappingService(api *card.CardAPIClient, logger *zap.Logger) *CardMappingService {
	ch := make(chan struct{})
	close(ch) // initially ready
	return &CardMappingService{
		api:           api,
		logger:        logger,
		codeToUID:     make(map[string]string),
		baselineCache: make(map[string]*card.DeviceBaseline),
		readyCh:       ch,
	}
}

// WaitReady blocks until the cache is ready after an invalidation cycle.
func (s *CardMappingService) WaitReady() {
	<-s.readyCh
}

// LoadCodeToUIDMap loads device_code → device_uid mapping from the API at startup.
func (s *CardMappingService) LoadCodeToUIDMap(ctx context.Context) error {
	m, err := s.api.LoadCodeToUIDMap(ctx, sleepadDeviceTypes)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.codeToUID = m
	s.mu.Unlock()
	s.logger.Info("loaded device_code → device_uid map", zap.Int("count", len(m)))
	return nil
}

// InvalidateCache clears both code→uid map and card cache, then reloads code→uid.
func (s *CardMappingService) InvalidateCache(ctx context.Context) {
	// Block callers of WaitReady until reload completes.
	ch := make(chan struct{})
	s.mu.Lock()
	s.readyCh = ch
	s.baselineCache = make(map[string]*card.DeviceBaseline)
	s.mu.Unlock()

	if err := s.LoadCodeToUIDMap(ctx); err != nil {
		s.logger.Error("reload code-to-uid map after invalidate", zap.Error(err))
	}
	close(ch)
	s.logger.Info("card cache invalidated")
}

// InvalidateByDeviceUID 删除单设备在 baselineCache 中的条目（device_uid 键）。
func (s *CardMappingService) InvalidateByDeviceUID(deviceUID string) {
	if deviceUID == "" {
		return
	}
	s.mu.Lock()
	delete(s.baselineCache, deviceUID)
	s.mu.Unlock()
}

// InvalidateByCardID removes cached entries belonging to a specific card.
func (s *CardMappingService) InvalidateByCardID(cardID string) int {
	s.mu.Lock()
	n := 0
	for uid, m := range s.baselineCache {
		if m.CardID == cardID {
			delete(s.baselineCache, uid)
			n++
		}
	}
	s.mu.Unlock()
	if n > 0 {
		s.logger.Info("card cache entries evicted", zap.String("card_id", cardID), zap.Int("count", n))
	}
	return n
}

// PutMapping adds a single device_code → device_uid entry to the in-memory map.
func (s *CardMappingService) PutMapping(deviceCode, deviceUID string) {
	if deviceCode == "" || deviceUID == "" {
		return
	}
	s.mu.Lock()
	s.codeToUID[deviceCode] = deviceUID
	s.mu.Unlock()
}

// CodeToUID returns device_uid for a given device_code, or "" if not found.
func (s *CardMappingService) CodeToUID(code string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.codeToUID[code]
}

// UIDToCode returns device_code for a given device_uid (reverse lookup).
func (s *CardMappingService) UIDToCode(uid string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for code, u := range s.codeToUID {
		if u == uid {
			return code
		}
	}
	return ""
}

// ResolveToDeviceUID 将 MQTT deviceId 入参（device_code 或 device_uid 二选一）规范化为 device_uid。
// 内部所有业务统一以 device_uid 为 identity key（v1 deviceID/UUID 概念已退役）。
func (s *CardMappingService) ResolveToDeviceUID(ctx context.Context, deviceKey string) string {
	if deviceKey == "" {
		return ""
	}
	if u := s.CodeToUID(deviceKey); u != "" {
		return u
	}
	// Fallback: ask the API
	b, err := s.api.LookupBaseline(ctx, deviceKey)
	if err != nil {
		return ""
	}
	return b.DeviceUID
}

// GetBaseline resolves MQTT deviceId（device_code 或 device_uid 二选一）→ *DeviceBaseline。
// 缓存命中走内存；miss 走 CardAPIClient.LookupBaseline 拉一次并写 baselineCache。
func (s *CardMappingService) GetBaseline(ctx context.Context, deviceKey string) (*card.DeviceBaseline, error) {
	s.mu.RLock()
	uid := s.codeToUID[deviceKey]
	s.mu.RUnlock()

	cacheKey := uid
	if cacheKey == "" {
		cacheKey = deviceKey
	}

	s.mu.RLock()
	cached, ok := s.baselineCache[cacheKey]
	s.mu.RUnlock()
	if ok {
		return cached, nil
	}

	info, err := s.api.LookupBaseline(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.baselineCache[info.DeviceUID] = info
	s.mu.Unlock()

	return info, nil
}

// ResolveBaseline 供 health_check 等场景按 device_uid（或 device_code）直查 API、**不写 baselineCache**，
// 避免 health-check 的临时性查询污染主缓存。命中失败时返回零值 + false。
func (s *CardMappingService) ResolveBaseline(ctx context.Context, deviceKey string) (card.DeviceBaseline, bool) {
	if s.api == nil || deviceKey == "" {
		return card.DeviceBaseline{}, false
	}
	b, err := s.api.LookupBaseline(ctx, deviceKey)
	if err != nil {
		return card.DeviceBaseline{}, false
	}
	return *b, true
}

// ListSleepadBaselinesForHealth 定时全量探测用。
func (s *CardMappingService) ListSleepadBaselinesForHealth(ctx context.Context, deviceTypes []string) ([]card.DeviceBaseline, error) {
	if s.api == nil {
		return nil, nil
	}
	if len(deviceTypes) == 0 {
		deviceTypes = sleepadDeviceTypes
	}
	return s.api.ListBaselines(ctx, deviceTypes)
}
