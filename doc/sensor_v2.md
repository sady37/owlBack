# wisefido-sensor v2 重设计 — 讨论稿

> 状态：**讨论中**，不是最终设计。
> 目的：在 device-gateway（qinglan/sleepace）与 cardagg v2 整改之后，重新审视 wisefido-sensor 的核心模型。
> 关联权威文档：[cardagg_sensor_split.md](cardagg_sensor_split.md) / [platform_agent_addressing.md](platform_agent_addressing.md) / [AI_fall_detect.md](AI_fall_detect.md)

---

## 0. 范围与约束

适用 owlBack/CLAUDE.md 全部 6 条编码规则 + #2 流分配纪律。本文档不改动以下既有契约：

- 数据链分流（radar → event_stream，sleepad/device-health → alarm_stream，sensor 派生回流 alarm_stream，producer=`fd00:0:fff1::1`）
- sensor 永不写 `alarm_events` 表（PR1 A11 红线）
- cardagg 永不当 producer / 永不订阅 event stream
- sleepace native alarm 可信不二次验

**v2 整改的可动手范围**：roomengine 内部模型 — track / verdict / cell / zone state / fall 规则 这五样东西彼此的契约。

---

## 1. 用户已对齐的根本性决定（2026-05-17 ~ 2026-05-18）

### 决定 16：room.kind binary 分类 — bathroom 显式标注，其它默认 bedroom（2026-05-18 vue 编辑器简化）

**v2 不需要完整 room.kind 枚举**（bedroom/living/kitchen/...）。Elder care suite 实际拓扑是 studio 公寓（起居 + 卫浴），只需 binary 分类：

```go
type Room struct {
    ID       string
    Kind     string  // "bathroom" 显式标注 / "" (默认) = bedroom
    ...
}
```

**约定**：
- vue layout 编辑器：room 属性面板加 checkbox / radio "Is bathroom"
- 显式勾选 → `room.kind = "bathroom"` → 走 BathroomGhostAdjudicator + BathroomFallRules
- 未勾选（默认）→ 视作 bedroom → 走 GeneralGhostAdjudicator + GeneralFallRules
- 房间内有 Bed cell（cell.AreaType=AreaBed）是 bedroom 的**充分非必要**特征 — 不强制要求，仅作为编辑器提示"建议作为 bedroom"
- v3 扩展为完整 kind 枚举时（living / kitchen / hallway），bathroom 标注语义不变

**PG 存储（2026-05-17 PR-4 勘误）**：**复用既有 `rooms.room_type` VARCHAR(50)** — 该列值域 (`bedroom`/`bathroom`/`livingroom`/`kitchen`/...) 与本决定语义兼容。**不新增 `rooms.kind` 列**（违反 CLAUDE.md #1.3 单源真相）。sensor 内部命名仍用 `RoomKind`（与本决定 doc 措辞一致），bootstrap SQL `r.room_type AS room_kind` 做 column → field 适配。详 §10.4。

**为什么不区分 living / kitchen**：
- v2 suite 拓扑是 studio 公寓（起居 + 卫浴），不存在独立 living / kitchen
- 即使有，elder care 场景下 fall 规则与 bedroom 同（lost / bedside / still）
- v3 unit-layout 演进时再细分（如 kitchen 加 still_fall 特殊处理 — 参考 [kitchen_lostfall_known_limitation](../../.claude/projects/-home-wisefido-owl/memory/kitchen_lostfall_known_limitation.md)）

### 决定 15：Layout 显式 bathroom_enter 标注（2026-05-18 简化拓扑）

**比决定 12 硬编码 suite 拓扑更通用**。Layout 编辑时显式标注 bathroom 入口位置，复用现有 `cell.AreaType=AreaEnter` 体系，加一个 `EnterTarget` 字符串属性：

```go
type Cell struct {
    ...
    Belief        []AreaBelief
    EnterTarget   string  // "" / "bathroom_<roomID>" / "outside" / ...
                          // 仅当 Belief[0].Type == AreaEnter 时有意义
}
```

**约定**：

| 约定 | 内容 |
|---|---|
| **单入口约束** | 每个 bathroom 在 layout 上**恰好 1 个连续 cells 区域**标注 `EnterTarget="bathroom_<self_id>"` |
| **Storage / 储衣室合并** | 储衣室仅通过 bathroom 出入 → 人工划入 bathroom zone，cells 标 AreaUnknown；不新增 AreaCloset 类型 |
| **Jack-and-Jill bathroom** | v2 不支持；部署时拆为两个 zone（v3 多 bathroom 户型再做）|
| **多 bathroom unit** | v2 支持（每个 bathroom 各自一个 EnterTarget）|
| **中部 bathroom** | layout 标注的 EnterTarget cell 不要求"在 bedroom 一侧"，可在任何外部 room（客厅 / 走廊）一侧 |

**Layout 编辑器校验**（不入 runtime）：
- 0 片 EnterTarget → WARNING + runtime 走 fallback（决定 12 suite 假设 + 自动学习推荐）
- 多片 EnterTarget → ERROR（必须合并或选 1 个）
- EnterTarget cell 必须与 bathroom zone 边界相邻 → ERROR if 不相邻

**Fallback 链**（layout 未标注 / 多片错误）：

```
1. 自动学习：累积"track 反复穿越 bathroom 边界"位置 → 推荐 EnterTarget 候选
   → 写 ai.log，前端编辑器拉建议给运维审核（不自动写 layout）
2. Runtime 兜底：决定 12 suite 拓扑假设（bathroom 入口仅通向 bedroom）
3. 共生律自校正（决定 14 规则 3）仍工作
4. Fall 触发仍可信（决定 14 ghost 不抑制 fall）
```

**架构价值（vs 决定 12 硬编码）**：

| 维度 | 决定 12 硬编码 | **决定 15 layout 标注** |
|---|---|---|
| 适用拓扑 | 仅 1 bedroom + 1 bathroom suite | 任意 |
| v3 unit-layout 演进 | 推翻重做 | **直接升 zone 边界，业务逻辑不变** |
| 新增概念 | suite 拓扑假设 | 仅 cell.EnterTarget 字符串属性 |
| Layout 编辑器 UI 改动 | 无 | 加一个下拉框 |
| 中部 bathroom / 多 bathroom | 不支持 | **支持** |

### 决定 14：Ghost-Real 共生律 — Ghost 必有真人作为反射源（2026-05-18 物理事实）

**用户拍板的物理论证**：

```
Ghost（镜面 / 金属反射 / 玻璃门多径）= 真人作为反射源产生的虚像
  → 没有真人 → 没有反射源 → 没有 ghost
  → 真人离开 bathroom → ghost 秒级消失（典型 < 10s）
  → "无主 ghost 单独存活" 物理上不可能
```

**可能状态**：

| 真人数 | ghost 数 | 可能? |
|---|---|---|
| 0 | 0 | ✓ |
| 1 | 0+ | ✓ |
| 2 | 0+ | ✓ |
| 0 | ≥1 | ✗ **物理不可能** |

**推论**：
- Ghost 残影不需特殊处理（出门事件 + 秒消足够，晚 10s 无影响）
- **Ghost 是真人存在的间接证据**，不是应该被压制的假信号
- Bathroom 内有 ghost 但 firmware 丢真人 track = 真人物理上仍在 bathroom（蒸汽 / 静止多普勒消失）
- "Ghost latch" 持久化机制不必要（ghost 自然随真人离开消失）
- Fall 触发不因 ghost verdict 而抑制（ghost 存在 = 真人在场）

**对设计的影响**：
- §4.A.3 规则 3 重写为"共生律自校正"
- §4.A.4 Ghost 身份持久化删除
- §6.A.3 BathroomLostFall 加最强信号"track == 0 立即 fire"
- §6.A 所有 fall 规则不因 ghost 抑制触发
- **bathroom 内镜像学习保留为 P0**（不是抑制 ghost，而是 disambiguate 哪个 track 是真人，用于 fall 归因）

### 决定 11/12/13：v2 范围 = Suite-specific（不是 unit-level）（2026-05-18 工程现实）

**决定 11 — v2 vs v3 边界**：
- **v2**：只做 **suite-specific**（单个 elder care suite 拓扑），不需要 unit layout / logicID 系统
- **v3**：才扩展到 unit 多房间（含客厅 / 餐厅 / 多卧室）+ 完整 unit layout + logicID

**决定 12 — Suite 拓扑（决定 15 已替代为 fallback）**：

⚠️ **已降级为 fallback**：决定 15（layout EnterTarget 标注）是 v2 主路径，决定 12 仅在 layout 未标注或多片错误时作为 runtime fallback 兜底。

```
Suite v2 默认拓扑假设（仅 fallback）：
  - 1 Bedroom（含 Bed + sleepad）+ 1 Bathroom
  - 唯一连接：bathroom 入口 → bedroom（不直通走廊）
  - 所有进 bathroom 的人必经 bedroom

部署前提（部署时审查）：
  - bathroom 入口 cell 必在 bedroom radar 的 RadarVisibleRegion 内
  - bedroom 必有 sleepad
  - bathroom 不必有 sleepad（v2 设计如此）
```

**决定 13 — 人类行为时序假设**：
- 护工 / 家属 / 访客进入 suite 后**不会立即把 elder 拉进 bathroom**
- 中间有 **≥ 2 分钟** 在 bedroom 过渡期（实际通常 5-10 分钟以上）
- 这个过渡期足够 sensor 在 bedroom 内确认"现在有几个人"
- bathroom 占用人数 = bedroom 过渡期 census + 入口流量派生
- **加固理由（Q-prereq-6 用户回复）**：需要护工帮助洗澡的老人都是**行动不便**的 → 转移过程慢 → ≥ 2 min 是物理下限，实际更长

**为什么这套约束成立**：
- elder care 业务场景就是 suite，不是公寓 / 别墅
- elder 不会一个人冲进 bathroom（如果是 ambulatory 老人，bathroom 内静止 = 真危险，简单判定就够）
- 协助型 visitor（护工/家属）从外部进 bathroom 必经 bedroom
- 这些约束让 §4.A 的"unit-level ActiveResidentSet"**降级为更简单的 SuiteBedroomCensus**

### 决定 7：不追求完美 — Unit-level 出入口计数法（2026-05-18 简化）

**工程哲学**：在 bathroom 内部做完美 ghost 判定是**反工程的方向**。
**正确方向**：在 **bedroom（决定 12 主判定房间）做 SuitePerson 识别**，用 **bathroom 入口流量转移身份** + **共生律物理约束** + **镜像学习 disambiguation**。

```
┌── Bedroom（"主判定房间"）— Suite v2 决定 12 ──────────────┐
│  radar 视野好 + sleepad 双源 + 空间大                    │
│  → SuiteBedroomCensus 识别 SuitePerson(s):              │
│    - resident   sleepad 双源锚定（强）                  │
│    - visitor    radar 2min anchor（决定 13 时序假设）   │
│                                                          │
│  Bedroom Fall 规则（消费 SuitePerson，决定 16 default）: │
│    silent_fall   sleepad LeftBed + BedState 矛盾        │
│    bedside_fall  BedState→Vacant + 床边静止 15min       │
│    lost_fall     SuitePerson 在 bedroom + 静默 cell-typed│
│    （still_fall v2 不实现，v3 扩展 living/kitchen 再加） │
└──────────────────┬───────────────────────────────────┘
                   │ SuitePerson 流向 bathroom 入口（决定 12 唯一通道）
                   ▼
┌── Bathroom 入口 cell（必在 RadarVisibleRegion 内）──────┐
│  zoneengine BathroomGate 子模块:                        │
│    bedroom→bathroom: BathroomCount +1                  │
│                      person.AnchorRoomKind = "bathroom" │
│    bathroom→bedroom: BathroomCount -1                  │
│                      AnchorRoomKind = "bedroom"        │
│  → BathroomCount 由入口累积差驱动（不靠内部 track 数）  │
└──────────────────┬───────────────────────────────────┘
                   │
                   ▼
┌── Bathroom 内部 — 决定 14 共生律（ghost 必有真人反射）──┐
│  规则 1: 新出生 cell 远离入口（≥30cm）→ Ghost          │
│  规则 2: 内部 track 数 > BathroomCount → 多余 = Ghost  │
│         + §4.A.5 镜像学习 disambiguate 哪个是真人       │
│  规则 3: 共生律自校正（track ≥1 → 期望 Count ≥1）       │
│         不一致 → 修正 BathroomCount（漏记 +1 兜底）     │
│                                                          │
│  Bathroom Fall（决定 14: ghost 不抑制 fall）:           │
│    BathroomStillFall                                    │
│      cell ∈ {Toilet,Shower} + Stand 静止 ≥10/12min     │
│    BathroomBedsideFall                                  │
│      Occupied 90s grace + 任意位置静止 ≥8min           │
│    BathroomLostFall（两档）                             │
│      最强: BathroomCount≥1 + track 数==0 ≥30s 立即 fire│
│      次强: BathroomCount≥1 + 真人 track 静默 7min       │
└────────────────────────────────────────────────────────┘
```

**前提**：bathroom 入口必在 RadarVisibleRegion 内（决定 2）。盲区入口 → fallback 到规则 1（出生地仍可用） + 规则 3（共生律自校正）；规则 2 失效。

**简化掉的内容**（与早期 v2 相比）：
- ❌ ~~Mirror reflection_line 在线学习降级~~ → **回升 P0 保留**（决定 14 修订，角色变更为 disambiguation）
- ❌ BATHROOM_STRICT_SINGLE / FILTER_STRICT / NORMAL 状态机（被 SuiteBedroomCensus 替代）
- ❌ Ghost-latched 24h 复杂规则（决定 14 共生律后不需要——ghost 跟随真人自然消失）
- ❌ "unit 其它房间活动 → bathroom 静止 = ghost" 跨房间一致性兜底（决定 14 共生律推翻——bathroom 静止 track = 真人在场 = fall 候选）

**保留的内容**：
- ✅ Bathroom 与 bedroom 分支（仍需要，规则不同）
- ✅ 4 个 bedroom fall + 3 个 bathroom fall = 7 类 fall 规则（详 §6）
- ✅ 镜像学习 P0（用于 fall 归因 disambiguation，详 §4.A.5）

代价对比：~860 行 → **~750 行**（镜像学习回升加约 ~400 行），3-4 周 → **3 周**。

### 决定 0：Bathroom 与其它房间走**独立分支**（2026-05-18）

bathroom 物理特性与其它房间根本不同，**强制走独立 ghost adjudicator + 独立 fall 规则**：

| 特征 | Bathroom | 其它房间（bedroom / living / kitchen） |
|---|---|---|
| 主 ghost 源 | **镜面 + Shower 金属把手 + 玻璃浴室金属支撑**（强反射体密集） | 多径散射 |
| 跨源信号 | **无 sleepad**（信号 d 失效） | bedroom 有 sleepad |
| 蒸汽 / 多普勒失效 | **是**（淋浴 60GHz 严重吸收 + 静止人多普勒消失） | 否 |
| Ghost 与真人 traverse | **同步**（镜像跟真人走） | 不同步（真人 traverse 高，多径 ghost 低） |
| Ghost 与真人 anchor | **同步**（同时出生同时存活） | 不同步 |
| 死亡风险 | **极高**（关门 + 长静止 + 滑倒 + 体位性低血压） | 中 |
| 单人假设 | **极强**（2 人共用极罕见） | 中 |

**根本含义**：在 bathroom，统计累积式 ghost adjudication（信号 a+e）会**双双失效**，必须改用**几何专用检测 + 跨房间先验**。bathroom 分支不复用其它房间的 ghost adjudicator 代码。

### 决定 1：三层流水线 — ghost → zone → fall（不是各 fall 各查 ghost）

原因：zoneengine 对 bed/room/bathroom 的状态判断**强于** 4 fall 各自的私有推断。让 zone 成为 fall 的**前置依赖**，让 ghost 成为 zone 的**前置过滤**。

```
Layer 1: Ghost adjudication
  输入：所有 track frames + EnterRoom/ExitRoom/InBed/LeftBed event + sleepad event
  内部：birth filter + Kalman + cell history + GhostPenalty 累积
  输出：每个 track 的 Verdict (Real / Ghost / Pending / Anchored)
                      │
                      ▼
Layer 2: Zone state (bed / room / bathroom)
  输入：**仅 Real / Anchored / Pending 的 track** + sleepad event + radar event
        + RadarVisibleRegion（新，§5）
  内部：ZoneEngine 三态状态机 (Vacant / Occupied / Leaving) + subset_invariant
  输出：BedState / RoomState / BathRoomState（zone-centric，作为下游唯一真相）
                      │
                      ▼
Layer 3: Fall detection (4 类推断 fall)
  输入：Zone state 翻转事件 + Real track 的当前位置 / 姿态 / 静止时长 + sleepad event
  内部：4 fall 规则全部**消费 ZoneState**，不再维护私有计数器
  输出：emit alarm (producer=fd00:0:fff1::1)

  ┌────── 反馈通道（Layer 2 → Layer 1）──────┐
  │  BedState=Vacant + sleepad 确认无人 + radar 突然 InBed                  │
  │     → 床区 track 高 ghost 嫌疑                                          │
  │  RoomState 长期 Vacant + track 突然出生远离 Enter                       │
  │     → ghost penalty 因子加成                                            │
  │  形态：FeedbackEvent → Ghost adjudicator                                │
  └─────────────────────────────────────────┘
```

**含义**：
- 之前讨论的"4 fall 各自跑 verifier 评分（Q-A3/Q-D1）"被这一架构替换 — 不再 fall 内查 ghost，而是**ghost 在最上游就过滤掉**，下游不可能看到 ghost track。
- 之前讨论的"BedFusion 子模块（Q-B1）"被吸收为 zoneengine 的标准职责 — zoneengine 本来就是融合层。
- Layer 间单向数据流（避免循环依赖）；反馈走专用 `FeedbackEvent`，不走回数据流。

### 决定 2：从根本上解决 Room 空间 vs Radar 空间不对齐

物理事实：
- **Room 几何** = layout 中 wall 围成的多边形（完整房间）
- **Radar 可观测区** = radar FOV cone ∩ (Room − wall 阻挡)
- 二者**几乎从不重合** — radar FOV 受限于安装位置，wall 还会进一步切掉一部分

由此产生的连锁问题（实测案例）：
| 现象 | 当前误处理 |
|---|---|
| 房间有 2 个门，radar 只看到 1 个 → 人从另一门进入，无 EnterRoom，track 出生在房间中央 | 当前 birth_score 判 ghost（误） |
| L 形房间，人走到拐角 radar 死角 → track 失锁但人未离开房间 | 当前 lost_fall 触发（误） |
| 人从 radar 看不见的门出去 → track 失锁，无 ExitRoom 事件 | 当前 lost_fall pending 入池等超时（误，pending 拖几分钟才取消） |
| 真 ghost（多径反射）从 wall 内侧"出生"远离 Enter | 当前 ghost penalty 扣分（正确，但与上述真人混淆） |

**v2 要解决**：layout 模型需要显式区分 "Room 几何" 与 "Radar 可观测区"，并标注 radar **盲区出入口**位置。详见 §5。

---

## 2. 当前架构缺陷（重新审视）

### 2.1 两个子系统几乎不互通

```
              iot:event:stream            iot:alarm:stream
              (radar)                      (sleepad)
                │                             │
                ▼                             ▼
   ┌────────────────┐         ┌─────────────────────────────────┐
   │  zoneengine    │         │         roomengine               │
   │  ZoneState     │ ←(无)→  │  TrackState + Cell.AreaType      │
   │  Vacant/       │         │  4 个 fall（私有计数器）          │
   │  Occupied/     │         │  - bedPersonCount               │
   │  Leaving       │         │  - lastLeftBedAt                │
   │                │         │  - lastNumberPeopleZeroMs       │
   │                │         │  - bathroomRealCount            │
   └────────────────┘         └─────────────────────────────────┘
```

zoneengine 不知道 track verdict；roomengine 不消费 ZoneState。
4 fall 各自有 ghost gate 的不一致（lost ✅ / silent 隐式 ✅ / bedside ❌ / still ❌）— **这一不一致在新架构下自动消失**：ghost 已在 Layer 1 过滤，Layer 3 不可能看到 ghost track。

### 2.2 当前 4 fall × ghost × zone state 耦合矩阵（仅作"为什么要重设"的证据）

| Fall | 当前 ghost gate | 消费 ZoneState? | 备注 |
|---|---|---|---|
| silent_fall | ✅ 隐式（BedSession 双源门控） | ❌ 自己重做一遍 sleepad+radar InBed 一致性 | **重做** zoneengine 的事，违反 #1.3 |
| lost_fall | ✅ 显式（[track_manager.go:1002](../wisefido-sensor/internal/roomengine/track_manager.go#L1002)） | ❌ 自己消费 number_people 原始事件 | 绕过 ZoneState |
| bedside_fall | ❌ **无** | ❌ 自己用 lastLeftBedAt | 看不到房间人数（护工陪同会误报）|
| still_fall | ❌ **无** | ❌ 用 roomName 字符串匹配 | 与 zonealarm.Stay 10min 双计时器 |

---

## 3. 目标架构 — 三层流水线的契约

### 3.1 Layer 1 → Layer 2 契约（Track Verdict 投影）

zoneengine 不再直接读 radar track frame；改读 **Layer 1 输出的 TrackStatus**（带 Verdict）：

```go
type TrackStatus struct {
    TrackID        int
    DeviceID       string
    RoomID         string

    Verdict        TrackVerdict   // Pending / Real / Ghost / Anchored

    // 当前位置 / 姿态 / 空间归属（grid 反查）
    X, Y, Z        int
    Pose           int
    CellAreaType   AreaType
    StillSec       int

    // 当前 zone 归属（grid → zoneengine zone 反查）
    InBedZoneID       string  // "" = 不在任何 bed
    InRoomZoneID      string  // 一般 == RoomID
    InBathroomZoneID  string

    UpdatedAt      int64
}
```

发流：`sensor:track:status:stream`（瞬态，不入库，符合 #2.1）
消费者：zoneengine（Layer 2）/ fall rules（Layer 3 间接通过 zoneengine）/ UI canvas（前端 dev tool）

**Ghost track 的处理**：仍写 TrackStatus 但 Verdict=Ghost；下游消费方按 Verdict 自行决定是否计入。
**默认契约**：zoneengine occupancy 计数**只接受 Verdict ∈ {Real, Anchored, Pending}**。Ghost 不计 zone 占用。

### 3.2 Layer 2 → Layer 3 契约（Zone Event）

fall 规则只消费 ZoneEvent（翻转沿）和 ZoneState snapshot，不再读 raw track / sleepad event：

| Fall | 新触发条件（Layer 3 消费） |
|---|---|
| **silent_fall** | BedState 仍 Occupied/Leaving + sleepad 直发 LeftBed 矛盾 → 等矛盾窗口 |
| **lost_fall** | RoomState=Occupied + 所有 Real track 失锁 + 无 ExitRoom + 无 RoomState→Vacant 翻转 → 等 cell-area-typed 窗口 |
| **bedside_fall** | BedState→Vacant 后 + Real track 在床边 ≤100cm 静止 ≥15min（夜间）+ RoomState.Count == 1 |
| **still_fall** | BathRoomState=Occupied + Real track 在 cell∈{Toilet,Shower} stand 静止 ≥ 15/18min |

私有计数器全删（bedPersonCount / lastLeftBedAt / lastNumberPeopleZeroMs / bathroomRealCount）；统一改读 ZoneState。

### 3.3 Layer 2 → Layer 1 反馈契约（FeedbackEvent）

zoneengine 检测到下列异常时给 Ghost adjudicator 发 FeedbackEvent：

| 触发 | Feedback 形态 |
|---|---|
| BedState=Vacant + sleepad 确认无人 ≥ 30s + radar InBed | `reason=bed_ghost_suspect`, affected=[radar], penalty=+30 |
| RoomState 长 Vacant ≥ 5min + track 出生远离 Enter | `reason=stale_room_ghost_birth`, penalty=+20 |
| BedState=Occupied (sleepad 锚定) + radar 第二 track 床区出生 | `reason=bed_multipath_ghost`, affected=[radar], penalty=+40 |

Ghost adjudicator 收到 FeedbackEvent → 给目标 track 加 GhostPenalty，加速翻 Ghost。
**避免循环**：FeedbackEvent 只走专用通道（不走 track:status:stream），Layer 1 也不读 zone:state:stream。

---

## 4. Layer 1 — Ghost adjudication（**双分支**）

### 4.0 分支选择 — 由 room.kind binary 分类决定（决定 16）

Layer 1 入口按 `room.kind` binary 分流：
- `room.kind == "bathroom"` → **BathroomGhostAdjudicator**（§4.A，几何先验为主）
- 默认（未标注 / 任何其它值）→ **GeneralGhostAdjudicator**（§4.B，统计累积为主）— 视作 bedroom

两套 adjudicator **不共享内部状态**（cell history / mirror line 在 bathroom 单独维护），但**共用** Verdict 类型、TrackStatus 输出契约、FeedbackEvent 消费接口。

**v3 扩展点**：未来 kind 枚举完整化（living / kitchen / hallway 等）时，default 分支会进一步细分；bathroom 分支语义不变。

### 4.A BathroomGhostAdjudicator — Suite Bedroom Census + 入口流量（v2 决定 7+11+12+13）

bathroom 内的 ghost 是镜面 / 金属反射 / 玻璃门多径主导，**统计累积式信号失效**且 mirror 几何检测投入产出比低。
**本分支不在 bathroom 内部做复杂判定**，而是依赖 **bedroom census + bathroom 入口流量**派生 bathroom 人数。

#### 4.A.1 SuiteBedroomCensus — Bedroom 内人数普查（核心数据结构）

```go
type SuitePerson struct {
    PersonID        string   // resident.id (来自 unit.residents) 或 "visitor_<n>" 临时 ID
    Role            string   // "resident" / "visitor" (visitor 不在 unit.residents 表)
    AnchorTrackID   int      // 当前主 anchor 的 radar track id（bedroom 内）
    AnchorRoomKind  string   // "bedroom" 或 "bathroom"
    AnchorSince     int64
    LastActiveMs    int64
    SleepadAnchored bool     // 床压双源锚定（仅 resident 可能为 true）
}

type SuiteBedroomCensus struct {
    SuiteID   string                    // = bedroom RoomID（v2 hard-code 1:1）
    Persons   map[string]*SuitePerson   // key=PersonID
    
    // 入口流量累积
    BathroomCount int  // bathroom 内当前人数 = 累积 gate 流量
    
    mu        sync.RWMutex
}
```

**Census 升格规则**（**只在 bedroom 内识别**）：

```
Resident（已知）：
  - 通过 sleepad 双源锚定（sleepad InBed + radar 同步）→ 直接 census persistent
  - 或 radar 在 bedroom 存活 ≥ 5 min + traverse ≥ 10 cells
  - PersonID = 来自 unit.residents 表

Visitor（未知，临时）：
  - bedroom 内"非 resident 区域"出现新 track
  - 存活 ≥ **2 min**（决定 13 时序假设：护工进 bedroom 不会马上走，2min 足够确认）
  - traverse ≥ 5 cells
  - → 临时分配 PersonID = "visitor_<timestamp>"
  - 离开 suite ≥ 10 min 后从 census 移除
```

**bathroom 不识别 person**：bathroom 内 track 只用于"占用一致性"判定，不写入 SuitePersons。

#### 4.A.2 Bathroom 入口流量（决定 15 layout 显式标注）

```
入口 cell 识别（决定 15 主路径）：
  - 扫描 bathroom 周围 cells where cell.EnterTarget == "bathroom_<this_id>"
  - 这些 cells 构成入口区域（约定 1 个连续区域）
  
Fallback（决定 12 兜底）：
  - layout 未标注 EnterTarget → 假设 suite 拓扑，bathroom 紧邻 bedroom 一侧的 AreaEnter cells 视作入口
  - 多片错误（layout error）→ 取第一片，audit log

zoneengine BathroomGate 子模块（监测入口 cells）：

外部 → bathroom 内（track 越过入口 cells 进入 bathroom zone）：
  - BathroomCount +1
  - 该 person.AnchorRoomKind = "bathroom"
  - 该 person.LastBathroomEntryMs = now

bathroom 内 → 外部（track 越过入口 cells 反向离开 bathroom zone）：
  - BathroomCount -1（floor 0）
  - 该 person.AnchorRoomKind = "<外部 room kind>"
  - v2 suite 拓扑下"外部"几乎总是 bedroom；v3 unit-layout 后可能是客厅 / 走廊

兜底 -1（应对 track 失锁导致流量未记到）：
  - bathroom 内无任何 track ≥ 30s
  - 入口附近 ≤ 50cm 出现新出生 track（在外部 room 内）
  - → 视作"从 bathroom 出来"推断 -1
```

**关键**：
- BathroomCount 由入口流量驱动，不由 bathroom 内 track 数驱动
- v2 设计**不关心**外部 room 是 bedroom 还是别的（决定 15 让 BathroomGate 拓扑无关）
- 这让 v3 unit-layout 演进时，BathroomGate 业务逻辑**几乎不需要重写**

**Storage / 储衣室处理（决定 15）**：
- 储衣室物理上是 bathroom 延伸（仅通过 bathroom 出入）
- layout 编辑时人工划入 bathroom zone（cells 标 AreaUnknown）
- 储衣室 cells 在 bathroom zone 内 → BathroomCount 自动覆盖
- 储衣室内 track 自动适用 bathroom fall 规则

#### 4.A.3 Bathroom 内部 ghost 判定（4 条简化规则）

```
规则 1（出生地判据 — 用户确认有效）：
  bathroom 内"新出生"track（出生 cell 不在入口附近 ≤ 30 cm）
  → 直接判 Ghost（GhostPenalty=100，绝对 ghost）
  → 镜面 / 金属反射 ghost 通常出生在洗手台 / 淋浴区中间，远离入口 ✓
  → 决定 12 保证入口是唯一物理通道，从入口外的位置出生 = 必非真人

规则 2（占用一致性）：
  bathroom 内活 track 数 > BathroomCount（来自入口流量）
  → 多余的 track 是 Ghost
  → 选择策略：保留"从入口轨迹连续"的，丢"远离入口轨迹"的

规则 3（Ghost-Real 共生律自校正 — 决定 14）：
  bathroom 内活 track 数 ≥ 1（不论 verdict）
  → 必有 ≥ 1 真人在 bathroom（反射源约束，决定 14）
  → 期望 BathroomCount ≥ ceil(track_count / max_ghost_per_person)
    max_ghost_per_person 默认 = 3（1 真人最多产生 3 个镜面 / 金属 / 玻璃 ghost）
  
  若入口流量 BathroomCount < 期望值（漏记 +1）：
    → 修正 BathroomCount = max(BathroomCount, expected_min)
    → audit log: gate_undercount_corrected
    → 不主动翻 ghost verdict（Ghost 已被 §4.A.5 镜像学习 disambiguate）

  反向场景：bathroom 内活 track 数 == 0 + BathroomCount > 0
    → 真人物理上仍应在 bathroom（出门事件未触发）但 sensor 完全看不到
    → 触发 §6.A.3 BathroomLostFall 最强信号（不在本节，详 fall 规则）

规则 5（Split-ghost 出生邻近性 — 决定 14 共生律的瞬时表现）：
  bathroom 内新出生 track T：
    - 与房间内已存在 track A 的最近距离 < 50-80cm（默认 60cm）
    - 出生时间窗内 A 在镜面 / 反射区方向有运动（可选加强）
  → 强烈疑似 split ghost（决定 14 共生律的物理胎记）
  → 直接判 Ghost（GhostPenalty=100），不等镜像学习累积证据
  
理由：
  - 决定 14 共生律：ghost 必有真人作为反射源
  - Split ghost 是从真人 track 上"分裂"出的衍生 ghost（雷达同时收直射 + 反射回波）
  - 出生瞬间紧贴真人 → 这是 split 物理本质的胎记
  - 与规则 1（远离入口出生）正交：split 出生位置不一定远离入口，但一定紧贴真人
  - 比镜像学习快：不需要等 reflection_line 累积，瞬时识别
  - 比规则 2 快：不需要等 BathroomCount 入口流量稳定
  
注：
  - "运动同步性持续验证"（窗口 30-60s motion correlation）作为出生瞬间没抓到的兜底
  - 1Hz 采样下噪声大，需多指标投票（位置相关 + 距离稳定性 + 镜像对称性）
  - Phase 2 实测决定是否启用持续验证

规则 4（fallback — 入口在 radar 盲区，**边缘 case ≤ 5%**，Q-prereq-5 答"绝大多数在视野内"）：
  若该 bathroom 入口不在 RadarVisibleRegion 内：
    → 入口流量不可获取，BathroomCount 不可信
    → 降级到出生地判据（规则 1 仍可用，bathroom 入口物理上还存在）
    → 占用一致性（规则 2）失效（无 BathroomCount 基线）
    → 共生律自校正（规则 3）失效（无 BathroomCount 可校正）
    → 镜像学习（§4.A.5）仍可用 disambiguate 真人 / ghost track
    → 同时 fire ai.log warning + 部署 escalate（建议调整 radar 安装位置）
```

#### 4.A.4 ~~Ghost 身份持久化~~ — **删除**（决定 14 后不必要）

决定 14 Ghost-Real 共生律论证：真人离开 bathroom → ghost 秒级消失。
"Ghost latch 24h" 解决的是"真人离开但 ghost 残留"的场景，**物理上不存在**，所以 latch 机制删除。

唯一保留的简化语义：track 被 §4.A.3 任一规则判 Ghost 后，**当前 verdict 持续到 track 自然消失 / BathRoomState→Vacant**——这只是 verdict 的正常生命周期，不算 latch。

#### 4.A.5 Bathroom 内镜像学习 — **P0 保留**（决定 14 后角色变更）

**用户三次确认**：bathroom 内镜像学习是 v2 P0，不降级到 Phase 2。

**新角色**：从"主动抑制 ghost 占用计数" → 改为 **disambiguation 哪个 track 是真人**

为什么需要 disambiguation：
- 入口计数 BathroomCount 知道"bathroom 内有几个真人"
- 但**不知道**当 bathroom 内 N 个 track 时，"哪个 track 是真人，哪个是 ghost"
- Fall 规则需要正确归因（"哪个 track 摔倒"）
- 入口轨迹追溯（规则 2 "保留离入口轨迹近的"）是一个判据，但不总可靠（蒸汽时入场轨迹被截断）
- **mirror reflection_line 提供独立几何判据**：知道 reflection_line 后，离雷达更远的 track 是 ghost

**离线学习 pipeline**（保留之前设计，每 bathroom 独立 daily 运行）：

```
输入：过去 N 天该 bathroom 所有 track frames（含历史 ghost 标记 — Q-prereq-3 镜面历史保留）

对每对长期共存 ≥ 60s 的 track (A, B)：
  1. 拟合直线 L 使 ‖mirror(A_t, L) - B_t‖ 在所有 t 上最小
  2. 验证残差 r < 15cm AND 距离差稳定性 σ_d < 10cm AND 运动学反对称
  3. 高置信 → 写入 room.reflection_lines[] = {line, confidence, last_confirm_ms, evidence_n}
  
衰退：30 天未确认 → stale；60 天 → 删除
冷启动：bathroom 入住前 N 天 reflection_lines 为空 → 仅靠规则 1+2+3 工作
```

**实时使用路径**（每帧 O(N²) 在 bathroom 内，N 通常 ≤ 3）：

```
对每对活 track (A, B)：
  for each reflection_line L in room.reflection_lines:
    if ‖mirror(A, L) - B‖ < 30cm AND 运动学反对称 ≈ 满足:
      → 该对是镜像 pair
      → 离雷达更远的标记 Verdict=Ghost
      → 不影响 BathroomCount（occupancy 由入口流量保证）
      → 仅影响"哪个 track 是真人"标识，用于 fall 归因
```

**与 §4.A.3 规则 1/2/3/5 的关系**：
- 规则 1（出生地）：bathroom 内新出生远离入口 → Ghost；**与镜像学习独立工作**
- 规则 2（占用一致性）：track 数 > BathroomCount → 多出的是 ghost；**用镜像学习辅助挑选哪个是 ghost**
- 规则 3（共生律自校正）：track 数 ≥1 → 反推 BathroomCount；**不依赖镜像学习**
- 规则 5（split-ghost 出生邻近性）：新出生紧贴已有 track → ghost；**瞬时识别，不等镜像学习累积证据**
- 镜像学习：disambiguate 哪个 track 是真人（fall 归因关键）；规则 5 是其**出生瞬间的快速 entry**

**双轨 disambiguation 体系**：
| 时机 | 工具 | 触发速度 | 准确度 |
|---|---|---|---|
| 出生瞬间 | 规则 5 出生邻近性 | 立即（< 1s）| 中（紧贴 ≠ 一定 ghost）|
| 累积长期 | 镜像学习 reflection_line | 离线 daily | 高（几何硬约束）|
| 失败兜底 | 规则 2 入口轨迹追溯 | 中 | 中 |

**为什么 v2 必须保留**（用户三次强调）：
- bathroom 是 elder care 头号事故场景
- fall 归因错（把 ghost 当摔倒主体）= 漏报真人 fall
- 入口轨迹追溯在蒸汽 / 多普勒失效时不可靠
- 镜像几何是物理硬约束，不被蒸汽 / 多普勒影响
- 历史数据保留（Q-prereq-3）已具备 → daily pipeline 可立即启用

#### 4.A.6 Public Bathroom Standalone Mode（决定 24）

Elder care 机构楼层公共 bathroom（活动室旁 / lobby 旁 / 走廊上）部署的 standalone 模式：

```
触发条件（2026-05-17 PR-7 修订：显式配置不推断）：
  room.kind == "bathroom" AND rooms.is_public_bathroom == TRUE
  → 走 Public Bathroom Standalone Mode

为什么显式配置而不是"推断无 sibling bedroom"：
  - 业务语义（多人共用 vs 私用）≠ 拓扑事实（有无 sibling bedroom）
  - 楼层 public bathroom 可能挂在含管理办公室 / 储物间的 /80 下 → 拓扑推断失败
  - elder suite bootstrap 时 bedroom 可能晚于 bathroom register → 启动顺序敏感
  - 错误成本不对称：suite→public 误判 → fall 归因仅 location，elder 救援延迟
  - 决定 16 已走显式（vue checkbox），决定 24 顺承

配置：
  - 只装 radar，不装 sleepad（用户确认 Q1）
  - 无 SuiteBedroomCensus（无附属 bedroom 不识别 person）
  - 无 BathroomGate 入口流量（入口不连任何 bedroom）
  - 无 SuitePerson 概念（不识别 elder 身份）
  - SuiteID = bathroom 自身 /128（不与 suite bathroom 的 unit /80 SuiteID 冲突）
```

**SuiteID 取值约定（PR-7 落地）**：

| 场景 | SuiteID | 理由 |
|---|---|---|
| suite bathroom (附属某 bedroom) | unit_spatial_prefix /80 | 与 sibling bedroom 共享 census bucket |
| public bathroom (standalone) | bathroom_spatial_prefix /128 | 自闭 census bucket，多 public bathroom 互不干扰 |

**适用 / 不适用的 §4.A 机制**：

| 机制 | Suite-bathroom 模式 | **Public bathroom 模式** |
|---|---|---|
| §4.A.1 SuiteBedroomCensus | ✅ | ❌ 不适用（无附属 bedroom） |
| §4.A.2 BathroomGate 入口流量 | ✅ | ❌ 不适用（入口不连 bedroom） |
| §4.A.3 规则 1 出生地判据 | ✅ | ✅ **保留** |
| §4.A.3 规则 2 占用一致性 | ✅ | ⚠️ 降级（BathroomCount 仅从内部 track 派生）|
| §4.A.3 规则 3 共生律自校正 | ✅ | ✅ **保留**（决定 14 物理事实） |
| §4.A.3 规则 4 fallback | ✅ | N/A |
| §4.A.3 规则 5 split-ghost 出生邻近性 | ✅ | ✅ **保留** |
| §4.A.5 镜像学习 | ✅ | ✅ **保留**（不分 suite/public） |

**BathroomCount 的派生（public mode 简化）**：

```
public mode BathroomCount = bathroom 内 disambiguate 出的真人 track 数
  - 通过 §4.A.5 镜像学习 + §4.A.3 规则 1/3/5 识别 ghost
  - 剩余 verdict ∈ {Real, Anchored, Pending} 的 track 数 = 真人数
  - 不依赖入口流量（无 BathroomGate）

规则 2 占用一致性在 public mode 失效（无独立 BathroomCount 基线对比）
  → 完全依赖 §4.A.5 镜像学习 + 规则 1/5 出生判据
```

**Fall 触发主体（public mode 简化）**：

```
触发主体 = bathroom 内 disambiguate 出的真人 track
  - 不是 SuitePerson（无 person 概念）
  - 不识别 elder 身份
  - Fall 归因仅到 location（如 "public bathroom #lobby-1F" 内有人 fall）

后续告警（v2 范围外）：
  - cardagg / data 服务接收后通知该楼层所有 admin / 护工
  - 由人类前往现场识别具体 elder
```

**部署前提**：
- public bathroom 入口仍约束在 RadarVisibleRegion 内（同 §4.A.2 前提）
- 否则降级到 §4.A.3 规则 4 fallback（仅规则 1 / 镜像学习可用）

### 4.B GeneralGhostAdjudicator — 统计累积式（default = bedroom，决定 16）

v2 范围下 default 分支 = bedroom（决定 16 binary 分类）。沿用现有 GhostPenalty 累积范式 + 补强信号。
v3 扩展时再细分 living / kitchen 等亚分支。

#### 4.B.1 信号 d — Sleepad 跨源排他（首推，bedroom 核心）

sleepad 锚定 "床上 1 人" 后，房间内若出生第二个 track：
- 单人 unit 场景：床上 + 另一处真人**几乎不可能** → +40 penalty
- 多人 unit：访客/家属可能在房间内 → +20 penalty

#### 4.B.2 信号 a — Anchor 互斥

第一个长存活 track anchored=Real 后，第二 track 出生默认 +30 penalty 起步。
**bedroom-specific 加成**：sleepad 锚定方向（床上 vs 床下）与新 track 位置一致性检查 → 不一致再 +20。

#### 4.B.3 信号 e — Traverse 差异（多时间尺度）

| 窗口 | 阈值 | 用途 |
|---|---|---|
| 10s | < 2 cells | 新生 ghost 早期识别 |
| 60s | < 5 cells | 主判据 |
| 5min | < 15 cells | anchor 升级 gate |

#### 4.B.4 信号 c — Motion correlation（Phase 2，实测后定）

1Hz 采样下 30s 窗口噪声大，需 prod 数据验证 corr 阈值是否分得开真双人 vs ghost-real 对。

### 4.C 共享部分（两个分支都用）

| 项 | 实现位置 |
|---|---|
| Verdict 类型：Pending / Real / Anchored / Ghost | track.go |
| TrackStatus 输出 contract | §3.1 |
| FeedbackEvent 消费（zone → ghost）| §3.3 |
| Pre-fall ghost rescan（fall trigger 前急刹车） | 共用，bathroom 触发更频繁（每 60s 周期化）|

### 4.D Open Q-A 系列（保留）

| 编号 | 问题 | 状态 |
|---|---|---|
| Q-A1 | Track-ID 是否解耦 firmware id（方案 1/2/3）| 倾向方案 2，但 bathroom 分支下次要（镜像 ghost 比 ID 复用更紧迫）|
| Q-A2 | Verdict 加 Anchored 一等态 | 倾向是；Bathroom 分支额外需要"Ghost latch 不可翻 Real" |
| Q-A3 | Layer 1 消费 FeedbackEvent 方式 | 倾向 A（直接加 penalty）|

---

## 5. Layer 2 — Zone state（议题 B + 新议题 E）

### 5.1 新议题 E — Radar 可观测区 & 盲区出入口

**新增 layout 元素**：

| 元素 | 几何 | 来源 | 用途 |
|---|---|---|---|
| `RadarVisibleRegion` | 多边形 | 配置 + 学习 | radar 真正能看到的区域；超出此区域 = 盲区 |
| `BlindEnter` | 矩形 | 配置 + 学习 | radar 看不见的房间出入口（门、过道、视野外的椅子区） |

**自动学习**：
- sensor 累积"track 反复出生 / 反复消失"的位置 → 推荐 BlindEnter 候选（待人工确认）
- sensor 累积"FOV 边缘 track 总是失锁"的弧线 → 推荐 RadarVisibleRegion 边界
- 学习写 ai.log，layout 编辑器从 log 拉建议（不自动写 layout）

**影响范围**：

| 既有逻辑 | 新行为 |
|---|---|
| `birthScore` 因子 1 (dEntry) | dEntry = min(visible Enter, BlindEnter) — track 出生在 BlindEnter 附近不再 ghost penalty |
| `birthScore` 因子 4 (no_enter_pair) | 若 BlindEnter 附近出生，pair 检查放宽（盲入口不可能有 firmware EnterRoom 事件） |
| `lost_fall` pending 入池前置检查 | track 失锁位置在 BlindEnter 附近 + 无 frozen run → 视作 ExitRoom 兜底，跳过 pending |
| `RoomState` Vacant 翻转 | RoomState 不依赖 firmware ExitRoom 单一信号；track 在 BlindEnter 附近失锁也算 vacate 候选 |
| `EnterRoom` 缺失补全 | 不补 EnterRoom 事件本身（避免污染审计），但 zoneengine.adapter_radar 直接给 RoomState +occupied delta |

**配置/学习取舍**：
- MVP：仅支持配置（layout 工具加 RadarVisibleRegion + BlindEnter 绘制）
- 二期：开启学习并提供 layout 编辑器建议
- 三期：自动学习 + 用户审核工作流

### 5.2 Open Q-B1 — Zone state 翻转的权威来源（已部分对齐）

通过决定 1，zoneengine 是 bed/room/bathroom 状态的**唯一真相**。
剩余子问题：
- **B1a**：✅ **决定 16 已拍**：`room.kind == "bathroom"` 显式标注（vue 编辑器 checkbox）是**唯一**权威源。旧候选**全部废弃**：
  - ❌ `room.name` 字符串匹配（"bathroom"/"restroom"/"toilet"）— 字符串脆弱
  - ❌ `Stay alarm 启用` 间接判据 — 违反 #1.3 单源真相
  - ❌ `cell.AreaType ∈ {Toilet, Shower}` 作为"是不是 bathroom"判据 — 仅保留作为 bathroom **内部** still vs bedside 细分判据
- **B1b**：BlindEnter 的存在让 RoomState 翻转规则需重写 — 详 §5.1

### 5.3 Open Q-B2 — 4 fall 私有计数器全删（已对齐）

通过决定 1，4 fall 全改读 ZoneState/ZoneEvent。涉及代码改动：
- 删 `bedPersonCount` / `lastLeftBedAt` / `lastNumberPeopleZeroMs` / `bathroomRealCount` / `BedSession` 大部分逻辑
- 保留 `BedSession.LeftBedHadHRRR` / `BedSession.LeftBedMaxPeople` 等 latch 字段（这些不是状态，是 LeftBed 时刻的证据快照）

---

## 6. Layer 3 — Fall detection（**双分支**）

### 6.0 分支选择 — room.kind binary 分类（决定 16）

| 条件 | 适用 fall 规则 |
|---|---|
| `room.kind == "bathroom"` | **§6.A BathroomFallRules**（3 类规则全部 Count 不依赖 / Count 降级模式）|
| 默认（未标注 / 视作 bedroom） | **§6.B GeneralFallRules**（4 类规则消费 ZoneState.Count 正常 gate）|

**v3 扩展点**：living / kitchen 细分时再加分支（如 kitchen 抑制 still_fall — 参考 [kitchen_lostfall_known_limitation](../../.claude/projects/-home-wisefido-owl/memory/kitchen_lostfall_known_limitation.md)）。

### 6.A 设计原则：Critical 精确 + Warning 兜底（PR-10 review 2026-05-17 锁定）

**两层告警协同**：

| 层 | 触发源 | 阈值 | level | 设计目标 |
|---|---|---|---|---|
| **Critical 精确** | §6.A 4 类规则 (PR-10) | 多档（30s lost-strong / 7min lost-weak / 8min bedside / 10-12min still）| Critical Fall | 救援响应快；多档覆盖不同物理场景 |
| **Warning 兜底** | zonealarm.Supervisor Stay rule | 10min bathroom occupied | Warning | Critical 失效（阈值未到 / wire 漏 / bug）时的 floor |

**协同语义**：

1. **Critical 多档故意重叠**：10b (bedside_silent, 8min) 和 10d (suite_person_silent_with_ghost, 7min) 在"bathroom 内非 Toilet/Shower 静止"路径上**会同帧 fire**——这是设计意图：
   - 10d 是"SuitePerson 主体"视角（person silent ≥ 7min，含 ghost 反射源仍存）
   - 10b 是"bathroom 房间状态"视角（90s grace + 决定 18 多人不 fire）
   - 各自携带不同 Reason 字段，cardagg 端可按 Reason group / dedup；不允许在 sensor 内 dedup（每条 Reason 对应不同救援响应优先级）
   - 重复 fire 不是 bug — 是同一物理事件的"多视角佐证"

2. **Warning 兜底始终 armed**：zonealarm.DefaultRules 含 `Stay / bathroom / occupied / 600s / WARNING`，不依赖任何 per-room flag。只要 zoneengine 子系统 wire 起来即生效。这是 sensor 整体 fall 检测的 floor：即使 PR-10 全部失效，bathroom 持续 10min 占用仍发 Warning 告警。

3. **bootstrap-wiring 不变量**（v2 main.go 必须同时满足，详 [sensor_v2_known_limitations.md L4](sensor_v2_known_limitations.md)）：
   - `wiring.NewSubsystem(...)` 启动 zoneengine + zonealarm.Supervisor → Warning 层 armed
   - `engine.SetBathroomFallRules(NewBathroomFallRules(...))` → Critical bathroom 层 armed
   - `engine.SetBedroomFallRules(NewBedroomFallRules(...))` → Critical bedroom 层 armed
   - bootstrap-PR 实施时**应 log.Fatal 拒启动**任一缺失场景（不接受降级运行）
   - 启动 log 应同时出现 `bathroom_fall_rules_initialized` + `bedroom_fall_rules_initialized` + `zone_alarm.yaml loaded`，缺任一行即漏接

### 6.A.0' outRoom / enterRoom 事件的不可靠性（layout 约定的副作用）

**根本原因 — Layout 编辑约定（2026-05-17 用户确认）**：
- Layout 开发时使用真实 radar 测量 room 物理边界
- **radar 盲区 / 信号丢失区被人工标记为 `enter` cells**（用作"track 失锁等价于离开"的判定基础）
- 走出 radar 识别区（包括真出门 + 走入盲区）→ firmware 触发 `enter/out room` event

**直接后果 — outRoom 事件无法物理区分以下两种情况**：

| 情况 | 物理事实 | outRoom event 表现 | 业务影响 |
|---|---|---|---|
| 真出门 | 人走出 room | 触发 outRoom | 期望行为 |
| 走入盲区 | 人仍在 room 内（信号被遮挡）| 触发 outRoom | 误报"离开" |

**对 §6.A 规则的影响**：
- §6.A 不能用 outRoom event 作 cancel trigger（盲区遮挡会假取消 Critical fall）
- §6.A.3 Lost-Fall 取消条件只能用：BathroomGate 反向流量（PR-5） + SuitePerson.AnchorRoomKind 翻 bedroom + 主动重新观测到 track —— 均为**正向证据**（看到人 = 取消），不接受 outRoom 这种"假阴性证据"
- zonealarm.Supervisor 在 outRoom event 上仍可走原有 cancel 路径（Warning 容错更高，假取消代价小）

**v3 规划**：unit-layout 后引入 cross-source consistency（多 radar 互相验证 + sleepad 床压补 outRoom 真假判定），届时 outRoom 才有"真"可言。v2 范围内**outRoom 是软证据**，只 inform 不裁决。

### 6.A BathroomFallRules — Suite Person lost（不是任意 track lost）

**Public bathroom 适用范围（决定 24 + §4.A.6）**：
- §6.A 各规则**同时适用 suite-bathroom + public-bathroom**
- 触发主体在 public mode 改为"bathroom 内 disambiguate 真人 track"（不是 SuitePerson）
- Fall 归因在 public mode 仅 location 不识别 elder 身份
- 取消条件中"SuitePerson 跨 BathroomGate"在 public mode 不适用（无 BathroomGate）→ 改用"track 完全离开 bathroom radar 视野"

**核心设计原则修订（决定 7+11+12+13+14）**：
- BathroomCount 由 bedroom census + 入口流量派生，**已可信**
- Fall 规则用 **SuitePerson lost** 而不是 **任意 track lost** —— ghost 滞留 / firmware coast 不能压制告警
- Lost / Still 触发前必跑**系统健康检查**（断网不当 fall）
- **决定 14 关键推论**：**Ghost 是真人存在的间接证据**，Fall 规则**不因 ghost verdict 而抑制触发**
  - bathroom 内有 ghost = 真人在场（反射源约束）
  - 镜像学习（§4.A.5）用于 disambiguate "哪个 track 是真人"，但 fall 触发不依赖 disambiguate 结果
  - 即使 disambiguate 失败（无 reflection_line），仍按"track 存在 = 真人存在"触发 fall
- **Storage / 储衣室处理（决定 15）**：
  - 储衣室合并到 bathroom zone（cells 标 AreaUnknown，layout 编辑时人工划入）
  - **BathroomStillFall** 仅在 cell ∈ {Toilet, Shower} 触发 → 储衣室**不触发**（Stand 取衣常态）
  - **BathroomBedsideFall** 90s grace 后任意位置静止 ≥8min → 储衣室**触发**（衣柜前 8min 静止异常）
  - **BathroomLostFall** track 失锁覆盖整个 bathroom zone（含储衣室）

```
触发链顶层：
  is_suite_healthy(suite_id) AND
  exists SuitePerson p where p.AnchorRoomKind == "bathroom" AND
  p.LastActiveMs older than threshold
```

#### 6.A.0 系统健康检查（共用前置，suite 版本）

```go
func isSuiteHealthy(suiteID string) bool {
    // bedroom radar 最近 30s 有心跳 + 数据帧
    // bathroom radar 最近 30s 有心跳 + 数据帧（若部署了）
    // sleepad 最近 60s 有心跳（bedroom）
    // server-suite 链路无积压
    // 不在维护窗口内
    // SuiteBedroomCensus 在过去 1h 内被更新过（防 redis flush 后空 census）
}
```

任一不满足 → fall 触发降级为"系统状态异常"通知（不当 fall 报警）。

#### 6.A.1 BathroomStillFall — 替代 still_fall

```
触发条件（用 SuitePerson，不用任意 track；决定 14 后 ghost 不抑制）：
  - 系统健康检查通过（§6.A.0）
  - 存在 SuitePerson p where p.AnchorRoomKind == "bathroom"
  - 该 p 对应的 bathroom 内 track 存在（verdict 不限—— ghost 也算，决定 14）
    优先取镜像学习 disambiguate 出的真人 track；若失败取任意 track（保守）
  - track 位置 cell.AreaType ∈ {Toilet, Shower}
    （bathroom 类型已在 §6.0 按决定 16 room.kind 分支 gate；本节内 cell.AreaType
     仅用于 bathroom 内部细分 still vs bedside 路径，不作为 bathroom 类型判据）
  - pose=Stand（坐马桶/淋浴的 Sit 不触发，这是正常长静止）
  - 静止时长 ≥ 10/12 min（风险时段 / 非风险时段，比卧室 15/18 严）

SuitePerson 数据一致性 audit（不是 ghost 判定）：
  - 决定 14 共生律之后，"bedroom 活动 → bathroom 静止 = ghost" 推论不再成立
    bathroom 静止 = 真人危险信号（fall 候选），无论 bedroom 内是否有其它 person 活动
  - 但同一 PersonID 不能在两处由 SuitePerson.AnchorRoomKind 单值字段保证；
    若发生（bedroom census + bathroom census 同 ID 矛盾）→ 数据健全性错误
    → 触发 audit log，按入口流量最近一次事件修正 AnchorRoomKind
  - suite 内多 person 场景：bathroom 内 p1 + bedroom 内 p2（不同 ID）是正常 case
```

#### 6.A.2 BathroomBedsideFall — bathroom 内 R4 等价

bathroom 没有床，R4 等价物 = "进入 bathroom 后任意位置长静止"：

```
触发：
  - 系统健康检查通过（§6.A.0）
  - BathRoomState 翻转 Vacant → Occupied 后 90s 内 → 不触发（grace period）
  - BathRoomState=Occupied 持续 ≥ 90s 后：
    - 存在 SuitePerson p where p.AnchorRoomKind == "bathroom"
    - p 对应 track 存在（verdict 不限，决定 14 ghost 不抑制）
      优先取镜像学习 disambiguate 出的真人 track
    - track 在同位置静止 ≥ 8 min
    - 非 cell.AreaType ∈ {Toilet, Shower}（这些有 BathroomStillFall 覆盖）
    - 夜间不要求 Stand pose（地上躺着算 Lie pose）
    → fire BathroomBedsideFall (Reason="bathroom_long_static")

BathroomCount 处理（决定 18 修订 — 真多人由人类处理，不 fire）：
  - BathroomCount==1（单一 SuitePerson）→ **fire Fall**
  - BathroomCount==2 但 1 个 ghost（镜像学习 disambiguate 出 1 真人 + 1 ghost，共生律）→ **fire Fall**
  - BathroomCount==2 且 2 个都识别为真人（真多人陪同）→ **不 fire**
    理由：决定 18 — 真多人时人类已在场处理（护工陪伴的核心目的就是处理 elder 异常）
  - BathroomCount≥3 真多人 → 同上 **不 fire**
```

#### 6.A.3 BathroomLostFall — 两档信号（决定 14 共生律后修订）

**决定 14 推论**：Ghost 是真人存在的间接证据，所以 fall 触发不因 ghost 抑制。
两档信号按"是否还有 ghost 间接证据"分级：

**最强档（立即 fire，30s 触发）— 真人完全失锁**

```
触发：
  - 系统健康检查通过（§6.A.0）
  - BathroomCount ≥ 1（入口流量显示有人在 bathroom）
  - bathroom 内活 track 数 == 0 持续 ≥ 30 s
    （连 ghost 都没有 = 真人最严重的物理状态异常）
  → fire BathroomLostFall (Reason="suite_person_completely_lost_no_ghost_proxy")
  → 这是最强 lost 信号，立即报警不等 7min
  
物理论证：
  - 决定 14：真人在 bathroom → 必有 ghost（除非完全无反射几何）
  - 现在连 ghost 都消失 = 真人物理状态极端异常
  - 或：真人离开 bathroom 但出门事件未触发（部署 bug）→ 仍需 escalate
```

**次强档（7 min 触发）— 真人静默但 ghost 还在**

```
触发：
  - 系统健康检查通过
  - 存在 SuitePerson p where p.AnchorRoomKind == "bathroom"
  - bathroom 内仍有 track（real 或 ghost 都算，决定 14 把 ghost 视为真人存在证据）
  - p 对应 track（按 §4.A.5 镜像学习 disambiguate 出的真人 track）静止 ≥ 7 min
  → fire BathroomLostFall (Reason="suite_person_silent_with_ghost_proxy")

取消：
  - p 通过 bathroom 入口 → bedroom（BathroomGate 反向流量）
  - p.AnchorRoomKind 翻 "bedroom"
  - p 对应 track 再次活动
  - BathRoomState→Vacant（BathroomCount 归零）
```

#### 6.A.4 BathroomSilentFall — 不适用

bathroom 无 sleepad，silent_fall 在 bathroom 分支不存在。

### 6.B GeneralFallRules — default = bedroom（决定 16 binary 分类）

#### 6.B.1 silent_fall

```
触发：BedState 仍 Occupied/Leaving + sleepad 直发 LeftBed + 60s 内 BedState 仍未→Vacant
依赖：Layer 2 BedState 翻转沿 + sleepad event
不需要 fall 内 ghost gate（Layer 1 已过滤）
```

#### 6.B.2 lost_fall（v2 suite bedroom 版本）

```
触发：
  - 系统健康检查通过（§6.A.0）
  - 存在 SuitePerson p where p.AnchorRoomKind == "bedroom"
  - p.LastActiveMs > cell-area-typed 时长（5min walkway / 60min bed / ...）
  - 期间无 p 跨房间转移事件（无 BathroomGate 通过事件）
  - 无 RoomState→Vacant 翻转（含 BlindEnter 兜底，§5.1）

取消：
  - p 重新活动（新 active frame）
  - p 跨房间转移到 bathroom（BathroomGate 流量）
  - p 离开 suite（resident 不会离开，visitor 可能）
  - 其它 SuitePerson 也在 bedroom（陪同场景，资源覆盖）

注意：不再用"任意 track lost"作为触发——ghost 滞留不能压制告警
v3 扩展：unit 多 room 时按 RoomID 分别判定，p.LastActiveMs 跨 room 维护
```

#### 6.B.3 bedside_fall

```
触发：
  - BedState transition: Occupied/Leaving → Vacant 后
  - Real track 在床边 ≤100cm 静止 ≥15min（夜间）
  - RoomState.Count == 1（多人 unit 时放宽到 ≤2 但降级 SuspectedFall）
```

#### 6.B.4 still_fall

**v2 范围内不实现**（决定 16 后）：
- v2 elder care suite 拓扑下，default = bedroom 路径中 Toilet/Shower cell 不应出现（这些 cell 仅在 bathroom 内学习生成）
- 旧"Stay alarm 启用 → 视作 bathroom-like" 间接判据**已废弃**（违反 #1.3 单源真相）
- v3 扩展 living/kitchen 等房间类型时，再加专门的 still_fall 规则（参考 [kitchen_lostfall_known_limitation](../../.claude/projects/-home-wisefido-owl/memory/kitchen_lostfall_known_limitation.md)）

v2 默认 bedroom 路径只有 3 个 fall：silent_fall / lost_fall / bedside_fall。

### 6.C Fall 之间 dedup 优先级（共用）

```
silent_fall  >  bedside_fall  >  bathroom_*  >  still_fall  >  lost_fall
```

bathroom_* 系列在 bathroom 内是头号优先（覆盖该房间）；同房间 60s 内不允许低优先级 fall fire。

### 6.D Open Q-C 系列

| 编号 | 问题 | 倾向 |
|---|---|---|
| Q-C1 | Fall 之间 dedup 优先级 | 见 §6.C |
| Q-C2 | silent_fall 在 Layer 3 剩余职责 | 见 §6.B.1（已对齐）|
| **Q-C3 (新)** | BathroomBedsideFall 8min 时长是否合理 | 待 prod 数据校准 |
| **Q-C4 (新)** | BathroomBedsideFall 90s grace 是否合理 | 短了误报 / 长了真摔倒漏报，待数据 |

---

## 6.5 ~~多模态融合~~ — **整章删除**（Q-prereq-2 答无传感器）

bathroom 多模态融合（湿度 / 水流 / PIR / 门磁）在 v2 **不实现**：
- 产品形态不支持这些传感器
- 即使部分客户安装，非同一系统不可接入 → 视为无
- v2 bathroom 保护**纯雷达** + 决定 14 共生律 + 决定 15 layout 标注 + 决定 16 room.kind

未来若产品引入辅助传感器，v3/v4 再加多模态融合层。

---

## 6.6 阈值矩阵（决定 19 — risk_factor = 1.5 简化）

**单一 `risk_factor = 1.5`**，合并 risk-time（夜间）+ multi-resident 两个维度为一个简化因子。

```
threshold = base_threshold × risk_factor

risk_factor = 1.0 (normal) / 1.5 (risk scenario)

risk scenario 当前定义（任一满足）：
  - multi-resident bedroom（unit.residents.count ≥ 2）
  - 是否包括 risk-time（夜间）？ 待 prod 校准
    （现行 fall_rules_three_classes 是 risk-time × 1.2 放宽 non-risk-time；
     决定 19 后简化为单一 1.5 因子，是否覆盖 risk-time 实施时再定）
```

### 各 fall 规则放宽策略

| Fall 规则 | base | risk_factor 放宽? | normal 值 | risk 值 |
|---|---|---|---|---|
| BathroomStillFall | 10min (risk) / 12min (non-risk) | ✅ | 10/12 | 15/18 |
| **BathroomBedsideFall** | **8min** | ✅ | 8 | **12**（约 1.5×；用户原话 8→15 是上限）|
| BathroomLostFall 次强 | 7min | ✅ | 7 | 10.5 (≈11) |
| **BathroomLostFall 最强（track=0）** | **30s** | ❌ **不放宽** | 30s | 30s（最强物理信号，保持）|
| bedroom lost_fall | cell-typed | ✅ | base | × 1.5 |
| bedroom bedside_fall | 15min | ✅ | 15 | 22.5 (≈23) |
| **silent_fall** | **60s 矛盾窗口** | ❌ **不放宽** | 60s | 60s（事件驱动，不依赖时长）|

**实施时校准**：上述值为初版默认，prod 部署后按实测调整。

---

## 6.7 三级告警状态机（决定 17）

### 6.7.1 状态机表

| 级别 | 状态 | Card(popAlarm) | Pending | AlarmBell | Resolved 历史 |
|---|---|---|---|---|---|
| **Critical**（vital / fall）| active | ⭐ 最高最新者 | ✓ | ✓ | — |
| Critical | acked（人工 ack 但未恢复）| — | ✓ | ✓ | — |
| Critical | auto_resolved（自动恢复但未 ack）| — | ✓ | ✓ | — |
| Critical | acked_auto_resolved（人工 handle 后）| — | — | — | ✓ |
| Critical | manually_resolved | — | — | — | ✓ |
| **Warning** | active | ⭐ 最高最新者 | ✓ | ✓ | — |
| Warning | auto_resolved | — | — | — | ✓ |
| **Error**（device 硬件）| active | (按级别参与竞争)| ✓ | ✓ | — |
| Error | auto_resolved | — | — | — | ✓ |
| ~~Expired~~ | — | 暂不用 | — | — | — |

### 6.7.2 Card / popAlarm 选择规则

只显示 **1 个**：
1. 高级别 cover 低（critical > warning）
2. 同级别，新 cover 旧
3. 结果 = 当前 active 中 "最高级别 + 最新" 那 1 个

### 6.7.3 多 Critical 同时挂载

- popAlarm 层面：新 cover 旧（UI 简洁）
- Pending / AlarmBell 层面：全部累积保留
- 人类点击 AlarmBell 时一并查看所有挂载 critical
- 场景罕见（v2 只有 vital + fall 两类 critical）

### 6.7.4 三级核心差异

| 级别 | 人工 ack 强制? | auto_resolved 离开 Pending+AlarmBell? |
|---|---|---|
| **Critical** | **必须**（否则永远挂在 Pending）| ❌ 不离开（等待人工 handle）|
| Warning | 不强制 | ✅ 立即归 Resolved |
| Error | 不强制 | ✅ 立即归 Resolved |

**理由**：Critical（fall / vital）是 elder care 核心事件，即使自动恢复（如老人自己爬起来）也必须人工审查 + 记录到 HIPAA 审计。Warning/Error 是辅助状态，自愈即可消除。

---

## 7. Open Questions 总览（v2 修订版）

### 7.1 已对齐决定

| 编号 | 议题 | 状态 |
|---|---|---|
| 决定 0 | Bathroom 与其它房间走独立分支 ghost/fall adjudicator | ✅ 已对齐 (2026-05-18) |
| 决定 1 | 三层流水线 ghost→zone→fall | ✅ 已对齐 (2026-05-17) |
| 决定 2 | RadarVisibleRegion + BlindEnter 区分 radar/room 空间 | ✅ 已对齐 (2026-05-17) |
| 决定 3 | Unit-level Single-Person 升级 — 独居老人 bathroom 风险升级保护 | ✅ 已对齐 (2026-05-18) |
| 决定 4 | Ghost-latched 不可自动翻 Real（防 ghost 反客为主） | ✅ 已对齐 (2026-05-18) |
| 决定 5 | bathroom fall 规则 Count 不可信 → 降级 SuspectedFall 而非取消 | ✅ 已对齐 (2026-05-18) |
| 决定 6 | Mirror reflection_line 在线学习 — **P0 保留**（用户三次确认）；角色由"抑制 ghost"改为"disambiguate 哪个 track 是真人，用于 fall 归因" | ✅ 已对齐 (2026-05-18 修订) |
| **决定 7** | 不追求完美 — 出入口计数法，bathroom 不内部 ghost 判定 | ✅ 已对齐 (2026-05-18) |
| **决定 8** | Lost = SuitePerson lost（不是任意 track lost），防 ghost 滞留压制告警 | ✅ 已对齐 (2026-05-18，ChatGPT P0 反馈采纳) |
| **决定 9** | Fall 触发前置系统健康检查（雷达心跳 / 链路 / 维护窗口）| ✅ 已对齐 (2026-05-18，ChatGPT P0 反馈采纳) |
| **决定 10** | ~~跨房间一致性兜底（bedroom 活动 → bathroom 静止 = ghost）~~ → **决定 14 推翻**：共生律下 bathroom 静止 = 真人在场（fall 候选）。同一 PersonID 不在两处由 SuitePerson.AnchorRoomKind 单值字段保证（数据健全性 audit，不是 ghost 判定）| ⚠️ 已修订 (2026-05-18 决定 14 替代) |
| **决定 11** | v2 范围 = suite-specific（1 bedroom + 1 bathroom），unit-level v3 才做 | ✅ 已对齐 (2026-05-18 工程现实) |
| **决定 12** | Suite 拓扑硬编码：bathroom 入口只通向 bedroom，bedroom 是"主判定房间" | ✅ 已对齐 (2026-05-18) |
| **决定 13** | 人类行为时序：护工/家属进 suite 不会马上冲 bathroom，bedroom 过渡期 ≥ 2min 足够 census | ✅ 已对齐 (2026-05-18) |
| **决定 14** | Ghost-Real 共生律：ghost 必有真人作为反射源；ghost 是真人存在的间接证据；fall 触发不因 ghost 抑制；ghost 残影不特殊处理 | ✅ 已对齐 (2026-05-18 物理事实) |
| **决定 15** | Layout 显式 bathroom_enter 标注（cell.EnterTarget 字符串属性）+ 单入口约定 + 储衣室合并 bathroom zone；决定 12 降级为 fallback | ✅ 已对齐 (2026-05-18) |
| **决定 16** | room.kind binary 分类 — bathroom 显式标注（vue 编辑器 checkbox），其它默认 = bedroom；v2 不分 living / kitchen；v3 再扩展 | ✅ 已对齐 (2026-05-18) |
| **决定 17** | 3 级告警状态机（详 §6.7）：Critical 必须人工 ack 才离开 Pending+AlarmBell；Warning/Error auto_resolved 即归 Resolved；popAlarm 单显示（高级别 cover 低 / 同级别新 cover 旧 / 多 critical 人类一并处理）| ✅ 已对齐 (2026-05-18) |
| **决定 18** | BathroomBedsideFall 多人场景：count=1 fire Fall / count=2 但 1 ghost（共生律 disambiguate）fire Fall / 真多人（≥2 真人）不 fire（人类已在场处理）| ✅ 已对齐 (2026-05-18) |
| **决定 19** | multi-resident bedroom v2 支持（不留 v3）；不否决 fire 只放宽阈值；单一 `risk_factor = 1.5` 简化合并 risk-time + multi-resident 两维度；最强信号（LostFall 30s / silent_fall 60s）不放宽 | ✅ 已对齐 (2026-05-18) |
| **决定 20** | EnterType 4 类语义：`""` inside_enter（默认）/ `"outside"`（必标）/ `"bathroom"`（必标，v2 单 bathroom 无需 id）；inside_enter 自学习仅限**单 device-layout 坐标系内**（v2 无跨 device 关联）；outside/bathroom 必须人工标 | ✅ 已对齐 (2026-05-18) |
| **决定 21** | P1 设计选项全采纳：Q-A1 track-ID 方案 2（瞬跳检测）/ Q-A2 Verdict 加 Anchored 一等态 / Q-A3 FeedbackEvent 直接加 penalty / Q-A4 Mirror 离线 daily + 在线兜底 / Q-A10 disambiguate 失败取离入口最近 track / Q-A11 split-ghost 出生邻近性 60cm 默认 | ✅ 已对齐 (2026-05-18) |
| **决定 22** | Q-D2 Still 计时不清零（max 已累积，老人挣扎不推迟告警）+ Q-D3 5min 内同 resident 同事件去重，保留最高级别 | ✅ 已对齐 (2026-05-18) |
| **决定 23** | Q-D1 分级告警 T1/T2/T3 不做（太复杂）；单档触发即全量 critical | ✅ 已对齐 (2026-05-18) |
| **决定 24** | Public bathroom（楼层公共）**v2 支持**（standalone mode，详 §4.A.6）：只装 radar 无 sleepad；无 SuiteBedroomCensus / 无 BathroomGate 入口流量；走 §4.A 通用 bathroom ghost adjudication（决定 14 共生律 + 镜像学习 + 出生地判据）；Fall 触发主体 = bathroom 内 disambiguate 真人 track；Fall 归因仅 location 不识别 elder 身份 | ✅ 已对齐 (2026-05-18) |

### 7.2 Prerequisite — 阻塞 §4.A / §6.A 细化的事实问题（**用户/数据需回答**）

这些不是设计选项，是事实问题；**不能凭空收敛**，需要 prod 数据 / 产品方向 / 既有架构事实回答：

| 编号 | 问题 | 阻塞什么 |
|---|---|---|
| **Q-prereq-1** | bathroom fall 占比？| ✅ **用户回答**：bathroom/stair 占 elder 危险 fall **50%+**（行业认知）；P0 投入合理 |
| **Q-prereq-2** | bathroom 多模态传感器可用性？| ✅ **用户回答**：**无**（产品形态不支持，即使部分客户安装也非同一系统不可接入）；§6.5 整章删除 |
| **Q-prereq-3** | 历史 track 数据保留多久？| ✅ **用户回答**：当前训练阶段默认 **180 天**；远超 mirror 离线 pipeline 需要的 7-14 天 |
| **Q-prereq-4** | bathroom 入口都只通向 bedroom？特殊户型例外？| ✅ **用户回答**：90% elder suite 是 1 bedroom + 1 bathroom 私有；约 10% 例外可由决定 15 layout EnterTarget 补救。**另外 public bathroom（楼层公共）也存在**，走决定 24 + §4.A.6 standalone mode |
| **Q-prereq-5** | bathroom 入口在 radar 视野内的比例？| ✅ **用户回答**：**绝大多数在视野内**；规则 4 fallback 是边缘 case (≤ 5%) |
| **Q-prereq-6** | 护工陪 elder 进 bathroom 的过渡时长？2 min 够吗？| ✅ **用户回答**：≥ 2 min 成立；需要护工帮助的 elder 都是**行动不便**的，转移过程慢 |

### 7.3 设计选项 — 待拍板

| 编号 | 议题 | 倾向 |
|---|---|---|
| Q-A1 | Track-ID 是否解耦 firmware id (方案 1/2/3) | 方案 2，bathroom 分支下次要 |
| Q-A2 | Verdict 加 Anchored 一等态 | 是，bathroom 额外需要 Ghost latch |
| Q-A3 | Layer 1 消费 FeedbackEvent 方式 | A（直接加 penalty）|
| Q-B1a | bathroom 卫浴属性权威来源（cell / room.name / layout flag）| ✅ **决定 16 已拍**：`room.kind == "bathroom"` 显式标注，vue 编辑器 checkbox |
| Q-B1b | BlindEnter 存在下 RoomState 翻转规则 | 待详细设计 |
| Q-C1 | Fall dedup 优先级 | 见 §6.C |
| Q-C3 | BathroomBedsideFall 8min 时长 | 待 prod 数据校准 |
| Q-C4 | BathroomBedsideFall 90s grace 时长 | 待 prod 数据校准 |
| Q-E1 | RadarVisibleRegion / BlindEnter 是配置还是学习 | MVP 配置，二期学习建议 |
| Q-E2 | layout 编辑器 UX | 留 owlFront |
| **Q-A4** | Mirror reflection_line 学习是离线 daily 还是在线增量？冷启动期（无 reflection_line）的工作模式？ | 倾向：离线 daily + 在线增量兜底；冷启动期纯靠规则 1+2+3 |
| ~~Q-A5~~ | ~~Ghost latch 撤销路径~~ | 撤回（决定 14 删除 ghost latch 机制）|
| **Q-A10 (新)** | Mirror 学习的失败兜底 — disambiguate 失败时 fall 归因用哪个 track？ | 倾向：取"离入口轨迹最近"的（规则 2 现有逻辑）作 fall 归因 fallback |
| **Q-A11 (新)** | 规则 5 split-ghost 出生邻近性阈值（50/60/80 cm） | 倾向 60cm 默认，prod 数据校准 |
| **Q-A12 (新)** | 运动同步性持续验证（30-60s 窗口）是否启用 | 倾向 Phase 2，需 1Hz 下分布实测 |
| **Q-L1 (新)** | cell.EnterTarget 实现方案（决定 15 选项 A 已拍）| ✅ 复用 AreaEnter + EnterTarget 字符串属性 |
| **Q-L2 (新)** | 自动学习 EnterTarget 候选 — 写回 layout vs 写 ai.log 给运维审核 | ✅ 写 log 审核（已拍 B）|
| **Q-L3 (新)** | Bathroom 单入口约定（决定 15 已拍）| ✅ 单入口；多入口 layout 错误 |
| **Q-L4 (新)** | Jack-and-Jill bathroom v2 支持？| ✅ 不支持，部署时拆为两个 zone；v3 多 bathroom 再做 |
| **Q-L5 (新)** | Storage 直通 bedroom + 直通 bathroom 罕见户型 | ✅ layout 人工选归属（划入 bedroom 或 bathroom） |
| **Q-L6 (新)** | 储衣室是否新增 AreaCloset cell 类型 | ✅ 不新增，复用 AreaUnknown |
| **Q-L7 (新)** | room.kind 字段类型（决定 16）| ✅ `string`（"bathroom" / ""），不用 bool（v3 扩展 living/kitchen/... 留空间）|
| **Q-L8 (新)** | vue 编辑器 bathroom 标注 UX | checkbox "Is bathroom"（默认未勾），有 Bed cell 时编辑器提示"建议作为 bedroom 默认（即不勾选）" |
| **Q-A6 (新)** | SuitePerson anchor 升格阈值（resident 5min / visitor 2min + 5-10 cells）是否合适 | 待 prod 数据校准 |
| ~~Q-A7~~ | ~~Bathroom 入口必须可见~~ | 改入 Q-prereq-5 |
| **Q-A8** | 决定 11 v2/v3 边界 — v2 是否真的只支持 suite (1 bed + 1 bath)？ | ✅ **决定 19 修订**：v2 也支持 multi-resident bedroom（不否决 fire，按 `risk_factor=1.5` 放宽阈值）。多 bedroom 户型仍留 v3 |
| **Q-A9** | v2 多 visitor 上限？ | ✅ **用户回答**：允许；visitor 影响 unit 单人/多人判断（单人时风险升级，决定 3）；上限默认 4（resident 1 + visitor 3），prod 校准 |
| **Q-D1 (新)** | 分级告警 T1/T2/T3（5min 软提醒 / 10min 中等 / 15min 全量）| 产品方向题，影响告警通道设计；倾向加，但留产品决策 |
| **Q-D2 (新)** | Still 计时撤销规则 — 微小活动是清零还是暂停 | 老人摔倒后挣扎不应推迟告警；倾向 max(已累积, 0) 不清零 |
| **Q-D3 (新)** | 告警去重 — still + lost 同事件触发是否合并 | 倾向去重（同 resident 5min 内只发一次最高级别）|
| **Q-D4 (新)** | 告警状态恢复 — 老人被扶起后是否撤销告警 | 倾向"标记 resolved 不撤销"（已发出的告警保留审计）|

---

## 8. 不在本文档讨论范围

- iot:event_log / monitor_stream 持久化 — [iot_v2_cutover_done](../../.claude/projects/-home-wisefido-owl/memory/iot_v2_cutover_done.md)
- platform agent IPv6 identity / envelope.producer = INET — [platform_agent_addressing](platform_agent_addressing.md)
- cardagg 退订 event stream / trust producer="wisefido-sensor" — [cardagg_sensor_split.md §5](cardagg_sensor_split.md)
- zonealarm 4 件套（Stay/LeftBed30min/NightAbsence/BedNightAbsence）— 已落 ✅
- sensor stdlib log → zap 迁移 — 独立 PR

---

## 9. 实施总结

决策层（§0-§8）+ 实施层（§10-§12）已最终。
v2 sensor 改造采用 **fresh checkout + selective copy** 策略（用户决定）：
- 老目录 `wisefido-sensor` → `wisefido-sensor-v1`
- 新目录 `wisefido-sensor` 全新建
- systemd `owlback.sensor` 不改名（指向新 binary）
- 直切 cutover（不双跑）
- 总改动量 ~3,200 行，3-4 周开发 + 2-3 周 prod 验证

---

## 10. 终态设计 — 三层契约形式化

### 10.1 数据结构（核心）

#### 10.1.1 TrackStatus（Layer 1 → Layer 2 契约）

```go
package sensor

type TrackVerdict uint8

const (
    VerdictPending  TrackVerdict = 0
    VerdictReal     TrackVerdict = 1
    VerdictAnchored TrackVerdict = 2  // 决定 21 (Q-A2)：与 Real 平级，但 LongSurvival/StartupGrace 锚定后不可翻 Ghost
    VerdictGhost    TrackVerdict = 3
)

type TrackStatus struct {
    TrackID    int
    DeviceID   string  // radar IPv6 /128
    RoomID     string  // room spatial_prefix

    Verdict       TrackVerdict
    GhostPenalty  int     // 0-100；≥80 → Ghost
    
    X, Y, Z       int     // 当前位置 (cm)
    Pose          int     // 1=Stand / 2=Sit / 3=Lie / 4=Walk / ...
    StillSec      int     // 当前静止时长
    CellAreaType  AreaType
    EnterTarget   string  // cell.EnterTarget（"" / "outside" / "bathroom"）

    InBedZoneID       string  // "" = 不在床区
    InRoomZoneID      string  // 一般 == RoomID
    InBathroomZoneID  string  // "" = 不在 bathroom

    PersonID        string  // 关联 SuitePerson.PersonID；"" = 未关联
    PersonRole      string  // "resident" / "visitor" / ""
    
    UpdatedAt       int64
}
```

#### 10.1.2 SuiteBedroomCensus（Layer 1 内部）

```go
type SuitePerson struct {
    PersonID         string   // resident.id 或 "visitor_<ts>"
    Role             string   // "resident" / "visitor"
    AnchorTrackID    int      // 当前 anchor 的 radar track
    AnchorRoomKind   string   // "bedroom" / "bathroom"
    AnchorSince      int64
    LastActiveMs     int64    // 最近"非静止"帧；决定 8 用作 lost 判据
    LastBathroomEntryMs int64
    SleepadAnchored  bool     // 床压双源锚定（仅 resident 可能为 true）
}

type SuiteBedroomCensus struct {
    SuiteID        string                     // = bedroom RoomID
    Persons        map[string]*SuitePerson    // key=PersonID
    BathroomCount  int                        // bathroom 内当前人数（入口流量累积）
    UpdatedAt      int64
    
    // Public bathroom standalone mode (决定 24 / §4.A.6) 时：
    //   SuiteID = bathroom 自己的 RoomID
    //   Persons = nil（公共 bathroom 不识别 person）
    //   BathroomCount = bathroom 内 disambiguate 出的真人 track 数（非入口流量）
    IsPublicBathroom bool
}
```

#### 10.1.3 ZoneEvent / FeedbackEvent（Layer 2 输出 + Layer 2→1 反馈）

```go
// 既有 zoneengine.ZoneEvent 保留（决定 1+11 复用），新增字段：
type ZoneEvent struct {
    // ... 既有字段 (CardID/ZoneType/ZoneID/Transition/NewState/PrevState/Trace/Ts)
    PersonIDs []string  // 触发本翻转的相关 SuitePerson（v2 新增）
}

// 既有 zoneengine.FeedbackEvent 保留（决定 1 反馈通道复用，决定 25 直接加 penalty）
type FeedbackEvent struct {
    // ... 既有字段 (CardID/ZoneType/ZoneID/Reason/Affected/Penalty/DurationMin/Ts)
    TargetTrackID  int  // 决定 25：明确指向哪个 track 加 penalty
}
```

#### 10.1.4 Cell 扩展（决定 15 + 决定 20）

```go
type Cell struct {
    // ... 既有字段 (Belief/Confidence/Source/FakeAlarmCount/DoorEventCount/...)
    
    EnterTarget   string  // 决定 15：""=inside / "outside" / "bathroom"
                          // 后端约定值，vue 编辑器 dropdown 三选项映射
    
    InsideEnterLearned    bool   // 决定 20 inside_enter 自学习标志
    InsideEnterEvidenceN  int    // 一周累积"track 失锁/重生"次数
}

// Cell 学习算法扩展：
//   - 复活 MarkDoorEvent (EnterRoom/ExitRoom event 来时调)
//   - 新增 MarkInsideEnterCandidate (track 失锁 + 3s 内同位置重生 时调)
//   - 解锁 AreaEnter 自动学习路径（仅对 inside_enter，不影响 layout 锁定的物理门）
```

### 10.2 流接口

#### 10.2.1 输入流（sensor 订阅）

| Stream | 持久化 | 用途 |
|---|---|---|
| `iot:event:stream` | iot → event_log 90d | radar EnterRoom/ExitRoom/InBed/LeftBed/Track + Fall verifier 输入 |
| `iot:alarm:stream` | iot → alarm_events 7y | sleepad native alarm（决定 7 sleepad LeftBed 输入 silent_fall） |
| `iot:monitor:stream` | iot → monitor_stream 90d | track frame raw 输入 |

#### 10.2.2 输出流（sensor 发布）

| Stream | 持久化 | 用途 |
|---|---|---|
| `iot:alarm:stream` (回流) | iot → alarm_events 7y | sensor 派生 alarm（fall / Stay / NightAbsence），producer=fd00:0:fff1::1 |
| **`sensor:track:status:stream`** (新) | 不入库（决定 1）| Layer 1 → Layer 2 TrackStatus 投影；瞬态投影流 |
| **`sensor:zone:state:stream`** (新) | 不入库 | Layer 2 → Layer 3 ZoneEvent；瞬态翻转沿 |
| **`sensor:feedback:stream`** (新) | 不入库 | Layer 2 → Layer 1 FeedbackEvent；专用反馈通道 |
| **`sensor:ai:log:stream`** (新或复用) | iot → event_log 90d? | inside_enter 学习推荐 / mirror reflection_line 推荐 / audit 日志 |

#### 10.2.3 持久化（PG）

| 表 | 改动 | 用途 |
|---|---|---|
| `alarm_events` | + `alarm_status='acked_auto_resolved'`（决定 17，2026-05-18 migration 已写）| Critical 自动恢复+人工 handle 终态 |
| `cells_v2`(room_cell_state) | + `EnterTarget` / `InsideEnterLearned` / `InsideEnterEvidenceN` / `reflection_lines: []` (JSONB) | 决定 15+20+6 持久化（复用 payload JSONB 不加新表） |
| `rooms_v2` | + `room.kind = "bathroom"/""`（决定 16）| layout 显式标注 |
| `units_v2`（无改动）| 复用 `units.residents.count_active` 派生 risk_factor | 决定 19 |

### 10.3 关键算法决策图

#### 10.3.1 Layer 1 ghost 分支选择

```
入口：每帧 track frame / event
  ├── room.kind == "bathroom"?
  │   ├── 是 → BathroomGhostAdjudicator (§4.A)
  │   │   ├── room 属于某 suite？(has 兄弟 bedroom)
  │   │   │   ├── 是 → Suite-bathroom 模式（§4.A.1-4.A.5）
  │   │   │   └── 否 → Public-bathroom standalone mode（§4.A.6 决定 24）
  │   └── 否 → GeneralGhostAdjudicator (§4.B, default = bedroom)
```

#### 10.3.2 Layer 3 fall 分支选择

```
入口：ZoneEvent 翻转沿 / SuitePerson 状态变化
  ├── room.kind == "bathroom"?
  │   ├── 是 → BathroomFallRules (§6.A)
  │   │   ├── BathroomStillFall (cell ∈ {Toilet, Shower})
  │   │   ├── BathroomBedsideFall (其它 cell, 决定 18 多人不 fire)
  │   │   └── BathroomLostFall 两档（决定 14 共生律）
  │   └── 否 → GeneralFallRules (§6.B)
  │       ├── silent_fall (sleepad LeftBed + BedState 矛盾)
  │       ├── bedside_fall (BedState→Vacant + 床边静止)
  │       └── lost_fall (SuitePerson 在 bedroom + 静默)
  │
  └── 共用前置：
      ├── 系统健康检查（决定 9 / §6.A.0）
      ├── Fall 主体 = SuitePerson（决定 8，非任意 track）
      └── 阈值 = base × risk_factor (决定 19，risk=1.0 / multi-resident=1.5)
```

### 10.4 PG 表新增字段一次性 migration

```sql
-- 2026-05-18_sensor_v2_cell_extensions.sql

-- Cell 扩展（room_cell_state.payload JSONB 内字段，schema 由 sensor 维护，PG 透明）
-- 不改 PG 列，仅文档化 JSONB schema 新增字段：
--   payload.cells[].enter_target: string
--   payload.cells[].inside_enter_learned: bool
--   payload.cells[].inside_enter_evidence_n: int
--   payload.reflection_lines: [{line_xy, confidence, last_confirm_ms, evidence_n}]

-- Room kind binary (决定 16) — **勘误 2026-05-17 PR-4 ready 前修订**
-- ~~ALTER TABLE rooms_v2 ADD COLUMN IF NOT EXISTS kind VARCHAR(20) DEFAULT '';~~
--
-- **复用既有 rooms.room_type VARCHAR(50) 列**。该列在 owlRD/dbv2/14_rooms.sql 已存在，
-- 值域 (bedroom/bathroom/livingroom/kitchen/dining/lobby/corridor/...) 完全 cover 决定 16 binary 分类。
-- 新加 kind 列会形成"两个写入口语义重叠" → 违反 CLAUDE.md #1.3 单源真相。
--
-- Sensor 内部仍用 RoomKind 字段名（与决定 16 doc 措辞一致），bootstrap SQL 做适配：
--   SELECT COALESCE(r.room_type, '') AS room_kind FROM rooms r ...
--
-- 跨服务（cardagg/data/owlFront）读这列时保持 PG 命名 room_type；
-- 仅 sensor 内部 domain layer 用 RoomKind 反映"v2 binary 收缩语义"。

-- alarm_status 新值（决定 17，前面已写 2026-05-18 migration）
-- 已在 2026-05-18_alarm_events_add_acked_auto_resolved.sql

-- 不需要 SuiteBedroomCensus 新表（Redis 持久化）
```

### 10.5 关键不变量（v2 必须维护）

| 不变量 | 守护机制 |
|---|---|
| Ghost 必有真人作为反射源（决定 14）| §4.A.3 规则 3 共生律自校正 |
| BathroomCount 来自入口流量，不来自内部 track 数 | §4.A.2 BathroomGate（suite mode）|
| 同一 PersonID 不在两处 | SuitePerson.AnchorRoomKind 单值字段 |
| Fall 触发主体 = SuitePerson（不是任意 track）| 决定 8 + §6.A/§6.B 所有 fall 规则 |
| Fall 触发前置系统健康检查 | 决定 9 + §6.A.0 |
| Critical alarm 必须人工 ack 才离开 Pending+AlarmBell | 决定 17 状态机 |
| Bathroom 与其它房间 ghost adjudicator 不共享状态 | 决定 0 双分支 |
| Layer 间单向数据流（无循环依赖）| FeedbackEvent 走专用 stream |
| sensor 永不写 alarm_events 表 | 通过 alarm_back_channel 让 cardagg 落库 |

---

## 11. PR 切分清单（16 PR / 5 Phase）

### Phase A — 数据结构 + 流契约（PR-1/2/3，~600 行）

#### PR-1：Cell 扩展 + Enter 自学习骨架复活

| Sub-task | 改动 | 行数 |
|---|---|---|
| 1a | `Cell.EnterTarget` 字段（决定 15）+ JSONB schema | ~50 |
| 1b | 复活 `MarkDoorEvent` — runEventLoop EnterRoom/ExitRoom callback 调 cell 学习 | ~30 |
| 1c | 新增 `MarkInsideEnterCandidate` — track 失锁 + 3s 内同位置重生 → 累计 InsideEnterEvidenceN | ~80 |
| 1d | 解锁 AreaEnter 自学习路径（区分 inside_enter learning vs layout-locked 物理门）| ~40 |
| 1e | `room.kind` binary 字段（决定 16）+ vue 编辑器 checkbox + AreaEnter target dropdown | ~80 owlBack + ~80 owlFront |
| 1f | Audit script：扫所有 unit 的 layout，检查 bathroom EnterTarget 在 RadarVisibleRegion 内 | ~60 |
| **合计** | | **~420** |

依赖：无
验收：现有 sensor 部署不破坏；新字段写入但 v2 逻辑还未消费。

#### PR-2：SuiteBedroomCensus + SuitePerson 数据结构

| Sub-task | 改动 | 行数 |
|---|---|---|
| 2a | `SuitePerson` / `SuiteBedroomCensus` Go struct + Redis 持久化 | ~120 |
| 2b | bedroom 内 resident 升格逻辑（sleepad 双源锚定 / 5min anchor）| ~80 |
| 2c | bedroom 内 visitor 升格逻辑（2min anchor + 决定 13 时序假设）| ~50 |

依赖：PR-1
验收：bedroom census 正确，bathroom census 字段为空。

#### PR-3：TrackStatus 输出 + sensor:track:status:stream

| Sub-task | 改动 | 行数 |
|---|---|---|
| 3a | `TrackStatus` Go struct + Redis stream publisher | ~80 |
| 3b | roomengine 每帧派生 TrackStatus（含 PersonID 关联）| ~100 |
| 3c | dev UI（playback 工具）订阅 stream 显示 | ~50 |

依赖：PR-1, PR-2
验收：dev 工具能看到完整 TrackStatus 流。

---

### Phase B — Layer 1 Ghost 双分支（PR-4/5/6/7/8，~1300 行）

#### PR-4：room.kind binary 分支选择入口

| Sub-task | 改动 | 行数 |
|---|---|---|
| 4a | engine.go 增加 ghost adjudicator 分支选择 | ~60 |
| 4b | BathroomGhostAdjudicator 接口骨架（空实现 fallback 到 General）| ~80 |

依赖：PR-1, PR-3
验收：room.kind="bathroom" 走分支但行为 = General（PR-5 之后才真切换）。

#### PR-5：BathroomGate 入口流量子模块

| Sub-task | 改动 | 行数 |
|---|---|---|
| 5a | BathroomGate Go struct + 监测 cell.EnterTarget="bathroom" 翻越事件 | ~120 |
| 5b | bedroom↔bathroom 双向计数 + 兜底 -1 逻辑 | ~60 |
| 5c | SuitePerson.AnchorRoomKind 状态机更新（bedroom↔bathroom 翻转）| ~40 |

依赖：PR-4
验收：track 跨入口时 BathroomCount 正确累积。

#### PR-6：规则 1/2/3/5 — Bathroom 内部 ghost 判定

| Sub-task | 改动 | 行数 |
|---|---|---|
| 6a | 规则 1 出生地判据（远离 EnterTarget cell 出生 → ghost）| ~50 |
| 6b | 规则 2 占用一致性（track 数 > BathroomCount → 多余 ghost）| ~80 |
| 6c | 规则 3 共生律自校正（决定 14）| ~70 |
| 6d | 规则 5 split-ghost 出生邻近性（< 60cm 紧贴）| ~80 |
| 6e | 规则 4 fallback（入口在 radar 盲区，边缘 case）| ~40 |

依赖：PR-5
验收：bathroom 内 ghost 识别率 prod baseline 建立。

#### PR-7：Public Bathroom Standalone Mode（决定 24 / §4.A.6）

| Sub-task | 改动 | 行数 |
|---|---|---|
| 7a | 检测 room 是否属于 suite (has 兄弟 bedroom) | ~40 |
| 7b | Public mode：跳过 SuiteBedroomCensus + BathroomGate | ~60 |
| 7c | Fall 触发主体改 "bathroom 内 disambiguate 真人 track"（非 SuitePerson）| ~80 |
| 7d | Fall 归因 location only（不识别身份）| ~30 |

依赖：PR-6
验收：public bathroom 部署能正常工作，fall 归因到 location。

#### PR-8：镜像学习离线 daily pipeline + 在线增量兜底

| Sub-task | 改动 | 行数 |
|---|---|---|
| 8a | 离线 daily pipeline（cron 03:00 UTC，扫过去 14 天 track frames）| ~200 |
| 8b | reflection_line 拟合（SVD + 残差验证 + 距离差稳定性）| ~150 |
| 8c | 写回 cells_v2.payload.reflection_lines (复用现有表，决定 6 持久化方案 A) | ~50 |
| 8d | 实时使用：disambiguate 镜像 ghost（决定 14 + 决定 21 Q-A4）| ~80 |
| 8e | 冷启动期：在线增量拟合作为兜底 | ~60 |

依赖：PR-6 (镜像学习需要 BathroomCount 基线)
验收：bathroom case A (D5F7) prod 实测漏检率 < 10%。

---

### Phase C — Layer 3 Fall 规则重写（PR-9/10/11/12，~700 行）

#### PR-9：删 v1 4 fall 私有计数器 + BedSession 多数

| Sub-task | 改动 | 行数 |
|---|---|---|
| 9a | 删 `bedPersonCount` / `lastLeftBedAt` / `lastNumberPeopleZeroMs` / `bathroomRealCount` 全部维护代码 | -200 |
| 9b | 保留 `BedSession.LeftBedHadHRRR` / `LeftBedMaxPeople` latch 字段（证据快照，非状态）| -100 不删 |
| 9c | grep `已退役|deprecated|TODO.*迁移` 自查（CLAUDE.md #1.6）| 0 |

依赖：v2 Layer 2 已就绪（PR-5 之后）
验收：sensor build 通过且无残留私有计数器。

#### PR-10：BathroomStillFall / BathroomBedsideFall / BathroomLostFall 双档

| Sub-task | 改动 | 行数 |
|---|---|---|
| 10a | BathroomStillFall (cell ∈ {Toilet, Shower}, Stand 静止 ≥ 10/12 min)| ~80 |
| 10b | BathroomBedsideFall (90s grace + 任意位置静止 ≥ 8 min, 决定 18 真多人不 fire)| ~100 |
| 10c | BathroomLostFall 最强档 (track==0 ≥ 30s 立即 fire) | ~70 |
| 10d | BathroomLostFall 次强档 (SuitePerson.LastActiveMs > 7 min) | ~60 |

依赖：PR-2 (SuitePerson), PR-6 (bathroom adjudicator)
验收：3 类 bathroom fall 在 dev 实测全部正确触发。

#### PR-11：bedroom silent_fall / bedside_fall / lost_fall 重写（SuitePerson）

| Sub-task | 改动 | 行数 |
|---|---|---|
| 11a | silent_fall (BedState 矛盾 + sleepad LeftBed, 不依赖 BedSession 多数逻辑) | ~70 |
| 11b | bedside_fall (BedState→Vacant + 床边静止 15 min, SuitePerson 主体) | ~80 |
| 11c | lost_fall (SuitePerson 在 bedroom + 静默 cell-typed 时长) | ~80 |

依赖：PR-9, PR-10
验收：bedroom 3 类 fall 替换原 v1 逻辑。

#### PR-12：Fall dedup 优先级 + 阈值矩阵 risk_factor=1.5

| Sub-task | 改动 | 行数 |
|---|---|---|
| 12a | Fall dedup（silent > bedside > bathroom_* > still > lost, 60s 内同房只 fire 最高级）| ~50 |
| 12b | 阈值矩阵 risk_factor 应用（决定 19, unit.residents.count_active ≥ 2 时 ×1.5）| ~40 |
| 12c | 最强信号不放宽（LostFall 30s / silent_fall 60s 矛盾窗口）| ~20 |

依赖：PR-10, PR-11
验收：阈值矩阵 dev 实测各 cell 各时段正确。

---

### Phase D — 告警状态机 + 系统健康（PR-13/14，~250 行）

#### PR-13：三级告警状态机（决定 17）

| Sub-task | 改动 | 行数 |
|---|---|---|
| 13a | alarm_status='acked_auto_resolved' 终态写入路径（sensor 端 emit + cardagg 落库）| ~80 |
| 13b | Stay alarm auto_resolve（BathRoomState→Vacant 时清）| ~40 |
| 13c | popAlarm 选择规则（高级别 cover 低，新 cover 旧）— 给 cardagg / owlFront 提供 selector API | ~50 |

依赖：alarm_status migration（已写）
验收：Critical 自动恢复后仍在 Pending+AlarmBell；Warning/Error auto_resolved 立即消失。

#### PR-14：系统健康检查前置（决定 9）

| Sub-task | 改动 | 行数 |
|---|---|---|
| 14a | `isSuiteHealthy()` 实现（雷达心跳 / sleepad 心跳 / 维护窗口 / SuiteBedroomCensus 新鲜度）| ~80 |
| 14b | 所有 fall 规则触发前调 isSuiteHealthy；不健康 → emit "system_status_abnormal" 通知 | ~30 |

依赖：PR-10, PR-11, PR-12
验收：dev 模拟雷达断网 → 无 fall 误报，发出系统状态异常告警。

---

### Phase E — cardagg 配合（PR-15/16，~250 行）

#### PR-15：cardagg 退订 event stream（cardagg_sensor_split 收尾）

| Sub-task | 改动 | 行数 |
|---|---|---|
| 15a | 删 cardagg 对 iot:event:stream 的订阅（已在 cardagg_sensor_split.md 拍）| -80 |
| 15b | 删 cardagg.alarm_handler 内 `case alarm.Fall, alarm.SittingOnGround` 分支 | -50 |
| 15c | 简化 enablement gate：trust producer="wisefido-sensor" 一律放过 | ~30 |

依赖：v2 sensor 完整切完（PR-1 ~ PR-14）
验收：cardagg 不再消费 radar firmware Fall，只接 sensor 派生 alarm。

#### PR-16：alarm_status 完整生命周期 + UI 三层（Card/Pending/AlarmBell）

| Sub-task | 改动 | 行数 |
|---|---|---|
| 16a | wisefido-data 服务 service 层 acked_auto_resolved 完整支持（之前 grep 出 3 处硬编码）| ~80 |
| 16b | owlFront UI 三层渲染（Card 选 1 个最高最新 / Pending 列 / AlarmBell 列）| ~120 (owlFront) |
| 16c | 决定 22 still 计时不清零 + 5min 同 resident 去重 | ~40 |

依赖：PR-13
验收：UI 完整呈现决定 17 状态机，老人挣扎不推迟告警。

---

### 时间线估算

| Phase | PR | 累计行数 | 累计周期 |
|---|---|---|---|
| A | 1, 2, 3 | ~600 | 1 周 |
| B | 4, 5, 6, 7, 8 | +1,300 | +1.5 周 |
| C | 9, 10, 11, 12 | +700 | +1 周 |
| D | 13, 14 | +250 | +0.5 周 |
| E | 15, 16 | +250 | +0.5 周 |
| **合计** | **16 PR** | **~3,100 行** | **~4.5 周开发** + 2-3 周 prod 验证 |

---

## 12. 迁移路径

### 12.1 目录改名 + cutover 步骤

```bash
# Step 1：v1 目录改名（保留作 ref + selective copy 源）
cd /home/wisefido/owl/owlBack
git mv wisefido-sensor wisefido-sensor-v1
git commit -m "rename: wisefido-sensor → wisefido-sensor-v1 (v2 fresh checkout 准备)"

# Step 2：新建 v2 目录骨架
mkdir wisefido-sensor
cd wisefido-sensor
# 创建 go.mod / cmd/wisefido-sensor / internal/ 骨架
# go module path = "wisefido-sensor"（与改名前一致，systemd unit 不动）

# Step 3：按 PR-1 ~ PR-16 顺序开发 + selective copy 自 wisefido-sensor-v1

# Step 4：dev 全量测试 + audit script (PR-1f) 跑过所有部署
# 4.1 stop owlback.sensor (v1)
# 4.2 跑 PG migration:
#     - 2026-05-18_alarm_events_add_acked_auto_resolved.sql
#     - rooms_v2 加 kind 字段（PR-1e）
# 4.3 deploy v2 binary
# 4.4 start owlback.sensor (v2)
# 4.5 监控 sensor:track:status:stream / sensor:zone:state:stream / ai.log

# Step 5：prod 验证 1 月

# Step 6：删除 wisefido-sensor-v1 目录
git rm -r wisefido-sensor-v1
git commit -m "cleanup: remove wisefido-sensor-v1 after 1mo prod validation"
```

### 12.2 Copy 清单（v1 → v2）

| v1 文件 / 模块 | v2 处理 |
|---|---|
| **直接 copy（几乎不动）** | |
| `roomengine/kalman.go` | copy |
| `roomengine/grid.go`（除 MarkDoorEvent 复活逻辑外）| copy + 微改 |
| `roomengine/cell.go` 基础部分（Belief / Decay / FakeAlarmCount / DoorEventCount / 持久化）| copy + 加新字段 |
| `roomengine/cell_learning.go` Auto-Deny BFS / 学习算法 | copy |
| `roomengine/layout_parser.go` | copy + 加 EnterTarget 解析 |
| `roomengine/grid_render.go` / `room_svg.go` | copy + 渲染新字段 |
| `roomengine/math_util.go` | copy |
| `roomengine/fall_rules_param.go` | copy + 加 risk_factor / split-ghost 阈值 |
| `roomengine/persist*.go` | copy + 持久化 reflection_lines |
| `zoneengine/` 全部 | copy（决定 1 复用 zoneengine 骨架）|
| `zonealarm/` 全部 | copy（决定 0 + Stay/LeftBed30min/NightAbsence/BedNightAbsence 已落）|
| `evaluator/` | copy |
| `consumer/` | copy + 调整 alarm_back_channel 输出新 stream |
| `config/`, `repository/`, `models/` | copy |
| **重写（v1 schema 引用 ≥3 处）** | |
| `roomengine/track_manager.go` 4 fall 私有计数器 / BedSession 多数 | **全删** |
| `roomengine/track_manager.go` 4 fall 触发逻辑 | **重写**（消费 SuitePerson + ZoneState）|
| `roomengine/engine.go` 主消费循环 | **重写**（加 room.kind 分支 + TrackStatus publisher）|
| `roomengine/fall_verify.go` | **修改**（与新 ghost adjudicator 集成）|
| `roomengine/track.go` Verdict 枚举 | **修改**（加 VerdictAnchored）|
| **新建** | |
| `roomengine/suite_census.go` | SuiteBedroomCensus + SuitePerson 实现 |
| `roomengine/bathroom_gate.go` | BathroomGate 入口流量 |
| `roomengine/public_bathroom.go` | Public Bathroom standalone mode |
| `roomengine/mirror_learning_offline.go` | 离线 daily pipeline |
| `roomengine/mirror_learning_online.go` | 在线增量兜底 |
| `roomengine/track_status_publisher.go` | sensor:track:status:stream |
| `roomengine/system_health.go` | 决定 9 系统健康检查 |
| `cmd/audit-bathroom-enter/main.go` | 一次性 audit script (PR-1f) |

### 12.3 v2/v3 兼容性保留

| 保留项 | v3 演进 |
|---|---|
| `cell.EnterTarget` string 类型 | v3 多 bathroom 时扩展为 `"bathroom_<id>"` 或 human-readable name |
| `room.kind` string 类型 | v3 加 `"living"` / `"kitchen"` / `"hallway"` 枚举 |
| `BathroomGate` 抽象 | v3 unit-layout 多 enter 时复用，加多入口监测 |
| `SuiteBedroomCensus` 接口 | v3 改为 `UnitActiveResidentSet`（跨多 room） |
| Mirror 学习 | v3 跨 device-layout 后可学 outside_enter / bathroom_enter 推荐位置 |

### 12.4 Rollback 策略

如果 prod 验证发现重大 bug：
1. stop owlback.sensor (v2)
2. PG migration 已经落地的 schema 不回滚（acked_auto_resolved + rooms.kind 都向后兼容）
3. start wisefido-sensor-v1 binary（仍在）
4. 重启 cardagg 重新订阅 event stream（PR-15 之前还在）

**Rollback 窗口**：PR-15（cardagg 退订 event stream）之前 = rollback 安全；PR-15 之后 = 需要 cardagg 也回滚。建议 PR-15 在 v2 prod 验证 2 周后再 merge。

---
#d2987a   #dcb894) 