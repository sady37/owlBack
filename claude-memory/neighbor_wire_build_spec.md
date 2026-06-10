---
name: neighbor_wire_build_spec
metadata: 
  node_type: memory
  type: project
  originSessionId: b27531f3-e8fb-4bf7-a017-a9d22b3eab5e
---

2026-06-08 死源审计收官后,用户拍板优先做 **#3 Neighbor wire**(定性"最具竞争力方向/限制条件下唯一解")。设计经与用户实时共议**已全部 pin 死**,待 shadow-first 落地。HEAD=d57240e(设计预审已 push 进 doc/feedback.md)。

**价值定位(用户原话)**:跨房耦合 = 丢轨歧义消解的**唯一外部信号**——本房雷达丢人时(盲区/冻结 ghost/信号丢失 = CABB/John.Y lost_track FP 主类),房内无证据区分"走去隔壁(→压 phantom fall)"vs"摔在本房(→保 fall)",**只有邻房占用能判**。直击全项目 FP 痛点。

**机制(用户定的核心)**:不读邻房 belief 后验(那才有反馈环),而是**新事件触发→60s 窗内查邻房原始事实**。每房已自算 `lastEnterMs`/`lastExitMs`(track_manager.go:131-132)+`BedOccupancyState`(:880)。流程:新事件→解析本房 MM-邻居→查邻居 60s 窗占用(room enter/exit ts **和** bed InBed,**两者都要**)→单占用门控→`neighborToObs`(belief_adapter.go:459 已在)→`ObsNeighbor`(likelihood.go:102 已在,R5-lock 许可压制清单成员)→`belief.Step`。`neighborToObs`/`ObsNeighbor` 似然**都已在,只是 tick 不喂**(B 审计 8/8 案 Neighbor 未populate=死管)。

**5 岔口(审查60 裁定 + 用户 steer,有一处待最终 reconcile)**:
- **N-1 占用源**:★**分歧待 reconcile**。用户 steer=邻房**内存原始 room+bed state**(lastEnterMs/lastExitMs 的 roomLedger + BedOccupancyState,60s 窗,**非 belief 后验**);审查60 批的是①=读兄弟 belief 后验 `1−P(Empty)−P(Left)`。**用户版更优**(确定性原始事实,**彻底消掉 N-5 跨 shadow 耦合**,且用户=principal)。**lean 用户 raw room+bed state**;新会话开工前向委员会过一轮 reconcile(用户版 N-5 moot,委员会版需 N-5 锁)。room state+bed state **两者都喂**(用户明确)。
- **N-2 邻居范围**=**MM neighbor 关系图(sensor 本地 create,不问 data)**。★用户纠正:MM 是 sensor 里 create 的(我曾误套 sensor_asks_data 铁律,已纠 [[sensor_asks_data_sync_not_db]])。引擎内 `RoomConfig.EnterTargets`(门→outside/bathroom)+`roomSuiteID`(engine.go:247)+`RoomType` 是门拓扑级邻接;优先用 sensor MM 更细 neighbor 图(关系解析做成可替换 seam)。审查60 批 N-2①(全同-unit),与 MM 兼容(单元内 MM 邻接 ⊆ 同-unit)。
- **★N-3 门(安全关键,漏报红线)——审查60 重裁,纠正我的预设**:门的真本质**不是"单元单占用计数",是审查㉚ 归因不变量**——压制须可归因于**同一丢轨人** + 无其他未交代人可能是 faller。多 resident 单元"邻房占用"可能是**另一个人**→归因不安全→**必不压**(否则压真摔=漏报)。**裁:复用现成 `SoleResidentRecaptureState` 的 sole-resident 语义**(已排 visitor/已 LOG skip-multiresident/P6.5 实证,hasSoleResidentInBedroom)。**禁新建第二个 census 阈**(#1.3 单源/#2.4 drift)——我原写的 SuiteCensusManager 新阈**错**,改复用 SoleResident gate。fail-safe:sole-resident 不成立→不压。
- **N-4 聚合**=多邻居取 **max**(sole-resident 前提下任一邻房占用=人在别处强证据)。审查60 批。
- **N-5 反馈环**=用户 raw-state 版**moot**(读原始 ts 非后验,无 shadow 耦合);若走委员会 belief-后验版则须落不震荡锁(1-tick 陈旧+damp<1 有界)。

**★★ wire 受理硬前置(审查60,先于 wire)**:
1. **P6.5 覆盖 delineation doc(★必须,落 SD double-count 教训)**:显式写清 P6.5 SoleResidentRecapture 覆盖 X(单 resident+sleepad-anchored 回床)/ Neighbor 增量 Y(兄弟房 radar 见人但**未 sleepad-锚床**,P6.5 够不到)/ 重叠 Z(双压,单-resident-FP 同向安全不增漏报)。**委员会亲查结论:Neighbor 非纯 double-count,有真增量** Y → 值得 wire(不同于 StandDuration 退役)。**这是受理 wire 的前置,先做。**
2. **双向 R5 锁**:r5_calibration_lock 现仅锁"Neighbor 压 fall 是许可压制(<1)";**须加多-resident 单元 case 证 ObsNeighbor gate-OFF 不压真摔**(漏报方向锁,对齐 P6.5 skip-multiresident)。
3. **N-5 不震荡锁**(若走 belief-后验版):测证跨房读序不 runaway。用户 raw-state 版此项 moot。
4. **corroboration≠substitution 锁**:ObsNeighbor 压 fall 须邻房**真变占用**(正证据 occ),非"邻房无信号"裸 absence;**Conf 反映真占用确定度而非缺信号**(同 np=0/NoDetect realness)。落测。
5. **整单元 redis-replay 真验**(ObsNeighbor 8 案 populated+R5+9 红 0 新增+多占用案不压真摔)——**待用户环境**(unit201 三设备谁跑);**shadow-first 可先建(R0 安全)不阻塞**。

**落地约束**:shadow-first(R0 log-only,不碰生产 #1/R1)。验证用 **synthetic 2-room Go 测试**先行(同 unit 两房,一房 fire enter/InBed→另一房 ObsNeighbor populated+压 fall;多 resident→gate-OFF 不压真摔;>60s stale 不喂),真实 replay 待用户。放行 bar:build/vet/belief 绿+R5-lock 绿+9 红 0 新增。

**审查60 裁决**:locus 批准(N-1①/N-2①/N-3 复用 SoleResident/N-4 max/N-5 1-tick),**wire 前置=P6.5 覆盖 delineation doc 必先做**。新会话开工顺序:① reconcile N-1(用户 raw-state vs 委员会 belief-后验)+ 写 P6.5 覆盖 doc(doc-only,push)→ ② 委员会受理 → ③ shadow-first 建机制+synthetic 测 → ④ 用户给整单元 replay 环境 → 真验。

**★进展 2026-06-08(commit 3c4f901,owlBack/doc/feedback.md 倒序顶)**:**步①完成已 push,待委员会受理(步②)**。两件 doc-only 交付:
- **N-1 reconcile 已提**:改裁 N-1①belief-后验 → **N-1-raw**(读兄弟房 roomLedger lastEnterMs/lastExitMs[track_manager.go:131-132,roomLedgerEmpty:502]+BedOccupancyState[:880],room+bed 双喂)。三论据:①**彻底 moot N-5**(读原始 ts 非后验,无 shadow 耦合,wire 前置#3 整条 moot);②**破委员会驳②"窄"**——room ledger 由 radar EnterRoom/ExitRoom 驱动[:1066-1070]非 sleepad-only,广度==①;③corroboration≠substitution 原生满足。**零新管**:engine 持 `e.rooms map[string]*TrackManager`[engine.go:158],同 `beliefShadowFor` 的 `(e *Engine)` 路径。**委员会若坚持①则 N-5 锁回 required**。
  - **★60s 语义(用户实时定档,改正早前"recency"误解)**:60s **不是邻房占用相对 now 的新鲜度**,是**同-unit 两房事件的因果相关度阈**——`|邻房事件ts − 本房触发ts| ≤ 60s = 相关(人从本房走到邻房)`;`>60s = 两房事件相互独立 → 不联动不压`。本房触发 ts = lost-track `st.lastSeenMs`[belief_shadow.go:348]。⟹ **仅 lost-track 分支**做相关度查邻房(**非每 tick 无条件喂**);`NeighborRoomSignal/NeighborBedSignal(triggerMs,window)` 判 `|邻房 lastEnterMs/BedStatusTs − triggerMs| ≤ window`。窗口可调 `FallParam.Neighbor.CorrelationWindowMs`(默认 60s,归 Lost 类,可调长)。
  - **★用户拍 A(纯相关度,先简化)**:>60s 即无相关→不压,**不加**"sole-resident 长躺邻房床"的静态人证兜底(B 作后续可选增强)。N-3 sole-resident 门仍在(复用 SoleResidentRecaptureState residentCount==1,归因安全:多 resident→不压),**与相关度正交**。
  - **★★ shadow-first 已落地(commit 88870d2,审查61→62→施工方交付链)**:审查62 已**销项 durable 挑战**(认 60s=correlation gate 非 freshness),与实现一致;「非全空间监控」铁律=correlation-not-durable 的安全法理([[partial_monitoring_fall_suppression_law]])。落地物:`FallRulesParam.Neighbor.{HandoffWindowMs:60_000,JitterMs:5_000}`(fall_rules_param.go)+ `NeighborRoomEnterMs`/`NeighborBedHandoff`(track_manager.go,自锁)+ `neighborHandoff`(belief_neighbor.go,N-2①/N-3/N-4/N-6 单合并 occ=room OR bed conf=max)+ beliefShadowTick lost-track 路径喂 ObsNeighbor + **stale_corr no-silent-caps LOG**(`belief_shadow_neighbor_stale_corr` 记留驻 gap)。**审查62 三建中条件全交**:①N-6 单合并 ②stale_corr LOG ③双向 R5(似然层 belief `TestR5LockNeighborCorroborationNotSubstitution` occ=0→1.0 + gate 层 roomengine synthetic⑤ multi-resident gate-OFF)。synthetic `TestNeighborHandoffSuppress` 6 case 全绿;build/vet/belief 绿/9 红 0 新增。**判据=fresh 有向 `|邻房ts−st.lastSeenMs|∈[−5s,60s]`,非 durable**。**生产-gate-前待**:#5 整单元 redis-replay(unit201 三设备待用户)/N-6 终验过压/DBN recall 真摔集。
  - **★审查63 验收 + 审查64 撤回 N-7(防再被带偏)**:审查63 验收 wire(R6 全绿/3 条件真交/死源 5/5),挑 N-7(N-3 只数 resident,1 resident+1 visitor 单元 rc==1 放行,room ledger 身份盲→visitor 进兄弟房误压 resident 真摔),提 headcount==1 硬门。**审查64 用户纠+委员会自纠撤回 N-7**:ObsNeighbor 是**软似然因子(damp0.7)非 hard 否决**,lost-track 从未知出发只是抬后验;visitor 巧合=低概率噪声(需恰在窗内有向进兄弟房);**headcount 硬门净负**(废掉任何有 visitor/常驻 caregiver 的单元=lost-track 二义最常见多人居所,违审查62「A 纯相关先简化」)。**N-7 降级=recall 验证噪声 caveat(非 defect 非前置,不改设计)**;plan-B 软 down-weight(headcount>1 降 conf 非硬 off)可选非必须。⟹ **别再加 headcount 硬门**。
  - **★验证 blocked on data + harness gap(2026-06-08 收口)**:生产-gate 3 前置(#5 整单元 replay/N-6 终验过压/DBN recall 真摔集)**全需 unit201 数据**(CD2B+1641+333B),无数据→验证暂停。**逮 harness 前置 gap**:`bReplay`(belief_b_replay_test.go:99)**只 RegisterRoom 单房**(:106)→ neighborHandoff 找不到兄弟房→ObsNeighbor 单房 replay 永不 fire(结构使然非死管);**Neighbor 真验须先扩整单元 `bReplayUnit`**(multi-room 同 suite+roomSuiteID+suiteCensus seed+per-device 路由+跨房 ts 合并)。验证 spec=`owlBack/doc/neighbor_verification_spec.md`(V1 populated/V2 stale_corr/V3 N-6 无过压 damp==0.7 非 0.49/V4 R5 多resident gate-OFF/V5 回归+recall 扣 N-7 visitor 噪声)。**loop 静默暂停=blocked on data 非 blocked on 设计**;数据到→实现 bReplayUnit→逐跑。
  - **★全-pipeline 证明已绿(belief_neighbor_pipeline_test.go,bReplayUnit 式合成整单元)**:raw record→生产 `handleMessage`/`handleEventMessage`→`ParseRadarTracks`/`ProcessFrame`→`SnapshotTrackStatuses`→真 `beliefShadowTick`→lost-sweep→`neighborHandoff`→`belief_shadow_neighbor_handoff` fire(非直驱 tick 单测)。借真 cabb layout 几何+合成 in-FOV 帧(时序可控);本房 A 5 帧 track#1→t0+66s decoy(id=2)→超 60s TTL 真 lost-sweep;兄弟房 B EnterRoom 经真 handleEventMessage→lastEnterMs 窗内。两 case 绿:sole→fire / multi→gate-OFF。正中审查65「bReplayUnit 现在就建别 block on data」建议。**踩坑(unit201 真 replay 复用)**:①device_addr 须 **canonical INET**(`a::1` 非 `a:0:0:1`,否则 FromStreamMap 规范化后 deviceRoom miss→dropped_unrouted)②radar `track_id` 须 **0–8**(track_parse.go:50,>8 被滤)③**filtered/空帧只 NoTargetTick 不进 beliefShadowTick**(engine.go:1925,触发 lost-sweep 须有效帧)。**仍非 unit201 真 redis-replay**(待用户数据+bReplayUnit 载真多房 fixture 跑 V1–V5+recall)。
  - thread-safety 已核:beliefShadowTick **不持 e.mu**(engine.go:1009 在 RLock:1015 前;且已调 SuiteIDForRoom 这个 RLock 方法)→ helper 可安全 e.mu.RLock 快照兄弟房(无重入死锁,避 [[radar_realtime_subscribe_not_delivered]] 坑);tm 信号访问器自锁 tm.mu(BedOccupancyState 同款)。锁序 e.mu→{census.mu|tm.mu} 无环。
- **P6.5 覆盖 delineation 已落(亲查坐实)**:P6.5 触发=`geom==GeomInToilet`硬门[belief_shadow.go:339]+`SoleResidentRecaptureState` recap(=residentCount1∧(SleepadAnchored∨AnchorRoomType==RoomTypeDefault bedroom))[suite_census.go]。**X**=单resident+浴室丢轨+回床/bedroom;**Y1**=非浴室 geom 丢轨→GeomInToilet 门不 fire(直击 CABB/John.Y);**Y2**=兄弟房站着无 sleepad 接触未回床→recap=false P6.5 不抑制;**Z**=重叠双压单-resident-FP 同向安全。**结论:非纯 double-count 有真增量 Y 值得 wire**(异于 StandDuration 退役)。
- **N-3 收下**:复用 SoleResidentRecaptureState sole-resident 语义,不新建第二 census 阈,多 resident→gate-OFF 不喂。
- **受理后(步③)**:shadow-first R0 log-only 建 + synthetic 2-room Go 测;放行 bar=build/vet/belief 绿+R5-lock 绿(待补多-resident 漏报 case=wire 前置#2)+9 红 0 新增;整单元 replay 待用户 unit201。

**协作协议**:施工方(本 agent)↔审核委员会(doc/feedback.md 倒序)。裁前不建/需新岔口列选项不擅决/直接 main 不开分支/脏文件勿动/commit+feedback 倒序+push/冲突 rebase --continue 勿 abort/skip(tenant3.owl.zone 地雷)/cd wisefido-sensor 跑 go。详见 [[belief_state_rule_engine_reframe]]。
