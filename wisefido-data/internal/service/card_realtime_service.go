package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

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
	GetCardStatusVersion(cardID string) uint64
}

const realtimeKeyPrefix = "vital-focus:card:"
const realtimeKeySuffix = ":realtime"
const maxPullRealtimeCards = 40

// CardChange 卡片变更记录
type CardChange struct {
	CardID string `json:"card_id"`
	Op     string `json:"op"` // "add" / "update" / "delete"
}

// SSEEventType SSE 推送事件类型
const (
	SSEEventStatus     = 1 // card_status（报警/设备状态/事件状态）
	SSEEventCardChange = 2 // card_change（卡片增删）
)

// SSEEvent 统一 SSE 推送事件
type SSEEvent struct {
	Type    int          // SSEEventStatus / SSEEventCardChange
	Status  *StatusEvent // Type=1 时有值
	Changes []CardChange // Type=2 时有值
}

// sseInitData 前端 POST /cards/stream/init 传入的初始化数据
type sseInitData struct {
	WatchIDs []string
	ViewIDs  []string
}

// sseConn 一个 SSE 连接的注册信息
type sseConn struct {
	userKey   string           // "tenantID:userID"
	tenantID  string
	userID    string
	watchIDs  map[string]bool  // 全部订阅卡（fan-out 范围），init 后填充
	viewIDsCh chan []string     // 切页时接收新 viewIDs
	eventCh   chan SSEEvent     // 统一事件推送 channel
	initCh    chan sseInitData  // 首次 init 数据（watchIDs + viewIDs）
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
	sseConns  map[string]*sseConn // connID → sseConn
	cardIndex map[string][]string // cardID → []connID（反向索引）
	userIndex map[string][]string // userKey → []connID（按用户路由 card_change）
	connSeq   uint64              // 自增连接 ID

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

// registerSSE 建立 SSE 连接（不注册 fan-out，等 InitSSE 后再注册）
func (s *CardRealtimeService) registerSSE(tenantID, userID string) (string, <-chan SSEEvent, <-chan []string, <-chan sseInitData) {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	s.connSeq++
	connID := fmt.Sprintf("sse-%d", s.connSeq)
	userKey := tenantID + ":" + userID

	eventCh := make(chan SSEEvent, 12)
	viewIDsCh := make(chan []string, 2)
	initCh := make(chan sseInitData, 1)
	s.sseConns[connID] = &sseConn{
		userKey:   userKey,
		tenantID:  tenantID,
		userID:    userID,
		watchIDs:  nil, // init 后填充
		viewIDsCh: viewIDsCh,
		eventCh:   eventCh,
		initCh:    initCh,
	}
	// userKey 反向索引
	s.userIndex[userKey] = append(s.userIndex[userKey], connID)

	return connID, eventCh, viewIDsCh, initCh
}

// activateSSE 在 InitSSE 后注册 fan-out 索引（需持有 connMu.Lock）
func (s *CardRealtimeService) activateSSE(connID string, watchIDs []string) {
	conn, ok := s.sseConns[connID]
	if !ok {
		return
	}
	watchSet := make(map[string]bool, len(watchIDs))
	for _, id := range watchIDs {
		watchSet[id] = true
	}
	conn.watchIDs = watchSet
	for _, cardID := range watchIDs {
		s.cardIndex[cardID] = append(s.cardIndex[cardID], connID)
	}
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
	for cardID := range conn.watchIDs {
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
	close(conn.eventCh)
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
				s.logger.Info("[FANOUT] status event",
					zap.String("card_id", evt.CardID),
					zap.Int("conn_count", len(connIDs)))
				for _, connID := range connIDs {
					if conn, exists := s.sseConns[connID]; exists {
						select {
						case conn.eventCh <- SSEEvent{Type: SSEEventStatus, Status: &evt}:
						default:
							s.logger.Warn("[FANOUT] eventCh full, dropped status",
								zap.String("conn_id", connID),
								zap.String("card_id", evt.CardID))
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
// 扫同一 tenant 的所有在线用户，重查 CardList（用户可能新增了对该 branch 的可见卡片）
func (s *CardRealtimeService) UpdateByBranch(ctx context.Context, tenantID, branchID string) {
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
			case conn.eventCh <- SSEEvent{Type: SSEEventCardChange, Changes: changes}:
			default:
				s.logger.Warn("[FANOUT] eventCh full, dropped card_change",
					zap.String("conn_id", connID))
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
	Data            map[string]json.RawMessage `json:"data"`                   // card_id -> realtime JSON
	SkippedCardIDs  []string                   `json:"skipped_card_ids"`       // 被跳过
	RefreshCardList bool                       `json:"refresh_card_list"`      // 建议前端重新调 GetCardList
	CardChanges     []CardChange               `json:"card_changes,omitempty"` // pending add/update/delete
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
	// CORS 由 AuthMiddleware 统一处理

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

	var lastRealtimeRef unsafe.Pointer
	var lastStatusRef unsafe.Pointer
	var lastPushMu sync.Mutex
	var messageCounter int64

	// 心跳 + 数据推送都在同一个 goroutine，避免并发写 ResponseWriter
	tickerHeart := time.NewTicker(30 * time.Second)
	defer tickerHeart.Stop()
	tickerData := time.NewTicker(1000 * time.Millisecond)
	defer tickerData.Stop()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("SSE connection closed", zap.String("card_id", cardID), zap.String("tenant_id", tenantID))
			return
		case <-tickerHeart.C:
			io.WriteString(w, ": heartbeat\n\n")
			flusher.Flush()
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

// SubscribeCardsStream 多卡 SSE 流
// URL 只传 token + interval，连接建立后前端 POST /init 传 watchIDs + viewIDs
// 切页时通过 POST /view 动态更新 viewIDs，无需重连 SSE
func (s *CardRealtimeService) SubscribeCardsStream(ctx context.Context, w http.ResponseWriter, interval int, tenantID, userID string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

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

	// 注册 SSE 连接（尚未注册 fan-out，等 InitSSE）
	connID, eventCh, viewIDsCh, initCh := s.registerSSE(tenantID, userID)
	defer s.unregisterSSE(connID)

	connData, _ := json.Marshal(map[string]interface{}{
		"conn_id":  connID,
		"interval": interval,
		"status":   "connected",
	})
	io.WriteString(w, ": SSE connection established\n")
	io.WriteString(w, "event: connected\ndata: "+string(connData)+"\n\n")
	flusher.Flush()

	s.logger.Info("SubscribeCardsStream connected, waiting init",
		zap.String("conn_id", connID),
		zap.String("tenant_id", tenantID),
		zap.String("user_id", userID),
		zap.Int("interval", interval))

	// === 阶段 1：等待前端 POST /init 提交 watchIDs + viewIDs ===
	var initData sseInitData
	// 等 init 期间仍处理心跳和断连
	tickerWait := time.NewTicker(30 * time.Second)
	defer tickerWait.Stop()
	initTimeout := time.After(60 * time.Second)
waitInit:
	for {
		select {
		case <-ctx.Done():
			return
		case <-initTimeout:
			io.WriteString(w, "event: error\ndata: {\"error\":\"init timeout\"}\n\n")
			flusher.Flush()
			return
		case <-tickerWait.C:
			io.WriteString(w, ": heartbeat\n\n")
			flusher.Flush()
		case initData = <-initCh:
			break waitInit
		}
	}
	tickerWait.Stop()

	watchIDs := initData.WatchIDs
	viewIDs := initData.ViewIDs

	// 通知前端 init 完成
	initAck, _ := json.Marshal(map[string]interface{}{
		"conn_id":     connID,
		"watch_count": len(watchIDs),
		"view_count":  len(viewIDs),
		"status":      "ready",
	})
	io.WriteString(w, "event: ready\ndata: "+string(initAck)+"\n\n")
	flusher.Flush()

	s.logger.Info("SubscribeCardsStream initialized",
		zap.String("conn_id", connID),
		zap.Int("watch_count", len(watchIDs)),
		zap.Int("view_count", len(viewIDs)))

	// === 阶段 2：推送 snapshot ===
	lastVersions := make(map[string]uint64, len(watchIDs))
	s.pushRealtimeSnapshot(w, flusher, connID, viewIDs, lastVersions)
	s.pushStatusSnapshot(w, flusher, connID, watchIDs)

	// === 阶段 3：持续推送 ===
	tickerHeart := time.NewTicker(30 * time.Second)
	defer tickerHeart.Stop()
	tickerData := time.NewTicker(time.Duration(interval) * time.Second)
	defer tickerData.Stop()
	currentViewIDs := viewIDs
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("SubscribeCardsStream closed",
				zap.String("conn_id", connID),
				zap.String("tenant_id", tenantID),
				zap.String("user_id", userID))
			return

		case <-tickerHeart.C:
			io.WriteString(w, ": heartbeat\n\n")
			flusher.Flush()

		case newViewIDs := <-viewIDsCh:
			currentViewIDs = newViewIDs
			s.pushRealtimeSnapshot(w, flusher, connID, currentViewIDs, lastVersions)
			s.logger.Info("[SSE] view updated",
				zap.String("conn_id", connID),
				zap.Int("new_view_count", len(currentViewIDs)))

		case <-tickerData.C:
			payload := make(map[string]interface{})
			for _, cardID := range currentViewIDs {
				curVer := s.streamProvider.GetCardRealtimeVersion(cardID)
				if curVer <= lastVersions[cardID] {
					continue
				}
				data := s.streamProvider.GetCardRealtimeData(cardID)
				if data == nil {
					continue
				}
				if rm, ok := toMap(data); ok {
					payload[cardID] = rm
				}
				lastVersions[cardID] = curVer
			}
			if len(payload) == 0 {
				continue
			}
			jsonData, err := json.Marshal(payload)
			if err != nil {
				continue
			}
			s.logger.Info("[SSE-PUSH] ticker realtime",
				zap.String("conn_id", connID),
				zap.Int("cards", len(payload)),
				zap.Int("bytes", len(jsonData)),
				zap.String("data_preview", truncStr(string(jsonData), 300)))
			io.WriteString(w, "data: "+string(jsonData)+"\n\n")
			flusher.Flush()

		case sseEvt, ok := <-eventCh:
			if !ok {
				continue
			}
			switch sseEvt.Type {
		case SSEEventStatus:
			if sseEvt.Status == nil {
				continue
			}
			statusMsg := buildCardStatusMessage(sseEvt.Status.CardID, sseEvt.Status.Data)
			if statusMsg == nil {
				s.logger.Warn("[SSE-PUSH] buildCardStatusMessage returned nil",
					zap.String("conn_id", connID),
					zap.String("card_id", sseEvt.Status.CardID))
				continue
			}
			jsonData, err := json.Marshal(statusMsg)
			if err != nil {
				s.logger.Warn("[SSE-PUSH] marshal card_status failed",
					zap.String("conn_id", connID),
					zap.Error(err))
				continue
			}
			s.logger.Info("[SSE-PUSH] card_status event",
				zap.String("conn_id", connID),
				zap.String("card_id", sseEvt.Status.CardID),
				zap.Int("bytes", len(jsonData)))
			io.WriteString(w, "event: card_status\ndata: "+string(jsonData)+"\n\n")
			flusher.Flush()
			case SSEEventCardChange:
				jsonData, err := json.Marshal(sseEvt.Changes)
				if err != nil {
					continue
				}
				io.WriteString(w, "event: card_change\ndata: "+string(jsonData)+"\n\n")
				flusher.Flush()
			}
		}
	}
}

// buildCardStatusMessage 将 card:status:stream 的消息转为 SSE card_status 事件
// raw 是 DataStreamSubscriber 缓存的 stream message.Values（PublishJSONToStream 包装：{data: "<CardStatus JSON>", timestamp: "..."}）
// 输出：直接透传 CardStatus 快照 + card_id，前端按快照字段处理
func buildCardStatusMessage(cardID string, raw map[string]interface{}) map[string]interface{} {
	if len(raw) == 0 {
		return nil
	}
	var cardStatus map[string]interface{}
	if dataStr, ok := raw["data"].(string); ok {
		_ = json.Unmarshal([]byte(dataStr), &cardStatus)
	} else if dataMap, ok := raw["data"].(map[string]interface{}); ok {
		cardStatus = copyMapShallow(dataMap)
	} else {
		cardStatus = copyMapShallow(raw)
	}
	if cardStatus == nil {
		return nil
	}
	cardStatus["card_id"] = cardID
	return cardStatus
}

func copyMapShallow(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// toMap 将任意类型转为 map[string]interface{}
func toMap(raw interface{}) (map[string]interface{}, bool) {
	switch v := raw.(type) {
	case map[string]interface{}:
		return v, true
	default:
		b, err := json.Marshal(raw)
		if err != nil {
			return nil, false
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, false
		}
		return m, true
	}
}

// pushRealtimeSnapshot 推送 viewIDs 的 CardRealTime 快照（data: 事件）
func (s *CardRealtimeService) pushRealtimeSnapshot(w http.ResponseWriter, flusher http.Flusher, connID string, viewIDs []string, lastVersions map[string]uint64) {
	snapshot := make(map[string]interface{}, len(viewIDs))
	for _, cardID := range viewIDs {
		data := s.streamProvider.GetCardRealtimeData(cardID)
		ver := s.streamProvider.GetCardRealtimeVersion(cardID)
		if data != nil {
			if rm, ok := toMap(data); ok {
				snapshot[cardID] = rm
			}
			lastVersions[cardID] = ver
		}
	}
	if len(snapshot) > 0 {
		jsonSnap, _ := json.Marshal(snapshot)
		s.logger.Info("[SSE-PUSH] realtime snapshot",
			zap.String("conn_id", connID),
			zap.Int("cards", len(snapshot)),
			zap.Int("bytes", len(jsonSnap)),
			zap.String("data_preview", truncStr(string(jsonSnap), 300)))
		io.WriteString(w, "data: "+string(jsonSnap)+"\n\n")
		flusher.Flush()
	}
}

// pushStatusSnapshot 推送 watchIDs 的 CardStatus 快照（event: card_status 逐卡透传）
// 优先从 subscriber 缓存读，缓存 miss 的批量 MGet Redis fallback
func (s *CardRealtimeService) pushStatusSnapshot(w http.ResponseWriter, flusher http.Flusher, connID string, watchIDs []string) {
	statusMap := make(map[string]map[string]interface{}, len(watchIDs))
	var missIDs []string

	for _, cardID := range watchIDs {
		data := s.streamProvider.GetCardStatusData(cardID)
		if data != nil {
			statusMap[cardID] = data
		} else {
			missIDs = append(missIDs, cardID)
		}
	}

	if len(missIDs) > 0 {
		batch := s.loadCardStatusBatch(missIDs)
		for id, data := range batch {
			statusMap[id] = data
		}
	}

	count := 0
	for _, cardID := range watchIDs {
		data := statusMap[cardID]
		if data == nil {
			continue
		}
		statusMsg := buildCardStatusMessage(cardID, data)
		if statusMsg == nil {
			continue
		}
		jsonData, err := json.Marshal(statusMsg)
		if err != nil {
			continue
		}
		io.WriteString(w, "event: card_status\ndata: "+string(jsonData)+"\n\n")
		count++
	}
	if count > 0 {
		flusher.Flush()
		s.logger.Info("[SSE-PUSH] status snapshot",
			zap.String("conn_id", connID),
			zap.Int("cards", count),
			zap.Int("from_cache", count-len(missIDs)),
			zap.Int("from_redis", len(missIDs)))
	}
}

// loadCardStatusFromRedis 批量从 Redis key 读取 CardStatus（缓存 miss 时 fallback）
func (s *CardRealtimeService) loadCardStatusBatch(cardIDs []string) map[string]map[string]interface{} {
	if len(cardIDs) == 0 {
		return nil
	}
	keys := make([]string, len(cardIDs))
	for i, id := range cardIDs {
		keys[i] = "vital-focus:card:" + id + ":status"
	}
	vals, err := s.kv.MGet(context.Background(), keys)
	if err != nil {
		s.logger.Warn("loadCardStatusBatch MGet failed", zap.Error(err))
		return nil
	}
	result := make(map[string]map[string]interface{}, len(cardIDs))
	for i, val := range vals {
		if val == "" {
			continue
		}
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(val), &data); err != nil {
			continue
		}
		result[cardIDs[i]] = data
	}
	return result
}

// GetCardStatus 从 Redis 读取单张卡片的 CardStatus（供 REST 轮询）
func (s *CardRealtimeService) GetCardStatus(ctx context.Context, cardID string) (map[string]interface{}, error) {
	key := "vital-focus:card:" + cardID + ":status"
	val, err := s.kv.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if val == "" {
		return nil, nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}
	return data, nil
}

// InitSSE 初始化 SSE 连接的 watchIDs + viewIDs（连接建立后前端 POST 调用）
func (s *CardRealtimeService) InitSSE(connID string, watchIDs, viewIDs []string) error {
	s.connMu.RLock()
	conn, ok := s.sseConns[connID]
	s.connMu.RUnlock()
	if !ok {
		return fmt.Errorf("SSE connection %s not found", connID)
	}
	if conn.watchIDs != nil {
		return fmt.Errorf("SSE connection %s already initialized", connID)
	}

	// 权限校验
	stored := s.getCardList(conn.tenantID, conn.userID)
	allowedSet := make(map[string]bool)
	if stored != nil {
		for _, ids := range stored.CardsByBranch {
			for _, id := range ids {
				allowedSet[id] = true
			}
		}
	}
	var filteredWatch []string
	for _, id := range watchIDs {
		if id != "" && allowedSet[id] {
			filteredWatch = append(filteredWatch, id)
		}
	}

	// 注册 fan-out
	s.connMu.Lock()
	s.activateSSE(connID, filteredWatch)
	s.connMu.Unlock()

	// 过滤 viewIDs（必须是 watchIDs 的子集）
	watchSet := conn.watchIDs
	var filteredView []string
	for _, id := range viewIDs {
		if id != "" && watchSet[id] {
			filteredView = append(filteredView, id)
		}
	}

	// 通知 SSE goroutine 开始推送
	select {
	case conn.initCh <- sseInitData{WatchIDs: filteredWatch, ViewIDs: filteredView}:
	default:
	}
	return nil
}

// UpdateSSEView 更新 SSE 连接的 viewIDs（切页时调用）
func (s *CardRealtimeService) UpdateSSEView(connID string, newViewIDs []string) error {
	s.connMu.RLock()
	conn, ok := s.sseConns[connID]
	s.connMu.RUnlock()
	if !ok {
		return fmt.Errorf("SSE connection %s not found", connID)
	}
	// 过滤：newViewIDs 必须是 watchIDs 的子集
	var filtered []string
	for _, id := range newViewIDs {
		if conn.watchIDs[id] {
			filtered = append(filtered, id)
		}
	}
	select {
	case conn.viewIDsCh <- filtered:
	default:
		s.logger.Warn("[SSE] viewIDsCh full, dropping update",
			zap.String("conn_id", connID))
	}
	return nil
}

// truncStr 截断字符串（日志用）
func truncStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
