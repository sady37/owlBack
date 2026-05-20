# Layout 视角的 backlog — 从 layout 几何知识 derive 业务信号

收集"可以从 layout 几何 + 雷达 XY 即时算出来"但当前未接线的字段 / 派生信号。这类工作的共同特征：
- 数据全在 sensor 端（grid 几何 + monitor track 坐标），不需要跨进程协调
- 接线代价小（adapter / translator 拨几个引用）
- gap 类型 = "字段在 schema 里、消费者在用、但永远 zero value，因为没人写"

---

## TD-001 RoomState.LastExitToOutside 接线

### 背景

`owl-common/card/card_types.go::RoomState.LastExitToOutside bool` 字段定义于 schema，注释说"留作 risk/alarm 原始信号"，**但当前永远 false** —— grep `LastExitToOutside\s*=` 命中 0，没人写。

### 数据已就位

| 组件 | 携带的信息 | file:line |
|---|---|---|
| Radar firmware → adapter_radar | ExitRoom event 携带 track 最后位置 `position_x/position_y`（canvas 坐标）| [zoneengine/adapter_radar.go:157 + applyRoomLike:188](owlBack/wisefido-sensor/internal/zoneengine/adapter_radar.go#L157) `fields` 透传到 `SignalEvidence.TriggerData` |
| Layout grid | 每个 Enter cell 的 `EnterTarget string` （`""`/`"outside"`/`"bathroom"`）| [roomengine/cell.go:151-157](owlBack/wisefido-sensor/internal/roomengine/cell.go#L151) + [grid.go::StampEnters L142](owlBack/wisefido-sensor/internal/roomengine/grid.go#L142) |
| ZoneEvent.Trace | 翻转触发的 SignalEvidence 列表（含上面 TriggerData）| [zoneengine/types.go:135-144](owlBack/wisefido-sensor/internal/zoneengine/types.go#L135) |
| RoomConfig | `cfg.Enters []radarutils.Rect` + `cfg.EnterTargets []string`（parallel array）| layout_parser.go 写入 |

### 计算公式

```
当 ZoneEvent.Transition == "vacant" 且 ZoneType==Room：
  exit_xy = ZoneEvent.Trace[last(kind=="leave")].TriggerData["position_x/y"]
  cell = grid.CellAt(exit_xy)
  LastExitToOutside = (cell != nil && cell.EnterTarget == "outside")
```

### 接线方案

**方案 A（推荐 — 在 adapter_radar 早 lookup）**：

1. adapter_radar 收 ExitRoom 时，从 `fields` 取 position_x/y，调 `grid.CellAt(x, y)`，把 `cell.EnterTarget` 写回 `fields["exit_target"]`
2. translator 在 TransitionVacant 分支读 `ZoneEvent.Trace[last].TriggerData["exit_target"]`，等于 "outside" → 设 `out.LastExitToOutside = true`
3. translator owner 注释更新（[translator.go:14](owlBack/wisefido-sensor/internal/zoneengine/translator.go#L14) 把 LastExitToOutside 从"非 sensor owner"挪到"sensor owner"列）
4. cardagg projector 维持"整段覆盖"语义不变（sensor 现在写就行）

**方案 B**（translator 自查 grid）：解耦更差，translator 当前刻意不引 grid，不推荐。

### 工作量

- adapter_radar.go: ~10 LOC
- translator.go: ~5 LOC + comment 订正
- types.go 增 `ExitTargetKey = "exit_target"` 常量
- Tier1 test: 5 case (vacant + outside cell → true / vacant + inside cell → false / vacant + bathroom cell → false / 缺 position 字段兜底 / 非 vacant transition 不动)

### 替代选项：删字段

若业务上确认 LastExitToOutside 无消费者（SceneState/risk/alarm 都不用），**直接从 schema 删**比接线更干净。需先 grep cardagg/data/FE 确认 0 reader。

---

## TD-002 FE mirror-learned cells 显示 + 用户确认回写 layout

### 背景

sensor 侧 L1 mirror pair ghost 检测 + 自学习已完工（commit `c360bbb` / `9efde6a`）：
- `mirror_detect.go` 三不变量 + radar 距离 tiebreaker
- 单 10cm cell ≥3 hits → `Belief[0]=AreaDeny+SourceLearned`（人工 SourceHuman 不覆盖）
- 写入 `roomengine_grid_snapshot` 表（含 `MBC`+`LMM` 字段）
- 每日 11:50 归档 `_history` 表保 365 天（[[grid_snapshot_history]]）

**AreaDeny+SourceLearned cells 已在 sensor 内部直接生效**（fall verifier 看到即判 ghost），不需要 FE 介入就工作。此 TD 是给安装员 / 客服**可视化审查 + 必要时纠正**的额外闭环。

### 现状（FE / data 侧 0 接线）

`grep roomengine_grid_snapshot|mirror|MBC` 在 `owlFront/` 与 `wisefido-data/`：0 命中。

| 段 | 状态 |
|---|---|
| sensor 检测 + 学习 + 持久化 | ✅ 完工 |
| wisefido-data API 暴露 grid_snapshot 给 FE | ❌ 未做 |
| FE 渲染学习态 cells（AreaDeny + MBC>0）| ❌ 未做 |
| FE 给用户确认 / 编辑（手画多边形覆盖学习态）| ❌ 未做 |
| 回写 `room_visual_layout.cfg.Interferes` | ❌ 未做 |

### 待拍板决策（开工前必须先定）

1. **UI 形式**：grid 画布（叠在 layout 上显示 cell color）/ 列表（按 cell 坐标 + MBC 排序）/ 两者结合
2. **用户操作语义**：
   - 一键确认整片学习态 → 升 SourceHuman
   - 编辑单 cell → toggle AreaDeny/AreaWalk
   - 画多边形覆盖学习态 → 写入 `cfg.Interferes`
3. **回写时机**：实时（每次编辑发请求）/ 提交时批量
4. **学习态是否暴露给一般用户**：还是仅 admin/installer 可见
5. **edit 优先级**：用户手画 > 学习 — 已在 sensor 端规则里（`SourceHuman` 不被 `SourceLearned` 覆盖），FE 层只需呈现「学习 vs 人工」即可

### 涉及 service / 工作量估算

- **wisefido-data**：新增 API（GET grid snapshot by room_id / POST layout interferes 更新）+ Tier1 — ~1.5h
- **owlFront**：grid 画布组件 / 列表 / 编辑交互 / API client — ~3-4h（UI 复杂度 dominate）
- **owlRD（可能）**：`room_visual_layout.cfg.Interferes` 字段已存在则不动；否则补 migration — ~30min
- **owl-common（可能）**：grid_snapshot 对外字段（避免暴露内部 Belief schema）— ~30min

总计 5-6h，跨 4 repo，**独立项目级**。

### 不做的代价

仅失去「可视化审查 / 用户 override」能力。sensor 自学自用的核心功能不受影响。优先级看产品对「黑盒学习不可解释」的容忍度。

---

## 通用模式：layout-anchored derivations

凡是"sensor 知道几何 + 雷达知道当前 XY，能即时 derive 出来的信号"，都应在 sensor 端落地（不要让 cardagg 跨进程查 layout）。例子：

- ✅ Track CellAreaType（已落 — TrackStatusBase.CellAreaType 由 grid.CellAt 填）
- ✅ Track EnterTarget（已落 — `base.EnterTarget = c.EnterTarget`）
- ❌ RoomState.LastExitToOutside（**本 TD-001**）
- 其它候选（待评估）：进 Toilet/Shower cell 翻转的 dedicated event；从 AreaSit cell 离开的 "stand-up after long sit" 信号；etc.

新条目按 TD-NNN 顺序追加。

---

## 关联

- [[v3_ToDo.md](v3_ToDo.md)] — v3 大改造期 backlog（multi-target tracking 等需要架构改的留 v3）
- [[card_display.md](card_display.md)] — card display spec（SceneState / RoomRiskLevel 等消费 RoomState 字段的位置）
- [owl-common/card/card_types.go](../owl-common/card/card_types.go) — RoomState schema 源
