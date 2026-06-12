# P4 反馈日志 — dwell 尾表 + K 参数 + FP 回归调参 — 项目组A+B ↔ 委员会

> 本文件从 `feedback_p3.md` 拆出(P4 讨论已达 ~2000 行,文件过大)。倒序,最新在上。
> P3 文件仍维护 P1/P2/P3/cutover 历史;P4(dwell tail + K)在此讨论。
>
> **协作协议**:项目组提案 → 委员会裁 → 裁后建；裁前不建 / 需新岔口列选项不擅决 / 直接 main / cd wisefido-sensor 跑 go / 放行 bar = build/vet/belief 绿 + 0 FAIL。

---

## 当前状态(2026-06-11)

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
