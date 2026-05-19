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
	logger *zap.Logger

	mu       sync.Mutex
	children map[string]*unitChildrenEntry
}

func NewUnitPicker(db *sql.DB, client *redislib.Client, writer *card.Writer, reader *card.Reader, logger *zap.Logger) *UnitPicker {
	return &UnitPicker{
		db:       db,
		client:   client,
		writer:   writer,
		reader:   reader,
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
	display := buildUnitDisplay(unitPref, winner, roomName, states, prevSelf)
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
	rows, err := p.db.QueryContext(ctx, `
		SELECT c.spatial_prefix::text,
		       masklen(c.spatial_prefix),
		       COALESCE(r.room_name, '')
		  FROM cards c
		  LEFT JOIN rooms r
		    ON r.room_id = network(set_masklen(c.spatial_prefix, 88))
		 WHERE c.spatial_prefix << $1::INET
		   AND masklen(c.spatial_prefix) IN (88, 96)`, unitPref)
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

// batchReadChildStates pipeline 读子卡 room_state + bed_state JSON。
// 返回 cardID → 解析后的 CardStatus（只填了 RoomState/BedState）。
func (p *UnitPicker) batchReadChildStates(ctx context.Context, children []unitChild) map[string]*card.CardStatus {
	if len(children) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, unitPickerSelfTimeout)
	defer cancel()

	pipe := p.client.Pipeline()
	cmds := make([]*redislib.SliceCmd, 0, len(children))
	for _, ch := range children {
		cmds = append(cmds, pipe.HMGet(ctx, card.HashKey(ch.CardID), "room_state", "bed_state"))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redislib.Nil {
		p.logger.Warn("pipeline child states", zap.Error(err))
		return nil
	}
	out := make(map[string]*card.CardStatus, len(children))
	for i, ch := range children {
		vals, err := cmds[i].Result()
		if err != nil || len(vals) != 2 {
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
		out[ch.CardID] = s
	}
	return out
}

// pickActiveChild 子卡挑选：
//
//	P1: RoomState.RiskLevel > 0 + TotalPeople > 0  → 最高 risk + 最近 updated_at
//	P2: RoomState.TotalPeople > 0                    → 最近 updated_at
//	P3: BedState.BedStatus == 0 (在床)               → 最近 updated_at
//	P4: 最近 updated_at（任意 state 信号）
//	无任何信号 → ("", "")
func pickActiveChild(children []unitChild, states map[string]*card.CardStatus) (string, string) {
	var pickID, pickRoom string
	var bestRisk int
	var bestUpdated int64
	var pickPriority int // 1..4

	for _, ch := range children {
		s := states[ch.CardID]
		if s == nil {
			continue
		}
		updated := stateUpdatedAt(s)

		// P1
		if s.RoomState != nil && s.RoomState.RiskLevel > 0 && s.RoomState.TotalPeople > 0 {
			if pickPriority > 1 ||
				s.RoomState.RiskLevel > bestRisk ||
				(s.RoomState.RiskLevel == bestRisk && updated > bestUpdated) {
				pickPriority = 1
				bestRisk = s.RoomState.RiskLevel
				bestUpdated = updated
				pickID = ch.CardID
				pickRoom = ch.RoomName
			}
			continue
		}
		if pickPriority == 1 {
			continue
		}
		// P2
		if s.RoomState != nil && s.RoomState.TotalPeople > 0 {
			if pickPriority > 2 || updated > bestUpdated {
				pickPriority = 2
				bestUpdated = updated
				pickID = ch.CardID
				pickRoom = ch.RoomName
			}
			continue
		}
		if pickPriority <= 2 && pickPriority > 0 {
			continue
		}
		// P3
		if s.BedState != nil && s.BedState.BedStatus == 0 {
			if pickPriority > 3 || updated > bestUpdated {
				pickPriority = 3
				bestUpdated = updated
				pickID = ch.CardID
				pickRoom = ch.RoomName
			}
			continue
		}
		if pickPriority > 0 && pickPriority <= 3 {
			continue
		}
		// P4
		if updated > bestUpdated {
			pickPriority = 4
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
func buildUnitDisplay(unitPref, winnerID, roomName string, states map[string]*card.CardStatus, prevSelf *card.CardStatus) *card.CardDisplay {
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

	d := BuildCardDisplay(merged)
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

func stateUpdatedAt(s *card.CardStatus) int64 {
	var m int64
	if s.RoomState != nil && s.RoomState.UpdatedAt > m {
		m = s.RoomState.UpdatedAt
	}
	if s.BedState != nil && s.BedState.UpdatedAt > m {
		m = s.BedState.UpdatedAt
	}
	return m
}
