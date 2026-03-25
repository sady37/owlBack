package service

import (
	"context"
	"sync"

	"owl-common/card"

	"go.uber.org/zap"
)

var sleepadDeviceTypes = []string{"Sleepad", "SleepPad", "sleepad"}

// CardMappingService provides MQTT deviceId → card info lookup.
//
// Two layers of in-memory cache:
//   - codeToUID: device_code → device_uid (loaded at startup, refreshed on config event)
//   - cardCache: device_uid → DeviceBaseline (populated on first lookup, cleared on config event)
type CardMappingService struct {
	cardDB *card.CardDB
	logger *zap.Logger

	mu        sync.RWMutex
	codeToUID map[string]string              // device_code → device_uid
	cardCache map[string]*card.DeviceBaseline // device_uid → full identity
}

func NewCardMappingService(cardDB *card.CardDB, logger *zap.Logger) *CardMappingService {
	return &CardMappingService{
		cardDB:    cardDB,
		logger:    logger,
		codeToUID: make(map[string]string),
		cardCache: make(map[string]*card.DeviceBaseline),
	}
}

// LoadCodeToUIDMap loads device_code → device_uid mapping from device_store at startup.
func (s *CardMappingService) LoadCodeToUIDMap(ctx context.Context) error {
	m, err := s.cardDB.LoadCodeToUIDMap(ctx, sleepadDeviceTypes)
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
	s.mu.Lock()
	s.cardCache = make(map[string]*card.DeviceBaseline)
	s.mu.Unlock()

	if err := s.LoadCodeToUIDMap(ctx); err != nil {
		s.logger.Error("reload code-to-uid map after invalidate", zap.Error(err))
	}
	s.logger.Info("card cache invalidated")
}

// InvalidateByCardID removes cached entries belonging to a specific card.
func (s *CardMappingService) InvalidateByCardID(cardID string) int {
	s.mu.Lock()
	n := 0
	for uid, m := range s.cardCache {
		if m.CardID == cardID {
			delete(s.cardCache, uid)
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

// ResolveToDeviceUID 将 device_code 或 device_id 转为 device_uid，内部业务统一用 device_uid。
func (s *CardMappingService) ResolveToDeviceUID(ctx context.Context, id string) string {
	if id == "" {
		return ""
	}
	if u := s.CodeToUID(id); u != "" {
		return u
	}
	uid, _, err := s.cardDB.ResolveDevice(ctx, id)
	if err != nil {
		return ""
	}
	return uid
}

// GetCardInfo resolves MQTT deviceId → DeviceBaseline.
// Uses in-memory cache; DB is only hit on first lookup per device_uid.
func (s *CardMappingService) GetCardInfo(ctx context.Context, mqttDeviceID string) (*card.DeviceBaseline, error) {
	s.mu.RLock()
	uid := s.codeToUID[mqttDeviceID]
	s.mu.RUnlock()

	key := uid
	if key == "" {
		key = mqttDeviceID
	}

	s.mu.RLock()
	cached, ok := s.cardCache[key]
	s.mu.RUnlock()
	if ok {
		return cached, nil
	}

	info, err := s.cardDB.LookupCard(ctx, key)
	if err != nil {
		info, err = s.cardDB.LookupDeviceOnly(ctx, key)
		if err != nil {
			info, err = s.cardDB.LookupDeviceStoreOnly(ctx, key)
			if err != nil {
				return nil, err
			}
		}
	}

	s.mu.Lock()
	s.cardCache[info.DeviceUID] = info
	s.mu.Unlock()

	return info, nil
}
