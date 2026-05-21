// unit_picker.go — /80 unit 卡的 display 投影：跨 /88 /96 子卡挑 active 一份显示。
//
// 触发：SensorStateProjector / AlarmRouter 写完 /88 /96 子卡 state 后调
// UnitPicker.RefreshParent(childCardID)；picker 自动派生父 /80，扫子卡 state，
// 选 winner（risk > people > inbed > recency），合成 /80 display 写入。
//
// 同时支持 /80 unit 卡自身收到 alarm/visitor 等顶层写入：RefreshSelf 重算 display。
//
// children 缓存：unitPref → []child（含 RoomName），TTL 60s；CardLifecycle 改/增/删时
// 调 InvalidateUnit 清缓存。
//
// 设计纪律：
//   - 唯一权威 writer = cardagg（rule #1.3）；picker 写 /80 display 不踩其它 owner 字段
//   - 不写 /80 的 bed_state / room_state（那是子 /88 /96 的事）
//   - 一次性 Pipeline 批读子卡 hash，单次 HSET 落 /80 display + Section1.down 标签

package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/netip"
	"sync"
	"time"

	"owl-common/card"
	"wisefido-cardagg/internal/service"

	redislib "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	unitChildrenTTL  = 60 * time.Second
	unitPickerSelfTimeout = 2 * time.Second
)

type unitChild struct {
	CardID   string // /88 or /96 CIDR text
	Masklen  int
	RoomName string // rooms.room_name (派生自 /88 prefix；/96 子 → 上一级 /88 room)
}

type unitChildrenEntry struct {
	children  []unitChild
	fetchedAt time.Time
}

type UnitPicker struct {
	db     *sql.DB
	client *redislib.Client
	writer *card.Writer
	reader *card.Reader
	meta   *service.DeviceMetaCache
	logger *zap.Logger

	mu       sync.Mutex
	children map[string]*unitChildrenEntry
}

func NewUnitPicker(db *sql.DB, client *redislib.Client, writer *card.Writer, reader *card.Reader, meta *service.DeviceMetaCache, logger *zap.Logger) *UnitPicker {
	return &UnitPicker{
		db:       db,
		client:   client,
		writer:   writer,
		reader:   reader,
		meta:     meta,
		logger:   logger,
		children: make(map[string]*unitChildrenEntry),
	}
}

// RefreshParent 子卡（/88 或 /96）状态变化后重算父 /80 display。childCardID 是 INET CIDR 文本。
// 父 /80 不存在（DB 无 /80 unit 卡） / 解析失败 → 静默 no-op。
func (p *UnitPicker) RefreshParent(ctx context.Context, childCardID string) {
	unitPref, ok := deriveUnitPrefix(childCardID)
	if !ok || unitPref == childCardID {
		return
	}
	p.refresh(ctx, unitPref)
}

// RefreshSelf /80 unit 卡自身的 state 变化（alarm / target / visitor）后重算 display。
// 仍走 children 路径 —— /80 display 永远是子卡 picker 的产物 + /80 自身 alarm/target。
func (p *UnitPicker) RefreshSelf(ctx context.Context, unitCardID string) {
	if !isUnitPrefix(unitCardID) {
		return
	}
	p.refresh(ctx, unitCardID)
}

// InvalidateUnit CardLifecycle 增/改/删卡后清 children cache。
func (p *UnitPicker) InvalidateUnit(unitPref string) {
	if unitPref == "" {
		return
	}
	p.mu.Lock()
	delete(p.children, unitPref)
	p.mu.Unlock()
}

// InvalidateAll cache 全清（reset op）。
func (p *UnitPicker) InvalidateAll() {
	p.mu.Lock()
	p.children = make(map[string]*unitChildrenEntry)
	p.mu.Unlock()
}

func (p *UnitPicker) refresh(ctx context.Context, unitPref string) {
	children := p.getChildren(ctx, unitPref)
	if len(children) == 0 {
		return
	}
	states := p.batchReadChildStates(ctx, children)
	prevSelf, _ := p.reader.ReadCardStatus(ctx, unitPref)

	winner, roomName := pickActiveChild(children, states)
	// unit /80 hasBed / isBathroom = winner 子卡静态属性（winner 是 bed/room 卡 → 从 meta 取）
	hasBed := false
	isBath := false
	if winner != "" && p.meta != nil {
		m := p.meta.GetOrLoad(ctx, winner)
		hasBed = m.HasBed()
		isBath = m.IsBathroom()
	}
	display := buildUnitDisplay(unitPref, winner, roomName, states, prevSelf, hasBed, isBath)
	if display == nil {
		return
	}
	if err := p.writer.WriteCardStatus(ctx, &card.CardStatus{
		CardID:  unitPref,
		Display: display,
	}); err != nil {
		p.logger.Warn("unit picker write", zap.String("unit", unitPref), zap.Error(err))
	}
}

// getChildren 查 unit /80 下所有 /88 /96 子卡（cache 60s）。
func (p *UnitPicker) getChildren(ctx context.Context, unitPref string) []unitChild {
	p.mu.Lock()
	entry := p.children[unitPref]
	if entry != nil && time.Since(entry.fetchedAt) < unitChildrenTTL {
		out := entry.children
		p.mu.Unlock()
		return out
	}
	p.mu.Unlock()

	children := p.queryChildren(ctx, unitPref)
	p.mu.Lock()
	p.children[unitPref] = &unitChildrenEntry{children: children, fetchedAt: time.Now()}
	p.mu.Unlock()
	return children
}

func (p *UnitPicker) queryChildren(ctx context.Context, unitPref string) []unitChild {
	// v2.5: cards 都是 /88，住在 unit /80 下；rooms.card_id FK 反挂
	rows, err := p.db.QueryContext(ctx, `
		SELECT c.card_id::text,
		       masklen(c.card_id),
		       COALESCE((SELECT room_name FROM rooms WHERE card_id = c.card_id ORDER BY room_slot LIMIT 1), '')
		  FROM cards c
		 WHERE c.unit_id = $1::INET`, unitPref)
	if err != nil {
		p.logger.Warn("query unit children", zap.String("unit", unitPref), zap.Error(err))
		return nil
	}
	defer rows.Close()
	var out []unitChild
	for rows.Next() {
		var c unitChild
		if err := rows.Scan(&c.CardID, &c.Masklen, &c.RoomName); err != nil {
			continue
		}
		out = append(out, c)
	}
	return out
}

// batchReadChildStates pipeline 读子卡 room_state + bed_state + display JSON。
// 返回 cardID → 解析后的 CardStatus（填 RoomState/BedState/Display）。
func (p *UnitPicker) batchReadChildStates(ctx context.Context, children []unitChild) map[string]*card.CardStatus {
	if len(children) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, unitPickerSelfTimeout)
	defer cancel()

	pipe := p.client.Pipeline()
	cmds := make([]*redislib.SliceCmd, 0, len(children))
	for _, ch := range children {
		cmds = append(cmds, pipe.HMGet(ctx, card.HashKey(ch.CardID), "room_state", "bed_state", "display"))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redislib.Nil {
		p.logger.Warn("pipeline child states", zap.Error(err))
		return nil
	}
	out := make(map[string]*card.CardStatus, len(children))
	for i, ch := range children {
		vals, err := cmds[i].Result()
		if err != nil || len(vals) != 3 {
			continue
		}
		s := &card.CardStatus{CardID: ch.CardID}
		if str, ok := vals[0].(string); ok && str != "" && str != "{}" {
			var rs card.RoomState
			if json.Unmarshal([]byte(str), &rs) == nil {
				s.RoomState = &rs
			}
		}
		if str, ok := vals[1].(string); ok && str != "" && str != "{}" {
			var bs card.BedState
			if json.Unmarshal([]byte(str), &bs) == nil {
				s.BedState = &bs
			}
		}
		if str, ok := vals[2].(string); ok && str != "" && str != "{}" {
			var dd card.CardDisplay
			if json.Unmarshal([]byte(str), &dd) == nil {
				s.Display = &dd
			}
		}
		out[ch.CardID] = s
	}
	return out
}

// pickActiveChild 子卡挑选：按 display.card_priority 取 MAX（spec card_display.md §3.2）；
// 同 priority 按 updated_at 取最新。无 display 或 priority=0 的子卡参与兜底（priority=0 默认）。
func pickActiveChild(children []unitChild, states map[string]*card.CardStatus) (string, string) {
	var pickID, pickRoom string
	var bestPriority int = -1
	var bestUpdated int64

	for _, ch := range children {
		s := states[ch.CardID]
		if s == nil {
			continue
		}
		var prio int
		if s.Display != nil {
			prio = s.Display.CardPriority
		}
		updated := stateUpdatedAt(s)
		if prio > bestPriority || (prio == bestPriority && updated > bestUpdated) {
			bestPriority = prio
			bestUpdated = updated
			pickID = ch.CardID
			pickRoom = ch.RoomName
		}
	}
	return pickID, pickRoom
}

// buildUnitDisplay 给 /80 卡合成 display：
//   - 拿 winner 的 RoomState/BedState 作为 Section2.left/Section3.right 派生输入
//   - /80 自身的 AlarmState/Target 仍用（visitor / Section1DownRight）
//   - Section1DownLeft = winner.room_name 或 "WholeUnit"
func buildUnitDisplay(unitPref, winnerID, roomName string, states map[string]*card.CardStatus, prevSelf *card.CardStatus, hasBedDevice bool, isBathroom bool) *card.CardDisplay {
	merged := &card.CardStatus{CardID: unitPref}
	if prevSelf != nil {
		merged.AlarmState = prevSelf.AlarmState
		merged.Target = prevSelf.Target
	}
	if winnerID != "" {
		if w := states[winnerID]; w != nil {
			merged.RoomState = w.RoomState
			merged.BedState = w.BedState
		}
	}

	d := BuildCardDisplay(merged, hasBedDevice, isBathroom)
	if d == nil {
		return nil
	}
	_ = roomName
	return d
}

// deriveUnitPrefix CIDR text → 父 /80 CIDR text。仅当输入 /88 或 /96 时返回 ok=true。
func deriveUnitPrefix(cardID string) (string, bool) {
	pfx, err := netip.ParsePrefix(cardID)
	if err != nil {
		return "", false
	}
	bits := pfx.Bits()
	if bits != 88 && bits != 96 && bits != 128 {
		return "", false
	}
	return netip.PrefixFrom(pfx.Addr(), 80).Masked().String(), true
}

func isUnitPrefix(cardID string) bool {
	pfx, err := netip.ParsePrefix(cardID)
	if err != nil {
		return false
	}
	return pfx.Bits() == 80
}

// stateUpdatedAt 用 max(per-field ts) 派生"最新刷新时刻"，用于 UnitPicker tie-break。
func stateUpdatedAt(s *card.CardStatus) int64 {
	var m int64
	if s.RoomState != nil {
		if v := s.RoomState.MaxTs(); v > m {
			m = v
		}
	}
	if s.BedState != nil {
		if v := s.BedState.MaxTs(); v > m {
			m = v
		}
	}
	return m
}
