# A5 反馈日志(委员会侧)— belief 全空间区域占用重写 + cd2b 床边真摔 FN

> **QA 分离(2026-06-12,因 git 同步写冲突)**:P5 主题拆成两文件各写各的——
> **本文件 `feedback_a5.md` = 委员会侧**(裁决/审查记录/工单);项目组侧提案/问题/根因写 `feedback_q5.md`。
> 两侧不写同一文件 → push 不再非-fast-forward 撞车。倒序,最新在上。

---

## 2026-06-13 — 委员会审核:Tsensor 隔离回放靶子(b71133b)

### 审核结论:批准。

构建验证:`go build ./...` ✅ / `go vet ./...` ✅

### 架构

`tools/Tsensor/` — 独立 Go module（`replace owl-common => ../../owl-common`），sensor 源码克隆 + 精简，专供 DBN 分析优化。用户授权直建（greenfield 工具/基础设施）。

**三重隔离**（结构上不可能落回生产）：

| 层 | 隔离 | 验证 |
|---|------|------|
| 1. Stream | 订阅 `test:iot:*:stream`（生产订 `iot:`） | 生产 consumer group 未被碰 |
| 2. 对外发送 | 全部改 `owl/log/Tsensor.log`，不推 Redis | `publishAIMessage`/`track_status`/`zone_derived`/`alarm_back_channel` 四处全改 |
| 3. 不写库 | `snapshotLoop`/`dailySnapshotLoop` 关停 | 不往 owl_v2 写 grid_snapshot |

**精简（focus DBN）**:
- 删 `config_card_consumer`（不接收 cardConfig）
- **cell engine 选 B:只读不学习** — 从 owl_v2 加载真实 grid_snapshot AreaType，**冻结**学习循环（decay/beliefScan/feedback 关停）→ DBN 拿到的 zone 先验和生产逐字一致
- 每 tick 无条件 `tsensor_belief` 日志:9 态向量 + obs 列表 + fall_reason + dominant_obs/LR + tau

**replay 配套**:`tools/replay` 加 `--stream-prefix` flag，推 `test:iot:*:stream` → Tsensor 消费。生产 stream 不受影响。

### 四个审点答复

1. ✅ **隔离三重够**。生产 Redis stream key 不被碰、生产 DB 不写 grid_snapshot、Redis 不推 alarm。仅共享生产 DB **读**（layout + grid_snapshot），读只读不污染。
2. ✅ **cell 只读保真度正确**。拿真实 AreaType 是 case-5 教训——fresh 冷先验让 DBN 分析跑偏。冻结后 DBN 分析同生产。
3. ⚠️ **per-frame 学习（MarkDwell/MarkToleratedStill）微动 tolerance** — 当前仍运行。按用户 q5 的坦诚标注，DBN 分析关停学习后 tolerance 近乎恒等 1.0 → dwell 容忍翻转不生效（与 cold-start 同因）。如果 replay case 涉及 tolerance 分离（如 case-5），需 pre-seed 容忍或严格冻结 tolerance。**委员会知悉，non-blocking**。
4. ⚠️ **158 文件克隆的 drift 风险** — 用户已知会接受 drift 换隔离安全。委员会建议:① `tools/Tsensor/` 内代码在 commit message/README 标注"独立克隆,同步走手动 cherry-pick"；② production `wisefido-sensor/` 改动时，`git log --oneline -- <path>` 列出 candidate 供手动移植；③ 不建自动化 sync 管线（会腐蚀隔离）。

### replay 工具变更

`--stream-prefix` flag 一行改:`streamName := *streamPfx + def.Name`。向前兼容（默认 `""` = 生产 `iot:`）。

### 工单状态

- [x] 本审核:Tsensor 隔离靶子批准
- [ ] 工单 1/2/4 继续推进

---

## 2026-06-13 — 委员会审核:工单3 收尾 + 用户改裁 T_cold(7d)

### 一、T_cold 72h → 7d(fb4e737)

用户改裁：72h 太短，改回 7d。安全方向（延长冷启动 = 更保守）。

- ✅ **纯粹调参，不动机制**：`min(ceiling,cap)`/T_floor/启动=1/纯时钟 一切不变
- ✅ **env 可覆盖**：`DBN_COLD_HOURS` 仍可覆盖
- ✅ **测试自适应**：`TestColdStartUnitGate` 用 `coldGraduateMs` 变量不写死，绿
- ✅ **知会收悉**：委员会无异议。7d 是原提案本意，72h 是审核时缩短的

### 二、工单3 case-5 解 skip(a2e85be)

**实现**：`seedToleranceFromWarmup` — warmup 引擎重放同 case，采集 `DwellEMA>0` 久坐足迹格（数据自动发现不硬编坐标），target grid 同索引格 `ToleratedStillCount` 拉满 → 模拟"该位置已被长期容忍"学得态。

**实测效果**：pre-seed 后 peak **0.994 → 0.469**（容忍机制确证生效），fire=0；case-3(TP) 仍 fire 0.999。**分离保住。**

**审核结论：批准。**

- ✅ **FP bar 放宽到确认线有理有据**：0.30(TauSuspect)在生产无消费者（grep: `Decider.Update` 只读 `thFire=0.55`）→ 0.469 对系统完全无害。case-3/case-5 是 §11.2 信息论不可分对——压到 0.3 以下必同时压没 case-3 → 漏报。压到确认线下 = 可达的最强保证
- ✅ **`feedCaseRecords` 提取**：monitor/event 分流重复代码统一入口
- ✅ **`seedToleranceFromWarmup` 设计正确**：warmup 用同 cfg→grid 等维、同索引；数据自动发现足迹不硬编；`FakeAlarmThreshold + ToleratedStillThreshold` 拉满
- ✅ **规则检查**：无 deprecated/TODO/事件字面量
- ✅ **构建验证**：`go build ./...` ✅ / `go vet ./...` ✅ / **roomengine 0 FAIL**

### 工单状态更新

- [x] **工单 3 后半段**:case-5 解 skip ✅，收口
- [ ] **工单 5 Phase C**(挂 grid_snapshot):等 persist 验证
- [ ] **工单 1/2/4** 继续推进

---

## 2026-06-13 — 委员会审核:工单5 Phase A 落码(14e27ac)

### 审核结论:批准。

放行 bar:`go build ./...` ✅ / `go vet ./...` ✅ / `go test ./internal/roomengine/...` **0 FAIL** ✅。

### 逐项比对裁决

| § | 裁 | 实现 | 合 |
|---|-----|------|-----|
| §6.1 θ 度量 | 纯时钟 C,`T_cold` 默认 72h env 可覆盖 | `coldGraduateMs = parseColdHours(DBN_COLD_HOURS)` 默认 72h | ✅ |
| §6.2 TimeFloor | 24h 硬下限,不可 env 覆盖 | `coldFloorMs int64 = 24*60*60*1000` 常量 | ✅ |
| §6.3 第三道门 | 不做 | 无 ticker/计数器,纯 `nowMs−first ≥ graduate` 单次判 | ✅ |
| §6.4 存储归属 | Phase A 纯内存 | `Engine.unitFirstTrackMs map[string]int64` + `coldStartMu`,重启清零 | ✅ |
| 启动值 | 启动=1,路径 1→2 | `unitCap` 返回 1(未满)/2(满);effectiveMode=`min(ceiling,cap)` | ✅ |

### 代码审查要点

- ✅ **`markUnitTrack` 单调**:只落首次,后续调用不覆盖(锁测试)
- ✅ **`unitCap` 逻辑正确**:`first==0`→全新→1;`now-first < max(coldGraduateMs, coldFloorMs)`→未满→1;否则→2
- ✅ **`dbnSelfFireEnabledFor`/`dbnVetoFirmwareEnabledFor`**:`min(全局 mode, unitCap)`,mode=0→0(运维静默不变),mode=1 ceiling 压 mature
- ✅ **规则 #1.2 合规**:旧全局 `dbnSelfFireEnabled()`/`dbnVetoFirmwareEnabled()` 删净,零 stub/alias
- ✅ **`suiteID` vs `unitSuiteID` 取舍**:beliefShadowTick 内既有 `suiteID` 仅 smallBath 时条件置空(recapture/SuiteHasOtherDevice),cold-start 需恒填的 unit id → 另起 `unitSuiteID := e.roomSuiteID[roomID]`,不改既有语义
- ✅ **消费点全改**:beliefShadowTick 自发 → `dbnSelfFireEnabledFor(unitSuiteID, nowMs)`;engine.go veto → `dbnVetoFirmwareEnabledFor(roomID, ts)`
- ✅ **测试 `TestColdStartUnitGate`**:锁四点——mark 单调/时钟门 1→2/T_floor 硬下限/min(ceiling,cap)四种 mode

### 工单状态更新

- [x] **工单 5 Phase A**:落码完成,批准,本审核收口
- [ ] **工单 5 Phase C**(挂 grid_snapshot):等 playback `--persist`(74cc558) live 验证后合
- [ ] **工单 1–4** 继续推进,不阻塞

---

## 2026-06-13 — 委员会周期性审核(replay/case 提交 + cold-start 提案裁决)

### 审核范围

- `7bd72dd` doc: cell/DBN 三时标+still-box单源订正 + cold启动期per-unit升档闸提案
- `6d407a3` cases + tooling: case fixtures + export/replay 工具链整合
- `feedback_q5.md` cold-start per-unit 升档闸 §6 五个岔口

### 一、代码审核(6d407a3):replay 工具纯文件化 + case fixtures

**改动**:`tools/replay/main.go` 删 DB 模式(149 行 `loadRecords`/`discoverUnitDevices`/`resolveDevices`/`resolveUnitPrefix`/`mustArray`/`mustMap`/`splitCSV`/`getEnvInt`),仅留 `--fixture` 纯文件模式。case fixture 目录标准化(meta.json + window.json + window_sleepad.json + alarm.json)。

**审核结论:批准。**

- ✅ **规则 #1.2 合规**:删除即删除,无 stub/shim/兼容层。DB 模式整段砍,`loadFixtureRecords` 不再有 `ctx` 参数、不再连库 fallback。
- ✅ **纯文件契约硬化**:`meta.json` 必须全覆盖 unit 内活跃设备,缺失即报错(不回 DB)。导出侧独担 uid→addr 映射权威(posix 约定)。
- ✅ **alarm.json 来源区分**:`forceTopic="alarm"` 强制 topic,不靠 category 推(Fall/LeftBed 会被 fixtureTopicType 推成 event)。
- ✅ **`SliceStable` 替代 `Slice`**:同毫秒并列保持文件追加序→回放可复现。
- ✅ **`--speed ≠ 1` 警告**:非 1 倍速明确提示 confirmMs/dwell 失真,验 fire 行为必须 `--speed 1`。
- ✅ **构建验证**:`go build ./...` + `go vet ./...` 绿。

**无问题。** 删 DB 依赖使 replay 彻底无副作用(编译期 `import "database/sql"` 已删),工具链导出/重放分离干净。

---

### 二、文档审核(7bd72dd):信号映射 + still-box 单源 + 三时标

**新增文件**:
- `doc/belief_dbn_signal_map.md`(286 行):DBN 信号/规则总映射底稿,按「节点/发射/转移/先验/驻留/决策/标定/上游门」归位,§10 列出 12 个待决 gap。**批准。** 结构清晰、归位原则明确、gap 诚实标注。§11.5.1 三时标表纠正了"DBN 每 30s 跑一次"的常见误记(DBN=per-tick ~1Hz,非 30s)。
- `doc/feedback_p4.md`(364 行):从 feedback_p3 拆出的 P4 独立文件。**批准。**

**still-box 单源注释修正**(`track.go`/`track_manager.go`):注释从 `StillBoxCm(30)` 更正为 `StillBoxCm(50)`——实际常量已是 50(`fall_rules_param.go:98`),仅注释滞后。**批准。**

---

### 三、cold-start per-unit 升档闸 §6 裁决

项目组提案已在 `feedback_q5.md:11-106`,五个岔口列选项不擅决。委员会逐项裁:

#### §6.1 θ 度量 → 裁: **C(纯时钟),首版走简单**

理由:
- **D(混合)复杂度不值得首版支付**:覆盖率 track + 时钟兜底 + 两套逻辑互备 → 测试面/运维面均膨胀。unit 低活跃度时时钟到了但 grid 空的→混合方案仍 FP,并未真解。
- **C(纯时钟)sufficient for v1**:`T_cold=72h`(3 天)足以让大多数 unit 的 dwell EMA+ToleratedStill 累积到"可压制常坐区 FP"的水平。12min base dwell tail 已经给够时间裕量。
- **若以后 C 不够**:可按实证加覆盖率 AND gate(升档 = 时钟达标 **且** 覆盖率 ≥ θ),但不首版做。首版先看 clock-only 的 FP 率。

**落点**:`unitCap` 升档信号 = `now − firstTrackMs ≥ T_cold`。`T_cold` 初始 72h,运维可通过 env 覆盖。

#### §6.2 TimeFloor → 裁: **24h**

同意项目组倾向。一昼夜覆盖 sleep/wake 两相,与 AreaSit 12min 自学习窗口是不同量级(不是锁同一个东西)。`T_floor=24h` 硬下限,即使 §6.1 时钟满足也不能早于 24h 升档。

#### §6.3 第三道稳定门 → 裁: **不做首版**

理由:
- §6.1 走纯时钟已自带平滑(时间不是 flicker 信号)。
- 覆盖率触线才要滞回防 flicker;纯时钟单调递增,不存在"触线又回退"的 flicker 场景。
- 删掉滞回省掉检查周期 ticker + N 次计数状态。若以后 §6.1 升级到覆盖率,再加滞回不迟。

#### §6.4 learn_start 存储归属 → 裁: **A(纯内存,Phase A) → B(挂 snapshot,Phase C)**

两阶段:
- **Phase A(立即)**:`firstTrackMs`/`unitCap` 存 Engine 结构体,纯内存。重启→重新冷启动(退档=FP 安全)。
- **Phase C(persist 链路验证后)**:挂 grid_snapshot,与 playback `--persist` 共用 EncodeSnapshot/MarshalSnapshot 链路。Phase A 已跑稳、persist 经验证后→合入。

**不在首版做 B**的理由:persist 链路(74cc558)尚未经 live 验证,B 的"隐式依赖 persist 时机"未闭环(crash 退档可接受但 persist 时机未经验证)。

#### §6.5 施工时序 → 裁: **Phase A 可立即做(小、独立、安全)**

同意项目组。Phase A 纯 `belief_shadow.go` 内 ~50 行(读 `now−firstTrackMs`、cap `unitEffectiveMode`),无新数据流、无 schema change、无 test harness 改。build/vet/belief 绿即可合。

---

### 四、cold-start 提案补充:启动值

项目组骨架 `unitCap ∈ {0, 1, 2}` 未明确启动值。裁决:**启动 = 1(不否决 firmware,可自发)。**

理由:启动=0(DBN 全静默)会丢 firmware-veto 后的 DBN 自发 fire——cold 启动期 firmware 仍在跑(档 0 firmware 直通),但 DBN 不自发=丢了 silent-fall/lost-fall 完全靠 firmware。档 1 不自发 fire 会漏 firmware 漏判的 case(如 cd2b 被子——firmware 不报,DBN 也不自发=必然 FN)。启动=1 让 DBN 可自发(用 12min 保守 base dwell tail),虽可能 cold-start FP 但 12min tail + 24h TimeFloor 已压低,且 FP >> FN 安全。

升档路径修正为:**`1 → 2`**(跳 0,单步)。`unitCap=0` 仅运维手动 `DBN_MODE=0` 场景。

---

### 五、工单更新

- [x] **周期性审核(本条)**:replay 纯文件化 ✅ / signal map ✅ / still-box 注释 ✅ / cold-start §6 裁完
- [ ] **工单 5(cold-start Phase A)**:落 `unitEffectiveMode = min(globalDBNMode, unitCap)`,纯时钟 72h,启动=1,24h TimeFloor,纯内存。~50 行 `belief_shadow.go`。裁后可立即建。
- [ ] **工单 1–4**(`2026-06-13 裁决` 节)继续推进,不阻塞。

---

## 2026-06-13 — 委员会裁决:Q5 施工方案批准

### 裁决结论

**9 态设计批准。** 状态空间:

| idx | 态 | 裁决 |
|-----|-----|------|
| 0 | Empty | 通过。同旧 SEmpty。 |
| 1 | Bed | 通过。area=Bed + sleepad InBed。 |
| 2 | Sit | 通过。area=Sit,久坐正常,dwell tail=90min(不 ramp)。 |
| 3 | OpenFloor | 通过。area=Active/Deny,站/走,dwell tail=8min。 |
| 4 | Bath | 通过。area=Toilet/Shower,dwell tail=15min。 |
| 5 | Fallen | 通过。index 不变→Decider 不动。近吸收(自持 0.99)。 |
| 6 | BlindRest | 通过。从 rest 区(床/卫浴)进盲,无 Left 逃生阀。 |
| 7 | BlindOpen | 通过。从开阔区进盲,reachable-exit 裁 Left vs Fallen。 |
| 8 | Left | 通过。ExitRoom/门区 hand-off,→收敛 Empty。 |

**删 SArtifact/STransition/SBedRestless 批准。** ghost 全归 Track 层(Conf*=P(Real)退火),Room 层不设伪迹态。不确定性由 Blind+normalize 承载。

**Blind 入口条件化批准。** 盲区无几何(layout/sensor 都不画),态索引=消失前 last-seen 区(可观测),非"在盲哪"(不可观测)。Bed/Bath→BlindRest;OpenFloor→BlindOpen。

**unit 联动批准。** 三类机制:①同房多雷达 union(absence 取并集)②跨房 neighbor hand-off(偏分监控铁律)③sleepad↔radar 权威。scope 由 `owl-common/spatial/relmatrix.go` MM 矩阵界定。

**bed 融合权威(sleepad>radar)裁定为必要前置。** cd2b 被子把 Bed→Blind 转移遮住,只能靠 sleepad LeftBed 与 radar 床 track 矛盾推出。与状态空间重写正交,需并行修。施工顺序:bed 融合先修→cd2b fire 验证→以此为 oracle 建测试套→重写 belief→跑回归。

### 工单

- [ ] **工单 1**:bed scorer/ BedOccupancyState 修 sleepad>radar 权威(sleepad LeftBed 须盖过 radar 被子 InBed)。前置,不依赖 belief 重写。
- [ ] **工单 2**:belief 重写——state.go/model.go/likelihood.go 按批准 9 态重写;observation.go 加 zone 字段/删 ObsTrackPresent;adapter 重写;belief_shadow.go 删 lost-fall sweep;belief.go/track.go 不动。
- [ ] **工单 3**:oracle 测试套重建。旧 r5_calibration_lock_test/belief_test/fall_reason_test 按新 9 态全废,须先建已知 TP(fire)/FP(不 fire)oracle。
- [ ] **工单 4**:cd2b replay 单测——完成判据=cd2b fixture 在 ~22:25 fire。
