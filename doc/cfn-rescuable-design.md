# C_FN 可救援数（rescuable count）设计方案

状态：**第一层已落码 + 09e7 验证**（2026-06-21，用户拍"先算对+记日志，纯 forensic 不碰报警"）。第二层"C_FN 融入 fire"（§7）仍是未落码大工单。
A（radar 融合）会话产出。

**落地（forensic-only，FN-safe）**：
- `belief.ArgmaxIsBed(sMarg)`（decide.go）= 共享"radar 在床"判据（与 B 切片2 同源）。
- `census.RescuableCount(inBed)`（census.go，A 范围）= present-real ∧ ¬inBed，应用同人折叠（与 Nr 同口径）。
- `Decision.RescuableCount`（decide.go forensic 字段）+ engine Tick loop 后算（engine/engine.go）+ xray `rescuable` 打点（main.go）。
- **FN-safe 已 grep 验**：`RescuableCount/ArgmaxIsBed` 不进 fire 判据（`inst := pFallen>=pFire` 不读它）；cFN/RiskContext **未动**（第二层再切）。
- **09e7 验证**：fire=0 零回归；rescuable≤n_r 恒成立；床排除触发正确（`n_r=2 rescuable=1` 帧 = 09E7 轨 argmaxS=Bed 被排、D523 轨 Fallen 算入）。
- ⚠️ **覆盖缺口（no-silent-caps）**：09e7 无"单人独自躺床"段（房级 top_s 仅 Fallen/Empty），故 `n_r=1→rescuable=0` 场景**未被本 case 覆盖**；床排除靠 `n_r=2→rescuable=1` 帧验证。要补该场景需另取含"单人在床"段的 case。

---

## （原方案，保留作设计依据）状态：方案（已落第一层）
关联契约 [`fusion-absorption-contract.md`](fusion-absorption-contract.md) §2/§7.4；memory [[fall_detection_risk_stratified_design]] / [[dbn_mode2_veto_via_autorecover]] / [[partial_monitoring_fall_suppression_law]]。

---

## 0. 一句话

把 fall 代价函数 `C_FN` 里"房内人数"的口径，从**所有真人**收窄成**能救摔倒者的人（可救援数）**——排除躺床/睡着的人（radar S=Bed + sleepad-detected）。少算救援者 → C_FN 更高 → fire 更激进 = **严格 FN-safe**。

---

## 1. 现状（落码前必须认清的两个事实）

### 1.1 C_FN 当前**不门控 fire**，只进 forensic

[`belief/decide.go`](../tools/Xsensorv1/internal/roomengine/belief/decide.go)：
- `inst := pFallen >= pFire(0.85)`（:104）——**fire 判据只看 `P^F(SFallen)` 单阈**，不经 C_FN / PeopleCount。
- `cfn := d.p.cFN(rc)`（:102）注释明写"risk 融入裁决待后续工单，**当前不门控 fire**"。
- 文件头（:7-13）："§26 55% 三分 + C_FN 仅两可窗 + Λ 作 gate **已废**……risk-time 如何不受限地融入裁决 = 后续工单"。

**推论**：本方案改 C_FN 口径 = **纯 forensic、零告警影响**。它的实际价值要等**独立的"C_FN/risk-time 融入 fire"大工单**才兑现——那才是碰 FN-safe 红线（降告警需 95% 置信，[[dbn_mode2_veto_via_autorecover]]）的高风险改动，**不在本方案范围**。

### 1.2 人数链路与 ordering 难点

[`engine/engine.go`](../tools/Xsensorv1/internal/roomengine/engine/engine.go) `Room.Tick`：
```
:304  r.census.Update(...)                              // realness / 折叠
:307  nr := r.census.Nr()                              // 折叠后真人数（A 已落地）
:310  rc := adapter.BuildRiskContext(fi, nr, alone)    // RiskContext.PeopleCount = nr   ← Tick 前组装
        ── per-track loop：每 track belief.Tick → SMarg（S 9 态边缘后验，:454）── ← S 在 Tick 后才有
:558  dec.PeopleCount = nr
```

**难点**：`RiskContext.PeopleCount` 在 per-track Tick **之前**组装（只有 `nr`，不知每 track 的 S）；而"可救援数"要排 `S=Bed`，**依赖 Tick 之后**的 `SMarg`。当前 `cFN()` 在 `decider.Tick(rc)` 内（loop 内）调用，用 loop 前的 `rc` —— 拿不到本帧的 S。

**本方案的解（利用 1.1）**：因为 C_FN 当前**只 forensic**，可救援数**也只需进 forensic**，可以在 per-track Tick loop **之后**单独算，不触碰 `decider.Tick` 的实时 fire 路径。ordering 难点被"forensic-only"绕开。真正的两遍-Tick/上帧-S-近似留给"融入 fire"工单（见 §7）。

---

## 2. 可救援数定义

```
rescuable = #{ present-real track : argmax(SMarg) ≠ SBed }   （应用 census 同人折叠）
            − uncovered-sleepad（不计；sleepad 本就不进 P2，见契约 §2）
```

- **present-real**：`lastTick==tick ∧ Online ∧ PReal≥0.5`（与 `census.Nr()` 同口径，realness 已排镜像 ghost）。
- **排 `SBed`**：`argmax(SMarg) == belief.SBed` 的真人 = 躺床，救不了摔倒者 → 不计。
  - **只排 `SBed`**（躺床）。`SSit/SOpenFloor/SBath/SBlind*` 不排——坐着/站着/盲区清醒者能救人；保守少排 = 更激进 fire = FN-safe。
- **折叠**：先按可救援 filter，再在可救援子集内应用 `census` 的同人对（`link.same`）折叠减量（2 雷达 1 人且该人不在床 → 仍数 1）。
  - 边界：同人对一端 S=Bed、一端 S≠Bed（同人两雷达判不同态）→ filter 后只剩一端，无折叠减量，算 1 可救援（正确：人实际可动则算可救援）。
- **不计 sleepad**：契约 §2 P2 已定 sleepad 不进风险口径；可救援数进一步排 sleepad-detected 的人（睡着）。

| 场景（09e7） | Nr（A 折叠） | rescuable |
|---|---|---|
| 1 人 2 雷达，**站立/走动** | 1 | 1 |
| 1 人 2 雷达，**躺床**（两 track S=Bed） | 1 | **0** |
| 1 人在床 + 1 人站立（2 雷达各见） | 2 | 1 |

---

## 3. 实现 locus（落码时）

### 3.1 共享 helper（单源，与 B 切片2 对齐）

B 切片2 判"radar 在床"用 `fr.Tracks[i].SMarg argmax == belief.SBed`。**可救援数必须用同一判据**，否则两套口径 drift。

→ 落 **一个共享 helper**，B 与本方案共用：
```go
// belief 包（SMarg 是 belief.Frame 的字段，归属 belief 最自然）
func ArgmaxIsBed(sMarg []float64) bool   // argmax(sMarg) == int(SBed)
```
登记进契约 §3 共享代码面（避免 A/B 撞车）。

### 3.2 算 rescuable + forensic

在 `Room.Tick` 的 per-track loop **之后**（`results[]` 已含每 track `SMarg`）：
```
rescuable := 0
for each present-real track t:  if !ArgmaxIsBed(t.SMarg) { rescuable++ }
rescuable -= 折叠减量(census.same 对中两端都 !ArgmaxIsBed 的)
```
- 塞进 `Frame`/`Decision` 的 forensic 字段 `rescuable_count`（main.go onRoomFrame 打进 xray，与 `n_r`/`real_people` 并列）。

### 3.3 cFN 口径切换

`decide.go` `RiskContext` 加 `RescuableCount int`；`cFN()` 的多人折扣 `if rc.PeopleCount>1 → disc=1/N` 改读 `RescuableCount`：
- **当前**：`cFN()` 只进 forensic（Decision.CFN），故可救援数也只 forensic——可在 loop 后重算一次 forensic `CFN`，**不动 `decider.Tick` 内的实时路径**（1.1：实时路径本就不读 cFN）。
- `RiskContext.PeopleCount` 保留（forensic 对照"全员 vs 可救援"差值）。

---

## 4. FN-safe 论证（review checklist）

- [ ] **方向**：可救援数 ≤ Nr（排床上人）→ C_FN ≥ 原值 → fire 阈更松 → **只会更少漏报**。无 fall FN。
- [ ] **当前零 fire 影响**：fire = `pFallen≥0.85` 单阈，不读 cFN（§1.1）。本方案落码后 grep `decide.go` 确认 `inst`/`fired` 不引用 `RescuableCount/CFN`。
- [ ] **只排 SBed**：不排 Sit/OpenFloor/Bath/Blind（少排=保守=FN-safe）。
- [ ] **折叠不压摔**：可救援折叠只动计数，两 track 始终各自独立进 OR-fire（契约 §1 铁律）。
- [ ] **helper 单源**：与 B 切片2 `ArgmaxIsBed` 同一实现，无两套 SBed 判据。

---

## 5. 范围与授权

- 落码点 = `belief/decide.go`（`RiskContext`/`cFN`）+ `engine/engine.go`（loop 后算 rescuable）+ `belief` 新 helper。
- **契约 §1 现 A/B 均不碰 belief** → 本方案落码须 (a) 用户授权扩范围，或 (b) 单立 fall 任务。**本文件只是方案，未落码**。
- 与 B 切片2 的执行序：B 的吸纳/SBed 判定先对齐 `ArgmaxIsBed` helper；可救援数复用之。

---

## 6. 验证计划（真 case，禁 unit test [[validate_real_case_no_unit_tests]]）

- **09e7-0620-2240**（2 雷达 1 人 + sleepad）：
  - 人躺床段 → 两 track `argmax(SMarg)==SBed` → `rescuable_count=0`（即便 `n_r=1`）。
  - 人站立/走动段 → `rescuable_count=1`。
  - **fire 零回归**（不门控 → fire 序列与现状逐 tick 相同）。
  - forensic `rescuable_count` 在 xray 可见。
- 工具：`tools/timeline_from_xray.py` 加 `rescuable` 列（per-room）。

---

## 7. 给"C_FN 融入 fire"大工单的接口预留（不在本方案）

当 risk-time 真要门控 fire 时，需解 §1.2 ordering：
- 选项 A：**两遍 Tick**（第一遍出 S，第二遍带 rescuable 进 cFN 门控）。
- 选项 B：**上一帧 S 近似**本帧可救援（1Hz，态变化慢；FN-safe 偏差方向需验）。
- 那时 `pFire` 阈由 `cFN` 调制（C_FN↑→阈↓），**且降阈/降告警门必须焊 95% 置信**（[[dbn_mode2_veto_via_autorecover]]）。本方案产出的 `rescuable_count` + `ArgmaxIsBed` helper 是其就位地基。
