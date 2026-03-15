package service

import (
	"context"
	"sync"

	"owl-common/card"
	"wisefido-qinglan/internal/repository"

	"go.uber.org/zap"
)

// DeviceBoundResolver returns device-bound room_id and bed_id (e.g. from device_store).
type DeviceBoundResolver interface {
	GetBoundRoomAndBed(ctx context.Context, deviceUID string) (roomID, bedID string)
}

// CardMappingService provides deviceUID → card info lookup via in-memory cache.
// On cache miss, queries DB through card.CardDB.
// If boundResolver is set, RoomID/BedID are filled from device binding.
type CardMappingService struct {
	cardDB        *card.CardDB
	logger        *zap.Logger
	boundResolver DeviceBoundResolver

	mu        sync.RWMutex
	cardCache map[string]*card.DeviceCardMapping
}

func NewCardMappingService(cardDB *card.CardDB, logger *zap.Logger) *CardMappingService {
	return &CardMappingService{
		cardDB:    cardDB,
		logger:    logger,
		cardCache: make(map[string]*card.DeviceCardMapping),
	}
}

// SetDeviceBoundResolver sets the optional resolver for RoomID/BedID (e.g. qinglan device_store).
func (s *CardMappingService) SetDeviceBoundResolver(r DeviceBoundResolver) {
	s.boundResolver = r
}

// InvalidateCache clears the entire in-memory card cache.
func (s *CardMappingService) InvalidateCache() {
	s.mu.Lock()
	s.cardCache = make(map[string]*card.DeviceCardMapping)
	s.mu.Unlock()
	s.logger.Info("card cache invalidated")
}

// InvalidateByCardID removes cached entries belonging to a specific card.
// Avoids clearing unrelated devices; no DB call needed.
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

// InvalidateByDeviceUID removes the cached DeviceCardMapping for a device (e.g. when device_store room/bed binding changes).
func (s *CardMappingService) InvalidateByDeviceUID(deviceUID string) {
	if deviceUID == "" {
		return
	}
	s.mu.Lock()
	_, ok := s.cardCache[deviceUID]
	if ok {
		delete(s.cardCache, deviceUID)
	}
	s.mu.Unlock()
	if ok {
		s.logger.Info("card cache entry evicted for device", zap.String("device_uid", deviceUID))
	}
}

// GetCardIDByDeviceUID resolves deviceUID → DeviceCardMapping (lazy-loaded).
func (s *CardMappingService) GetCardIDByDeviceUID(ctx context.Context, deviceUID string) (*card.DeviceCardMapping, error) {
	s.mu.RLock()
	cached, ok := s.cardCache[deviceUID]
	s.mu.RUnlock()
	if ok {
		return cached, nil
	}

	info, err := s.cardDB.LookupCard(ctx, deviceUID)
	if err != nil {
		return nil, err
	}
	if s.boundResolver != nil {
		info.RoomID, info.BedID = s.boundResolver.GetBoundRoomAndBed(ctx, info.DeviceUID)
	}

	s.mu.Lock()
	s.cardCache[info.DeviceUID] = info
	s.mu.Unlock()

	return info, nil
}

// deviceBoundResolver implements DeviceBoundResolver from device_store (BoundRoomID/BoundBedID).
type deviceBoundResolver struct {
	repo repository.DeviceRepository
}

// NewDeviceBoundResolver returns a resolver that fills RoomID/BedID from device_store.
func NewDeviceBoundResolver(repo repository.DeviceRepository) DeviceBoundResolver {
	return &deviceBoundResolver{repo: repo}
}

func (r *deviceBoundResolver) GetBoundRoomAndBed(ctx context.Context, deviceUID string) (roomID, bedID string) {
	dev, err := r.repo.GetDeviceByUID(ctx, deviceUID)
	if err != nil || dev == nil {
		return "", ""
	}
	if dev.BoundRoomID.Valid {
		roomID = dev.BoundRoomID.String
	}
	if dev.BoundBedID.Valid {
		bedID = dev.BoundBedID.String
	}
	return roomID, bedID
}

var _ DeviceBoundResolver = (*deviceBoundResolver)(nil)
