package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"owl-common/card"
	"wisefido-data/internal/store"

	"go.uber.org/zap"
)

// writeServerTick SSE keepalive：每 ~30s 推 server_ts 给前端维护 serverTimeRef。
// FE 用 serverTimeRef - section.anchor_ms 渲 DisplayTime（不用浏览器本地时钟），
// 解耦 monitor 流后卡静默时也能持续更新时长显示。
func writeServerTick(w io.Writer, flusher http.Flusher) {
	payload := `{"kind":"tick","server_ts":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + `}`
	io.WriteString(w, "event: tick\ndata: "+payload+"\n\n")
	flusher.Flush()
}

// sseHotPathLog：SSE 高频路径默认 Debug；设置 SSE_VERBOSE_LOG=true 时在 Info 级别可见（便于不改全局 level 排障）。
func (s *CardRealtimeService) sseHotPathLog(msg string, fields ...zap.Field) {
	if os.Getenv("SSE_VERBOSE_LOG") == "true" {
		s.logger.Info(msg, fields...)
		return
	}
	s.logger.Debug(msg, fields...)
}

// StatusEvent status 变更事件（service 包内定义，避免循环引用 subscriber 包）
type StatusEvent struct {
	CardID string
	Data   map[string]interface{}
}

// RealtimeDataProvider 供 SSE 使用，从 DataStreamSubscriber 缓存 + Redis Hash 读取
type RealtimeDataProvider interface {
	GetCardRealtimeData(cardID string) map[string]interface{}
	GetCardRealtimeVersion(cardID string) uint64
	GetCardStatusData(cardID string) map[string]interface{}
	GetCardStatusVersion(cardID string) uint64
	ReadCardStateSnapshot(ctx context.Context, cardID string) map[string]interface{}
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
	userKey     string // "tenantID:userID"
	tenantID    string
	userID      string
	clientIP    string
	userAgent   string
	connectedAt time.Time
	cancel      context.CancelFunc // 同用户新开 SSE 时用于结束旧连接（勿在持有 connMu 时调用）
	watchIDs    map[string]bool    // 全部订阅卡（fan-out 范围），init 后填充
	viewIDsCh   chan []string      // 切页时接收新 viewIDs
	watchIDsCh  chan []string      // card_change 后热更新 watchIDs，无需断 SSE
	eventCh     chan SSEEvent      // 统一事件推送 channel
	initCh      chan sseInitData   // 首次 init 数据（watchIDs + viewIDs）
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

	// GetCardStatus 用：device online 来自 card:state，alarm 来自 alarm_db
	stateReader *card.Reader
	db          *sql.DB
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

// SetGetCardStatusDeps 注入 GetCardStatus 用依赖：device online 用 card:state，alarm 用 alarm_db
func (s *CardRealtimeService) SetGetCardStatusDeps(stateReader *card.Reader, db *sql.DB) {
	s.stateReader = stateReader
	s.db = db
}

// StateReader 暴露注入的 cardagg state reader（card_static_service 给设备读 Redis device:status 用）
func (s *CardRealtimeService) StateReader() *card.Reader {
	return s.stateReader
}

// SetStatusEventChan 注入 status 事件 channel（由 main.go 桥接 subscriber → service）
func (s *CardRealtimeService) SetStatusEventChan(ch <-chan StatusEvent) {
	s.statusEventCh = ch
}

// sseSupersedePrevious：仅当 SSE_SINGLE_SSE_PER_USER=true 时，新开 SSE 会 cancel 同 tenant:user 下已有连接（治「僵尸长连接」）。默认 false：多标签、多网页、多设备可同时挂多条 SSE。
func sseSupersedePrevious() bool {
	return os.Getenv("SSE_SINGLE_SSE_PER_USER") == "true"
}

// registerSSE 建立 SSE 连接（不注册 fan-out，等 InitSSE 后再注册）。
// 返回的 connCtx 由 WithCancel(parentCtx) 派生，用于替换旧连接时仅结束本路 goroutine。
func (s *CardRealtimeService) registerSSE(parentCtx context.Context, tenantID, userID, clientIP, userAgent string) (connID string, connCtx context.Context, eventCh <-chan SSEEvent, viewIDsCh <-chan []string, watchIDsCh <-chan []string, initCh <-chan sseInitData) {
	userKey := tenantID + ":" + userID
	var superseded []string
	if sseSupersedePrevious() {
		s.connMu.Lock()
		if ids := s.userIndex[userKey]; len(ids) > 0 {
			superseded = append([]string(nil), ids...)
		}
		s.connMu.Unlock()
		for _, oid := range superseded {
			var fn context.CancelFunc
			s.connMu.Lock()
			if oc, ok := s.sseConns[oid]; ok && oc != nil && oc.cancel != nil {
				fn = oc.cancel
			}
			s.connMu.Unlock()
			if fn != nil {
				fn()
			}
		}
	}

	connCtx, cancel := context.WithCancel(parentCtx)
	s.connMu.Lock()
	defer s.connMu.Unlock()

	s.connSeq++
	connID = fmt.Sprintf("sse-%d", s.connSeq)
	now := time.Now()

	eventCh2 := make(chan SSEEvent, 12)
	viewIDsCh2 := make(chan []string, 2)
	watchIDsCh2 := make(chan []string, 2)
	initCh2 := make(chan sseInitData, 1)
	s.sseConns[connID] = &sseConn{
		userKey:     userKey,
		tenantID:    tenantID,
		userID:      userID,
		clientIP:    clientIP,
		userAgent:   userAgent,
		connectedAt: now,
		cancel:      cancel,
		watchIDs:    nil,
		viewIDsCh:   viewIDsCh2,
		watchIDsCh:  watchIDsCh2,
		eventCh:     eventCh2,
		initCh:      initCh2,
	}
	s.userIndex[userKey] = append(s.userIndex[userKey], connID)

	if len(superseded) > 0 {
		s.logger.Info("SSE superseded previous connection(s) for same user",
			zap.String("conn_id", connID),
			zap.Strings("superseded_conn_ids", superseded),
			zap.String("tenant_id", tenantID),
			zap.String("user_id", userID),
			zap.String("client_ip", clientIP))
	}

	return connID, connCtx, eventCh2, viewIDsCh2, watchIDsCh2, initCh2
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
	conn.cancel = nil
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
				s.sseHotPathLog("[FANOUT] status event",
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

	var lastRealtimeVer, lastStatusVer uint64

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
			writeServerTick(w, flusher)
		case <-tickerData.C:
			if curVer := s.streamProvider.GetCardRealtimeVersion(cardID); curVer > lastRealtimeVer {
				if data := s.streamProvider.GetCardRealtimeData(cardID); data != nil {
					if jsonData, err := json.Marshal(data); err == nil {
						io.WriteString(w, "data: "+string(jsonData)+"\n\n")
						flusher.Flush()
						lastRealtimeVer = curVer
					}
				}
			}
			if curVer := s.streamProvider.GetCardStatusVersion(cardID); curVer > lastStatusVer {
				if data := s.streamProvider.GetCardStatusData(cardID); data != nil {
					if jsonData, err := json.Marshal(data); err == nil {
						io.WriteString(w, "event: card_status\ndata: "+string(jsonData)+"\n\n")
						flusher.Flush()
						lastStatusVer = curVer
					}
				}
			}
		}
	}
}

// SubscribeCardsStream 多卡 SSE 流
// URL 只传 token + interval，连接建立后前端 POST /init 传 watchIDs + viewIDs
// 切页时通过 POST /view 动态更新 viewIDs，无需重连 SSE
func (s *CardRealtimeService) SubscribeCardsStream(ctx context.Context, w http.ResponseWriter, interval int, tenantID, userID, clientIP, userAgent string) {
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

	connID, streamCtx, eventCh, viewIDsCh, watchIDsCh, initCh := s.registerSSE(ctx, tenantID, userID, clientIP, userAgent)
	sseOpenedAt := time.Now()
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
		zap.String("client_ip", clientIP),
		zap.String("user_agent", truncStr(userAgent, 240)),
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
		case <-streamCtx.Done():
			return
		case <-initTimeout:
			io.WriteString(w, "event: error\ndata: {\"error\":\"init timeout\"}\n\n")
			flusher.Flush()
			return
		case <-tickerWait.C:
			writeServerTick(w, flusher)
		case initData = <-initCh:
			break waitInit
		}
	}
	tickerWait.Stop()

	watchIDs := initData.WatchIDs
	viewIDs := initData.ViewIDs

	// 通知前端 init 完成；带 server_ts 供前端时钟同步（毫秒）
	initAck, _ := json.Marshal(map[string]interface{}{
		"conn_id":     connID,
		"watch_count": len(watchIDs),
		"view_count":  len(viewIDs),
		"status":      "ready",
		"server_ts":   time.Now().UnixMilli(),
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
	watchIDsLive := watchIDs
	sseRealtimeTickerLogCount := 0
	for {
		select {
		case <-streamCtx.Done():
			s.logger.Info("SubscribeCardsStream closed",
				zap.String("conn_id", connID),
				zap.String("tenant_id", tenantID),
				zap.String("user_id", userID),
				zap.String("client_ip", clientIP),
				zap.String("user_agent", truncStr(userAgent, 120)),
				zap.Duration("uptime", time.Since(sseOpenedAt)))
			return

		case <-tickerHeart.C:
			writeServerTick(w, flusher)

		case newViewIDs := <-viewIDsCh:
			currentViewIDs = newViewIDs
			s.pushRealtimeSnapshot(w, flusher, connID, currentViewIDs, lastVersions)
			s.sseHotPathLog("[SSE] view updated",
				zap.String("conn_id", connID),
				zap.Int("new_view_count", len(currentViewIDs)))

		case newWatch := <-watchIDsCh:
			watchIDsLive = newWatch
			watchSet := make(map[string]bool, len(watchIDsLive))
			for _, id := range watchIDsLive {
				if id != "" {
					watchSet[id] = true
				}
			}
			prevView := append([]string(nil), currentViewIDs...)
			var filteredView []string
			for _, id := range prevView {
				if watchSet[id] {
					filteredView = append(filteredView, id)
				}
			}
			currentViewIDs = filteredView
			viewSet := make(map[string]bool, len(currentViewIDs))
			for _, id := range currentViewIDs {
				viewSet[id] = true
			}
			for id := range lastVersions {
				if !viewSet[id] {
					delete(lastVersions, id)
				}
			}
			s.pushRealtimeSnapshot(w, flusher, connID, currentViewIDs, lastVersions)
			s.pushStatusSnapshot(w, flusher, connID, watchIDsLive)
			s.sseHotPathLog("[SSE] watch updated",
				zap.String("conn_id", connID),
				zap.Int("new_watch_count", len(watchIDsLive)),
				zap.Int("view_count", len(currentViewIDs)))

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
			sseRealtimeTickerLogCount++
			if sseRealtimeTickerLogCount%20 == 0 {
				s.sseHotPathLog("[SSE-PUSH] ticker realtime",
					zap.String("conn_id", connID),
					zap.Int("cards", len(payload)),
					zap.Int("bytes", len(jsonData)),
					zap.String("data_preview", truncStr(string(jsonData), 300)))
			}
			io.WriteString(w, "data: "+string(jsonData)+"\n\n")
			flusher.Flush()

		case sseEvt, ok := <-eventCh:
			if !ok {
				continue
			}
			switch sseEvt.Type {
			case SSEEventStatus:
				if sseEvt.Status == nil || len(sseEvt.Status.Data) == 0 {
					continue
				}
				// device_status 由 cardagg 写入 device:status:{device_addr} hash；直接下发并附 server_ts 供前端时钟同步（毫秒）
				sendData := make(map[string]interface{}, len(sseEvt.Status.Data)+1)
				for k, v := range sseEvt.Status.Data {
					sendData[k] = v
				}
				sendData["server_ts"] = time.Now().UnixMilli()
				jsonData, err := json.Marshal(sendData)
				if err != nil {
					continue
				}
				s.sseHotPathLog("[SSE-PUSH] card_status event",
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
		s.sseHotPathLog("[SSE-PUSH] realtime snapshot",
			zap.String("conn_id", connID),
			zap.Int("cards", len(snapshot)),
			zap.Int("bytes", len(jsonSnap)),
			zap.String("data_preview", truncStr(string(jsonSnap), 300)))
		io.WriteString(w, "data: "+string(jsonSnap)+"\n\n")
		flusher.Flush()
	}
}

// pushStatusSnapshot 首次 init 后推送所有 watchIDs 的全量 card:state 快照（从 Redis 读）
// 推送格式：card_id + 具体 Data（target/room_state/bed_state/alarm_state/device_status/message）+ server_ts
func (s *CardRealtimeService) pushStatusSnapshot(w http.ResponseWriter, flusher http.Flusher, connID string, watchIDs []string) {
	ctx := context.Background()
	count := 0
	serverTs := time.Now().UnixMilli()
	// CardStatus 各字段 key，与 card_types.go 一致，便于前端按 key 合并
	dataKeys := []string{"target", "room_state", "bed_state", "alarm_state", "device_status", "message"}
	for _, cardID := range watchIDs {
		data := s.streamProvider.ReadCardStateSnapshot(ctx, cardID)
		if data == nil {
			continue
		}
		// Phase A：device_status 已迁出 card:state 独立 Hash，ReadCardStateSnapshot 不再含此键。
		// 这里补上 enrichDeviceStatus，让 SSE init snapshot 与 HTTP polling 行为一致。
		if err := s.enrichDeviceStatus(ctx, cardID, data); err != nil {
			s.logger.Warn("pushStatusSnapshot enrichDeviceStatus", zap.String("card_id", cardID), zap.Error(err))
		}
		payload := make(map[string]interface{}, len(dataKeys)+2)
		payload["card_id"] = cardID
		for _, k := range dataKeys {
			if v, ok := data[k]; ok && v != nil {
				payload[k] = v
			}
		}
		payload["server_ts"] = serverTs
		jsonData, err := json.Marshal(payload)
		if err != nil {
			continue
		}
		io.WriteString(w, "event: card_status\ndata: "+string(jsonData)+"\n\n")
		count++
	}
	if count > 0 {
		flusher.Flush()
		s.sseHotPathLog("[SSE-PUSH] status snapshot",
			zap.String("conn_id", connID),
			zap.Int("cards", count))
	}
}

// GetCardStatus Detail 页每 30 秒查询：device online 复用 FillDeviceOnlineStatusFromCardagg，alarm 用 alarm_db。
// 依赖（stateReader / db）由 main.go SetGetCardStatusDeps 启动期 wire。
func (s *CardRealtimeService) GetCardStatus(ctx context.Context, cardID string) (map[string]interface{}, error) {
	return s.getCardStatusFromQuery(ctx, cardID)
}

func (s *CardRealtimeService) getCardStatusFromQuery(ctx context.Context, cardID string) (map[string]interface{}, error) {
	nowMs := time.Now().UnixMilli()
	var out map[string]interface{}
	// 以 cardagg 全量 CardStatus 为基准：targets/room_state/bed_state/message 来自 Redis，不再动；仅下面对 device_status / alarm_state 覆盖
	if s.stateReader != nil {
		status, err := s.stateReader.ReadCardStatus(ctx, cardID)
		if err == nil && status != nil {
			out = cardStatusToMap(status)
		}
	}
	if out == nil {
		out = map[string]interface{}{"card_id": cardID}
	}
	out["timestamp"] = nowMs

	// device_status：DB 设备列表 + cardagg device:status hash（含 signal_poor/angle_abnormal/sensor_detached/last_seen_ms）
	if err := s.enrichDeviceStatus(ctx, cardID, out); err != nil {
		return nil, err
	}

	// 覆盖 alarm_state：DB 完整 AlarmState（与 CardStatus.AlarmState 对齐）
	if cardState, err := card.QueryCardAlarmState(ctx, s.db, cardID); err != nil {
		s.logger.Warn("GetCardStatus QueryCardAlarmState", zap.String("card_id", cardID), zap.Error(err))
	} else if cardState != nil {
		if as := cardState.ToAlarmState(); as != nil {
			out["alarm_state"] = cardAlarmStateToMap(as)
		}
	}
	return out, nil
}

// enrichDeviceStatus 把 cards.devices 列表 × device:status:{ipv6} hash 拼成完整 device_status map，
// 写入 data["device_status"]。供 SSE init (pushStatusSnapshot) 与 HTTP polling (getCardStatusFromQuery) 共用。
//
// 字段单一职责：`offline` 仅反映网络/心跳维度（= raw `ds.Offline`）；signal_poor / angle_abnormal /
// sensor_detached / last_seen_ms 作为 omitempty 明细字段暴露，前端按需决定渲染策略：
//   - Sleepad 卡片 SensorDetached 显示"Sensor Detached"文字（替代 Offline 文字）
//   - Detail 页 Online 字色随 degraded 状态变绿/黄
//   - 老客户端 JSON 解码默认忽略未知 key，看 `offline` 二态保持原行为；SensorDetached 漏报场景属于已知妥协，待 app 升级独立加图标
func (s *CardRealtimeService) enrichDeviceStatus(ctx context.Context, cardID string, data map[string]interface{}) error {
	if s.db == nil {
		return nil
	}
	// 范围用 /80 unit pool（不严格 LPM 到 /96），与 card_static_service.fillDevicesV3 对齐：
	//   fillDevicesV3 套娃归属：bed-anchor → room-anchor 吸收唯一 bed → /80 收口。
	//   例：room "Main" 只 1 bed card /96 "John.Y"；同 room 内 byte11=00 的 device D523 被吸收到该 /96。
	//   若 enrichDeviceStatus 严格 <<= /96 查，吸收进来的 D523 查不到 → 默认 offline，跟 FE 设备列表不一致。
	//   改 /80：返回 unit 内全部 monitor-on 设备的 status；FE 按 device_addr 查 map，多余 entry 自然忽略。
	rows, err := s.db.QueryContext(ctx, `
		SELECT host(d.device_addr), COALESCE(dfm.device_type::text, '') AS device_type
		FROM devices d
		JOIN device_factory_meta dfm ON dfm.device_uid = d.device_uid
		WHERE d.device_addr <<= network(set_masklen($1::INET, 80))
	`, cardID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var devices []struct{ deviceAddr, deviceType string }
	var deviceAddrs []string
	for rows.Next() {
		var addr, deviceType string
		if err := rows.Scan(&addr, &deviceType); err != nil {
			return err
		}
		devices = append(devices, struct{ deviceAddr, deviceType string }{addr, deviceType})
		deviceAddrs = append(deviceAddrs, addr)
	}

	nowMs := time.Now().UnixMilli()
	dsMap := FillDeviceStatusFromCardagg(ctx, s.stateReader, deviceAddrs, s.logger)
	deviceStatus := make(map[string]interface{}, len(devices))
	for _, d := range devices {
		entry := map[string]interface{}{
			"device_id":   d.deviceAddr, // FE map key 兼容历史命名（B1 待 FE 同步窗口改 device_addr）
			"device_type": d.deviceType,
		}
		if ds := dsMap[d.deviceAddr]; ds != nil {
			entry["offline"] = ds.Offline
			if ds.SignalPoor != 0 {
				entry["signal_poor"] = ds.SignalPoor
			}
			if ds.AngleAbnormal != 0 {
				entry["angle_abnormal"] = ds.AngleAbnormal
			}
			if ds.SensorDetached != 0 {
				entry["sensor_detached"] = ds.SensorDetached
			}
			if ds.LastSeenMs > 0 {
				entry["last_seen_ms"] = ds.LastSeenMs
			}
			if ds.UpdatedAt > 0 {
				entry["updated_at"] = ds.UpdatedAt
			} else {
				entry["updated_at"] = nowMs
			}
		} else {
			// hash 不存在 → 视为 offline
			entry["offline"] = 1
			entry["updated_at"] = nowMs
		}
		deviceStatus[d.deviceAddr] = entry
	}
	data["device_status"] = deviceStatus
	return nil
}

func cardStatusToMap(st *card.CardStatus) map[string]interface{} {
	if st == nil {
		return nil
	}
	b, _ := json.Marshal(st)
	var m map[string]interface{}
	if json.Unmarshal(b, &m) != nil {
		return map[string]interface{}{"card_id": st.CardID}
	}
	return m
}

func cardAlarmStateToMap(as *card.AlarmState) map[string]interface{} {
	if as == nil {
		return nil
	}
	b, _ := json.Marshal(as)
	var m map[string]interface{}
	if json.Unmarshal(b, &m) != nil {
		return nil
	}
	return m
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

// UpdateSSEWatch 热更新 watchIDs（card_change 后无需重连；与 InitSSE 相同权限过滤 + cardIndex 维护）
func (s *CardRealtimeService) UpdateSSEWatch(connID string, watchIDs []string) error {
	s.connMu.Lock()
	conn, ok := s.sseConns[connID]
	if !ok || conn == nil || conn.watchIDs == nil {
		s.connMu.Unlock()
		return fmt.Errorf("SSE connection %s not ready", connID)
	}

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

	oldSet := conn.watchIDs
	newSet := make(map[string]bool, len(filteredWatch))
	for _, id := range filteredWatch {
		newSet[id] = true
	}
	for cardID := range oldSet {
		if !newSet[cardID] {
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
	}
	for cardID := range newSet {
		if !oldSet[cardID] {
			s.cardIndex[cardID] = append(s.cardIndex[cardID], connID)
		}
	}
	conn.watchIDs = newSet
	ch := conn.watchIDsCh
	s.connMu.Unlock()

	select {
	case ch <- filteredWatch:
	default:
		s.logger.Warn("[SSE] watchIDsCh full, dropping watch update",
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
