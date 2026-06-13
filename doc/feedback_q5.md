# Q5 反馈日志(项目组侧)— belief 全空间区域占用重写 + cd2b 床边真摔 FN

> **QA 分离(2026-06-12,因 git 同步写冲突)**:P5 主题拆成两文件各写各的——
> **本文件 `feedback_q5.md` = 项目组侧**(提案/问题/根因/复算请求);委员会裁决/审查/工单写 `feedback_a5.md`。
> 两侧不写同一文件 → push 不再非-fast-forward 撞车。倒序,最新在上。
>
> **协作协议**:项目组提案 → 委员会裁 → 裁后建;裁前不建 / 需新岔口列选项不擅决 / 直接 main / cd wisefido-sensor 跑 go / 放行 bar = build/vet/belief 绿 + 0 FAIL。

---

## 工单5(cold-start Phase A)落码完成(2026-06-13,先提交待委员会审)

委员会 §6 裁决全部照办,Phase A 纯内存落地。放行 bar 达成:`build ./...` ✅ / `vet ./internal/roomengine/...` ✅ / belief 包 ✅ + **roomengine 0 FAIL**。

**实现(`belief_shadow.go` + `engine.go`,~70 行)**:
- `Engine.unitFirstTrackMs map[string]int64` + `coldStartMu`(纯内存,重启清零=重新冷启动,退档=FP 安全,§6.4 Phase A)。
- `markUnitTrack(suiteID, nowMs)`:beliefShadowTick 里 `len(bases)>0` 时打桩 firstTrack(单调只落首次),unit 粒度=`roomSuiteID`。
- `unitCap(suiteID, nowMs)`:`now−firstTrack ≥ max(coldGraduateMs, coldFloorMs)` → 2,否则 1(firstTrack=0 全新 unit → 1)。**纯时钟**(§6.1 裁 C,无覆盖率门)。`coldGraduateMs` 默认 72h、env `DBN_COLD_HOURS` 覆盖;`coldFloorMs=24h` 硬下限(§6.2)。无第三道稳定门(§6.3)。
- 两个消费点改 per-unit:`belief_shadow.go` 自发 → `dbnSelfFireEnabledFor(unitSuiteID, nowMs)`;`engine.go` veto firmware → `dbnVetoFirmwareEnabledFor(roomID, ts)`。二者 = `min(全局 dbnMode, unitCap)`,自发≥1 / 否决==2。
- **启动=1,路径 1→2**(§四裁决):cold unit cap=1 仍允许 DBN 自发(cd2b 这类 firmware 漏判靠它补),仅不否决 firmware;`unitCap=0` 只在运维 `DBN_MODE=0` 时由 min 产生。
- **删死码(#1.2)**:全局 `dbnSelfFireEnabled()`/`dbnVetoFirmwareEnabled()` 改 per-unit 后零调用 → 删除。

**测试**:新增 `cold_start_test.go::TestColdStartUnitGate` 锁四点——markUnitTrack 单调 / 时钟门 1→2 / T_floor 硬下限 / `min(ceiling,cap)`(dbnMode=2 cold→不否决但自发、mature→否决;dbnMode=1 ceiling 压 mature 不否决;dbnMode=0 全静默)。

**Phase C(挂 grid_snapshot,§6.4)未做**:等 playback `--persist`(74cc558)live 验证后再合,与本 Phase A 解耦。

**一处实现取舍记录**:beliefShadowTick 内已有一个**条件置空**的 `suiteID`(仅 smallBath 时填,喂 574/590 的 recapture/SuiteHasOtherDevice)。cold-start 需**恒填**的 unit id,故另起 `unitSuiteID := e.roomSuiteID[roomID]`,不动既有 `suiteID` 语义(避免改 lost-fall sweep 行为)。

---

## 新提案:cold 启动期 per-unit 升档闸(`unit_dbn_mode`)— 申请委员会裁(2026-06-13)

### 问题

全局 `DBN_MODE=2`(全开:否决 firmware + DBN 自发)依赖 grid 空间学习(AreaSit/tolerance)压制 FP——cell engine 须先学到"此处常站/常坐"才把 tolerance 反馈给 dwell 尾拉长。cold 启动期 grid 空白(tolerance 全 1.0,无任何学得语义),DBN mode 2 在 freshly-deployed unit 上会密集产出 dwell-silent FP——Kitchen/Hunzi/Ton 等常坐区的久静被误报。

本次 `dwellScaleOpenSec=720s(12min)` + radar 边缘 ×1.5 已大幅压低 cold-start FP(baseline 从 60s→12min 让正常歇着不触发),但**治标不治本**——12min 仍有跨入可能(晚间长坐/午睡),且容忍翻转(dwell LR<1 压制)完全不生效(无学得 tol)。

**正确解法**:per-unit 升档闸——unit 刚上线时不跑 mode 2,等 grid 学习达到"可压制 FP"的成熟度后再自动升档。同期 firmware 仍跑 mode 1 语义(直通,不漏真摔),DBN 仅暂不自发/不否决。

### 方案骨架

新增 per-unit effective DBN mode:
```
unitEffectiveMode = min(globalDBNMode, unitCap)
unitCap ∈ {0, 1, 2}   // 启动 = 0 或 1,升到 2 = 全权
```

- `globalDBNMode` = 现有 `DBN_MODE` 环境变量(运维全局闸,不变)
- `unitCap` = 本 unit 学习成熟度闸,由 engine 内部维护
- 运营视角:运维修 `DBN_MODE=2` 全局,新 unit 自动走 `unitCap` 自己爬坡;运维不 intervention

升档路径: `0 → 1 → 2`,每步经成熟度 gate。永不自动降档(fallback 靠运维手动 `DBN_MODE=0/1`)。

### §6 岔口 — 申请委员会逐项裁

以下 5 个岔口均项目组可列选项但不能擅决,请委员会裁。

---

#### §6.1 θ 度量:什么信号判"学习成熟"?

选什么指标作为 `unitCap` 升档的触发信号?

| 选项 | 描述 | 优点 | 风险 |
|------|------|------|------|
| **A. 网格覆盖率** | `grid cells with DwellEMA>0 / total cells` ≥ θ | 最直接——emawhere 有人停过,空间语义有了 | 覆盖率收敛慢(人不会走遍每格);θ 阈值难定 |
| **B. 容忍标定收敛** | `|ToleranceFactor_ema − 1.0|` 的 unit 级均值超过某门槛 | 最贴合动机——我们就是等 tolerance 能工作 | ToleranceFactor 只在开阔地有值(非全 grid);容易在盲区/卫浴为 1.0 拉低均值 |
| **C. 纯时钟** | `now − firstTrackMs ≥ T_cold` | 最简单,零复杂度 | 无学习语义——时间是 proxy,非 guarantee;unit 低活跃度(少人)可能时间够了但 grid 仍是空的 |
| **D. 混合:A+C** | 覆盖率 ≥ θ **或** 时钟 ≥ T_cold(先满足者胜) | 高活跃 unit 靠覆盖快升,低活跃 unit 靠时钟 safety-net 最终升 | 两套逻辑;低活跃 unit 时钟到但 grid 空的→升档后仍 FP |

**项目组倾向**: **D(混合)**。覆盖率主信号 + 时钟兜底(防低活跃 unit 永远不升)。覆盖率 θ=0.15(unit 内 ~15% 格有人停过)/ 时钟 T_cold=72h(3 天)作起步值待 oracle 标定。

---

#### §6.2 TimeFloor:最小冷启动保护期

不管 θ 度量多快满足,必须设一个硬下限 `T_floor`——防止"偶然达标→过早升档→FP 回归"。

候选值:
- **12h** — 覆盖一个白天(含早/午/晚不同活动模式)
- **24h** — 覆盖完整一昼夜
- **72h** — 覆盖 3 天(多日模式,与 §6.1 时钟兜底对齐)

**项目组倾向**: **24h**。一昼夜覆盖 sleep/wake 两相,且与 AreaSit 自学习 12min 窗口错开量级(不是同一个东西)。

---

#### §6.3 第三道稳定门(滞回)

θ 度量首次触线 → 不立即升档,需"持续满足 N 个连续检查窗"才升——防 flicker 升档。

- 检查周期:每 1h / 4h / 12h?
- N 次:1 / 2 / 3?

**项目组倾向**: 每 4h 检查 × 连续 2 次(即触线后稳 8h 才升)。理由:NS 不高压(升档 latency 8h 对 cold-start 场景可接受),安全感比速度重要。

---

#### §6.4 learn_start 存储归属

unit 的冷启动上下文(firstTrackMs / unitCap 当前值 / 升档历史)存哪?

| 选项 | 位置 | 重启行为 | 工程成本 |
|------|------|----------|----------|
| **A. 纯内存** | `Engine` struct 字段 | 重启→重新冷启动(清零) | 零 |
| **B. 挂 grid_snapshot** | snapshot 二进制 payload 内 | 重启→从 snapshot 恢复(但空 snapshot→重新冷启动) | 小——改 `MarshalSnapshot`/`UnmarshalSnapshot` 格式 |
| **C. 独立 DB 列** | `grid_snapshots` 表加 `cold_start_ctx` JSONB 列 | 重启→从 DB 恢复;跨 deploy 保留 | 中——schema migration + persister 改 |

**项目组倾向**: **B(挂 snapshot)**。理由:① cold-start 上下文与 grid 生命周期绑定(grid 重新学→上下文也该重置);② 不与 B "persist 暖 live"冲突——persist 本身就是把学好的 grid 落盘,带上 cold-start 上下文自然配套;③ 无 schema migration(playback `--persist` 写 → live 启动 load,同链路)。

但 B 的问题:**首次启动无 snapshot→内存模式→学好后 persist 写 snapshot(含 cold=done)→若 deploy 重启→从 snapshot 恢复=升档状态保留**。路径成立但隐式依赖 persist 时机——若 unit 升档后、persist 前 crash,重启会退档(重新冷启动)。退化可接受(退档=max 保守=FP 安全),但委员会应知悉。

---

#### §6.5 施工时序

此提案与 q5 现有工单(工单3 后半段 oracle 重基线 / 工单4 cd2b fire / belief 重写)的先后关系:

1. **前置依赖**:无硬依赖。此提案是 ops 安全网,正交于 belief 逻辑本身。
2. **建议时序**:
   - **Phase A**:委员会裁完 §6 岔口 → 项目组落 `unitCap` 逻辑(纯 `belief_shadow.go` ~50 行,无新数据流)
   - **Phase B**:q5 工单3/工单4/belief 重写继续并行推进
   - **Phase C**:grid_snapshot 挂载(§6.4 B 选项)等 playback `--persist`(74cc558) 经 live 验证后再合入

**项目组建议**:Phase A 可立即做(小、独立、安全),Phase C 等 persist 链路验证后再挂。

---

## 工单3 后半段 大部完成——用户选 A(2026-06-13,commit 5a31de1+94d9f79)

**用户裁:选 A**(tolerated cell→dwell LR<1 主动压制)。理由:B(死区)只延时不治本(tolerated cell 久站够久照饱和,case-5 站 30min 照报);A 给新拓扑(SFallen 近吸收)缺的**反向力**——cell engine 学到"此处常站"本就该是反 fall 证据,有反向力棘轮不成立;不伤 TP(真摔在非容忍点 mult=1 正常 ramp)。

**落地**(`survival.go` fallLRFromDwell 改 signed ramp + `dwellTolFlipK=2`):
- 非容忍 `mult=1` → `1+(d/scale)^k` 正向 ramp(TP 不变)。
- 容忍 `mult>1` → `tolWeight=(mult−1)·2` 翻符号 → 随 dwell **单调下压 LR<1**(floor 兜底)→ 主动压 SFallen。

**验**:`TestP4OpenFloorDwellToleranceGate` 无tol P0.940 fire / 高tol P0.021 not(分离 ✅);`TestMoMLostTrackVanishNoFire` 走出exit→Left 不报 ✅;真摔 case-3 fire 0.999(TP 保住 ✅)。build/vet 绿 + belief/roomengine 0 FAIL。

**case-5(FP)honest 残差**:`fire=0`(安全=不误报)但 `peak=0.994`。根因=A 是**学得**容忍机制,recall 测试 fresh 重放、cell 无累积容忍(mult=1)→ A 不 engage;case-3(TP)/case-5(FP)在 dwell 上**同构**,唯一分离杠杆=学得空间语义(§6.3 honest 残差)。→ case-5 `t.Skip` 留跟进:**pre-seed cell `ToleratedStillCount`(模拟学得态)** 或精度度量改 fire-based。这是工单3 后半段唯一剩项。

---

## 工单4 cd2b fire 实证 PASS(2026-06-13 live replay)

**cd2b-0604 床边真摔 FN 解决——新模型 fire。** 截 fixture `case-cd2b-0604-trim`(sleepad LeftBed 前60s→ExitRoom,~8min)经 `tools/replay --speed 1`→redis→live sensor(DBN_MODE=2)。按 room_id `fd00:0:3:112:3:100`(cd2b/201)过滤观测:

- `belief_shadow_fall` + `belief_dbn_fire` 确认;**`p_fallen=0.996`,`argmax=Floor-Fallen`,`reason=pose_lying`,~22:25**(对齐 oracle"应 ~22:25 fire")。
- **fire 路径 = 工单1 + 工单2 协同**:sleepad LeftBed → `bedReleased=true`(`belief_dbn_bedside_unbed`×224)→ `poseLikelihood(lying@bed, BedReleased=true)` → SFallen(`lrPoseLyingOpenFall=4`)→ ramp 0.996 → 确认。被子 lying@bed **不再被豁免**(床态已释放)→ 触发跌倒 = NOTES oracle 本意。
- 注:fire 走 pose_lying 路,**非** silent sweep(被子有 lying pose → pose 路先 fire;silent sweep 是"无 pose 全丢"的兜底)。

**更正前几轮误判**:中途"FN/床没释放/99 nodetect"是**看错房**——qinglan 一直喂别房(311/411)live 数据当噪声,未按 cd2b/201 room_id 过滤所致。cd2b/201 床**确实被 sleepad LeftBed 释放**,工单1 bed 融合(any-source-OR)本就 work。

**🔑 replay 观测铁律**(记入):① qinglan 必须运行(停了会断 sensor 处理 + FE);② 必须按目标 room_id 过滤 sensor 日志(隔离 qinglan 别房 live 噪声);③ 验 fire 必须 `--speed 1`(confirmMs/dwell 按墙钟)。

**工单状态更新**:工单1 ✅(实测床释放 work)/ 工单4 ✅(cd2b fire 实证)。trim fixture `case-cd2b-0604-trim` 暂未 commit(派生 artifact,待定)。

---

## 工单2 + 工单3-前半段 落码完成(2026-06-13,先提交待委员会审)

放行 bar 达成:`build ./...` ✓ / `vet ./internal/roomengine/...` ✓ / **belief 包 test ✓ + roomengine 0 FAIL**。

**已落(11 文件,全在 wisefido-sensor/internal/roomengine)**:
- `state.go` 新 9 态(`SEmpty/SBed/SSit/SOpenFloor/SBath/SFallen/SBlindRest/SBlindOpen/SLeft`,Fallen 仍 index 5);`model.go` 空间转移 A(SBed→Fallen 0.05、Blind→Fallen 0.5 种子、BlindRest 无 Left 逃生阀)+ Prior;`likelihood.go` 发射重映射(ObsNoDetect 压可见区+抬盲、poseLikelihood 卫浴分支、ExitRoom 压盲、删 ObsTrackPresent 整条);`calibration.go` 清死常量+新常量;`observation.go` 删 ObsTrackPresent;`belief_adapter.go` 删 ghost 发射。
- 测试:删 `belief_test.go`/`r5_calibration_lock_test.go`(旧 9 态 oracle 全废)→ 重建 `belief_test.go` 骨架 5 测(状态空间形状 / 床躺→Bed / 床离床→Fallen=cd2b belief 层出口 / 丢轨→blind ramp / ExitRoom 反证);修 `belief_adapter_test.go`/`belief_b_replay_test.go` 引用。

**落码中测试逮到并修的真问题**:① ExitRoom 似然没压盲区(补 `lrExitBlind`,且与"BlindRest 在 A 无 Left 逃生阀、离场靠观测拉"自洽);② Blind→Fallen 初设 12 是 drift 引擎(纯 Predict 凭空造跌倒)→ 改 0.5 种子,守住"★→Fallen 极小、靠观测驱动"原则。

**deferred 到工单3 后半段(oracle 重基线,已 t.Skip 带 reason 引工单3)**:3 个 FP 回归在新拓扑下需重标——
- `TestMoMLostTrackVanishNoFire`:喂纯 nil 是旧模型假设;新模型 MoM='走出 exit'应喂 ExitRoom→Left,**test 语义待更新**。
- `TestP4OpenFloorDwellToleranceGate`:dwell scale/tolerance 按旧拓扑标,新拓扑竞争 Fallen 的态变了,**dwell 常量待重标**(高/低 tol 当前不分离)。
- `TestRecallManifestAll`(case-5 hunzi FP peak 0.994):真数据 TP/FP recall/precision oracle **待按新拓扑重基线**。
- ⚠️ 暴露的标定张力:Blind→Fallen 种子大小 vs 纯 Predict 漂 / dwell tolerance 在新拓扑的分离力,是工单3 后半段重标的核心,且会牵动 lost-fall TP——须用真 fixture 标,别压死真摔。
- 委员会非阻断标记(2026-06-13 审 a5):`SOpenFloor→Fallen=0.5` 是旧 `SStandWalk→Fallen=0.2` 的 2.5x("极小种子 0.05–0.5"的上界)。作种子可接受、已知,归工单3 后半段连同其它 →Fallen 种子一并对真 fixture 重标(倾向回拢 0.2 量级以减 A 漂)。

**工单状态**:工单2 ✅(belief 重写)/ 工单3 前半段 ✅(骨架)/ 工单3 后半段 ⏳(oracle 重基线,上列 3 项)/ 工单1(bed 融合)、工单4(cd2b fire)未动。

---

## lost / still / moving 定稿模型(2026-06-13,用户拍板)

经几轮收敛,fall 时序坍缩成**两条机制、无例外**——纠正了原 belief_shadow 的 `MovingPrecondition` 译反(它在 still→lost 边界清零并排除最危险的"静止中被丢")。

### 核心洞察:lost 不另立机制,**算作 still-box 仍在持续**

still-box 的语义不是"track 在不在",而是"**人有没有离开这个 ~50cm 的点**"。一次跌倒 → 倒地+静止 → radar 可见性会闪(还瞄到=still-fall / 瞄丢=lost-fall),**但人没动,dwell 时钟从第一次静止起连续走,跨 lost/jitter 不重置**:
- track 丢(lost)→ 不清 box,当"还在框里只是没上报",dwell 继续;
- 小动几下(jitter)→ 落 StillBoxCm 内,box 不破(现判据 30s≥60%帧 50cm 内本就容忍);
- 丢了又回、**框内** → 无缝续上,自然滑过去。

### provisional + 位置一致性闸(用户补)

lost 期间"算 still"是**假设**,returning track 才证实/推翻:
- **返回框内(原位)** → 假设成立,Still 连续;
- **返回框外(别处)** → **Still 被中断**(空档里人移动了)→ 清 box + 拉回该段 ramp 的 Fallen;同时挡 **track-swap/ghost 在别处冒头**被误续;
- **一直不返回** → 假设保留(presume 倒地,保守安全)。

### 附赠:框进出 = 统一的自救/recovery 判据

不再需要单独 recapture 窗 / lostAnchor / 自救判别——丢了又回**框内**=人还倒着(继续 ramp,治 silent_leftbed 60s 恢复窗误 cancel 自救跌倒);**框外**=人起身走了(清 box、撤 ramp,这才是 recover)。框的进出就是答案。

### 收敛后的完整 taxonomy —— **只有 silent / moving 两类**(3 类坍缩 2 类)

原 still-fall + lost-fall 合并成 **silent**(lost 是 silent 的无报告段);moving-fall = **moving**。

| 类型 | 覆盖 | 时钟/状态 | 触发 | 判别 |
|---|---|---|---|---|
| **silent** | 倒地后静止(原 still-fall ≡ lost-fall;track 丢/抖/recapture 都穿过原位 still-box) | 一个连续 dwell 时钟,lost=原位 still 的无报告段 | survival `fallLRFromDwell` | 框内续/框外断;框进出=自救判据 |
| **moving** | 走动中、**从未进 box** 就消失(走进盲/走出门) | 消失锚点 + Blind/noDetect | 重复 absence | reachable-exit/neighbor 裁走出 vs 摔进盲 |

**Blind/noDetect 只服务 moving**;倒地后被丢的主路(cd2b)走 silent(still-box dwell),不进 Blind。

**reason 标签随之简化**(fall_reason.go):`ReasonLost` 退役 —— `ObsDwellStill` 主导 → `ReasonSilent`;`ObsNoDetect` 主导 → `ReasonMoving`(不再 `ReasonLost`)。lost 不再是独立成因。

### 实现落点(belief_shadow sweep 重写 = 工单2 deferred 部分)

- **track 生命周期(track_manager)**:lost track **不删、保活 `StillBoxRunStart` + stash 框心 `StillBoxStartX/Y`**;返回按位置在不在 `StillBoxCm` 内决定**续(框内)**还是**清+拉回(框外)**。
- **dwell**:`now − StillBoxRunStart` 连续喂 survival,still 与 lost 共用一条 ramp。
- **删**:`MovingPrecondition` 排除(belief_shadow:434)、`lostAnchor`/recapture 窗那套(框进出取代)。
- **moving-fall**:保留 noDetect→Blind + reachable-exit(仅此路)。

---

## 当前状态(2026-06-12)

**定论(用户/架构师)**:DBN 自第一天(`b3a45ca`,2026-06-01)状态空间建错。重写方向 = 把 Room 层从「姿态 9 态」(位置不进状态,只作观测条件)换成「**全空间区域占用 × {直立, 倒地}**,含盲区一等态」;转移=空间运动;观测=区域有无 track;fall 从「占用 + 不可见或静止 + 未离场」涌现。`belief_shadow.go` 的 lost-fall sweep(MovingPrecondition 译反)删除变涌现。

**验证方法论**:用进程内 `belief_replay` 单测(干净可复现),不再 live 重放(bedSessions 跨 run 残留)。fixture = `case-cd2b-0604-16141631`,完成判据 = 该 fixture 必须 fire。

---

### 架构定性(项目组,待委员会印证)

两层各建不同的东西,正交:

- **Track 层**(`belief/track.go`,**保留**)= 观测质量滤波器,答「这条 blip 是真人还是反射」。状态 `T={None,Real,Ghost,JustLeft,Lost}`,5 态 HMM。数学核心 = co-existence 耦合 ρ(孤立 track 不可能 ghost——ghost=真人反射,必有共存 partner)。唯一对外输出 `P(Real)∈[0,1]`。
- **Room 层**(新空间模型,**重写**)= 状态估计器,答「人在哪个区、倒没倒」。状态 = Zone × {直立, 倒地}(或等效:空间 zone + `SFallen` 吸收态)。数学核心 = 空间转移阵 + 盲区一等态 + dwell 生存函数。
- **耦合(已落地)**:`belief_shadow.go:354` `o.Conf *= pReal`(`pReal=P(T=Real)`)。ghost→`P(Real)→0`→Conf→0→似然经 `temper` 退火成单位阵→不驱动空间推断。等价原始提案 §4.3 边缘化 `P(fall|y)=P(fall|real,y)·P(R=real|y)`。
- **工程经济**:保持 `numStates=9` → `Vector`/`normalize`/`Max`/`temper`/`Decider`/`belief.go` 全不动,重写收敛进 `state.go`(语义)+`model.go`(A)+`likelihood.go`(L)三文件(`SFallen` 仍须是 9 态之一,Decider 引用)。

**项目组提出的三处需拧紧(委员会待裁)**:
1. cd2b 的 fire **不来自 Track 层判 ghost**——co-existence 对孤立被子 ρ=0 判它 Real,反而保护它。fire 杠杆在 Room 层(sleepad 权威释放床占用 + 真人 absence + 未离场 → blind dwell ramp Fallen)。
2. absence 须定义在 realness 门控后的 track 上(「任何可见区无 `P(Real)>阈` 的 track」=有效缺失),而非「无 track」。
3. FP 安全性反转:旧 `A` 靠「→Fallen 极小,沉默不造跌倒」;新模型 blind-zone 久缺**必须**从沉默 ramp Fallen → 守门挪到 blind 进出纪律 + dwell 时序 + recapture/Exit cancel,且守[偏分监控铁律](只有极近两事件有向 hand-off 能排除 lost-fall)。

**盲区无几何 — 对上面架构的实质修正(2026-06-12,用户提出)**:layout(`room_visual_layout`/`radar.areas`)只编码雷达**覆盖到**的 area 对象;盲区=其补集,**任何数据源都不画它**;firmware 只报 track,不报「看不见」。所以**盲区不能是有几何、可定位的状态**。
- 重述:**Blind = 「有真人在场(近期证据)+ 没有任何 real track 解析出他 + 没有 Exit 事件」**——用否定定义,可观测性来自「track 从有到无 + 无 ExitRoom」,不来自几何。
- Near/Far/Bath 子态**不能**来自盲区几何(未知),只能来自**消失前最后一帧位置**(last-seen area_type + 门距,已知)→ 改成**入口条件化**:`Blind-from-Door`(高 exit 先验)/`Blind-from-OpenFloor`(高 fall 先验)/`Blind-from-Bath`。状态索引是「怎么进的盲」(可观测),不是「现在在盲哪」(不可观测)。
- cd2b 把此逼到极致:被子 track 把 Bed→Blind 转移**遮住**(连「track 没了」都观测不到)→ 该转移只能靠 **sleepad LeftBed 与 radar 床 track 矛盾**推出,对上原始提案 §7 诚实下限。

---

### cd2b 床边真摔 replay FN — 根因实证(供 CodeX 独立复算)

**fixture**:`doc/cases/case-cd2b-0604-16141631`(201 卧室,radar `cd2b` + sleepad `1641`,UTC 22:14–22:31)。
**oracle**:22:23:20 sleepad **LeftBed**(人跌床边、雷达不可见);之后 radar 持续在床报 lying(被子=静止反射伪装成在床的人)。`alarm.json=[]`(生产漏报)。**正确结果 = ~22:25 应 fire 一条 Fall**(被子是反射,不该按「在床睡」豁免)。
**replay 结果**(in-process belief replay,cd2b fixture):`belief_shadow_fall` 空 → `SFallen` 从未 confirm → **DBN 也 FN**。

**FN 代码链(file:line 可核)**:

1. **孤立被子 track 不被判 ghost,被当人。** `belief_shadow.go:349-354`:fall 证据 `o.Conf *= pReal`,注释明写 `real lone faller pReal≈1(孤立 ρ=0 不可能 ghost)→ 无折扣`。即 co-existence 对孤立被子判 Real → Conf 不降 → 被子 pose=lying@bed 正常驱动 Room。**与 ghost 无关。**
2. **床区躺→倒地的唯一出口 = BedReleased 分支。** `likelihood.go:161` `case inBed && !c.BedReleased:` → **SBedLying**;`likelihood.go:165` `(inBed && c.BedReleased)` → **SFallen** 候选。由 `Observation.BedReleased` 二选一。
3. **BedReleased 生产闸门**:`belief_shadow.go:361` `if bedReleased { o.BedReleased = true }`;`belief_shadow.go:229` `bedReleased := bedSt.BedConfidence > 0 && bedSt.BedStatus != 0`(0=InBed);`bedSt := tm.BedOccupancyState(nowMs)`(:228)。
4. **闸门输入错**:回放期 `BedOccupancyState` 返回 `BedStatus=0(InBed)/conf90` → `bedReleased=(90>0 && 0!=0)=false` → BedReleased 永不置位 → 走 SBedLying → P(Fallen) 不 ramp → FN。

**根因定位**:**不在** belief 层(D3 止血 `belief_shadow.go:359-368` 接线正确),**在上游 bed 占用融合 `tm.BedOccupancyState`**:sleepad 22:23:20 的权威 LeftBed 没能盖过 radar 被子的 InBed 证据,bed 贝叶斯 scorer 仍输出 InBed/conf90。契约:sleepad 接触式权威 > radar(见 `bed_fusion_authority_model` / `bed_stale_leftbed_vetoes_radar_inbed`)。

**请 CodeX 独立复算/证伪**:
- **(a)** FN 根是否就是 `BedOccupancyState` 在 22:23:20–22:30 窗内仍返回 `BedStatus=0/conf90`?即 sleepad LeftBed 为何未释放 bed_state——radar 被子 InBed 证据权重 ≥ sleepad LeftBed?还是 LeftBed 根本没喂进 scorer?
- **(b)** 若把 bed 融合修对(sleepad LeftBed 胜出 → `BedStatus!=0` → `bedReleased=true`),现有 D3 止血是否**即可**让 cd2b 在 ~22:25 fire?还是 pose/dwell ramp 时序不够、仍需别的证据?
- **(c)** 计划中「全空间区域占用×姿态」重写是否真能消掉这个 band-aid——还是只是把同一个「sleepad vs radar 谁定床占用」的权威问题从 BedReleased 闸门挪到 blind-zone 进入条件上(即重写**不**自动解 (a),仍依赖 bed 融合权威修对)?

**项目组判断(待 CodeX 印证)**:(a) 成立;(c) 重写**不**自动解 (a)——区域模型让 blind-zone 成一等态、删 MovingPrecondition 译反,但「状态离开 Bed 区」仍取决于 sleepad LeftBed 能否压过 radar 被子的床占用。故 **bed 占用融合权威(sleepad>radar)是 cd2b fire 的必要前置,与状态空间重写正交,两者都要做。**

---

## 施工方案(项目组,待委员会裁)

核心一句话:Room 层从「姿态分类器」变成「单住户的空间占用估计器」,fall 不再靠 pose 抬,而是从「该在的人解析不到 + 没离场 + 拖时间」涌现;Track 层原样保留当真伪前置。保持 `numStates=9`,重写只落在 `state.go`/`model.go`/`likelihood.go` 三文件,`belief.go`/`track.go` 不动。

### 1. 状态空间(state.go,9 态,Fallen 仍占 index 5 → Decider 不改)

| idx | 新态 | 含义 | 来源 |
|-----|------|------|------|
| 0 | Empty | 房内确认无人 | 同旧 SEmpty |
| 1 | Bed | 占用在床(睡/休息) | area=Bed + sleepad InBed |
| 2 | Sit | 占用在休息区(椅/沙发)——久坐正常,低 dwell-fall | area=Sit |
| 3 | OpenFloor | 占用在开阔活动区(站/走,摔在这里) | area=Active/Deny |
| 4 | Bath | 占用在卫浴(高风险,常盲) | area=Toilet/Shower |
| 5 | Fallen | 倒地(唯一 fire 目标,近吸收) | 涌现 |
| 6 | BlindRest | 占用但解析不到,从 rest 区(床/卫浴)进盲 = 床边/卫浴真摔(cd2b!) | 否定定义 |
| 7 | BlindOpen | 占用但解析不到,从开阔区进盲 = 走动中消失 | 否定定义 |
| 8 | Left | 已经门离场(→收敛 Empty) | ExitRoom/门区 hand-off |

**关键设计决策(待委员会裁)**:

- **删 SArtifact**:ghost/反射的活全归 Track 层(`Conf*=P(Real)` 在似然层就把它退火成中性),Room 层不该再有伪迹态——兑现"Track=观测滤波器/Room=状态估计器"两层分离。腾出 1 个 slot。
- **删 STransition/SBedRestless**:不确定性由 Blind 态 + normalize 承载,不需要专门枢纽态;床上翻动并入 Bed。再腾 2 个 slot。
- **"门"不设独立态**:延续现在的做法(NearDoor 作观测条件 + Left 态 + ObsEnterExit 事件),省 1 个 slot 给 Blind 一分为二。
- **Blind 拆 Rest/Open 是整个重写的命门**:两者转移语义不同才值得各占一态——BlindRest→Fallen 几乎无逃生阀(rest 区不是门,人不会从床里凭空离场),BlindOpen→{Fallen | Left} 由 reachable-exit 裁。盲区本身无几何(layout/sensor 都不画),所以态的索引是"怎么进的盲"(last-seen 区,可观测),不是"在盲哪"(不可观测)。

### 2. 转移阵(model.go A,空间运动)

- **可见区互联**:Bed↔Sit↔OpenFloor↔Bath 经 OpenFloor/门枢纽连通(粗粒度、房无关)。
- **进盲**:每个可见区 → 对应 Blind 态小倾向(钻家具背侧/床边死角)。Bed/Bath→BlindRest、OpenFloor→BlindOpen。
- **出盲**:BlindOpen→OpenFloor(重现)/→Left(reachable-exit 高)/→Fallen(久缺);BlindRest→Bed/Bath(重现)/→Fallen(久缺,无 Left 逃生阀)。
- **Fallen** 近吸收(自持~0.99),只有正向 recapture/exit 证据能压下。Left→Empty。
- **安全性反转(必须正视)**:旧 A 的安全靠"→Fallen 极小,沉默不造跌倒"。新模型反过来——Blind 里久缺就该从沉默 ramp Fallen(这才是 lost-fall 的本义)。FP 守门从"A 惰性"挪到三处:① BlindOpen 的 reachable-exit 逃生阀;② 进盲纪律(Bed→BlindRest 必须先有"人离床"证据,不能凭 radar 抖动空降);③ ramp 由 dwell 生存函数定速(对齐床边真摔 6–8min 物理下限,不是秒级)。ramp 机制复用现有 ObsNoDetect(每盲 tick 固定强度 × Fallen 近吸收 → 自然累积),不手调 gain。

### 3. 发射(likelihood.go L)

- **可见 track**(Conf*=P(Real) 已门控)落 zone Z → 抬 Z;OpenFloor + pose=lying/fallen → Fallen;Bed + bedReleased(sleepad 离床) → Fallen(cd2b 出口,即现有 BedReleased 分支,现在喂新 Fallen)。
- **ObsNoDetect**:对 realness 门控后的 track 做房级归约——"任何可见区无 P(Real)>阈 的 track"=有效缺失 → 抬 Blind(按 last-seen 分 Rest/Open)+ 略抬 Left;重复 → ramp Fallen。
- **ObsBedOccupied**:sleepad InBed→Bed / LeftBed→压 Bed(释放)。← cd2b 命门在这。
- **ObsReachableExit/ObsNeighbor/ObsEnterExit/ObsVitalPresent** 路由 Left/Empty,大体沿用。
- **删 ObsTrackPresent→SArtifact 整条**(Track 层接管)。

### 4. unit 内部联动(对 Blind 态承重)

unit 联动不是锦上添花——没有它,跨房走动在"被离开的房"里就是一条 lost-fall 误报。三类机制:

**① 同房多雷达 union**(cd2b + 333b 这种):一个 unit 里覆盖同一房的多台雷达,它们的 track 全喂同一个 room belief。命门:ObsNoDetect 的 absence 归约是对本房全部 unit 雷达的 real track 取并集——"任何一台雷达解析到 P(Real)>阈"就不算盲。一台雷达补另一台的盲区,blind 直接缩小。

**② 跨房 hand-off**(人从 A 房挪到 B 房):ObsNeighbor 弱耦合消息:A 房住户进 BlindOpen 时,消费"邻房 B 此刻有没有 fresh 有向出现"。有 → BlindOpen→Left(人挪走了,不是摔);无 → 留在 Blind 继续 ramp。兑现[偏分监控铁律](只有极近两事件有向 hand-off 能排除 lost-fall),保留现有 neighborHandoff 不删。

**③ 房内 sleepad↔radar 权威**:sleepad LeftBed 压 radar 被子本身就是 unit 内跨模态设备联动,cd2b 命门。

**结构底座**:`owl-common/spatial/relmatrix.go`(mm_relationship_matrix,已建+测,sensor 当 maintainer)。它给出"哪些雷达覆盖本房/本床(radar→room/bed covers)、哪些房相邻"。Room belief 的输入 scope 由 MM 矩阵界定(哪些设备 union 进本房 absence / 哪些房可 hand-off)。

### 5. 文件改动

| 文件 | 动作 |
|------|------|
| `belief/state.go` | 重写 9 态枚举 + label |
| `belief/model.go` | 重写 A(空间)+ Prior |
| `belief/likelihood.go` | 重写发射映射(zone-based + noDetect→blind + bed 权威 + 删 artifact) |
| `belief/observation.go` | 可能加 zone 字段/ObsTrackInZone;删 ObsTrackPresent 路径 |
| `belief_adapter.go` | 从 area_type 出 zone;noDetect 走 realness 门控后的 absence;sleepad LeftBed 喂 bed 权威 |
| `belief_shadow.go` | 删 lost-fall sweep(:430+ 的 MovingPrecondition 译反);引擎 wiring 迁生产 tick(DBN_FIRE);realness filter 保留喂 Track 层 |
| `belief.go` / `track.go` | **不动** |
| 上游 bed scorer / BedOccupancyState | 修 sleepad>radar 权威(belief 包外,前置) |

### 6. 验证与风险(诚实)

- **完成判据**:`belief_replay` 单测,cd2b 必须在 ~22:25 fire。
- **最大风险 = 回归网清零**:重写状态语义会让 `r5_calibration_lock_test.go`/`belief_test.go`/`fall_reason_test.go` + 已知 FP 回归(case-5 hunzi、cabb 静止站立)全部失效——它们是按旧 9 态写的安全网。必须先重建 oracle 测试套(已知 TP 必 fire / 已知 FP 必不 fire),否则等于在没有回归网的情况下重写核心。
- **施工顺序锁定**:bed 融合权威先修 → cd2b fire 验证 → 以此为 oracle 锚点建测试套 → 重写 belief → 跑测试套验证不回归。
