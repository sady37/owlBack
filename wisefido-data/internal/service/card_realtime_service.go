package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"wisefido-data/internal/store"

	"go.uber.org/zap"
)

// StatusEvent status 变更事件（service 包内定义，避免循环引用 subscriber 包）
type StatusEvent struct {
	CardID string
	Data   map[string]interface{}
}

// RealtimeDataProvider 供 SSE 使用，仅从 DataStreamSubscriber 缓存读取（不直接订阅 Redis）
type RealtimeDataProvider interface {
	GetCardRealtimeData(cardID string) map[string]interface{}
	GetCardRealtimeVersion(cardID string) uint64
	GetCardStatusData(cardID string) map[string]interface{}
}

const realtimeKeyPrefix = "vital-focus:card:"
const realtimeKeySuffix = ":realtime"
const maxPullRealtimeCards = 40

// CardChange 卡片变更记录
type CardChange struct {
	CardID string `json:"card_id"`
	Op     string `json:"op"` // "add" / "update" / "delete"
}

// sseConn 一个 SSE 连接的注册信息
type sseConn struct {
	userKey  string             // "tenantID:userID"
	cardIDs  map[string]bool    // 该连接订阅的 cardID 集合
	statusCh chan StatusEvent   // per-connection status 事件推送 channel
	changeCh chan []CardChange  // per-connection 卡片增删事件推送 channel
}

// CardRealtimeService 卡片实时数据服务
// userCardLists 保存每个在线用户的 CardList 原始结构（不展开）
type CardRealtimeService struct {
	kv              store.KV
	allowedProvider AllowedCardIDsProvider
	logger          *zap.Logger

	streamProvider RealtimeDataProvider // 可选，SSE 从订阅器缓存拉取（不直接订阅 Redis）

	mu            sync.RWMutex
	userCardLists map[string]*CardList    // "tenantID:userID" → stored CardList
	userTypes     map[string]string       // "tenantID:userID" → userType（UpdateByBranch 重查用）
	cardChanges   map[string][]CardChange // "tenantID:userID" → pending changes

	// SSE 连接注册表（fan-out 用）
	connMu    sync.RWMutex
	sseConns  map[string]*sseConn  // connID → sseConn
	cardIndex map[string][]string  // cardID → []connID（反向索引）
	userIndex map[string][]string  // userKey → []connID（按用户路由 card_change）
	connSeq   uint64               // 自增连接 ID

	// status 事件输入 channel（由 main.go 桥接 subscriber.CardStatusEvent → StatusEvent）
	statusEventCh <-chan StatusEvent
}

// NewCardRealtimeService 创建卡片实时数据服务
// streamProvider 可为 nil；非 nil 时才能使用 SubscribeRealtimeStream（仅读缓存，不订阅 Redis）
func NewCardRealtimeService(kv store.KV, allowedProvider AllowedCardIDsProvider, logger *zap.Logger, streamProvider RealtimeDataProvider) *CardRealtimeService {
	return &CardRealtimeService{
		kv:              kv,
		allowedProvider: allowedProvider,
		logger:          logger,
		streamProvider:  streamProvider,
		userCardLists:   make(map[string]*CardList),
		userTypes:       make(map[string]string),
		cardChanges:     make(map[string][]CardChange),
		sseConns:        make(map[string]*sseConn),
		cardIndex:       make(map[string][]string),
		userIndex:       make(map[string][]string),
	}
}

// SetSSEDependencies 延迟注入 SSE 依赖（仅从 DataStreamSubscriber 缓存读）
func (s *CardRealtimeService) SetSSEDependencies(streamProvider RealtimeDataProvider) {
	s.streamProvider = streamProvider
}

// SetStatusEventChan 注入 status 事件 channel（由 main.go 桥接 subscriber → service）
func (s *CardRealtimeService) SetStatusEventChan(ch <-chan StatusEvent) {
	s.statusEventCh = ch
}

// registerSSE 注册 SSE 连接，返回 connID、statusCh、changeCh
func (s *CardRealtimeService) registerSSE(cardIDs []string, tenantID, userID string) (string, <-chan StatusEvent, <-chan []CardChange) {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	s.connSeq++
	connID := fmt.Sprintf("sse-%d", s.connSeq)
	userKey := tenantID + ":" + userID

	cardSet := make(map[string]bool, len(cardIDs))
	for _, id := range cardIDs {
		cardSet[id] = true
	}
	statusCh := make(chan StatusEvent, 8)
	changeCh := make(chan []CardChange, 4)
	s.sseConns[connID] = &sseConn{userKey: userKey, cardIDs: cardSet, statusCh: statusCh, changeCh: changeCh}

	// cardID 反向索引
	for _, cardID := range cardIDs {
		s.cardIndex[cardID] = append(s.cardIndex[cardID], connID)
	}
	// userKey 反向索引
	s.userIndex[userKey] = append(s.userIndex[userKey], connID)

	return connID, statusCh, changeCh
}

// unregisterSSE 注销 SSE 连接
func (s *CardRealtimeService) unregisterSSE(connID string) {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	conn, ok := s.sseConns[connID]
	if !ok {
		return
	}
	// 清理 cardID 反向索引
	for cardID := range conn.cardIDs {
		conns := s.cardIndex[cardID]
		filtered := conns[:0]
		for _, c := range conns {
			if c != connID {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) == 0 {
			delete(s.cardIndex, cardID)
		} else {
			s.cardIndex[cardID] = filtered
		}
	}
	// 清理 userKey 反向索引
	if conn.userKey != "" {
		conns := s.userIndex[conn.userKey]
		filtered := conns[:0]
		for _, c := range conns {
			if c != connID {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) == 0 {
			delete(s.userIndex, conn.userKey)
		} else {
			s.userIndex[conn.userKey] = filtered
		}
	}
	close(conn.statusCh)
	close(conn.changeCh)
	delete(s.sseConns, connID)
}

// StartStatusFanout 启动 status 事件 fan-out goroutine
// 消费 statusEventCh，按 cardID 查反向索引，推送到对应连接的 per-connection channel
func (s *CardRealtimeService) StartStatusFanout(ctx context.Context) {
	if s.statusEventCh == nil {
		s.logger.Warn("StartStatusFanout: statusEventCh is nil, skipping")
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-s.statusEventCh:
				if !ok {
					return
				}
				s.connMu.RLock()
				connIDs := s.cardIndex[evt.CardID]
				for _, connID := range connIDs {
					if conn, exists := s.sseConns[connID]; exists {
						select {
						case conn.statusCh <- evt:
						default:
							// per-connection channel 满，丢弃（避免阻塞 fan-out）
						}
					}
				}
				s.connMu.RUnlock()
			}
		}
	}()
	s.logger.Info("StartStatusFanout started")
}

// UpdateCardList 保存用户的 CardList（不展开）
func (s *CardRealtimeService) UpdateCardList(cardList *CardList, userType string) {
	if cardList == nil || cardList.UserID == "" || cardList.TenantID == "" {
		return
	}
	userKey := cardList.TenantID + ":" + cardList.UserID

	s.mu.Lock()
	s.userCardLists[userKey] = cardList
	s.userTypes[userKey] = userType
	s.mu.Unlock()

	s.logger.Debug("UpdateCardList",
		zap.String("tenant_id", cardList.TenantID),
		zap.String("user_id", cardList.UserID),
		zap.Int("branch_count", len(cardList.CardsByBranch)),
	)
}

// RemoveCardList 用户登出时清除
func (s *CardRealtimeService) RemoveCardList(tenantID, userID string) {
	userKey := tenantID + ":" + userID
	s.mu.Lock()
	delete(s.userCardLists, userKey)
	delete(s.userTypes, userKey)
	delete(s.cardChanges, userKey)
	s.mu.Unlock()
}

// UpdateByBranch 某 branch 卡片变更后调用
// Step1: 扫 userCardLists 找含该 branchID 的用户，重查 CardList
func (s *CardRealtimeService) UpdateByBranch(ctx context.Context, tenantID, branchID string) {
	// 找受影响的用户（CardsByBranch 含该 branchID）
	type affected struct {
		userID   string
		userType string
		oldList  *CardList
	}
	var targets []affected

	s.mu.RLock()
	for userKey, cl := range s.userCardLists {
		if cl.TenantID != tenantID {
			continue
		}
		if _, ok := cl.CardsByBranch[branchID]; !ok {
			continue
		}
		userType, ok := s.userTypes[userKey]
		if !ok || userType == "" {
			continue
		}
		targets = append(targets, affected{userID: cl.UserID, userType: userType, oldList: cl})
	}
	s.mu.RUnlock()

	if len(targets) == 0 {
		return
	}

	for _, t := range targets {
		// 重查该用户的 CardList
		newList, err := s.allowedProvider.GetCardList(ctx, tenantID, t.userID, t.userType)
		if err != nil {
			s.logger.Warn("UpdateByBranch re-query failed",
				zap.String("user_id", t.userID),
				zap.Error(err),
			)
			continue
		}

		// diff oldList vs newList → []CardChange
		changes := diffCardList(t.oldList, newList)
		if len(changes) == 0 {
			continue
		}

		// 更新本地存储 + 追加 cardChanges
		userKey := tenantID + ":" + t.userID
		s.mu.Lock()
		s.userCardLists[userKey] = newList
		s.cardChanges[userKey] = append(s.cardChanges[userKey], changes...)
		s.mu.Unlock()

		s.logger.Info("UpdateByBranch user affected",
			zap.String("tenant_id", tenantID),
			zap.String("branch_id", branchID),
			zap.String("user_id", t.userID),
			zap.Int("changes", len(changes)),
		)

		// fan-out card_change 到该用户的 SSE 连接
		s.fanOutCardChanges(tenantID+":"+t.userID, changes)
	}
}

// fanOutCardChanges 将卡片变更推送到指定用户的所有 SSE 连接
func (s *CardRealtimeService) fanOutCardChanges(userKey string, changes []CardChange) {
	s.connMu.RLock()
	defer s.connMu.RUnlock()

	connIDs := s.userIndex[userKey]
	for _, connID := range connIDs {
		if conn, ok := s.sseConns[connID]; ok {
			select {
			case conn.changeCh <- changes:
			default:
				// channel 满，丢弃旧的再写入
				select {
				case <-conn.changeCh:
				default:
				}
				select {
				case conn.changeCh <- changes:
				default:
				}
			}
		}
	}
}

// diffCardList 对比新旧 CardList，返回变更列表
// 新增 → add，删除 → delete，两边都有 → update（card 内容可能变了）
func diffCardList(oldList, newList *CardList) []CardChange {
	oldSet := make(map[string]bool)
	newSet := make(map[string]bool)
	if oldList != nil {
		for _, ids := range oldList.CardsByBranch {
			for _, id := range ids {
				oldSet[id] = true
			}
		}
	}
	if newList != nil {
		for _, ids := range newList.CardsByBranch {
			for _, id := range ids {
				newSet[id] = true
			}
		}
	}
	var changes []CardChange
	for id := range newSet {
		if !oldSet[id] {
			changes = append(changes, CardChange{CardID: id, Op: "add"})
		}
	}
	for id := range oldSet {
		if !newSet[id] {
			changes = append(changes, CardChange{CardID: id, Op: "delete"})
		}
	}
	return changes
}

// ConsumeCardChanges 取出并清空 pending 的卡片变更（GetCardRealtime / SSE 消费）
func (s *CardRealtimeService) ConsumeCardChanges(tenantID, userID string) []CardChange {
	userKey := tenantID + ":" + userID
	s.mu.Lock()
	changes := s.cardChanges[userKey]
	delete(s.cardChanges, userKey)
	s.mu.Unlock()
	return changes
}

// getCardList 取出用户存储的 CardList（nil 表示未登录/未推送）
func (s *CardRealtimeService) getCardList(tenantID, userID string) *CardList {
	userKey := tenantID + ":" + userID
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userCardLists[userKey]
}

// GetCardRealtimeRequest 拉取实时数据请求
// 客户端提交 CardsByBranch（与 CardList 同格式），一次最多 40 个 card
type GetCardRealtimeRequest struct {
	TenantID      string              `json:"tenant_id"`
	UserID        string              `json:"user_id"`
	CardsByBranch map[string][]string `json:"cards_by_branch"` // branchID → cardIDs
}

// GetCardRealtimeResponse 拉取实时数据响应
type GetCardRealtimeResponse struct {
	Data            map[string]json.RawMessage `json:"data"`                        // card_id -> realtime JSON
	SkippedCardIDs  []string                   `json:"skipped_card_ids"`            // 被跳过
	RefreshCardList bool                       `json:"refresh_card_list"`           // 建议前端重新调 GetCardList
	CardChanges     []CardChange               `json:"card_changes,omitempty"`      // pending add/update/delete
}

// GetCardRealtime 从 Redis 拉取指定 card 的实时数据
// 客户端提交 CardsByBranch，服务端按 branch 对照存储的 CardList 做权限校验
func (s *CardRealtimeService) GetCardRealtime(ctx context.Context, req GetCardRealtimeRequest) (*GetCardRealtimeResponse, error) {
	if req.TenantID == "" || req.UserID == "" {
		return nil, fmt.Errorf("tenant_id and user_id are required")
	}

	stored := s.getCardList(req.TenantID, req.UserID)

	resp := &GetCardRealtimeResponse{
		Data:            make(map[string]json.RawMessage),
		SkippedCardIDs:  make([]string, 0),
		RefreshCardList: stored == nil,
	}
	if stored == nil {
		return resp, nil
	}

	// 构建存储侧的 branchID → cardID set（用 map 做 O(1) 查找）
	storedSets := make(map[string]map[string]bool, len(stored.CardsByBranch))
	for bid, ids := range stored.CardsByBranch {
		m := make(map[string]bool, len(ids))
		for _, id := range ids {
			m[id] = true
		}
		storedSets[bid] = m
	}

	// 按客户端提交的 CardsByBranch 逐 branch 校验
	total := 0
	for branchID, cardIDs := range req.CardsByBranch {
		allowedSet := storedSets[branchID] // 该 branch 不存在则 nil
		for _, cardID := range cardIDs {
			if cardID == "" {
				continue
			}
			total++
			if total > maxPullRealtimeCards {
				return nil, fmt.Errorf("total cards exceeds limit %d", maxPullRealtimeCards)
			}
			if allowedSet == nil || !allowedSet[cardID] {
				resp.SkippedCardIDs = append(resp.SkippedCardIDs, cardID)
				continue
			}
			key := realtimeKeyPrefix + cardID + realtimeKeySuffix
			raw, err := s.kv.Get(ctx, key)
			if err != nil {
				resp.SkippedCardIDs = append(resp.SkippedCardIDs, cardID)
				if !errors.Is(err, store.ErrMiss) {
					s.logger.Warn("get realtime cache failed", zap.String("card_id", cardID), zap.Error(err))
				}
				continue
			}
			if !json.Valid([]byte(raw)) {
				resp.SkippedCardIDs = append(resp.SkippedCardIDs, cardID)
				continue
			}
			resp.Data[cardID] = json.RawMessage(raw)
		}
	}

	// 消费 pending cardChanges
	changes := s.ConsumeCardChanges(req.TenantID, req.UserID)
	if len(changes) > 0 {
		resp.CardChanges = changes
		resp.RefreshCardList = true
	}

	if len(resp.SkippedCardIDs) > 0 {
		resp.RefreshCardList = true
	}
	return resp, nil
}

// SubscribeRealtimeStream 建立 SSE 流：心跳 + 0.5Hz 从 streamProvider 缓存取 realtime/status（不订阅 Redis）
func (s *CardRealtimeService) SubscribeRealtimeStream(ctx context.Context, w http.ResponseWriter, cardID, tenantID, userID string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if s.streamProvider == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Service not available\"}\n\n")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Streaming not supported\"}\n\n")
		return
	}

	io.WriteString(w, ": SSE connection established\n")
	io.WriteString(w, "event: connected\ndata: {\"card_id\":\""+cardID+"\",\"status\":\"connected\"}\n\n")
	flusher.Flush()

	s.logger.Info("SSE connection established",
		zap.String("card_id", cardID),
		zap.String("tenant_id", tenantID),
		zap.String("user_id", userID))

	tickerHeart := time.NewTicker(30 * time.Second)
	defer tickerHeart.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-tickerHeart.C:
				io.WriteString(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}()

	var lastRealtimeRef unsafe.Pointer
	var lastStatusRef unsafe.Pointer
	var lastPushMu sync.Mutex
	var messageCounter int64

	tickerData := time.NewTicker(250 * time.Millisecond)
	defer tickerData.Stop()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("SSE connection closed", zap.String("card_id", cardID), zap.String("tenant_id", tenantID))
			return
		case <-tickerData.C:
			// 实时数据：从 DataStreamSubscriber 缓存取
			cachedRealtime := s.streamProvider.GetCardRealtimeData(cardID)
			if cachedRealtime != nil {
				lastPushMu.Lock()
				ref := unsafe.Pointer(&cachedRealtime)
				if ref != lastRealtimeRef {
					messageCounter++
					if messageCounter%2 == 0 {
						lastRealtimeRef = ref
						lastPushMu.Unlock()
						jsonData, _ := json.Marshal(cachedRealtime)
						io.WriteString(w, "data: "+string(jsonData)+"\n\n")
						flusher.Flush()
					} else {
						lastPushMu.Unlock()
					}
				} else {
					lastPushMu.Unlock()
				}
			}

			// 状态数据：从 DataStreamSubscriber 缓存取（与 realtime 同一订阅源，不另开 Redis 订阅）
			cachedStatus := s.streamProvider.GetCardStatusData(cardID)
			if cachedStatus != nil {
				lastPushMu.Lock()
				ref := unsafe.Pointer(&cachedStatus)
				if ref != lastStatusRef {
					lastStatusRef = ref
					lastPushMu.Unlock()
					statusMsg := buildCardStatusMessage(cardID, cachedStatus)
					if statusMsg != nil {
						jsonData, err := json.Marshal(statusMsg)
						if err == nil {
							io.WriteString(w, "event: card_status\ndata: "+string(jsonData)+"\n\n")
							flusher.Flush()
						}
					}
				} else {
					lastPushMu.Unlock()
				}
			}
		}
	}
}

// SubscribeCardsStream 多卡 SSE 流：前端传入 cardIDs + interval（秒）
// interval=1 → 1Hz，interval=2 → 0.5Hz
// 每次 tick 遍历 cardIDs，从 streamProvider 缓存取 realtime，组装 map[cardID]data 一次推送
// 权限校验：对比 stored CardList，过滤非法 cardID
func (s *CardRealtimeService) SubscribeCardsStream(ctx context.Context, w http.ResponseWriter, cardIDs []string, interval int, tenantID, userID string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if s.streamProvider == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Service not available\"}\n\n")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "event: error\ndata: {\"error\":\"Streaming not supported\"}\n\n")
		return
	}

	if interval < 1 {
		interval = 1
	}
	if interval > 10 {
		interval = 10
	}

	// 权限校验：取 stored CardList，构建 allowed set
	stored := s.getCardList(tenantID, userID)
	allowedSet := make(map[string]bool)
	if stored != nil {
		for _, ids := range stored.CardsByBranch {
			for _, id := range ids {
				allowedSet[id] = true
			}
		}
	}
	// 过滤出有权限的 cardIDs
	var validIDs []string
	for _, id := range cardIDs {
		if id != "" && allowedSet[id] {
			validIDs = append(validIDs, id)
		}
	}
	if len(validIDs) > maxPullRealtimeCards {
		validIDs = validIDs[:maxPullRealtimeCards]
	}

	connData, _ := json.Marshal(map[string]interface{}{
		"card_count": len(validIDs),
		"interval":   interval,
		"status":     "connected",
	})
	io.WriteString(w, ": SSE connection established\n")
	io.WriteString(w, "event: connected\ndata: "+string(connData)+"\n\n")
	flusher.Flush()

	// 注册 SSE 连接（fan-out 用）
	connID, statusCh, changeCh := s.registerSSE(validIDs, tenantID, userID)
	defer s.unregisterSSE(connID)

	s.logger.Info("SubscribeCardsStream connected",
		zap.String("conn_id", connID),
		zap.String("tenant_id", tenantID),
		zap.String("user_id", userID),
		zap.Int("card_count", len(validIDs)),
		zap.Int("interval", interval))

	// --- Snapshot：连接建立后立即推送当前缓存的最新数据 ---
	lastVersions := make(map[string]uint64, len(validIDs))
	snapshot := make(map[string]interface{}, len(validIDs))
	for _, cardID := range validIDs {
		if data := s.streamProvider.GetCardRealtimeData(cardID); data != nil {
			snapshot[cardID] = data
			lastVersions[cardID] = s.streamProvider.GetCardRealtimeVersion(cardID)
		}
	}
	if len(snapshot) > 0 {
		jsonSnap, _ := json.Marshal(snapshot)
		io.WriteString(w, "data: "+string(jsonSnap)+"\n\n")
		flusher.Flush()
	}

	// 心跳
	tickerHeart := time.NewTicker(30 * time.Second)
	defer tickerHeart.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-tickerHeart.C:
				io.WriteString(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}()

	tickerData := time.NewTicker(time.Duration(interval) * time.Second)
	defer tickerData.Stop()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("SubscribeCardsStream closed",
				zap.String("conn_id", connID),
				zap.String("tenant_id", tenantID),
				zap.String("user_id", userID))
			return

		case <-tickerData.C:
			// 定时推送：仅推送 version 有变化的 realtime 数据（dirty push）
			payload := make(map[string]interface{})
			for _, cardID := range validIDs {
				curVer := s.streamProvider.GetCardRealtimeVersion(cardID)
				if curVer > lastVersions[cardID] {
					if data := s.streamProvider.GetCardRealtimeData(cardID); data != nil {
						payload[cardID] = data
						lastVersions[cardID] = curVer
					}
				}
			}
			if len(payload) == 0 {
				continue
			}
			jsonData, err := json.Marshal(payload)
			if err != nil {
				continue
			}
			io.WriteString(w, "data: "+string(jsonData)+"\n\n")
			flusher.Flush()

		case evt, ok := <-statusCh:
			// status 事件即时推送（不等 ticker）
			if !ok {
				continue
			}
			statusMsg := buildCardStatusMessage(evt.CardID, evt.Data)
			if statusMsg == nil {
				continue
			}
			jsonData, err := json.Marshal(statusMsg)
			if err != nil {
				continue
			}
			io.WriteString(w, "event: card_status\ndata: "+string(jsonData)+"\n\n")
			flusher.Flush()

		case changes, ok := <-changeCh:
			// 卡片增删事件即时推送
			if !ok {
				continue
			}
			jsonData, err := json.Marshal(changes)
			if err != nil {
				continue
			}
			io.WriteString(w, "event: card_change\ndata: "+string(jsonData)+"\n\n")
			flusher.Flush()
		}
	}
}

// buildCardStatusMessage 将缓存中的 status（DataStreamSubscriber 存的是 stream message.Values）转为 SSE 的 card_status 格式
func buildCardStatusMessage(cardID string, raw map[string]interface{}) map[string]interface{} {
	var dataValue map[string]interface{}
	if dataStr, ok := raw["data"].(string); ok {
		_ = json.Unmarshal([]byte(dataStr), &dataValue)
	} else if dataMap, ok := raw["data"].(map[string]interface{}); ok {
		dataValue = dataMap
	}
	if dataValue == nil {
		dataValue = make(map[string]interface{})
	}
	eventType := "unknown"
	if et, ok := dataValue["event_type"].(string); ok {
		eventType = et
	}
	var timestamp int64
	if tsStr, ok := raw["timestamp"].(string); ok {
		if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
			timestamp = ts
		}
	}
	if timestamp == 0 {
		if tsVal, ok := dataValue["timestamp"]; ok {
			if ts, ok := tsVal.(float64); ok {
				timestamp = int64(ts)
			}
		}
	}
	return map[string]interface{}{
		"card_id": cardID, "event_type": eventType, "timestamp": timestamp, "data_value": dataValue,
	}
}
