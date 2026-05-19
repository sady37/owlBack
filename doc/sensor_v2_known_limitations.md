# sensor_v2 Known Limitations & PR-X Backlog

PR-1 ~ PR-11.1 落地后的已知简化点。每条 limitation 包含 **why**（设计取舍）、**how to apply**（实施路径）、**trigger**（启动 PR-X 的数据/事件阈值）。

不是 TODO 列表 — 是**用数据触发工程决策**的 trigger-based backlog：到达阈值才动手，否则当前简化已够。

设计原则：CLAUDE.md 规则 #1.2 "做不了就别动 schema" 的延伸 — 真做 PR-X 之前不留 stub / no-op 桩。

---

## L1. BedroomLostFall cell-typed 阈值未实现

**模块**：[bedroom_fall.go](../wisefido-sensor/internal/roomengine/bedroom_fall.go) `evaluateLostFall`

**当前实施**：单一 10min 阈值 + sleepad active-bed-session gating

**spec §6.B.2 原意**：`5min walkway / 60min bed / ...` cell-area 分层

**Why**：sleepad gating 已覆盖最危险的 bed 主导场景（睡觉每晚 22:10 误报）。AreaSit (沙发久坐 nap) 偏早 fire / AreaActive (走廊伫立) 偏晚 fire 是次要场景。

**How to apply (when triggered)**:
- 通过 SuitePerson.AnchorTrackID → 当前帧 bases → CellAreaType
- 阈值表：`{AreaBed: 60min, AreaSit: 30min, AreaActive: 5min, default: 10min}`
- **注意 grid 量化误差**：10cm 精度 + radar 位置噪声让 cell 判定不可靠 — 人在床上 track 可能落到隔壁 cell。阈值表不能仅靠 CellAreaType，要复合 sleepad gating（PR-11.1 已就位）

**Trigger PR-X (cell-typed)**：
- 沙发 nap 误报 ≥1 次/晚 (per resident) → 必做（影响 elder UX）
- 走廊伫立漏报（实际跌倒晚 5min 发现）≥1 次/月 → 必做（影响救援）
- 否则单一 10min + sleepad gating 已足够

---

## L2. BedroomLostFall sleepad staleness guard 未实现

**模块**：[bedroom_fall.go](../wisefido-sensor/internal/roomengine/bedroom_fall.go) `hasActiveBedSession`

**当前实施**：仅查 `InBedSinceMs > 0 && LeftBedAtMs == 0`，不查时戳新鲜度

**风险**：sleepad 失联（device offline）时 BedSession 永远停在"in-bed"状态 → lost_fall 永久被错误抑制

**Why**：sleepad 可靠性问题需协调 wisefido-sleepace 设备健康信号，跨服务改动。PR-11.1 范围内单一服务修复优先，跨服务留 PR-X。

**How to apply (when triggered)**:
- 加新鲜度阈值（如 InBedSinceMs 超过 8h 未刷新视作 stale）
- 协调 wisefido-sleepace 提供 device-health 信号（offline event / heartbeat timestamp）
- 在 §6.A.0 "系统健康检查"扩展时一并处理（决定 9）

**Trigger PR-X (staleness guard)**：
- sleepad 失联场景误抑制 ≥1 次/月 → 必做
- 否则当前 gating 足够

**Lock-in test**：[TestBedroomFall_LostFall_StaleBedSession_SuppressionLockedIn](../wisefido-sensor/internal/roomengine/bedroom_fall_test.go) 锁定"stale BedSession 当前会持续抑制"行为；PR-X 实施时此测试需更新

---

## L3. outRoom / enterRoom 事件不可靠（layout 约定的内在产物）

**模块**：全 sensor fall 路径

**根本原因**（决定 2026-05-17 用户确认）：
- Layout 开发用 radar 测量真实 room 物理边界
- **radar 盲区 / 信号丢失区被人工标记为 `enter` cells**（用作"track 失锁等价于离开"的判定基础）
- 走出 radar 识别区（**包括真出门 + 走入盲区**）→ firmware 触发 enter/out room event

**直接后果**：outRoom 事件无法物理区分以下两种情况

| 情况 | 物理事实 | outRoom event 表现 |
|---|---|---|
| 真出门 | 人走出 room | 触发 outRoom ✓ |
| 走入盲区 | 人仍在 room 内（信号被遮挡）| 触发 outRoom ✗ 误报"离开"|

**Why**：v2 单 radar 部署下没有物理方法区分。需要 v3 cross-source consistency（多 radar 互相验证 + sleepad 床压补判定）。

**How to apply**：
- §6.A 各 fall 规则的 cancel 条件不能用 outRoom event（盲区遮挡会假取消 Critical fall）
- cancel 条件只接受**正向证据**：BathroomGate 反向流量 / SuitePerson.AnchorRoomKind 翻 bedroom / 主动重新观测到 track
- zonealarm.Supervisor (Warning 兜底) 仍可走 outRoom cancel（Warning 容错更高，假取消代价小）
- v3 unit-layout 后引入 cross-source consistency 才能让 outRoom 有"真"可言

**Trigger v3**：unit-layout 演进 + 多 radar 部署 case → 系统重构机会，不是 PR-X 范围

---

## L4. Bootstrap-wiring 不变量 (v2 main.go 待写)

**模块**：v2 `cmd/wisefido-sensor/main.go`（PR-bootstrap，未写）

**不变量**（详 sensor_v2.md §6.A 设计原则）：
```
v2 main.go bootstrap 必须同时满足：
  1. wiring.NewSubsystem(...) 启动 zoneengine + zonealarm.Supervisor → Warning 层 armed
  2. engine.SetBathroomFallRules(NewBathroomFallRules(...))           → Critical bathroom 层 armed
  3. engine.SetBedroomFallRules(NewBedroomFallRules(...))             → Critical bedroom 层 armed
```

**任一缺失结果**：
- 缺 zoneengine：bathroom 内 Critical 漏报时无 Warning 兜底（10min Stay 不发）
- 缺 BathroomFallRules：bathroom 内 Critical 全部失效，仅靠 Warning 一档（太晚）
- 缺 BedroomFallRules：bedroom lost/bedside fall 完全失效

**How to apply (PR-bootstrap 写 main.go 时)**:
```go
func main() {
    engine := roomengine.NewEngine(...)
    
    // 启动后 invariant 检查（任一缺失 log.Fatal 拒绝启动）
    if engine.bathroomFall == nil {
        log.Fatal("BathroomFallRules not wired — refusing to start (sensor_v2_known_limitations.md L4)")
    }
    if engine.bedroomFall == nil {
        log.Fatal("BedroomFallRules not wired — refusing to start (sensor_v2_known_limitations.md L4)")
    }
    if zonealarmSupervisor == nil {
        log.Fatal("zonealarm.Supervisor not wired — Warning floor lost")
    }
}
```

启动 log 应同时出现 `bathroom_fall_rules_initialized` + `bedroom_fall_rules_initialized` + `zone_alarm.yaml loaded`，缺任一行 = 漏接。

---

## L5. PR-12 Fall dedup + 阈值矩阵未实现

**当前实施**：10b 和 10d 在 bathroom 内可能同帧 fire（不同 Reason）

**spec PR-12 计划**：
- Fall dedup 优先级：`silent > bedside > bathroom_* > still > lost`，60s 内同房只 fire 最高级
- 阈值矩阵 risk_factor 应用（决定 19，unit.residents.count_active ≥ 2 时 ×1.5）
- 最强信号不放宽（LostFall 30s / silent_fall 60s 矛盾窗口）

**Why**：sensor_v2.md §6.A 设计原则 "Critical 多档故意重叠" 锁定了 sensor 端不 dedup（每条 Reason 携带不同救援响应优先级）。dedup 在 cardagg 端按 Reason group / dedup 更合适。

**How to apply (when triggered)**:
- prod 实测重复 fire 噪声 — 如果 cardagg dedup 已经处理好，PR-12 可推迟或退化为纯阈值矩阵
- risk_factor 矩阵需要 unit.residents.count_active 信号，bootstrap 时加载

**Trigger PR-12**：
- cardagg dedup 不到位 + 同事件多 Reason 影响 elder 救援响应 → 必做
- multi-resident unit 部署后阈值矩阵需求显化 → 必做
- 否则 PR-12 可推到 Phase D/E 之后

---

## 备注：PR-X memory 同源

每条 L1-L5 在 Claude 项目记忆系统里都有对应 memory entry（`.claude/projects/.../memory/*.md`）。本文档是给团队成员（非 Claude session 用户）查阅。两边内容应保持同步 — 修改 limitation 描述时记得两边都改。

未来如果新增 limitation：
1. 在 Claude session 中写 memory entry（type=project）
2. 同步追加到本文档（保持团队可见性）
3. 实际触发 PR-X 时，commit message 引用本文档段号 + memory entry
