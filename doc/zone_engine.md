# Zone Engine 业务逻辑设计（v2）

**状态**：2026-05-20 整理（落 Bug 2 dedup + Vacant trust 守卫后回补）
**代码入口**：[`wisefido-sensor/internal/zoneengine/`](../wisefido-sensor/internal/zoneengine/)
**关联 memory**：[[zoneengine_phase1_done]] · [[zoneengine_adapters_done]] · [[bed_presence_fusion]]

---

## 1. 顶层架构

Zone Engine 是 sensor 的统一空间占用状态机，把"床/房间/卫生间"三类空间的多源信号融合成
**三态 ZoneState**（Vacant/Occupied/Leaving），下游派生消费。

```
┌─────────── Inputs (Redis streams) ───────────────────────────────┐
│ iot:event:stream                                                  │
│ ────────────────                                                  │
│ radar:    EnterRoom / ExitRoom / NumberPeople / InBed / LeftBed   │
│ sleepace: inBedStatus / bedStatus change（**非**厂家 alarmInBed） │
│                                                                   │
│ iot:monitor:stream   ──► realtime raw track + bed_status         │
│                          (sensor zoneengine **不消费**)           │
│ iot:alarm:stream     ──► sleepace native alarmInBed / LeftBed    │
│                          (sensor zoneengine v2 **不消费**         │
│                           2026-05-20 改路径)                       │
└────────────┬─────────────────────────────────────────────────────┘
             │ (两 adapter 都订阅 iot:event:stream)
             ▼                                  旁路记账（Bug 2 dedup）：
┌────────────┐ ┌──────────────┐               ┌────────────────────┐
│ adapter_   │ │ adapter_     │ ───SetRadar──▶│ BedPresenceFusion  │
│ radar      │ │ sleepace     │ ◀──SetSleepad-│ (per-/96 X/Y bool) │
└─────┬──────┘ └──────┬───────┘               └────────────────────┘
      │ Apply()       │ Apply()    ───SetZ──▶ ┌────────────────────┐
      │               │                       │ RadarRoomCountCache│
      │               │                       │ (per-/88 raw Z)    │
      ▼               ▼                       └────────────────────┘
   ┌──────────────────────┐
   │   Engine             │  states: map[StateKey]→zoneInstance
   │   ─────────          │  StateKey = (CardID, ZoneType, ZoneID)
   │   - Scorer           │  zoneInstance = {Scorer, StateMachine, ZoneState}
   │   - StateMachine     │
   │   - subset_invariant │
   │   - self_contradict  │
   └──────────┬───────────┘
              │  emitEvent(ZoneEvent)
              ▼
   ┌──────────────────────┐
   │ Listeners            │
   │ - StreamPublisher ──┼──▶ sensor:derived:stream (room.state / bed.state / target.state)
   │ - TargetAggregator   │
   │ - Zonealarm.Supervisor ─▶ alarm 派生 (Stay/LeftBed/NightAbsence/BedNightAbsence)
   │ - SleepStageClearAdp │
   └──────────────────────┘
```

**4 个 listener 全部订阅同一 ZoneEvent 流**，各自派生自己的下游产物。

---

## 2. 三态状态机（核心）

[`types.go`](../wisefido-sensor/internal/zoneengine/types.go#L42-L63) `ZoneStatus`：

```
                       enter ≥ +threshold
            ┌─────────────────────────────────────┐
            │                                     ▼
       ┌────┴─────┐                       ┌────────────┐
       │  Vacant  │                       │  Occupied  │
       │  ╳ 无人  │                       │ ✓ 已确认在 │
       └────▲─────┘                       └─────┬──────┘
            │                                   │
            │ timer expires                     │ leave ≤ -threshold
            │ (Leaving N s 无回弹)              │ (老人坐起/下床信号)
            │                                   │
       ┌────┴──────────────────┐                │
       │     Leaving           │ ◀──────────────┘
       │   ⌛ 软离开（计时中） │
       │   IsPresent = true    │
       └───────────┬───────────┘
                   │
                   │ enter ≥ +threshold (老人坐回)
                   │
                   └────────▶  Occupied  (TransitionReturned)
```

**Leaving 的意义**（老人友好）：老人坐起 / 试探下床 / 短时压感消失，几秒内回弹的话不该
误报 LeftBed。Leaving 是 Occupied 与 Vacant 之间的**软离开**中间态，timer 窗内可回滚。

**subset_invariant 视角**：`Leaving.IsPresent() = true`（仍算占用），跟 Occupied 同等处理。

---

## 3. Score 模型（Scorer）

[`scorer.go`](../wisefido-sensor/internal/zoneengine/scorer.go)（详 [zone_rules_eval.md](../wisefido-sensor-v1/docs/zone_rules_eval.md)）

```
pos = max(enter.strength_decayed, sustain.strength_decayed)
       + recent_enter_bonus (只 sustain 活时加 +15)
neg = leave.strength_decayed
score = pos - neg     ∈ [-100, +100+bonus]
```

**关键设计**：
- **多源同方向取 max 不 sum** — 防 sleepace + radar 同向叠加假强
- **反向 latch** — 高优先级 evidence（sleepace 90 / radar 80）在 latch 窗（10s）内拒绝反向信号
- **decay 线性** — 120s 走完一档 strength（DecayWindowMs）
- **enter / leave 互斥清零** — 一方触发清零另一方累积
- **sustain 独立通道** — HR+RR > 0 = "在床"持续证据，与 enter 同档（80）

StateMachine.Evaluate(score, now) 把 score 跟 `entry_threshold / exit_threshold` 比，
决定要不要翻 transition。

---

## 4. 三种关键 transition path

### 4.1 enter/leave/sustain（走 Scorer + StateMachine）

```
SignalEvidence{Source, Kind, Delta} ──▶ Scorer.Apply ──▶ newScore
                                                          │
                                            StateMachine.Evaluate(newScore)
                                                          │
                                                          ├── Flipped: applyTransitionToState
                                                          │              + emit ZoneEvent
                                                          └── 不翻：score 更新返回
```

### 4.2 count_change（直写，**不经过 Scorer/StateMachine**）

[`engine.go:163-181`](../wisefido-sensor/internal/zoneengine/engine.go#L163)：

```
SignalEvidence{Kind="count_change", Count=N}
        │
        ▼
z.state.Count = N
z.state.UpdatedAt = now
        │
        ▼
emit ZoneEvent{Transition="count_change", PrevState, NewState}
```

**注意**：count_change 不影响 Status（Occupied/Vacant/Leaving），只更 Count 字段。
NumberPeople=2→1 同房间内不算"新进/新离"。

### 4.3 Vacant transition 的 Count 强 set

[`engine.go applyTransitionToState`](../wisefido-sensor/internal/zoneengine/engine.go) 强语义：

```
case TransitionVacant:
    s.LastExitTs   = nowMs
    s.LeavingSince = 0
    s.Count        = 0    ◀── 显式归零，无视外部输入
```

Vacant 是 ground truth — engine 内部权威。**任何下游消费者都不能用 stale cache 推翻它**
（参 §6 Vacant trust 守卫）。

---

## 5. 子集不变量 + 自相矛盾反馈

### 5.1 subset_invariant — bed ⊆ room

[`engine.go maybeReconcileSubset`](../wisefido-sensor/internal/zoneengine/engine.go#L527)：

```
              bed transitions
                    │
                    ▼
        bed.IsPresent = true?
          (Occupied 或 Leaving)
                    │ yes
                    ▼
         room.IsPresent = false?  ── no ──▶ 一致，pass
                    │ yes（不一致）
                    ▼
         强抬 room → Occupied
         room.Count = max(prev, 1)
         emit ZoneEvent{ZoneType: Room, Transition: occupied}
```

**反向（周期 Tick 巡检）** [`engine.go repairSubsetInvariant`](../wisefido-sensor/internal/zoneengine/engine.go#L384)：
- `bed.IsPresent + room.Vacant` 持续不一致：
  - bed 信号 staleness（无新 evidence > N 秒）→ 强降 bed = Vacant
  - bed 新鲜 → 抬 room = Occupied

### 5.2 self_contradiction — 翻转后短窗内自打脸

```
zone 翻 Occupied
        │
        ├── 登记 pendingValidation
        │
        ▼
  window_sec 内（Tick 检查）
        │
        ├── score 仍 ≥ require_sustain_at → 通过
        │
        └── score 跌破 → ROLLBACK + emit FeedbackEvent
                          (Affected 信号源 Penalty 临时降权)
```

**仅 Occupied 方向登记**（review fix P0-1）：Vacant 是默认安全态，证据自然老化不视为矛盾。

---

## 6. Bug 2 dedup pathway（2026-05-20 新增）

旁路解决"radar NumberPeople 漏数床上人"的偏差，**不进 engine 状态决策**，仅在 publish 时
修正发到下游的 TotalPeople。

```
adapter_sleepace InBed/LeftBed                adapter_radar
        │                                     applyBed(InBed/LeftBed)
        │  presence.SetSleepad(bedCIDR, X)            │
        ▼                                             │ presence.SetRadar(bedCIDR, Y)
┌────────────────────┐                                ▼
│ BedPresenceFusion  │ ◀────────────────────────  per-/96 bed X(sleepad) + Y(radar)
│ (per-/96 X/Y bool) │
└─────────┬──────────┘
          │ ExtraPeopleInRoom(roomCIDR)
          │ = Σ(bed where X && !Y)
          │
adapter_radar                                          │
applyCount(NumberPeople)                              │
        │                                              │
        │  roomCount.SetZ(roomCIDR, count)            │
        ▼                                              │
┌────────────────────┐                                │
│ RadarRoomCountCache│  GetZ(roomCIDR) ──┐            │
│ (per-/88 raw Z)    │                   ▼            ▼
└────────────────────┘             ┌────────────────────────┐
                                   │ StreamPublisher        │
                                   │ applyRoomDedupInPlace: │
                                   │                        │
                                   │ if rs.TotalPeople==0:  │
                                   │   ┌──────────────────┐ │
                                   │   │ Vacant trust 守卫│ │
                                   │   │ 直接 return ◀────┼─┼─◀ 关键 Bug 修复
                                   │   │ (engine 权威)    │ │   (f61e8o 6min stale)
                                   │   └──────────────────┘ │
                                   │ else:                  │
                                   │   z = cache.GetZ()     │
                                   │   extras = fusion.Extra│
                                   │   rs.TotalPeople = z + extras
                                   └────────────────────────┘
```

**dedup 公式**：
- `Z` = radar `NumberPeople`（raw，来自 cache）
- `X` = sleepad InBed bool (per /96)
- `Y` = radar bed-area InBed bool (per /96)
- `published TotalPeople = Z + Σ(X for bed where X && !Y)`

**Vacant 守卫的必要性**：radar EnterRoom/ExitRoom alarm 路径不带 NumberPeople=0 心跳，
cache.Z 留 1 stale；engine 处理 Vacant 后 Count=0，但 dedup 用 stale cache 反向覆写。
守卫确保 Vacant ground truth 不被推翻。

---

## 7. Listener 输出契约

### 7.1 StreamPublisher → sensor:derived:stream

| Category | 触发 | Subject | 内容 |
|----------|------|---------|------|
| `room.state` | OnZoneEvent(Room/Bathroom transition) | /88 room CIDR | `card.RoomState{TotalPeople, RoomType, LastEnter/Exit, StaySec, RiskLevel}` |
| `bed.state` | OnZoneEvent(Bed transition) | /96 bed CIDR | `card.BedState{BedStatus 0/1, BedEvent, StartTime, DurationSec, TrackNumber, BedConfidence}` |
| `bed.sleepstage` | SleepStageConsumer ladder | /96 bed CIDR | `card.BedState{SleepStage, SleepConfidence}` 仅这两字段 |
| `target.state` | 60s ticker pull aggregator | /88 或 /96 | `card.TargetState{LastActiveTs, StandingMin, WeakBio}` |

OnZoneEvent 是**event-driven 触发**（zone transition 瞬间）；target.state 是**60s 周期 pull**
（聚合器纯 state holder，dirty 检查跳过无变化）。

### 7.2 Zonealarm.Supervisor → iot:alarm:stream

派生 4 条 zone-level alarm（不重做 engine 工作）：
- `Stay` / `LeftBed` / `NightAbsence` / `BedNightAbsence`

订阅 ZoneEvent + alarm-firer 按 yaml 规则配 timer/threshold。

---

## 8. 边界场景 cheat-sheet（改 zone engine 前必看）

| 场景 | 期望行为 | 触发路径 |
|------|---------|----------|
| 老人坐起→压感消失 5s→坐回 | 不报 LeftBed（Leaving timer 内回弹）| Leaving timer 内 enter signal → TransitionReturned |
| 老人下床 30min 没回 | 报 LeftBed（Vacant transition）| Leaving timer 超时 → TransitionVacant |
| sleepad 离线，radar 仍报 InBed | bed 仍 Occupied（radar 80 evidence 足）| adapter_radar InBed → Score → StateMachine |
| 床有人，radar 报 NumberPeople=0 | room subset_invariant lift 到 Occupied Count=1 | bed.IsPresent + room.Vacant → maybeReconcileSubset |
| 浴室人走 → radar ExitRoom + NumberPeople=0 | TransitionVacant，Count=0，**dedup 守卫确保 publish=0** | applyCount→engine→OnZoneEvent→applyRoomDedupInPlace skip |
| ghost track 让 radar Count=1 持续 | engine Count=1，dedup +extras（如有 sleepad InBed 矛盾）| 待 Layer 1 ghost adjudicator 接入（v2 决定 1） |

---

## 9. 已知改前必读清单（feedback rule）

[[feedback_read_design_before_modify]]：改 zone engine 任何代码前先读：
1. `types.go` 顶部（三态机 ASCII + Transition 常量定义）
2. `engine.go applyTransitionToState`（字段更新规则，特别是 Vacant Count=0 强 set）
3. `translator.go TranslateRoomState/TranslateBedState`（发到 stream 前的字段填充）
4. 本文档 §8 边界场景 cheat-sheet

改完枚举 §8 每行场景人工走一遍验证。Bug 2 dedup 当时漏看 Vacant Count=0 强语义，导致
bathroom 卡 6min 才被发现。
