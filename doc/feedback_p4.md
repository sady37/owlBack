# P4 反馈日志 — dwell 尾表 + K 参数 + FP 回归调参 — 项目组A+B ↔ 委员会

> 本文件从 `feedback_p3.md` 拆出(P4 讨论已达 ~2000 行,文件过大)。倒序,最新在上。
> P3 文件仍维护 P1/P2/P3/cutover 历史;P4(dwell tail + K)在此讨论。
>
> **协作协议**:项目组提案 → 委员会裁 → 裁后建；裁前不建 / 需新岔口列选项不擅决 / 直接 main / cd wisefido-sensor 跑 go / 放行 bar = build/vet/belief 绿 + 0 FAIL。

---

## 当前状态(2026-06-12)

**主线**:geom 全量退役 **已实施**(commit `7dd48de`,D1-D4 签字后施工)。

- bedside FN 止血 `5ef9ede` 已部署 test1(DBN_MODE=2)。
- geom 退役 `7dd48de`:删净 belief.Geom(grep=NONE),逐 tick 等价零新回归,红线回归 TestBedsideFallBedReleased 锁 D3。详 `doc/geom_retirement_mapping.md`§7 + 审查记录 [2026-06-12]。
- **已部署 test1**:`7dd48de` 09:24 重启上线(DBN_MODE=2,无 panic),替换止血版 `5ef9ede`。

**待清(non-blocking,与 geom 迁移并行,不挡签字)**:
1. `ObsTimeContext` —— likelihood `case ObsTimeContext: return lk(nil)` 空发射(只调 prior/θ_fire,不在 diag 更新),确认归并/清理。
2. `loggedVeto` —— no-op 守卫待删(DBN cutover task② 遗留)。
3. `ghost_*` log 统一 —— 9 种 `ghost_*` 收敛 `ghost_veto` + reason/veto_evidence(DBN cutover task② 遗留)。

---

### dwell 尾表 + K(历史,2026-06-11)

**已建(项目组A,待委员会 R6 签字)**:
- `ea459c9` feat(belief): `dwellTailFor`→roomType×areaType 尾表(60s/12min/20min/90min/排除)
- `e467be8` fix(adapter_sleepace): 删 SubjectEntity gate

**委员会 R6 亲跑结果**:build/vet ✅ belief ✅ roomengine **3 FAIL FP 回归**——tail 表 default→60s 让已知 FP 被 DwellStill 推 fire:

| 测试 | 旧(tail 前) | 新(tail 后) |
|---|---|---|
| case-5(hunzi CABB lost FP) | P=0.022 | **P=0.993** |
| cabb-fall-A(静止站立 FP) | P=0.024 | **P=0.989** |
| 高 tol 开阔地久站 | 不 confirm | P=0.896 |

**根因**:旧 `dwellTailFor` default→false 让 GeomUnknown 的 DwellStill 数学上零作用→FP 靠没有 dwell 贡献被抑制。新 tail 表「其余→60s」让这些 zone 的 DwellStill 全面再生效→FP 被推 fire。

**待讨论**(本文件):
1. tail 表 default 60s 太激进?改为更大的 scale(120s/180s)?用 cell tolerance learning?
2. K 参数(45-120s 按 unit)怎么与 tail 表交互?
3. case-5/cabb-fall-A 的 FP 是该被 suppress 还是该 fire?是数据问题还是模型问题?
4. 高 tol 开阔地 tolerance gate 是否应单独修复?

---

## 审查记录（倒序）

### [2026-06-12] 施工方 → 委员会:**cd2b-0606 no-fire 根因 = 确认窗时长(非漏报/非证据不足),工单 23ef71a 完成**

按工单逐案排查 no-fire。结论**推翻了「bedside FN / 证据不足」的旧假设**——证据从不缺,是 90s 确认窗在过滤人为测试的短躺。

**先澄清一个概念**:`Stand-Walk / Bed-Lying / Floor-Fallen` 是 **belief DBN 的 9 个隐状态**(`belief/state.go:24` 的 `State.String()` 标签,`SFallen`=Floor-Fallen),**不是 gate-list 概念**。gate-list 的 pose(`StandStatic/Lying`)在 track_manager,同形不同层。trace 的 `argmax_state` 是 DBN `b.Max()`,无 gate-list 泄漏。

**逐案表(cd2b-0606,本轮可重放的唯一真实 case;p_fallen 来自 `--speed 30` 压缩重放,时间用帧计数还原——原始帧率实测 1Hz)**:

cd2b 全程**两次** Floor-Fallen 抬头,均 <90s confirmMs → no-fire,但**两段性质不同**:

| 段 | 帧区间 | 时长 | p_fallen 峰 | 结局 | 性质 |
|---|---|---|---|---|---|
| 第一段 | 467→470 | 4s | 0.891 | 帧471 崩回 Bed-Lying(0.89→0.32) | ✅ 短躺/翻身,confirmMs 正确过滤 |
| 第二段 | 614→664 | 51s | **0.9952** | **维持 0.99 到 trace 末帧 664,无崩回** | ⚠️ 疑似倒地;**录制 51s 处结束**未达 90s |

| case | track 数 | coExist | ghost 分支 | confirmMs | fire | 根因 |
|---|---|---|---|---|---|---|
| cd2b-0606 | 1(孤立) | false | 不进(铁律 ρ=0) | 90s | **no** | 两段持续证据(4s/51s)均 < 90s 确认窗 |

**根因(诚实版,逐帧 trace 核对后**修正了初版「测试员躺~51s 起身 P 崩回」的臆测——第二段实际维持到录制末,不崩回**)**:

1. **第一段 = 真·短躺过滤(正确行为)**:4s Floor-Fallen 后帧471 立即崩回 Bed-Lying。这是翻身/短躺,4s≪90s,confirmMs 正确清零。`Decider` 要求 `SFallen>θ_fire` 连续维持 90s,中途崩回即清零(belief.go:105-119)。

2. **第二段 = 持续证据时长不足(数据局限,非抑制 bug)**:p_fallen 稳定 0.99 维持 **51s 到录制结束都没崩回**。no-fire 真因 = **录制窗口只覆盖 51s,confirmMs 要 90s 持续证据,数据不够长**。51s 之后人起身还是继续躺≥90s——**本数据未覆盖,无法判断**。结合「跌倒数据全人为测试,躺~1min 就站」,51s 后很可能起身,但这是推测非数据。⚠️ `belief.go:94` 注释「cd2b 类~47s 内崩回」只匹配第一段,不匹配第二段(疑历史标定),建议复核。

3. **重放工具 speed artifact**:`ProcessFrame.nowMs=frames[0].TMs`(track_manager.go:1057-1059),replay 把 frame TMs 重写成灌入墙钟(main.go:223)+`rel=(ts-t1)/speed`(:205)。`--speed 30` → frame TMs 间隔压成 33ms → confirmMs(按 TMs 算)被压 30 倍 → 加速重放下 confirmMs 必不满足。⚠️ **验 fire confirm 必须 `--speed 1`**;加速重放只能验 track/belief 演化(p_fallen 轨迹/ghost_veto)。已记入 tools/README。

4. **coExist/ghost 与 no-fire 无关**:单 track→`coExist`=false→ghost 铁律不进分支(P(Ghost)=0),ghost_veto 纯 forensic。

**阈值/后续建议**:
- **motion/mirror 阈值不需要动**——第二段 p_fallen 已达 0.9952,发射/realness 证据从不缺;问题在持续证据时长,不在阈值。
- **confirmMs=90s 是否合适本案无法定论**——第二段是「躺 51s 录制中断」,无法证明它该 fire(51s后未知)也无法证明不该。真长躺(≥5min)与短躺的分界标定**需要真实跌倒数据**(memory:现数据全人为测试,时长由测试行为定非真实分布),推迟。
- **真正需要**:`--speed 1` 重放一段**录满 ≥90s 持续倒地**的真实数据,验证 confirm 在真实时间语义下正常触发——本轮无此类数据。

**收尾**:TEMP trace(Info→Debug 排查口)已回退;build/vet 绿;binary 重建,test1 已 restart 上线(PID 2148796,无 panic,trace 回 Debug)。

---

### [2026-06-12] 施工方 → 委员会:**still-box 单源对齐完成（执行工单 721f40b）+ 一处工单前提澄清**

按工单 721f40b 把 cell engine 久静量改读 box、删 StillSince。施工中查清一个**工单前提需修正**：StillSince 和 StillBoxRunStart **不是「同一指标的两套计算（drift）」，是两个不同语义**——`StillSince`=帧间位移**即时态**（本帧动没动）；`StillBoxRunStart`=30s box **稳态**（窗内 max-min≤50cm）。直接合并会丢即时态/久静态之一。

**方案（用户「同步计算」洞察，单源一处算多处读）**：
- `updateContinuousIndicators`（算 box 的唯一点）**同步算** box 起点位置 `StillBoxStartX/Y`(=History[0]) + break 时暂存 dwell 时长 `StillBoxBreakDurMs` —— 解决删 StillSince 后「绑位置」「移动块 dwell 时序」两缺口。
- **拆两套语义**：即时态（StandStatic/Lying/Sit refresh、AdjustScore、速度）**保留帧间 `d` 判据=行为零变化**；久静量（LongStill/MarkDwell/MarkToleratedStill/MoveActive/StillSec）**改 box 单源**。
- 删 `StillSince/StillX/StillY`；顺手删 `computeRisk` 的 `stillSec` 死参数（传入即弃）。

**为何拆而非全 box**：`d`（即时）与 box（30s 窗）判据不同步——box 可能在 `d` 判静止的同帧 break，全 box 会让即时态（StandStatic/score）也滞后 30s、动 verdict。拆开后久静量单源 box（委员会要的），即时态等价。

**影响查证（删 StillSince/MoveActive 改 box 滞后 30s 的下游）**：
- **LeftBed**(sleepad bed_status event)/ **LeftRoom**(RoomLedger ExitRoom event)/ **card 生命迹象 LastActiveTs**(firmware activity stat 独立链)——**三条都不依赖 radar stillness，零影响**。
- MoveActive→realness（摔前在动）：box 滞后=刚停仍 active=**更安全**；→DecayInactive：阈值 4/10min/6h，30s 可忽略；→co-existence motion symmetry：minDispSqCm gate 挡。
- 久静量 box：判据更抗抖动；时机晚~30s 对 LongStill(15min)/Dwell EMA 可忽略。

**验收**：build/vet 绿；roomengine 仅 2 预存 dwell tail FP 红（belief 层 ObsDwellStill，与 cell engine still-box 不同层=无新回归）；测试无引用删字段。**未部署**（待裁）。

---


### [2026-06-12] 施工方 → 委员会:**★case fixture 三件套交付（导出再重放，请先审本条，再回 still-box 工单 721f40b）**

用户拍板「导出再重放」架构，三工具统一 `tools/`（commit 链 …`285d66e`）。围绕统一 fixture 契约（`window.json` + `meta.json{device_uid,device_addr,device_type}`），导出与重放**分离**，fixture 两源（PG 真实 / lego 合成），重放只吃文件、不管来源。DB/Redis 密码**单源 `.env`**（不硬编码）。

- **export**（原 `scripts/export_case_v2.sh` → `tools/export/`）：**unit 级**——从主设备推 /64，导整个 unit 全活跃设备（多 radar + sleepad，非单设备）；`case-<uidlast4>-<MMDD>-<startHHMM><endHHMM>` 名解析（last4→uid / HHMM+`--tz`→ms）；产 window.json(radar) + window_sleepad.json + meta.json{uid,addr,type} + 各 radar layout（主设备→room_layout.json）。
- **simulate-make**（新建 `tools/simulate-make/`）：lego 块合成 → 同 fixture 契约；内置 `open-floor-fall`/`ghost-jump`；pose 单源 `owl-common/observation`；**不连 DB/redis**，AI 灵活造 fall/ghost 场景。
- **replay**（原 `tools/redis-replay` 改名 `tools/replay`）：加 `--fixture <case目录>` 文件源（读 fixture→rebase ts→灌 redis→live sensor 真消费跑**完整链** track+belief shadow）；`meta.device_type` 权威定 dtype（sleepad 的 event 不误标 Radar）；meta + DB 补缺混合；`loadDotEnv` 密码不硬编码；保留 `--device/--unit` DB 模式向后兼容。

**端到端验证**：cd2b-0606（unit201 = cd2b + 333b radar + 1641 sleepad，3 设备）→ fixture→redis→live sensor(新代码)→belief_shadow 完整链 → **`ghost_veto` 统一日志（reason=shadow_realness_jump）生产 emit 正确**（= 工单 9fa233d 产物在 live 验证）；纯文件模式（meta 全覆盖）不连库；simulate→replay 闭环（合成 fixture 可被 replay 读）。

**请委员会先审本三件套**，通过后我再回 still-box 工单（721f40b）。**留下一轮**：cd2b-0606 为何未判 fall（单 track shadow_realness_jump、DBN 孤立 ρ=0 不否决；bedside fall 是否 FN）。doc 历史引用（feedback_p3 命令示例）的 redis-replay 名按历史保留，当前工具索引已更新。

---

### [2026-06-12] 施工方 → 委员会:**ghost_* 日志统一 → `ghost_veto`+reason 完成(执行工单 9fa233d，cutover 4 任务收口)**

按委员会工单 9fa233d 逐条执行。注:委员会给的行号是其视图,我先前 geom 退役/loggedVeto 清理已使行号漂,故按 **key 名**定位(实际位置一致)。

**生产 9 处 log key 全改 → `ghost_veto` + `zap.String("reason",...)`**:

| 旧 key | 文件 | reason |
|---|---|---|
| mirror_pair_detected | mirror_detect.go | `mirror_pair_l1` |
| track_verdict_ghost(penalty 路径) | track_manager.go | `penalty_accumulated` |
| track_verdict_ghost(score 路径) | track_manager.go | `low_score` |
| ghost_motion_symmetry_hit | track_manager.go | `motion_symmetry` |
| ghost_mirror_symmetry_hit | track_manager.go | `mirror_symmetry` |
| adjudicator_anchored_to_ghost_rejected | engine.go | `anchored_reject` |
| belief_shadow_veto_evidence | belief_shadow.go | `shadow_realness_jump`/`frozen`(原 veto_reason 字段→reason) |
| belief_dbn_veto_ghost | belief_shadow.go | `dbn_coexist` |
| belief_dbn_veto_risk | belief_shadow.go | `risk` |

- **forensic 不丢**:原有业务 reason(track_verdict 的 BirthReason、adjudicator 的 d.Reason)改名 `birth_reason`/`detail` 保留,新 `reason` 字段=委员会子类。
- **测试迁移(3 文件)**:`belief_dbn_veto_ghost` 断言→`FilterMessage("ghost_veto").FilterField(zap.String("reason","dbn_coexist"))`(保原语义不误数别的 veto);veto_harness `case "belief_shadow_veto_evidence"`→`"ghost_veto"`、`veto_reason`字段→`reason`。

**验证**:`grep '(Info|Warn)\("ghost_'` 全仓只命中 **ghost_veto**(8 Info + 1 Warn);旧 8 key 生产+测试残留=NONE;其它 `ghost_*`(ghost_penalty/ghost_track_id/ghost_decay/ghost_low_score 等)均为字段名/reason 常量值,同委员会「ghost_penalty 字段名不改」规则保留。build/vet 绿,roomengine 仅 2 预存 dwell tail FP 红(`peak=0.993`,veto=0=迁移零影响),belief 包绿=**零新回归**。

**★cutover 4 任务全部闭合**:①gate-list 删 ②DBN_MODE/Option A ③DBN 自有 ghost 检测 ④日志统一(本条)。未部署(test1 仍跑 geom 退役版,这批 log 改待裁部署)。

---


### [2026-06-12] 施工方 → 委员会:**non-blocking 待清(geom 退役后顺手清理)**

按当前状态「待清3项」处理,用户指示不深挖自查(易漏)、先提交委员会审:

- **① ObsTimeContext 删**:零构造点(全仓 grep 无 `Kind: ObsTimeContext`),likelihood case 返回 lk(nil) 空发射,Model.prior 不用 time/night(night 走 `ObsDwellStill.Night`,房型走 `RoomType`)→ 死 ObsKind。删常量+obsKindLabel+case+过时注释。ObsKind iota 重排无影响(belief 内存模型不落库,引用全用名)。
- **② loggedVeto 删 no-op 守卫**:`:323` 每 tick present 都 `reset false`→`:327` 守卫 `!tl.loggedVeto` 永真 = 确证 no-op(部署时见 veto_evidence 每秒打即此)。删字段+reset+守卫条件+set,改无条件 `if jumpGhost||frozenGhost` 每 tick forensic emit(per-tick forensic 保留,语义不变)。
- **③ ghost_* log 统一**:按用户指示**不深挖**(自查易漏,留委员会)。现状日志已大部分收敛:`belief_dbn_veto_{ghost,firmware,risk}` + `belief_shadow_veto_evidence` + `lostfall_{provisional,escalate,suppressed,cancel}`,无散落的裸 `ghost_X`。memory task② 说的「9种 ghost_* 收 ghost_veto」应在 DBN cutover 时已部分完成。细节收敛(reason 字段统一 / veto_evidence 迁 emitDecision flip-only)留委员会审定。

**验证**:build/vet 绿,roomengine 仅 2 预存 dwell tail FP 红(无新回归),belief 包绿。**未部署**(test1 仍跑 geom 退役版,这批 log 清理待裁部署)。

---


### [2026-06-12] 施工方 → 委员会:**geom 全量退役实施完成（commit `7dd48de`，D1-D4 签字后施工）**

承接 `d2c2197` 映射表 + 委员会签字 D1-D4。按 `doc/geom_retirement_mapping.md` 施工，详「实施记录」§7。

- **范围**:删 `belief.Geom` 类型/常量、Observation/TObservation.Geom、geomFromArea/geomFromGrid、beliefShadow{geom,lastLostGeom} 字段、所有死 Geom 赋值/死参数;`GeomConf→AreaConf`、`geomConfFromGrid→areaConfFromGrid`(D4 保留 blend);日志 `last_geom→last_area`。代码 geom 残留 grep=**NONE**。
- **D1-D4 全采纳推荐**:AreaCtx 结构体(精化为 `{AreaType,NearDoor,BedReleased}`,RoomType 不进——无分支用)/权威解析归构造层一次/止血换 BedReleased 载体不改阈值/保留 provenance blend。
- **关键工程约束(主动标注)**:belief 分层不能 import roomengine.AreaType(dwell 既定),新建 `belief/area.go` 值契约镜像;门优先(Room 层 geomFromGrid 含门距→保留 NearDoor;Track 层 geomFromArea 无门距→TObsAbsent 不加 NearDoor)。
- **批次微调**:发现真读 geom 仅 3 逻辑点 + 日志(余皆死字段),批1+2 合并为一次切换用 test 验等价;批3(删类型)放测试迁移后(删 Geom 类型前测试须先脱离常量)。
- **等价验证**:全 sensor build/vet 绿,逐 tick 等价;仅 3 预存红(2 dwell tail FP + 1 sleepace,均与 geom 无关,git stash 验基线一致)=**零新回归**。
- **红线回归**:新增 `TestBedsideFallBedReleased` 锁 D3(床区躺×bed_state:占用→SBedLying/离床→SFallen=开阔地躺;门优先)=承接止血缺的「床区躺+离床→fire」逻辑。
- **待补 follow-up(non-blocking)**:engine 层 plumbing e2e(beliefShadowTick 由 bedReleased 设 obs.BedReleased)——机制层已锁 D3 核心,plumbing 3 行被现有 engine 测试路径执行(未断言),带 bed_state mock 的断言测试 setup 较重,留 follow-up。
- **已部署 test1**:用户裁立即部署,`7dd48de` 09:24 重启上线(新二进制+DBN_MODE=2+无 panic),替换止血版 `5ef9ede`。

---

### [2026-06-12] 施工方 → 委员会:**bedside FN 止血部署 test1 + geom 退役映射表交付（plan-first 待签字）**

承接 `a22b6c1` 委员会裁定(geom→bed_state 迁移 plan-first + 先交 bedside FN 快补止血)。本条记两件交付。

**① bedside FN 止血 `5ef9ede` — 已部署 test1**

- **改法**:beliefShadowTick 顶部取一次 bed_state;`lying@床区(area_type=Bed) 且 bed_state 有数据且离床(BedStatus≠0,conf>0) → 翻 geom InBed→OpenFloor`,让 SFallen:4 竞争(脱离睡觉豁免)。= 把 silent_leftbed 判据搬进 DBN、由 bed_state 驱动。**无 sleepad(无解)/床占用时不动**(防睡觉 FP)。配套 `dampBedFallen 0.7→0.2`。
- **机理**:堵的是床边 FN 根因——`lying@残留 GeomInBed 盖掉 bed_state`,真摔倒在床边躺归 cell_area=Bed + in_bed_conf 高→判 SBedLying 非 SFallen。止血用 bed_state 轴翻 geom,孤立单 track 也能让 SFallen 浮出。
- **部署**:test1 本机,`build-owlback.sh sensor` 绿 → `systemctl restart owlback.sensor`。新二进制 06:31 PDT,PID 1759653,`DBN_MODE=2`(全开)+ `DBN_VETO_RECOVERY=0` 注入进程确认,无 panic/fatal,shadow 日志正常。**床边 FN 红线在生产护上。**
- **如实标注(2 项)**:
  1. **e2e 回归未补**:尚无"床区躺+离床→DBN fire"专用回归测(fixture 重)。止血当前靠逻辑 + bed_state 轴成立,端到端 fire 未实测。→ 随 geom 退役批 4 补。
  2. build/vet 绿,TestDBNFireSwitch 过,仅剩 2 预存在红测(非本次回归)。

**② geom 退役映射表 plan `d2c2197`(`doc/geom_retirement_mapping.md`)— 只产出未动代码**

逐条覆盖委员会要求的四块,每条列"替代权威源 + 改法":poseLikelihood pose×geom 全分支(§3.A)/Track 层 TObsAbsent last-geom 分流(§3.B)/所有 obs.Geom 字段(§3.C)/geomFromGrid·geomFromArea·geomConfFromGrid 全调用点(§3.D)。

- **核心结论**:geom 在服务器侧 ≈ `cell.area_type + 门距` 的纯函数(`geomFromArea`/`geomFromGrid`);dwell 腿已迁完(`fallLRFromDwell` 已吃 RoomType/AreaType),残留即 pose / TObsAbsent / 一批 Geom 字段。退役 = 各触点直读已存在权威源。
- **架构建议**:权威解析(含 bed_state 优先级 + 止血翻转)**只在观测构造处(adapter/shadowTick)做一次**,落成 Observation 字段;likelihood 改消费已解析的 `AreaCtx{AreaType,RoomType,NearDoor,BedReleased}`,保持纯函数。止血天然落在构造层。
- **工程细节(主动标注)**:`GeomInEnter` 是**唯一无法由 area_type 独立编码**的 geom 值(来自动态门距≤30cm,非静态格类型),退役须保留独立 `NearDoor` 信号(PoseWalking + TObsAbsent 两处用)。

**待委员会裁的 4 个决策点(§4,各附推荐)**:
- **D1** likelihood 签名 → `poseLikelihood(pose, AreaCtx)` 结构体化(推荐:是)
- **D2** 权威解析归位到构造层一次性做(推荐:是)
- **D3** 床区躺睡/摔二义由 `BedReleased` 驱动 = 止血换载体不改阈值(推荐:保留现语义)
- **D4 ⚠️风险最高** `GeomConf` provenance blend(FE画1.0/feedback0.6/学0.4)去留——推荐**保留**,删则自学床区全额抑制跌倒,与已签字 layout 权威模型冲突

**施工序(签字后)**:§5 四批——批1 加字段双写/批2 切 likelihood 读侧 + shadow 比对等价/批3 删净 geom(规则#1.2 grep 自查)/批4 重写单测 + 补"床区躺+离床→fire" e2e 回归。

---

### [2026-06-11] 施工方 → 委员会:**死源 #4 radar-vital 接通 bed scorer**（独立线程,非 dwell/K 主题;来自矩阵输入完整性审计）+ 3 个待裁

**性质**:本会话做了一轮「矩阵输入完整性」审计(消费侧 likelihood dispatch 闭合,逐腿核生产侧是否真有数据流)。结论:**无物理信号被静默丢弃**,但有 4 处"消费侧建好/生产侧缺失"的死腿,其中 **radar-vital = 死源 #4** 已接通。`b80...`(本卷下方 commit)。

**死源 #4 坐实(设计有/消费侧建/生产侧从没 wire)**:
- 设计 doc `bed_bayesian_review.md` **§5 line160** 早写了路由 `vital_radar:sustain → OnRadarVital`;消费侧整条腿也建好(`OnRadarVital` / `radarVitalLastTs` / `radarClusterLRLocked:605` 独立贡献 `lrRadarVital=ln(1.75)≈0.56` / γ 冲突 / multiSource sticky)。
- **但生产侧从没接**:`MonitorVitalSource:55` `if !isSleepad continue` 排除 radar,理由是**错误前提**「radar 能测沙发上人 HR/RR 故不可信」。
- **firmware 事实推翻该前提(用户拍板,见 memory `radar_hr_rr_bed_enter_gated`)**:qinglan radar **仅 track enterBed 才开 sleep_monitor 返 HR/RR;沙发不是床→不返回**。故 HR/RR 存在 ⟺ 在床,无沙发污染。

**已接(producer 侧 5 处;消费侧一行未动)**:
1. `vital_source.go` 删 radar 排除,radar device 也扫 HR/RR。
2. radar HR/RR 经注入 `BedResolver(radarAddr, roomPref)` 归床(单床房 covers=1.0 干净解析;多床候选 `""` → 不发,**让位每床自己的 sleepad**=接触式权威)。加 `roomPrefixFromDeviceAddr` /88。
3. `adapter_vital.go` emit 回调加 `source` 维度:sleepad→`"vital"`(OnSleepadVital,不变),radar→`"radar"`(Kind `sustain`)。
4. `IngestEvidence` 加 `radar:sustain → OnRadarVital` + 路由表注释。
5. 改正 `vital_source.go` 错误注释。

**归属为何不用 area_id 或位置多边形**(委员会可能会问,先答):
- 位置多边形:radar XY 床边有误差 → 会把床边误判在床(用户指出);且 firmware 已门控在不在床,server 不该拿噪声 XY 再判。
- area_id:track 同帧带 `area_id`(MonitorBuffer 里**有**,`area_type` 没透传需补 gateway)→ 理想但要额外管线;BedResolver 现成且多床 `""` 恰好实现"让位 sleepad",够用。

**置信层级守住(接触式 > 雷达)**:radar-vital LR 是所有正向腿最弱(ln1.75 < sleepad-vital ln2 < radar-InBed ln4 < sleepad ln9/19)+ γ 冲突 tempering + CoversWeight → 永远压不过有把握的 sleepad。R5 守(只调 Empty/Occupied 分布,vital 不进 SFallen)。

**验收 bar**:`go build ./...` + `go vet` 净;6 个 MonitorVitalSource 测试过(新增 `RadarResolvedToBed` / `RadarUnresolvedSkipped`,删除断言旧错误行为的 `RadarSkipped`);bed scorer + vital adapter 测试过;已重启 sensor。`TestSleepaceAdapter_UnboundDeviceSkipped` 失败经 stash 基线确认是 `f4187a6` **预存在**,非本次引入。

---

### [2026-06-12] ★委员会指令：统一 ghost_* 日志 → `ghost_veto` + reason（cutover 4 任务最后一项）

**当前残留（9 种 ghost-related log key）**：

| # | 旧 key | 文件:行 | 新 key | reason |
|---|--------|---------|--------|--------|
| 1 | `mirror_pair_detected` | mirror_detect.go:350 | `ghost_veto` | `mirror_pair_l1` |
| 2 | `track_verdict_ghost` | tm.go:1282,1313 | `ghost_veto` | `penalty_accumulated` / `low_score` |
| 3 | `ghost_motion_symmetry_hit` | tm.go:1586 | `ghost_veto` | `motion_symmetry` |
| 4 | `ghost_mirror_symmetry_hit` | tm.go:1596 | `ghost_veto` | `mirror_symmetry` |
| 5 | `adjudicator_anchored_to_ghost_rejected` | engine.go:875 | `ghost_veto` | `anchored_reject` |
| 6 | `belief_shadow_veto_evidence` | belief_shadow.go:330 | `ghost_veto` | `shadow_realness_jump` / `shadow_realness_frozen` |
| 7 | `belief_dbn_veto_ghost` | belief_shadow.go:805 | `ghost_veto` | `dbn_coexist` |
| 8 | `belief_dbn_veto_risk` | belief_shadow.go:810 | `ghost_veto` | `risk` |

**不改**：`belief_dbn_fire`（非 veto，是 fire 事件）、`ghost_penalty` 字段名（JSON schema，非 log key）。

**规则**：全部用 `zap.String("reason", ...)` 区分子类。改后 grep `"ghost_` 只命中 `ghost_veto` 和 `ghost_penalty` 字段名。build/vet/test 绿。**此指令为 cutover 4 任务收口——做完 4 任务全部闭合。**

**⚠️ 验证 caveat**:**当前部署仅设备 9e7 可能返回 radar HR/RR,且需专门测试才能验**(故 journal 被动观察不到 radar sustain = 正常,非 bug)。端到端实证 blocked on 9e7 专测。

**留给委员会的 3 个待裁**:
1. **`bed_bayesian_review.md` §6.1 line172「radar vital +0.56 是否过低」现在变可验** —— 接通前是纸面问题。委员会要不要在 9e7 专测后重估这个权重?还是先按设计值跑、实证再调?
2. **source 命名**:doc 写 `vital_radar`,我用 `radar`(Kind=`sustain`)与既有 `radar:enter/leave/pose_lying` 一致。确认用 `radar` 还是回退 doc 的 `vital_radar`?(倾向 `radar`:同源命名一致,且 `vital_sleepad` 实际代码里也从没采用、用的是 `vital`。)
3. **另 2 死腿待处置**:`sleepadAdapter`(room 层 vital-presence,零调用)/ `ObsTimeContext`(likelihood 中性、效应走 prior、无 producer)。前者疑似同 radar-vital 是合法待 wire,后者疑似真空壳可删。请委员会裁:逐个查设计意图还是一并清?

### [2026-06-11] ✅ 委员会裁——收死源#4 + 3 待裁答复

**收 radar-vital 接通**(`b80...`):设计有/消费侧建/生产侧从没 wire 的 gap 已闭合,producer 侧 5 处改动、消费侧零动 。Bar:build/vet ✅ 6 测试过 ✅ bed scorer 绿 ✅ 。验证 blocked on 9e7 专测(已知,不阻塞)。

**3 待裁答复**:
1. **radar vital +0.56 权重**→ **先按设计值跑,9e7 实证再调**。ln(1.75)≈0.56 是最弱正向腿,又是 radar 非接触式传感器,初始保守合理。专测有数据后再议。
2. **source 命名**→ **用 `radar`**(施工方倾向)。同源命名一致(`radar:enter/leave/pose_lying/sustain`),且 doc 的 `vital_radar` 从未在代码里用(sleepad 也只用 `vital` 而非 `vital_sleepad`)。doc 后续更新即可。
3. **另 2 死腿**→ **`sleepadAdapter` 保留待 wire**(同 radar-vital 是合法待设计,非 dead-on-arrival),**`ObsTimeContext` 可删**(likelihood 中性、效应走 prior、无 producer、无消费计划)。#1.2 删即删,不留 stub。

### [2026-06-11] 施工方 → 委员会:**DBN cutover**(删 gate-list firing → DBN 接管 + DBN_MODE 三档 + firmware-veto Option A + 死代码清理)

**性质**:把推断型 fall 从 gate-list 整体迁到 DBN(用户拍板 cutover)。gate-list 三类推断 fall(z_drop/silent_leftbed/lost_fall)firing 退役,DBN 接管(自发 + 可否决 firmware),配套清理 gate-list 时期死代码。

**已做(分两段)**:
- **已 push** `30a6c0b`:删 gate-list 推断 fall firing(`engine_z_drop`/`engine_silent_leftbed`(+vanished)/`lost_fall` 整套 pendingLostFalls 池+扫描+取消+`checkLostFall`/`lostFallWaitMs`+`PendingLostFall`/`LostFallStats`/`SilentFallLeftBedStats`+playback stats)。保留:firmware `radar_direct`(`RecordRadarAlarm`)、`AnomalyPathBreak`、`AnomalyStillTooLong`、DBN belief shadow。
- **工作树未提交(本提交一起 commit)**:
  - `DBN_MODE` 三档替换旧 `DBN_FIRE`:0=否决firmware+DBN不自发 / 1=不否决firmware+DBN自发(union地板) / 2=全开。`.env=2`,`run-service` 兜底 1。
  - **firmware-veto = Option A(延 tick 现算,非缓存)**:档0/2 firmware fall 暂存 → 下一 `beliefShadowTick` 用**当帧 bases** 现算 co-existence/ghost/τ* 裁决(与 self-fire 同源,结构消竞态);孤立必发(铁律);雷达静默由 `firmwarePendingDrainLoop`(2s,服务器墙钟 age,5s 超时)force-forward fail-safe。
  - 死代码清理:`emitDecision` 删;`fall_verify.go` **整删**(verifier 纯 informational 不 gate、读 gate-list 信号);`RecordRadarAlarm` 砍成纯 forward;`weakBioSource`/`SetWeakBioSource`/`WeakBioSource`/`FallVerifyStats` 删(`service.WeakBioScore` 保留喂 target.state);`BedSession` 砍到 {DeviceUID,InBedSinceMs,LeftBedAtMs}+死字段链(HasHRRR/LeftBedHadHRRR/LeftBedMaxPeople/SilentFallAlerted/RadarInBed*Ms/RadarSawTrackMs/RadarDeviceAddr/InBedConfidence)+`bedInBedConfidence`/`bedVanishMinConf`/`bedConfEMIFloor`/`minInt`/`maxInt`/`BedSessionLatch`/`SnapshotBedSessions`;删 LeftBed 5min gate(如实报 LeftBed);`silentFallParam` 整删、`lostFallParam` 11→5、`Still.StaticPosCm` 删;test helper(`newTestTM`/`newTestGrid`)迁 `test_helpers_test.go`。

**验收 bar**:`go build ./...` ✅ `go vet ./...` ✅。`go test ./...` 只剩 **3 个 stash 基线确认的预存在红测**(`TestP4OpenFloorDwellToleranceGate` / `TestRecallManifestAll/case-5` / zoneengine `TestSleepaceAdapter_UnboundDeviceSkipped`,非本次引入)。`TestDBNFireSwitch` 过(dbn_fire=1/veto=0)。

**in-session 对抗评审(Workflow committee)5 轮纪要(逐个 blocker→修)**:
1. firmware-veto 缓存式 co-existence 竞态(cache-time 旧共存,partner 消失后误否孤立真摔)→ 重做 **Option A**(延 tick fresh bases 现算)。
2. `tm==nil` 静默丢 firmware fall ×3 → 非 nil **by contract**(handler 顶 `if tm==nil return` 已 guard;stash 传参)+ `T_fire` 与转发**原子耦合**。
3. orphan-on-radar-silence(暂存后雷达彻底静默→pending 永不裁决)→ **clock-based** `firmwarePendingDrainLoop`。
4. `arrivedMs` 用 firmware 时钟(偏移算错 age)→ 改**服务器墙钟**;drain `tm==nil` 静默(rule 1.4)→ 去守卫(房从不注销→不可达)+ drain loop `recover` 兜响亮。
（第 4 轮修复**未再 in-session 复审**,提交本委员会/实测裁。）

**部署状态**:`30a6c0b` 已 live 在 test1(gate-list 删 + 旧 `DBN_FIRE=1` 语义);本批 DBN_MODE/Option A/清理**未部署**(待 commit+push+restart `DBN_MODE=2`)。test1 post-restart 日志:0 error/panic、gate-list 残留 0、`belief_dbn_fire` 11(dbn_silent/lost/moving)。

**⚠️ 待裁(诚实列未决)**:
1. **矩阵修 `lrPoseLyingBedFall`(床边 FN 红线)未做** —— `poseLikelihood` 的 lying@`GeomInBed` 仍零 `SFallen`;cutover 已 live 但 **DBN 接不住床边真摔**(原 silent_leftbed 兜底已删)。本轮做否?方案=几何降级(contact-authority:sleepad LeftBed 时床区躺按 OpenFloor 竞争 SFallen)还是别的?
2. **`dampBedFallen` 本仓库是 `0.7`**(清单标 0.2 但**不在 origin**,origin=HEAD=`30a6c0b`,无改它的 commit)。要改 0.2 吗?
3. **`DBN_MODE` 默认部署档**:`.env=2`(全开)。确认 2,还是先 1(保 firmware 地板)观察?
4. **firmware-veto Option A 复杂度**(5 轮才收敛)→ 委员会要不要重估"是否值得"vs 仅 risk-veto / 暂 mode 1?

### [2026-06-12] 施工方 → 委员会:**申请——大范围 geom 退役迁移(geom→room_type+cell.area_type / 床占用→bed_state):plan-first 还是直接开改?**

**背景(用户拍)**:`belief.Geom`(`geomFromArea`/`geomFromGrid` 把 track XY∩layout多边形 自算成 InBed/OpenFloor/Enter/Toilet)是**残留,已弃**。正确轴:
- **区域(在哪类区)** → `room_type + cell.area_type` 直传(**已成型**:`ObsDwellStill→fallLRFromDwell` 用 roomType×areaType 查尾表,`observation.go` 注"不经过 Geom intermediate");
- **床占用(在不在床上)** → **`bed_state`**(接触式权威 > 雷达 `track.area_id` 几何 > 服务器 XY∩多边形残留);
- `track.area_id` 现可得可信(`track_manager.go:46`"engine 不信 area_id"是过时注释,非设计拒用)。
- `poseLikelihood` 整套 geom-conditioned(lying/fallen/walking 各 geom 分支)= **最大的 geom 残留**,未迁。

**与上条 4 待裁的关系**:待裁#1(矩阵修=床边 FN 红线)= 本大迁移的**起点/子集** —— 床边 FN 根因正是 `lying@(残留)GeomInBed→SBedLying ×6` 用最弱又错轴的服务器自算几何,盖掉了手里本有的 `bed_state`(LeftBed 时 ObsBedOccupied 已释放却被 pose 几何抬回)。床/摔判别脱 geom 由 bed_state 定,就是修床边 FN。

**申请裁决(本次只问"怎么推进",不夹带改动)**:大范围迁移动到 **belief 发射面、安全敏感**(误改=漏/误报真摔),且施工方前几轮已把模型理错多次(geom 当成活的、绕 bed_state)。请委员会裁推进方式:
- **(A)施工方倾向**:先出"**哪个 geom 用法 → 换哪个权威源**"的**映射表 plan**(poseLikelihood 各 pose×geom 分支、其它 obs 的 Geom 字段、geomFromGrid/geomFromArea 调用点逐条列源),委员会**签字**后再动代码——降"边猜边写、再被打回"的一次性误改风险。
- **(B)**:不出映射表,直接按"**床/摔→bed_state、区域→room_type+cell.area_type**"开改。

**附**:本迁移与 cutover(77254a6,已 audit-收)解耦但同向;矩阵修/床边 FN 不部署不影响 test 现状(cutover 已 live、行为不变),但**床边真摔漏报口子在此迁移落地前一直开着**(原 silent_leftbed 兜底已删)。

### [2026-06-12] ✅ 委员会裁(用户拍,见 feedback.md last-audited a22b6c1)

1. **geom→bed_state 迁移 = plan-first (A)**,两条件:① 映射表逐条覆盖 `poseLikelihood` 全分支 ② **先交 bedside FN 快补止血**。
2. **`DBN_MODE=2`(全开)** 确认。
3. **`dampBedFallen` 补推**(0.7→0.2)、**`lrPoseLyingBedFall` 做**、**Option A 不改**。

### [2026-06-12] 施工方 → 委员会:**bedside FN 快补止血已交(止血,非全量迁移)**

**做了(止血,bed_state 驱动)**:`beliefShadowTick` 顶部取一次 `bedSt=BedOccupancyState`,`bedReleased = BedConfidence>0 && BedStatus!=0`(有床数据且离床)。pose obs 处:**lying@`GeomInBed` 且 bedReleased → 翻 `geom→OpenFloor`**,使 `lying@OpenFloor` 给 `SFallen:4` 竞争(脱离 `lying@InBed` 的睡觉豁免)。等价 silent_leftbed 判据(sleepad 离床 + radar 床区躺=床边真摔)进 DBN。无床数据(无 sleepad=无解)/占用 → 不动(防睡觉 FP)。`bedAdapter` 复用同一 `bedSt`。日志 `belief_dbn_bedside_unbed`。
**dampBedFallen 0.7→0.2**:床占用对 fall 的压制减弱(配合止血,床区躺别把 fall 压死)。

**性质**:止血借现有 geom plumbing(给 lying 的 pose-geom 按 bed_state 翻),**不是**全量 geom 退役;全量(geom→room_type+cell.area_type,poseLikelihood 全分支)走映射表 plan(待签字)。

**bar**:build/vet ✅;`TestDBNFireSwitch` 过;roomengine 仅 2 预存在红测。
**⚠️ 未补 e2e 测**:尚无"床区躺+离床→DBN fire"的专用回归测(fixture 重,需 bedroom+bed cell+sleepad LeftBed 事件)。本止血靠逻辑+bed_state 轴(委员会/用户多轮确认)成立,**端到端 fire 未实测**——请委员会裁:止血先收(测随全量迁移补)还是要先补 e2e 测再收?未部署(待 DBN_MODE=2 restart)。

---

### [2026-06-12] 项目组工单：still-box 单源对齐（cell engine 改读 StillBoxRunStart）

**当前问题**：track 层有两个独立 stillness 计算：

| 字段 | 用在哪 | 判据 |
|------|--------|------|
| `StillSince` | cell engine（`MarkLongStill`/`MarkDwell`/`MarkToleratedStill`） | 简单位移判断 |
| `StillBoxRunStart` | DBN（`ObsDwellStill`/lost-fall precondition） | box 判据：30s 窗 50cm 方框 |

**问题**：`belief_cell_contract.go:13` 宣称"双读同一个 StillBoxRunStart，谁都不重算"——但实际 cell engine 读的是 `StillSince`。两个计算并存 = drift 风险 + 双倍计算。

**要求**：cell engine 的 stillness 消费端（`MarkLongStill`/`MarkDwell`/`MarkToleratedStill`/stand-static 自学习）改为读 `StillBoxRunStart`（box 判据），删 `StillSince` 独立计算。DBN 侧不动（已在读 `StillBoxRunStart`）。规则 #1.3 单源真相。

**验收**：build/vet ✅，cell engine 相关测试全绿，cell learning（LongStill/Dwell/ToleratedStill）行为等价。

---

### [2026-06-12] 项目组工单：replay case 排查 no-fire 根因（重点：co-existence ≥2 track 才有 ghost）

**背景**：co-existence 铁律——`coExist = len(bases) >= 2`，孤立 1 track 永不判 ghost（ghost=真人反射必有共存 Real partner）。这护住了 long-lie 真受害者（越不动越像 frozen，但孤立证明真人），但同时也意味着：在单 track 场景下 DBN 没有 ghost veto 能力。

**任务**：
1. 用三件套 replay 回放已知 case（cd2b-0606 等），对 **no-fire** 的 case 逐案排查根因：
   - 是孤立 track（coExist=false）→ 无 ghost 不否决 → **正确行为**（铁律）
   - 还是 ≥2 track 但因其他原因未 fire（pFallen 不够 / ghost 误判 / risk veto / bed suppression）
2. 重点排查 ≥2 track 场景下的 ghost 判定是否正确——motion/mirror 对称是否覆盖了真实多径场景
3. 量化 **孤立 track 的 no-fire 占比**：如果大部分 no-fire 是正确孤立，铁律成立；如果 ≥2 track 也大量 no-fire → motion/mirror 对称覆盖率不够
4. 产出一个表格：case / track count / coExist / ghost / pFallen / fire / 根因

**产出**：no-fire 根因分析表 + 是否需调 motion/mirror 对称阈值的建议。
