# zone_rules.yaml — Provisional 数值来源 + Tuning Guide

> **状态**：所有数值是 placeholder，最终由用户对照真实信号 trace + log 调参。本文档解释**为什么**当前数值是这些（而不是别的），让 tuning 时知道每个旋钮的预期影响范围。
>
> **相关代码**：[`internal/zoneengine/scorer.go`](../internal/zoneengine/scorer.go) / [`state_machine.go`](../internal/zoneengine/state_machine.go) / [`engine.go`](../internal/zoneengine/engine.go)
>
> **设计图**：[`docs/zone_engine_design.html`](./zone_engine_design.html)

---

## 1. 评分模型回顾

```
pos = max(enter.strength_decayed, sustain.strength_decayed) + recent_enter_bonus
neg = leave.strength_decayed
score = pos - neg                  // ∈ [-100, +100+bonus]
```

- 多源同方向取 **max** 不是 sum（防 sleepace + radar 同向叠加成 170 假强）
- 反向 **latch**：高优先级 evidence 在 latch 窗内拒绝反向信号
- **decay**：线性，120s 走完一档 strength（[`scorer.go:48`](../internal/zoneengine/scorer.go#L48) `DecayWindowMs`）
- enter / leave **互斥清零**；sustain 独立通道（[`scorer.go:140-178`](../internal/zoneengine/scorer.go#L140-L178)）
- **recent_enter_bonus +15** 仅 sustain 活时加（[`scorer.go:218`](../internal/zoneengine/scorer.go#L218)）

---

## 2. Bed zone 数值来源

### enter_evidence

| Source | Strength | Latch (s) | 推理 |
|---|---|---|---|
| `sleepace` | 90 | 10 | 厂家压感原生信号，硬件级权威 → 设最高一档（90，留 +10 给未来需要 100% 强信号的场景） |
| `radar`    | 80 | 10 | 雷达派生 InBed（无 sleepad 时主源） → 比 sleepace 低 10，反映"间接推断"性质 |

Latch=10s 因为：sleepace InBed 一般跟着稳定持续状态（人上床后不会立即下来），10s 内压制反向信号是合理的厂家权威窗口。

### sustain_evidence

| Item | Value | 推理 |
|---|---|---|
| `hr_rr_present_any.strength` | 80 | 与 radar enter 同档；HR+RR 同时 > 0 是"在床"的强物理证据 |
| `hr_rr_present_any.sources`  | `[sleepad, radar]` | 任一源接受 |
| `recent_enter_bonus.window_sec` | 60 | enter event 后 60s 内视为"刚上床"（sleep transition 期普遍 < 60s） |
| `recent_enter_bonus.bonus`   | 15 | 不能太大避免 enter 单独 105；+15 让 80→95 是"刚上床更稳"的微强化 |

### leave_evidence — `bed_size_dependent: true`

理由：床型决定 sleepad 压感的覆盖率。Twin/Hospital 床 sleepad 覆盖率 ≈ 100%；Queen/King 床中央有 dead zone，老人睡边缘 sleepad 漏报概率高。

| Bucket | Bed kinds | sleepace | radar |
|---|---|---|---|
| `small_bed` | standard, hospital, twin | 80 / latch 2s | 70 / latch 2s |
| `large_bed` | full, queen, king, california_king | **70** / latch 2s | 70 / latch 2s |

`small_bed.sleepace=80` 因为压感 100% 覆盖时是"权威离床"信号；`large_bed.sleepace=70` 降到与 radar 同档，避免边缘睡眠假离床。

Latch=2s 是关键 — 不阻塞老人正常的"坐起→坐回"行为（< `leaving_window_sec=8s`），让 Returned 路径生效。原 5s 会阻塞 4s 内 Returned。

参考床型规格：[`USA_ElderCare_Hospital_Bed_Specs.md`](../../doc/USA_ElderCare_Hospital_Bed_Specs.md)

### state_machine

| Item | Value | 推理 |
|---|---|---|
| `enter_threshold` | +50 | 单一弱源（radar=80）即可越过 → 不卡正常事件；多源同向 max=80 也触发 |
| `exit_threshold`  | -50 | 对称设计 |
| `hysteresis_sec`  | 3   | 翻转后 3s 冷却（仅作用于反向翻转；Returned 不受约束） |
| `leaving_window_sec` | 8 | **中敏感度**：老人 4-6s 内"坐起→躺回"不假翻；> 8s 视为真离床 |

Leaving 窗口选择：医学文献老人 sit-to-stand 平均时间 ~3-5s（含犹豫）；8s 留 50% buffer 给迟疑性回床。

---

## 3. Room / Bathroom zone 数值（按 Bed 段比例外推）

```yaml
enter: { radar: 80/2s, polygon: 70/2s }
leave: { radar: 80/2s, polygon: 70/2s }
state_machine: { enter:+50, exit:-50, hyst:2s, leaving_window:5s }
```

- `radar=80`：与 Bed enter 同档 — firmware EnterRoom 是雷达侧已确认的事件
- `polygon=70`：自实现多边形穿越（待上线），降一档因为是软件派生
- `latch=2s`：room 边界事件不需要长 latch（人很快可能再次穿越）
- `leaving_window=5s` < bed 的 8s — room 边界相对清晰，老人不会在门口反复试探

**关于 NumberPeople**：统一以 `Kind="count_change"` 走旁路（只改 Count，不动 occupied）。yaml 不配 `number_people_0` 因为：
- count=0 ≠ "人离开"，可能只是静止 / 检测失败 / 雷达盲区
- 真离开应由显式 ExitRoom event 驱动
- 若未来要支持持续 count=0 辅助判离，应在 `Tick` 中加复合规则（如"连续 60s count=0 且无 sustain"），不能简单 `leave`

Bathroom 与 Room 同 schema，可独立 calibrate；目前同值。

---

## 4. negative_feedback 数值

### self_contradiction

| Item | Value | 推理 |
|---|---|---|
| `window_sec` | 15 | 翻转后 15s 检测窗 — 短于 leaving_window 二倍，长于 hysteresis 5x |
| `require_sustain_at` | 30 | 最小持续 score —— 若翻转后 score 跌破 30 即视为"翻得太草率" |
| `penalty_weight` | 5 | 触发源 confidence 临时 -5（不至于一次失败永久失能） |
| `penalty_duration_min` | 60 | 1 小时降权 — 等同医院"留观一节" |

**仅 occupied 方向登记**（[`engine.go:197`](../internal/zoneengine/engine.go#L197)）：
- Occupied 是正向维持态，需 score 持续证明
- Vacant 是默认安全态，证据老化不算"自相矛盾"（防"幽灵床位"翻回）
- Leaving 本身就是 self_contradiction 的"软形态"（timer 自动确认/取消）

### subset_invariant

| Item | Value | 推理 |
|---|---|---|
| `enabled` | true | bed⊆room 是物理约束，无理由禁用 |
| `mode` | `lift_parent` | bed 真在 → 抬升 room；不用 `drop_child`（会砍掉真信号） |
| `repair_interval_sec` | 10 | 周期巡检 10s，足够补偿"room 已 vacant 但 bed 仍 occupied"漂移 |
| `stale_bed_threshold_sec` | 300 | bed 5min 无证据视 stale（远长于任何正常事件间隔） |

巡检使用 [`Scorer.LastEvidenceTs()`](../internal/zoneengine/scorer.go#L89) 而**不**用 `state.UpdatedAt`（Tick 每秒刷新，永远显示新鲜，不能反映真证据老化）。

---

## 5. Tuning Guide — 调参怎么走

### 现象：老人坐起后假翻 Vacant
**症状**：sleepace LeftBed 触发后 4-6s 内 ZoneEvent.Transition=`vacant`，但下条 sleepace InBed 紧接 8s 后  
**调整**：
1. `leaving_window_sec` 8 → 15（让 Returned 路径有更长窗口）
2. 检查是否 `latch_sec` 误设大值（应 ≤ 2s）

### 现象：床位"幽灵"，明明无人床仍 Occupied
**症状**：sleepace LeftBed 触发后 ZoneEvent 不发，state 仍 Occupied  
**可能原因**：
- enter latch 阻塞了 leave —— 检查 `enter.latch_sec` 是否 ≥ leave 间隔（默认 10s）
- `subset_invariant` 周期巡检失效 —— 验 `repair_interval_sec` > 0 且 logs 有 `invariant_repair_*`

### 现象：room 频繁假翻
**症状**：低 confidence 的 polygon 或 number_people 触发误翻  
**调整**：
1. 提 `enter_threshold` 50 → 60（要求更强信号）
2. 降 `polygon.strength` 70 → 60

### 现象：HR/RR 在床但 score 还在下行
**症状**：sleepace InBed 翻 Occupied 后，sustain 应保持但 score 下行触发 Leaving  
**可能原因**：
- `hr_rr_present_any.strength` 80 < `exit_threshold` 绝对值 50 + leave.strength 80 = 130（leave 强信号 + sustain 80 = -50 仍触发 exit）
- 调整：sustain.strength 80 → 90；或 ExitThreshold -50 → -60

### 现象：cell-history 假报压不下来（后期）
当前规则 self_contradiction 只针对单事件级；cell-level 假报历史是 Phase 2。届时 yaml 会增加 `cell_history` block。

---

## 6. 不在 yaml 的硬编码（约定不动）

| Constant | Where | Value | 说明 |
|---|---|---|---|
| `DecayWindowMs` | [`scorer.go:48`](../internal/zoneengine/scorer.go#L48) | 120_000 | 一档 strength 完全衰减时间 |
| `SustainStaleMs` | [`scorer.go:51`](../internal/zoneengine/scorer.go#L51) | 10_000 | sustain evidence 失效阈值 |
| `traceWindowLen` | [`engine.go:65`](../internal/zoneengine/engine.go#L65) | 8 | 每 zone trace buffer 长度 |
| `DefaultLeavingWindowSecFallback` | [`state_machine.go:33`](../internal/zoneengine/state_machine.go#L33) | 8 | yaml 漏配 leaving_window 时的兜底 |

这些是工程上的安全默认，业务侧不应该需要改；改前请先 review。

---

## 7. Hot Reload 工作流

引擎已实现 `ReloadRules`（[`engine.go:102`](../internal/zoneengine/engine.go#L102)），运行时状态保留，仅切换规则表。配套的 yaml 监听 + `config:zone_rules:stream` 触发待 wiring。

调参流程（建议）：
1. 改 yaml
2. 通过 admin 工具 publish 到 `config:zone_rules:stream`
3. sensor 收到后调 `Engine.ReloadRules`
4. 观察 5-10 分钟 ZoneEvent + Score trace
5. 不行回退（保留旧 yaml 副本）

---

## 8. 关键决策的承上启下

| 决策 | 当前实现 | 后续 backlog |
|---|---|---|
| 三态状态机 Vacant/Occupied/Leaving | ✓ done | — |
| sustain 不清 + recent_enter_bonus 仅 sustain 活时 | ✓ done | — |
| self_contradiction 仅 occupied 方向 | ✓ done | — |
| subset_invariant 双向 + scorer.LastEvidenceTs | ✓ done | — |
| 同方向多源 max 非 sum | ✓ done | — |
| **cell_history**（cell-level 假报历史自适应阈值） | 后期 | 见 [`fall_rules_three_classes`](../../../.claude/projects/-home-wisefido-owl/memory/fall_rules_three_classes.md) memory |
| **operator_feedback** | 后期 | Phase 5 ground truth |
| **behavioral_baseline** | 后期 | 老人个体节律 baseline |
| **spatial_config 长前缀覆盖**（tenant-level rules override） | 待 wiring | yaml structure ready |
