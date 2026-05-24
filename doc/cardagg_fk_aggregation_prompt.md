# cardagg FK 反查 + 三层卡聚合治理 — 下次会话接续提示词

**前置 commit**: `03d2521 feat(sensor): Bayesian bed FSM + 自不否定物理原则 + 多源 LeftBed sticky`

**用户拍板的核心原则**（2026-05-23 讨论收口）：

1. **掩码是 type discriminator** — `/48 tenant / /56 branch / /64 site / /80 unit / /88 card or room / /96 bed / /128 device`，不需分多套 Go 类型，统一 `Entity{Prefix, Name, Kind}` 即可
2. **cardagg 必须知道 FK 归属** — 当前只 load `devices.card_id`，缺 `rooms.card_id` / `beds.card_id`，所以 sensor 来的 `/96` event 只能 LPM spatial prefix 猜父卡，不是 DB 真相
3. **event-driven 单一最新事件** — CardStatus 只发同级 latest event，FE 客户端按 (card_id, entity_id) 累积；**不需要 arrays in state struct**
4. **三层卡各自只显本级 latest event** —
   - `/96 bed`：own bed 最新事件
   - `/88 card`：own room state + 卡下任一 bed 最新事件 (event 带 bed_id 标定)
   - `/80 unit`：unit 级聚合 (alarm summary / total people sum)，**不持 bed_state / room_state**
5. **sensor 端已物理化干净** — `bed.state` 只发 /96、`room.state` 只发 /88、/80 sensor 完全不发任何 state；不要再动 sensor 端

---

## 现状证据（已经过实测）

### sensor 端清白（已 commit）

最近 500 条 `sensor:derived:stream`：
- `bed.state` 100% /96
- `bed.sleepstage` 100% /96
- `room.state` 100% /88
- `target.state` 100% /88
- `/80` 完全 0 条

### cardagg /80 unit 卡 hallucinate

cardagg `unit_picker.buildUnitDisplay` ([owlBack/wisefido-cardagg/internal/consumer/unit_picker.go:267-286](owlBack/wisefido-cardagg/internal/consumer/unit_picker.go#L267)) 错位 pin：
```go
merged.RoomState = w.RoomState   // ❌ /88 winner state pin 到 /80 unit
merged.BedState  = w.BedState    // ❌ /96 winner state pin 到 /80 unit
```
导致 /80 单元卡 FE 显 "InBed 1h 12m" 假象——/80 物理上不含床，是把某张 /96 床伪装成单元状态。

### cardagg FK 知识缺失

`CardMeta` ([owlBack/wisefido-cardagg/internal/service/device_meta.go:75](owlBack/wisefido-cardagg/internal/service/device_meta.go#L75)) 只 load `Devices map`，**完全没加载 rooms 表和 beds 表的 FK**。

sensor /96 event 来时只能 LPM spatial 猜：
```go
// owlBack/wisefido-cardagg/internal/consumer/sensor_state_projector.go:108
if resolved := p.meta.LookupCardByPrefix(ctx, msg.SubjectEntity); resolved != "" {
    destCardID = resolved   // ← LPM 猜，不是查 beds.card_id FK
}
```

LPM 用 spatial CIDR containment 猜，跟 DB `beds.card_id` 真相可能不一致。

### state struct 缺 entity_id

`owl-common/card.BedState` ([owlBack/owl-common/card/card_types.go:258](owlBack/owl-common/card/card_types.go#L258)) 和 `RoomState` ([line 389](owlBack/owl-common/card/card_types.go#L389)) **没有 BedID / RoomID 字段** — FE 接到事件不知道是哪张床。

Identifier 结构 `BedIdentifier{BedID, BedName}` / `RoomIdentifier{RoomID, RoomName}` 在 [card_types.go:5-15](owlBack/owl-common/card/card_types.go#L5) 已定义但 **未 inline 到 State struct**。

---

## 治理路线图

### Phase A — 打 FK 知识基础桩（这一步是地基，余下都依赖它）

**A.1** owl-common 加通用 Entity 类型（取代 BedInfo/RoomInfo/DeviceInfo 三套零碎结构；或共存渐进迁移）：

```go
// owlBack/owl-common/card/entity.go (新建)
type Entity struct {
    Prefix netip.Prefix  // CIDR；MaskLen 是 type discriminator
    Name   string
    Kind   int           // 同 masklen 内细分（room_type 0/1/2、device_type Sleepad/Radar 等）
}
```

**A.2** cardagg `CardMeta` 扩 children：

```go
// owlBack/wisefido-cardagg/internal/service/device_meta.go
type CardMeta struct {
    Self     Entity       // 本卡（masklen 告诉 /80 还是 /88）
    Children []Entity     // 子属：rooms+beds+devices 一锅，按 masklen 自然分类
    
    // 原有标量保留
    TenantPref      netip.Prefix
    UnitType        int
    UnitProperty    int
    HasBedSnap      bool
    HasBathroomSnap bool
    HasKitchenSnap  bool
    Devices         map[string]*DeviceMeta  // 兼容保留；与 Children 中 mask=128 项重叠
}
```

**A.3** cardagg `EntityOwnership` 反查表：

```go
// owlBack/wisefido-cardagg/internal/service/entity_ownership.go (新建)
type EntityOwnership struct {
    mu       sync.RWMutex
    ownerMap map[netip.Prefix]netip.Prefix  // child prefix → owning card prefix
}

func (o *EntityOwnership) OwnerOf(child netip.Prefix) (netip.Prefix, bool) { ... }
```

Loader: startup + `config:card:stream` invalidate 时 SQL union 三表 FK：
```sql
SELECT room_id::text AS child, card_id::text AS owner FROM rooms WHERE card_id IS NOT NULL
UNION ALL
SELECT bed_id::text,  card_id::text FROM beds  WHERE card_id IS NOT NULL
UNION ALL
SELECT device_addr::text, card_id::text FROM devices WHERE card_id IS NOT NULL
```

**A.4** sensor_state_projector 改 FK 反查替代 LPM：

```go
// owlBack/wisefido-cardagg/internal/consumer/sensor_state_projector.go
childPrefix, err := netip.ParsePrefix(msg.SubjectEntity)
if err != nil { return nil }
ownerCardPrefix, ok := p.ownership.OwnerOf(childPrefix)
if !ok {
    p.logger.Warn("entity has no owner card FK", ...)
    return nil  // 真孤立 → drop + warn，不再 LPM 猜
}
destCardID = ownerCardPrefix.String()
```

**A.5** 单测：
- `EntityOwnership` loader & query
- `sensor_state_projector` FK 反查替代 LPM (新 mock ownership)

### Phase B — state struct 加 entity_id

**B.1** owl-common 加 inline identifier：

```go
type BedState struct {
    BedIdentifier `json:",inline"`  // 加入 BedID + BedName
    BedStatus int
    BedStatusTs int64
    // ... 其余字段不变
}

type RoomState struct {
    RoomIdentifier `json:",inline"`  // 加入 RoomID + RoomName
    TotalPeople int
    // ... 其余字段不变
}
```

**B.2** sensor 翻译填 ID（[owlBack/wisefido-sensor/internal/zoneengine/translator.go](owlBack/wisefido-sensor/internal/zoneengine/translator.go)）：

```go
func TranslateBedState(e ZoneEvent) *card.BedState {
    out := &card.BedState{
        BedIdentifier: card.BedIdentifier{BedID: e.ZoneID},  // /96 prefix
        // ...
    }
}
```

**B.3** cardagg merge 保留 ID：sensor_state_projector 的 mergeBedStateSensorOwner 等函数确保 BedID/RoomID 透传。

**B.4** 单测：sensor translator 填 ID + cardagg merge 保留 ID

### Phase C — /80 unit_picker 拆 pin

**C.1** `unit_picker.buildUnitDisplay` 删除：
```go
merged.RoomState = w.RoomState  // 删
merged.BedState  = w.BedState   // 删
```

/80 unit display 只填 unit-level aggregates：
- TotalPeople = sum of children /88 cards' room_state.total_people
- AlarmSummary = unit own alarm count
- ActiveState = max(children active anchor) 或单独 logic

**C.2** `card_display_builder.BuildCardDisplay` 加 mask-aware 分支：
- masklen=80 → 走单元级聚合（不读 BedState/RoomState）
- masklen=88 → 读 own + 卡下 child entities union (按 A 步 FK 知道 children) 选 latest
- masklen=96 → 读 own bed_state

**C.3** FE Overview.vue Section2/Section3 单元卡 redesign：
- /80 卡：不显 bed_status / sleep_stage / bed_anchor_ms
- 改显：unit total people / unit alarm count summary

**C.4** 单测：单元卡 display 不再有 BedStatus 字段；不同 masklen 走对应分支

### Phase D — FE 适配

**D.1** FE 接 BedState 后用 `.bed_id` 找 BedInfo 渲具体床名（替代猜测）

**D.2** FE 卡 header Section1.down.left：用 state.bed_id / state.room_id 直查 source space 名，去掉 cardCoverageLabel 的复杂 fallback 链

---

## 实施顺序与硬约束

1. **必须先 Phase A**（FK 知识地基）—— 其他都依赖 ownership 真相
2. Phase B / Phase C 可并行（B 改 schema + sensor，C 改 cardagg display path）
3. Phase D 等 B + C 完成后做 FE

**硬约束**：
- 任何 schema 改动 (owl-common Entity / inline identifier) 必先 mock data + grep `BedState{` / `RoomState{` 全 callsite，确保 marshal 兼容
- 跨服务 contract 改动 ([[feedback_producer_first]])：先 sensor 端发新 schema → cardagg consume → FE 用
- Hash field 改动需 cleanup script 清 stale data
- 保留向后兼容期：BedState inline BedIdentifier 时旧 marshal 仍 work（字段默 omitempty）

---

## 切入点指令（下次会话粘贴接续）

> 接续 commit 03d2521 的下游治理。读 `doc/cardagg_fk_aggregation_prompt.md` 全文，Phase A 第一根桩开干：
> 1. `owl-common/card/entity.go` 新建 `Entity{Prefix, Name, Kind}` 通用类型
> 2. `wisefido-cardagg/internal/service/entity_ownership.go` 新建反查表 + loader (SQL union 三表 FK)
> 3. 改 `device_meta.go` `CardMeta` 加 `Children []Entity`
> 4. 改 `sensor_state_projector.go` 用 `EntityOwnership.OwnerOf()` 替代 `LookupCardByPrefix` LPM
> 5. 单测：ownership loader + projector FK 反查
>
> 完成后停下来等审查再做 Phase B/C。
> 
> 不允许跳过 Phase A 直接动 unit_picker pin 或 state schema —— 没 FK 真相所有判断都是空中楼阁。

---

## 已知坑

1. **rooms/beds 表当前数据可能 card_id 为空** —— 历史数据迁移问题，需先扫一遍 `SELECT count(*) FROM beds WHERE card_id IS NULL`，决定是补 FK 还是允许孤立
2. **/96 bed 不在 cards 表** —— 这是 v2.5 设计 (cards 全 /88)；ownership 是 bed→/88，不是 bed→/96 self
3. **wisefido-data 已用 CardStatic.Bed singleton** —— 若 Phase B inline identifier 改 schema，wisefido-data 那边 marshal 也要确认兼容
4. **FE Overview.vue cardCoverageLabel 现有 fallback 链** —— Phase D 简化时不能直接砍，要 graceful 迁移（先加新路径，旧的保留兜底，灰度后再删）
