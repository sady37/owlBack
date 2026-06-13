# 接续 owlBack 跌倒误报治本 + belief 自学习闭环

> 新会话直接读本文件开工。最后更新 2026-06-01。

## 背景与北极星
owlBack/wisefido-sensor 跌倒检测误报治本。最初设计意图(北极星):
**持续输入设备实时值 → 白盒信念滤波算出连续"危险系数" → 人工 feedback 离线再估计概率表 → 模型间接自学习。**
关键约束:要白盒贝叶斯(HMM/DBN),不要黑箱 AI;能注入人类规则、每个数字可解释。

## 先读(权威上下文,按序)
内存(auto-load 的 MEMORY.md 里有索引,重点这几条):
- belief_state_rule_engine_reframe —— 总纲,信念状态机替 gate-list;含本会话所有进度
- lost_track_ghost_npzero_door_exit —— lost_track 误报修复 + np=0 铁律 + 转移矩阵结论
- feedback_no_dynamic_threshold_modulation —— 铁律:派生信号不进实时 alarm 决策
- fall_fp_roots_and_todo / d5f7_ghost_cases / radar_layout_device_invariant

文档(owlBack/doc/):
- belief_dbn_completeness.md —— **完备性蓝图**:`L×A=state` 为何不完备 + 缺的三段(① per-track 存在/真伪隐链 ② P(无观测|s) 发射 ③ HSMM 驻留)+ 15 个现存 gate→DBN 转移/发射/驻留对照表 + P1-P5 迁移分期
- room_belief_state_machine.md / belief_input_normalization.md / belief_gate_to_matrix.md / fall_rule_inventory_and_conflicts.md

代码:wisefido-sensor/internal/roomengine/belief/(纯模型) + belief_adapter.go / belief_shadow.go(production shadow,log-only) / belief_replay_test.go(真机回放 oracle,TestReplayOracle)

## 上一会话已部署(别重做)
1. bathroom_fall.go evaluateLostFallStrong:method-2(失锁前全 ghost 不 fire `lastBasesAllGhost`)+ 门区 exit 推断(`inferredDoorExit`:门区 AreaEnter + np=0 合取)。
2. track_manager.go:记 `LastNumberPeopleZeroMs`(np=0 只落 ts,不直接取消);池入场已显式 Real/Pending-only(:1124)。
3. belief_shadow.go:92 ghost 帧 `delete(sh.tracks)`(real→ghost→消失不误发 lost)。
4. belief_replay_test.go:加 d5f7-bathroom-fp-0601 fixture(101 浴室真实 FP)+ ghost-verdict 注入验证 + lost-sweep ghost guard。
全部 build/vet/test 绿,sensor 已 restart 部署(shadow 仍 log-only,未碰 fire 路径)。

## 关键结论(已锁定)
- 转移矩阵 →Empty 列**已对**(占用态全 0、SFallen 近吸收);要调的是 likelihood np=0"强证据空房"(降弱+SFallen 中性),但**优先级低**——本 case 真防线是 ghost verdict 不是 np=0。
- np=0 是**确认证据非替代证据**:count=0≠人离开(镜面 ghost/水气衰减都假报);只能与门区空间证据合取。
- belief 与 gate **完全同构**:治本三件套 = ghost adjudicator verdict + "ghost 消失≠倒地" + 真人照报,缺一不可。

## 下一步要做的事

**【2026-06-01 DBN P1 已实施+部署】** 用户拍板**先补 DBN 再自学习**。Track 层 `T_t` 只读 shadow 已落地:`belief/track.go`(T 五态+A_T+TrackBelief)+`track_test.go`+`TestTrackLayerOracle`+`belief_shadow.go` wire(只 log `belief_shadow_track_lost` 不 fire)。三结构核心(method-2=A_T Ghost→Lost floor / 门区 exit=absent geom+TObsExit / absent 不分 Lost-vs-None 靠先验)。oracle:D5F7真摔 Lost✓ / ghost+exit+D523+cd2b →None/JustLeft✓ / cabb+Hunzi 静止站立 Lost(P1 不分,留 P3)。build/vet/test(-race)绿,sensor 已 restart 0 panic。doc `belief_dbn_completeness.md` §6.1。详见 memory 最新 bracket。

**待用户测试数据(2026-06-02 计划)** → 导入 oracle:用户将做 4 组标注测试,**填好 `scripts/test_cases_0601.manifest`(device_uid+时间) 后跑 `scripts/export_cases_manifest.sh`** 自动导 fixture+生成 oracle 条目:
- 101/201 bathroom(镜像)ghost+lost-fall = **FP** → 验证 P1 Track 层(进 TestTrackLayerOracle+TestReplayOracle)。
- 101 bedroom 边缘静坐**假 firmware fall = FP**(新类!):当前 `ObsFirmwareFall→SFallen×10` 会继承假报,且 firmware Fall **未 wire 进 shadow**(engine.go:1594 只喂 Enter/Exit)→ 需 P2 geom-条件化 firmware-fall 似然 + wire。
- 201 bedroom **bedside-fall = REAL** must-fire(新类!):近床 geom 若判 InBed 会压成漏报(likelihood.go:132 PoseFallen@InBed×1.5)→ 测床边界 geom 精度。101bed(FP)+201bed(real)是标定**对**。

**下次开工候选**:
1. **离线对账** production `belief_shadow_track_lost` log vs gate-list lost_track alarm(攒几天真机数据,逐条 triage T 层 vs gate 分歧)。
2. **P2 absence 发射**:把 `lostWhileMovingToObs` 合成 ramp 换成正式 `P(no-detect|T,S)` 发射,删 ramp。
3. 扩 oracle fixture(cabb-fall-B/C、更多 ghost 段)+ 标定 A_T。
4. **P3 HSMM**:解静止-vs-走动丢失(cabb/Hunzi 在 Track 层都 Lost,FP 性=dwell);吃掉 confirmMs/MovingPreconditionMs/StillSec。

---

## (历史)自学习设计

**【2026-06-01 已完成】** `owlBack/doc/belief_feedback_learning.md` 已写(用户拍板 3 决策:标签需人工 ground-truth 闸 / 学习先定义安全子集可在当前模型学 / cell-counter 拆 (a) 保留升级为 belief geom 发射标定源+(b) inline modulation 退役)。详见 memory belief_state_rule_engine_reframe 最新 bracket。

**下次开工候选**(doc §8 v1 PR scope,动手前确认):
1. §2 摄取改造:`feedback.go` 去无勾选兜底(:373) + 加标签权重字段 + 导人工确认 fixture 进 TestReplayOracle。
2. §6 (a):cell 计数器接闸后只作账本,断开 → `toleranceFactor` 的运行时调用(belief 侧不复刻;gate-list 现役保留到 cutover)。
3. §4 geom-conditioned + 基础姿态发射的离线再估计器(计数→Dirichlet posterior→版本化产物),**只产候选表+报告,不 wire fire**。
4. §7 promote 流程 + shadow 对账脚本。
**不做(标注解锁阶段)**:转移 A/HSMM 学习(DBN 补全 P1-P3 后)、absence 发射学习(P2)、gate-list (b) 退役(cutover 时)、θ_fire/confirmMs 自动学(先人工攒样本)。

另一条线(若先推 DBN 而非自学习):belief_dbn_completeness.md 的 P1(Track 层 A_T 只读 shadow)——自学习的结构前置。

## 红线
- 派生信号不进实时 alarm 决策(feedback_no_dynamic_threshold_modulation)
- 白盒不黑箱;与现行一致=与 fixture 标签一致(不是逐行复刻 gate)
- shadow-mode cutover;producer-first;schema 改动先过 owlRD/dbv2 CREATE 提审
- 开发阶段可直接 sudo systemctl restart owlback.sensor;改前读对应 design doc

## 访问
DB: localhost:5432 postgres/postgres/owl_v2(raw 帧在 monitor_stream,ts=timestamptz,device_addr=inet;报警 alarm_events;事件 event_log.event_kind)
Redis: localhost:6379 pw TeLunSu-36kr(ghost verdict 在 ai:track:verdict:stream,全局仅留 ~500 条)
导出 fixture: tools/export/export.sh(unit 级 PG→window.json+window_sleepad+alarm+meta，读 owl_v2;老 scripts/export_case*.sh 已退役删除)。重放: tools/replay --fixture(纯文件零 DB)
D5F7=E598A2ACD5F7=fd00:0:3:111:3:300:a2ac:d5f7(101 浴室镜面雷达,card /80)

**开工:先读上面的 doc 和 memory,然后从 belief_feedback_learning.md 起;动手前先跟用户确认大纲。**
