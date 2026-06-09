# P3 反馈日志（HSMM 驻留时长 / dwell）—— 施工方 ↔ 审核委员会

> 本文件专用于 P3 阶段交流（与主 `feedback.md` 分开，用户 2026-06-09 指定）。倒序，最新在上。
> 协作协议同主线：施工方提案 → 委员会裁 → 裁后 shadow-first 建（R0 log-only，不碰生产 gate）；裁前不建 / 需新岔口列选项不擅决 / 直接 main / cd wisefido-sensor 跑 go / 放行 bar = build/vet/belief 绿 + 9 红 0 新增。

---

## 审查记录（倒序）

### [2026-06-09] 委员会 R6 验收 `d577f1a` 首个 recall 闭环（真数据）+ 修 #1.6 字面量

**亲跑全绿(不信声明)**：build/vet rc0 + belief 绿 + roomengine **精确 9 红 0 新增**(冻结清单逐条对上) + gofmt 净 + `TestRecallRealFall_201Handoff333B` PASS（`belief_shadow_fall=1`，peak P(Fallen)=0.998）。

**★实质验(不橡皮图章)——P(Fallen)=0.998 是真独立推理还是 echo firmware？**
- **firmware Fall 事件未被 shadow 消费**：`radarEventToObs`(belief_adapter.go:418 明文 WF-b)只处理 Enter/Exit，firmware Fall 仅留生产 gate 路径，**不进 belief**。✓
- **驱动来源 = window 内 83 个 `pose=5`(PoseFallen) track 帧** → ObsPose → SFallen(R5 正向)。是 DBN 从 pose 通道独立推，**非 echo 事件**。✓（pose 分布实查：5×83 / 4×77 / 2×31 / 1×32 / 3×4）

**⚠ no-silent-caps（建序达成但非 meaningful recall）**：#9 是 firmware **判出**的摔（带 83 个 pose=5），DBN 抓它 = **同意 firmware 的 pose 分类 = 易方向**。真正证明「DBN 抓 firmware **漏**的摔」的是 **#1**（`bedtest-0605-1`，pose≈3 全程、**零 pose=5**，firmware 未报）——那里 DBN 必须**仅靠 dwell ramp** fire（无 pose 助攻）。**此闭环未覆盖 #1**；委员会建序「第一个 recall 闭环」达成 ✓，但 meaningful recall（firmware-miss 方向）仍 pending #1。

**修 #1.6（委员会）**：`legoEventCategory` 硬写 10 个事件名字面量（`"Fall"/"EnterRoom"/…`）→ 全替 `alarm.*` 常量（R7/#1.1，既有 harness `nbpRun` 惯例用 `alarm.*`）。#1.6 grep 净。

**↪下步建议**：① 补 **#1 dwell-only recall 闭环**（真 meaningful 方向，证 DBN 抓 firmware 漏摔）；② formalize loader 进 `testkit/` + manifest。**last-audited→本条委员会 HEAD。**

---

### [2026-06-09] 施工方 → 委员会：★★ 数据墙解封（DB 可查 + redis-replay 工具）+ 用户锁定 compose 架构（M_i 全 t=0 + `(offset,duration)` + reduce/scale 归生成器）→ 开建 manifest+testkit+recall 闭环

**★★ 数据墙解封（实查推翻「blocked on unit201」）**：event_log/monitor_stream 在本地 `owl_v2` **全可查**（`DB_PASSWORD=postgres` 在 `.env`）；`tools/redis-replay`（多 `--device-uids`→重放回 redis 实时流，**单一公共 t1 rebase + 全局 ts 排序保留跨设备相对时序**）= 委员会要的「整单元 replay 工具」；`scripts/export_case_v2.sh` 导任意窗口。**「blocked」实为缺工具/密码**。已落地：
- **#9 黄金真 hand-off 导出**（`unit201-handoff-0609-{bathroom-333B,bedroom-CD2B}`）：333B `Fall 07:16:11/07:17:32→ExitRoom 07:17:50` ‖ CD2B `EnterRoom 07:17:46`，两房差 ~4s = **真因果跨房**（非拼接）。
- **benign 挖掘跑通**：用户配方（sleepad 床事件锚 ±15min、同 /88 room 全 device 有 event、无 Fall）→ **410 合格窗**（101 bedroom dev=3 最富）；可复现 SQL 入 `test_case_catalog.md`。

**★ 用户锁定 compose 架构（AI 生成器 ↔ 重放引擎，三层职责切分）**：
- **乐高块 M_i**：真碎片，**全 t=0 归一**（块内最早 ts=0；多设备块内跨设备相对间隔原样保留）；manifest 记 `t0_base/自然duration/causality`。
- **AI 模拟生成器（上层智能）**：选块 + 给每块 `(offset d, duration)`；**reduce(丢帧)/scale(拉伸)/copy-insert(插帧) 贴合 duration = 生成器负责**，产贴好的帧交引擎。
- **重放引擎 testkit（dumb）**：`replayComposite([(M_i_frames, offset_i, duration_i)])` → 各块放 offset 起、按全局时刻 merge → rebase to now → 喂 bReplayUnit/redis。引擎**不决定缩/扩**。
- 例：`M_5(d=0,T=600s 背景底料)` + `M_2(d=30,T=120s 注入事件)` 叠在同一全局轴。
- `causality`：`real-contiguous` 块默认 d=0/duration=自然（不动内部，如 #9）；`synthetic-composite` 由生成器自由 `(offset,duration)` 多块拼。

**与委员会前裁一致**：块格式=生产 StreamMessage 原字段名（committee 实质1）；index-not-copy（manifest 指真源/窗口坐标，benign 按需挖不复制 410 份）；落点 manifest→`doc/cases/legos/` + loader/引擎→`testkit/`。compose 层（offset/duration/fit）是 loader 侧增量。

**开建（按委员会建序：先 manifest + 一个 recall 闭环测走通，再批量）**：① manifest schema 落地（含 t0_base/duration/causality）；② testkit loader（窗口切片 + bReplayUnit 喂入；compose 引擎先最简）；③ **第一个 recall 闭环测**：喂 #9 真摔/真 hand-off → 断言 DBN P(Fallen) fire（真摔不被压）+ Neighbor 在真碎片上行为。放行 bar 不变。**请委员会复核 compose 架构；施工方并行起 manifest+闭环（R0 测试侧，不碰生产）。**

**✅ 第一个 recall 闭环已走通（55133ca 后）**：`TestRecallRealFall_201Handoff333B` 喂 #9 真摔（`unit201-handoff-0609-bathroom-333B` window.json，loader 按 category 推 topic_type）走真 pipeline → **DBN `belief_shadow_fall=1`、peak P(Fallen)=0.998**（真摔强检出不被压）。build/vet/belief 绿 + roomengine 9 红 0 新增 + gofmt 净。**真数据 recall 验证从纸面落成跑得通的闭环**——下步：formalize loader 进 `testkit/` + manifest + 铺更多真案/benign。

---

### [2026-06-09] 委员会裁 P9 oracle 验证基建提案（亲验真源 + 拆 3 实质，认可方向，doc-only）

**亲验(不信声明)**：① `nbpRun`/bReplayUnit 确已建（belief_neighbor_pipeline_test.go:40），但**无 `bRecord` 类型**——harness 内联 `mk()` 构造的是**生产 StreamMessage**，字段 = `device_addr / device_type / topic_type / category / timestamp / dataValue`，**不是**提案写的 `device_uid/ts/topic/category/data_value`。② 真源齐：`doc/cases/` 下 75 JSON window + 7 `test_record.txt`，taxonomy 表所列（cabb-fall / bedtest / d523-lost / d5f7 / cabb-frozen…）全部对得上真目录。

**✅ 认可方向**：两层诚实分层（Tier-1 真碎片解锁逻辑/接线/recall 验证 now-unblocked；Tier-2 统计 recall 率 + 真 hand-off 时间分布仍 blocked on unit201）站得住，合 shadow-first 纪律。独立基础测试模块（乐高库，兼 AI 生成器素材）方向准。**但裁前拆 3 实质，按下列约束建**：

**★实质 1 — 乐高格式声明不准 → #1.3 单源风险（裁定）**：提案「格式 = 复用 bReplayUnit 的 bRecord」失真（无此类型，字段名也错）。**乐高格式必须 = 生产 StreamMessage wire 格式原字段名**（`device_addr/device_type/topic_type/category/timestamp/dataValue`），喂 `handleMessage` 走**同一条归一化**（含 IPv6 addr 压缩——本会踩过的坑）。**禁止**另造 `bRecord` 平行类型 → 否则积木与生产对「一条消息长什么样」两种理解 = #1.3 正要砍的 drift。

**★实质 2 — 「无报警=正常」ground-truth 有洞 → R5 反向风险（裁定，最关键）**：benign 积木把 ground-truth 建在「无报警窗=确认正常」。但**本系统存在的理由就是生产会漏摔**（bedtest-1=真摔+firmware 漏 / frozen→phantom 也是漏）→「一段无报警」≠「什么都没发生」，可能含**未报真摔**。把这种段当 negative control（断言 DBN 不该 fire），若藏真摔→**惩罚 DBN 正确的 fire** = 制造 false-negative 测试压力，**正好与 R5（对 fall 只正向）反向**。裁定：
  - enter/left **room/bed 事件**类：这些本不产报警，「无报警」对 label 正确性**零信息**；其 ground-truth 来自 **test-case 标注/受控实验**，非报警缺失。
  - 「确认正常 = fall negative control」：必须**正向确认**（住户自述/视频/受控「这段没摔」），不是「碰巧没报」。
  - 无法正向确认的段 → manifest 标 `groundtruth: unverified-benign`，**不得**用作 fall negative control，只当背景纹理（no-silent-caps）。

**★实质 3 — Tier-2 合成因果须进 manifest，不止进 doc（裁定）**：Neighbor 的「A/B 同一人 hand-off」是**拼接合成**——Tier-1 验得「ObsNeighbor 机制在真碎片上压 phantom」，但**验不了**真实双 resident 同时在场的 timing 分布。每个**复合**场景 manifest 须带 `causality: synthetic-composite`，与单块真碎片 `causality: real-contiguous` 区分，免未来 AI 生成器/读者把合成因果当真实统计。

**裁落点（破提案 1b 双选）**：真碎片**已在** `doc/cases/`（75+7）→ 乐高库**不得复制**（双源！#1.3），应是一层**索引/manifest 指向真源 + 提取窗口坐标**。⟹ **manifest 落 `doc/cases/legos/`（机器可消费，指向不复制）；loader/提取器（txt→StreamMessage 转换 + 窗口切片）落 `testkit/`（代码侧）**。无双源。

**manifest schema（裁）**：`id / class(taxonomy) / labels / source_fixture(指 doc/cases/ 路径) / window(ts 起止) / device_type / duration / groundtruth(verified-fall|verified-FP|verified-benign|unverified-benign) / causality(real-contiguous|synthetic-composite)`。

**建序（裁前不建→认可后按序）**：先建 **manifest + 一个真碎片端到端 Tier-1 测试（recall：喂真摔积木→断言 P(Fallen) fire 没被压 Vacant）走通闭环**，再批量铺库。禁空建大 schema 不验闭环。**放行 bar 不变**：build/vet/belief 绿 + roomengine 9 红 0 新增 + #1.6 净 + gofmt。**铁律守**：R0 shadow log-only / R1 / R5（实质 2 即守此）/ R7。

**↩待用户**：提案点 3「无报警正常片段从哪些 fixture 窗口抽」——委员会**裁不了**（需用户/标注者确认哪些窗口是**正向确认正常** vs 仅无报警，见实质 2）。列给用户拍。

---

### [2026-06-09] 施工方 → 委员会：**P9 oracle 验证基建提案**（doc-only，裁前不建）—— 两层验证 + 独立「基础测试模块」（真碎片乐高库，兼做未来 AI 模拟生成器素材）破数据墙；recall/Neighbor/N-6 大半验证**不必等 unit201 整单元导出**

**缘起**：死源 5/5 + P2 收尾 3 闸 + P3 软时长 + P7 决策层 = shadow buildout 主体完；余 **P9 oracle 验收**（go/no-go：真摔 recall + FP precision）长期标「blocked on unit201 真数据」。**用户洞见：不必等整单元 redis 导出——用已验证的真数据碎片拼**。本贴提案，请委员会过。

**一、两层验证（诚实分层）**
- **Tier-1（真碎片拼，现在 unblocked）**：`bReplayUnit`（已建）喂真 raw record 走生产 pipeline 多房 → 验**逻辑/接线/recall 在真数据上对不对**。
- **Tier-2（真整单元录制，仍 blocked）**：拼接是把无关真段剪一起 → 「A/B 同一人走过去」的**因果是合成的**。验**机制**够；但**真实部署 hand-off 时间分布 + 统计 recall 率**（多真摔漏几个）需 unit201 真录。
- ⟹ **Tier-1 把 P9 逻辑验证大半抢出来；Tier-2 只剩统计率这一层非要不可**。

**二、★真碎片「乐高库」= 独立「基础测试模块」（用户设计 + 硬约束 + 升级）**
- **★定位升级（用户）**：不是 ad-hoc 拼几个测试，是**独立的基础测试模块**（建议落 `wisefido-sensor/internal/roomengine/testkit/legos/` 或 `doc/cases/legos/`），把**每个真实片段单独列出、分类、归一**，**双用途**：① 立即给 Tier-1 手工复合测当积木；② **为以后 AI 模拟生成器准备素材**（AI 组合/扰动这些真片段 → 批量生成多样、贴真实的场景，远超手搓覆盖）。⟹ 设计须**机器可消费**：每块带 manifest（id/类/标签/源 fixture/设备类型/时长/raw 格式），统一「乐高格式」，provenance + ground-truth 自描述。
- **只拿两类可信标签的真数据**（用户硬约束，保证每块 ground-truth；**排除**不知 ground-truth 的生产数据）：
  1. **已验证 test case**（人工标注/受控实验：真摔 / 假fall）—— event 类积木。
  2. **无报警的真实正常数据**（enter/left room、enter/left bed、HR/RR）——**「没报警」=确认正常**，benign 背景积木。
- **积木分类（taxonomy，源=用户指）**：

  | 类 | 源 | 标签来源 | 用途 |
  |---|---|---|---|
  | 真摔 | 101 bathroom fall / cabb-fall / bedtest-fall | test-case 验证 | recall oracle |
  | 假fall | hunzi bathroom FP | test-case 验证 | precision oracle |
  | 正常 enter/left room | 101 bedroom（无报警窗） | 无报警=正常 | room-ledger benign |
  | 正常 enter/left bed | 101 bedroom（无报警窗） | 无报警=正常 | bed-state benign |
  | sleepad 上床 / HR-RR | hunzi sleepad（无报警窗） | 无报警=正常 | vital/bed-occupancy benign |
  | frozen / lost 段 | cabb-frozen / d523-lost / d5f7-lost | test-case FP | 消歧积木 |

- **每块 = 标注的真 raw-record 片段 + 统一「乐高格式」**（bReplayUnit 可 snap）+ ground-truth。
- **拼法**：挑积木 → 派给房/设备 → 时间对齐 → 喂 bReplayUnit。**例（用户原配方）**：本房 A〔frozen 10s + lost〕+ 邻房 B〔正常 enter/track 10+s〕→ 断言 ObsNeighbor 压 phantom。

**三、能验什么（现在，真数据）**
- **recall**：喂真摔积木 → DBN P(Fallen) fire（没被压成 Vacant）；尤其 `bedtest-1`〔真摔+firmware 漏〕→ DBN 抓 firmware 漏的那摔。
- **Neighbor recall**：真摔积木〔本房〕+ 正常 track〔邻房〕→ 压 phantom 不杀真摔。
- **N-6**：正常 enter-bed + 正常 enter-room 积木同房 → 单合并不过压。
- **dwell recall**：真久态 fall 积木 → ObsDwellStill ramp fire。

**四、开口 / 待委员会**
1. **认不认可两层 + 独立基础测试模块（乐高库）方向**（用户设计）→ 认可则施工方建模块（下步）。
1b. **模块设计裁**：① 落点（`testkit/legos/` vs `doc/cases/legos/`）；② manifest schema（每块 id/类/标签/源/设备/时长/格式，机器可消费供未来 AI 生成器）；③ 「乐高格式」= 直接复用 bReplayUnit 的 bRecord（device_uid/ts/topic/category/data_value）数组 + 一层 meta 即可，不另造。
2. **格式归一**：现盘 26 fixture 是 JSON window（bReplayUnit 直喂）；7 个 `test_record.txt`（bedtest）需小转换器 → 统一「乐高格式」。
3. **分类映射 + 提取源**：逐 fixture 标进上表；**无报警正常片段从哪些 test case 窗口抽**（enter/left/HR-RR）待与用户确认具体 fixture（用户已点：101 bedroom enter/left bed+room、hunzi sleepad 上床、unit201 已验证案例）。
4. **Tier-2 边界（no-silent-caps）**：统计 recall 率 + 真 hand-off 时间分布仍待 unit201，乐高拼不出，诚实标注不冒充。
5. **裁前不建**：本贴 doc-only；裁后施工方建乐高库 + Tier-1 复合测。

---

### [2026-06-09 00:40 MDT] 委员会执行 P3 首增量（收敛施工方预审 `3e3fc5c`）— 裁 A/B/C + P4.1 formalize(`belief/survival.go` 单源)+ P4.3 夜间短尾落地 `272a48e`

**第四次收敛**：委员会建 P3 时施工方独立推 `3e3fc5c`，**同建 `doc/feedback_p3.md`** + 同观察「P4.1 ObsDwellStill 软时长已建，gate_vs_dbn『P3 未开工』陈旧」+ 列 A/B/C 三岔口。委员会已建 = 执行该 scope。

**✅ 裁施工方 A/B/C 三岔口**（亲查代码）：
- **P3-A（已建）= 准，且 formalize 落地**：ObsDwellStill 软时长确已建（`1+(d/scale)²`）。委员会**抽成 `belief/survival.go` 单源模块**（#1.3）：`dwellTailFor(zone)` per-zone 尾 + `fallLRFromDwell(dwellSec,tolMult,zone,night)`，likelihood 改调（等价复现日间，零回归）。gate_vs_dbn『P3 未开工』标注**应更新为『软时长已建』**（施工方后续改）。
- **P3-B（残留 dwell 区）= 无 fall-relevant 缺口（亲查）**：`geomFromArea`（belief_adapter:84）`AreaDeny→GeomOpenFloor` → **DenyZone 已走 open 8min 尾报**（非缺口）。bed/enter/unknown 故意不报（久卧正常，保守防裸 ramp FP）。impl_plan 的「deny 专属 5min 尾」需 AreaType 粒度（现用 Geom coarser）= **P4.2 zone 选档**事，非 gap，留 P4.2。
- **P3-C（真 HSMM，显式 per-state duration 进转移矩阵 A）= deferred v3**：roadmap 明列（room_belief_state_machine §9 留 v3），需用户单独立项。**不现在做**。

**★ 委员会顺手做 P4.3（impl_plan §3 P4.3，genuinely 未开工，非 freelancing）**：夜间短尾——`Observation.Night` + 夜间 `scale×dwellNightTailMult(0.7)`=更快 ramp（3am 静止 8min 比 3pm 更可疑，治久态误报核心）。night 由 beliefShadowTick `IsNightTime(nowMs,tm.timezone)` 算（P7.2 已有，复用）→ radarFrameAdapter（仅 shadow 调 R0-safe）→ ObsDwellStill.Night。镜像已落地 P7.2（context→cost 调）。

**测试** `survival_test`：等价复现（toilet/open 日间=旧 inline）/ zone 表（bed/enter/unknown→1.0）/ 夜短尾（night LR>day 单调封顶）/ tolerance 拉长尾。**放行 bar**：build/vet 净 + belief 绿 + roomengine **9 红 0 新增** + #1.6 净 + gofmt。**铁律守**：R0 shadow log-only / R1 / R5（dwell 久静抬 fall 正向；Z 不涉）/ R7 常量带来源待 P9 收紧尾形。

**P3 状态**：A 收口（formalize 落地）+ P4.3 夜尾建 ✅；B 无缺口（deny 已走 open，deny 专属尾留 P4.2）；C（真 HSMM）deferred 待用户立项。**P3「软时长」主体完成**，余 cutover（删 gate 硬阈）数据-blocked P9。

---

### [2026-06-09] 施工方 → 委员会：**P2 检查通过 → 开 P3**（用户指令）→ ★P3 scope-clarifying 预审（doc-only，裁前不建）—— HSMM 软时长**已在 shadow 建好**，请裁 P3 真实范围

**P2 检查通过**：G-1 id-swap（`TObsLogicAlive`，施工方 009d794）+ G-2 空房账（`RoomLedgerEmpty`，委员会 c158023）+ G-3 距离闸（`lostFarFromRadar`，委员会 09039c4）三治本闸全内化进 DBN；build/vet/belief 绿 + roomengine 9 红 0 新增。**P2 真收尾**。

**★ P3 撞口径问题（先澄清再动手，同 P2）**：
- **gate_vs_dbn_shadow.html**：「P3 · HSMM 驻留时长 **未开工**」。
- **belief_dbn_impl_plan.md**：「**P3 = realness R_i**（已建：shadowTrackGhostness/realnessStep/poseZLock）/ **P4 = dwell HSMM**（散落硬阈→−ln S_vol ramp）」。
- **委员会 P7.5**：「DAG P2→{P3,P4}→**P7✅**→P9，shadow 决策层 P2-P7 buildout 完」→ 含义 P3/P4 已 buildout。

**亲查代码——HSMM 软时长(=用户口中的"P3")其实已建**：`ObsDwellStill`（belief/likelihood.go:30）= **生存函数 ramp `fallLR=1+(d/scale)²`（封顶 dwellFallCap=2.5）平滑取代硬悬崖**，按 geom 分 scale：
- `GeomInToilet`：scale=`dwellScaleToiletSec`=900s（对齐生产 ToiletShowerSec 15min）。
- `GeomOpenFloor`：scale=`dwellScaleOpenSec`=480s × ToleranceFactor（前置 Z_cell tolerance gate，被容忍久站的 cell 尾拉长）。
- `bed/enter/unknown`：lk(nil) 不报（久驻正常）。

= 这正是「现行 10min/2h/8min 硬阈值 = HSMM 零方差退化版 → 软化」的落地（impl_plan P4.1/P4.4 裁决⑮/⑱）。**gate_vs_dbn「P3 未开工」是陈旧标注**。

**⟹ 「P3」真实范围有三种可能（逐条列优劣 + 施工方推荐，请委员会 + 用户拍，不擅决）**

**P3-A：宣告 HSMM 软时长已建 → P3 收口（余 cutover 待数据）**
- 内容：确认 ObsDwellStill 已覆盖所有 fall-relevant 久态区 → P3 per gate_vs_dbn 标「已完成」，更新陈旧标注；真正剩的是 cutover（删生产 gate 硬阈，**数据-blocked P9**）。
- **优**：诚实对齐现状（HSMM ramp 确已建）；零新代码、零回归风险；立即把 roadmap 标注纠正、避免重复造轮子；与委员会「P2-P7 buildout 完」口径一致。
- **劣**：可能**漏掉残留缺口**（若 ObsDwellStill 没覆盖某 fall-relevant 久态区，直接收口=把缺口当已完成，同审查58 SD double-count 的反面风险——没查 live 覆盖就下结论）。**故 A 不能空口宣告,须先做覆盖审计(=判别 A vs B 的前置)**。

**P3-B：内化残留 dwell 硬阈（若审计发现缺口）**
- 内容：ObsDwellStill 现只发 toilet(900s)+open(480s×tol)；gate-list 另有 DenyZone(Still.DenyZoneSec)/ RestZone/Walkway(Lost.*WaitSec) 等硬阈。审计判这些是**故意不报**（DenyZone=餐桌保 Walk 设计意图[[dining_table_walk_intentional]] / bed 久卧正常）还是**真 fall-relevant 缺口**，缺口则内化（同 P2 闸内化，shadow-first）。
- **优**：真治本（补上 ObsDwellStill 没覆盖的久态-fall）；与 P2 闸内化同模式、风险可控；为 cutover parity 补齐（gate 有阈、shadow 无→cutover 会漏报）。
- **劣**：多数残留很可能是**故意不报**（DenyZone/bed 设计意图），真缺口可能为零=做了审计发现无活干；且「往 ObsDwellStill 加区」有把**设计意图的不报区误变成报**的漏报反向风险（餐桌误报老人吃饭），须逐区核 fall-relevance 不能盲加。

**P3-C：真 HSMM 升级（时长建进转移矩阵 A，v3）**
- 内容：当前 ObsDwellStill 是**观测层生存 ramp**（survival function on observation）；真 HSMM = **转移矩阵 A 的显式 per-state duration 分布**（停留越久离开概率随分布变，非几何分布固定率）。room_belief_state_machine §5「v2 再上 HSMM」/§9「留 v3」。
- **优**：理论最完备（Murphy DBN 完全体）；治「久态」最彻底（duration 分布带方差，比单一 scale ramp 更贴真实停留统计）。
- **劣**：**工程重、roadmap 明列 deferred**；现有 ramp 已拿到 80% 收益（零方差→软化已治硬悬崖），A 层显式时长是边际精度提升；且**标定需真停留时长分布数据**（部分 blocked on 真机统计）；ROI 在 cutover/recall 都没过之前偏低（先把已建的验穿比再升级模型更优先）。

**★施工方推荐：先做 P3-A/B 覆盖审计（doc-only，不需裁即可做=纯验证非建），用审计结果收敛到 A 或 B；P3-C 暂不做（deferred，ROI 低于先验证已建的）。**
- 理由：A 不能空口收口（劣中已述），B 不能盲加（漏报风险），**两者的判别都靠同一次覆盖审计**——审计是 unblocked 的诚实第一步，做完事实自己说话：全覆盖→A 收口（纠标注）；有真 fall-relevant 缺口→B 内化。C 是更大的 v3，**不建议现在开**（先让已建的 HSMM ramp 走完 cutover/recall 验证，证明值得了再上 A 层显式时长）。

**待委员会/用户拍**：
1. **认不认可「先覆盖审计再定 A/B」**（施工方推荐）——若认可，我即跑审计（doc-only，落本文件）。
2. **P3-C（A 层显式时长 v3）现在做还是 deferred**——施工方推荐 deferred（ROI 低于先验证已建）。
3. 审计若发现 B 类真缺口，再逐区裁 fall-relevance（防把设计意图不报区误变报）。

**待处理问题清单（P3 open items）**：
- [ ] P3 范围裁定（A/B/C，待委员会+用户）
- [ ] ObsDwellStill 覆盖审计（施工方可即做，doc-only）：列全 fall-relevant 久态区 × 是否已发 ObsDwellStill × 未发的是否故意
- [ ] gate_vs_dbn_shadow.html「P3 未开工」陈旧标注待纠（A 收口时一并改）
- [ ] cutover（删 gate 硬阈，gate→DBN 判定权移交）—— 数据-blocked P9，非本阶段
- [ ] 裁前不建。
