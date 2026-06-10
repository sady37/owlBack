---
name: lost_track_fall_detection_envelope_gate
description: "lost_track 误报治本=丢轨平面距 > d_fall(实测) 则抑制;fall 检测半径 < person 因贴地弱回波(非几何),按支架实测标定"
metadata: 
  node_type: memory
  type: project
  originSessionId: 9d69b66d-2014-414e-a786-d3a3658c2ecb
---

2026-06-03 诊断 Denver-101/E598A2ACD523 一条 CRITICAL Fall:实为 **lost_track 误报**(producer=fd00:0:fff1::1 是 wisefido-sensor roomengine,非 firmware;payload reason=lost_track,context=track_lost_no_exit_room_no_recovery)。人在距雷达 ~5.8m 边缘静止后丢轨(np=0),D523 不发 ExitRoom→等待超时硬推跌倒;6min 后 np=1 站姿复现,证伪。alarm_events 两时间=triggered_at(still_box锚点)+alerted_at(engine开火),非 start/end。

**核心(已纠偏)**:能"检测到人"(胸肩~150cm)≠ 能"检测到跌倒"(地面0–30cm),且 **fall 检测半径恒 < person**。原因是**贴地/躺姿目标回波弱**(RCS 小+地面杂波/多径),有效信号半径腰斩——**不是几何/FOV 锥效应**。雷达=加特兰 Calterah 4T4R,vfov 非 layout 标的 120、波束非干净对称锥;**纯几何 cone 公式会得出 fall≥person 的反结论,已废弃**。d_person/d_fall 必须**按设备+支架实测标定**。

**简化标准值**(A 支架实测取整,c=800斜距信号半径):d_person≈**700cm** / d_fall≈**500cm**(恒 d_fall<d_person;跌倒等效斜距√(500²+200²)≈538≪800,证明受弱回波限非斜距限)。各支架可微调,工程默认用此对。角度约定 `a=90−与墙夹角`。仍成立的几何只有斜距↔平面转换 `d=√(R²−(h−z)²)`。

**Why**:lost_track 凭"消失"反推,丢轨点超出 d_fall 时雷达本就给不出地面证据(无 pose=5),是结构性盲猜。

**浴室近距 FP 是另一机制**(非距离):firmware track_id 不稳→真人 track 空间跳变/分裂漂成 ghost,出生合法→一路 Real 到死触发 lost_track(D5F7 0603 案:track0 09:43:19 跳到反射区冻成 ghost,真人重编 track1;rule1 出生地只查 newBorn 漏判)。**治法=logicID(uidlast4+track_id+mmssms 稳定身份)+ 跨 track_id 数据关联(新 track 无 enter→沿用最近 track 的 logicID)+ 跳变即重判 + 空房窗口闸(ExitRoom/np=0 后 track 死亡不报)**。完整解题思路+fixture 见 doc/cases/d5f7-bathroom-lost-0603-0941-jump-FP/README.md。

**2026-06-03 已部署(基础设施)**:① logicID 出生分配(makeLogicID,uidlast4+track_id+mmssms)+ 进 verdict payload(observation.Track.LogicID/`logic_id`)；② **sensor_decision_log 表**(owlRD/dbv2/66_,hypertable 30d)落 sensor 全部决策审计(verdict_change/lostfall_pending/cancel×3/suppress/fire,带 logic_id+evidence),为 DBN 复盘"决策+特征"层；管道=sensor emit→`ai:track:verdict:stream`→**iot 第二消费者** InsertSensorDecision→表(cardagg 按 category=track_verdict 跳过 sensor_decision);CLAUDE.md 规则 2.1 已改(反转"不入库")。e2e 注入验证通过。③ **logicID 最近邻继承已部署**(2026-06-03):出生时若 `!hasRecentEnterRoom(3s窗)` → `nearestAliveLogicID`(Kalman 画布坐标最近的存活 track)继承其 logicID,否则全新(makeLogicID);firmware track 重用/跳变/分裂的同一逻辑目标身份连续。日志 `logic_id_inherited_no_enter`。② **最终安全版已部署**(2026-06-03,**铁律:ghost 不进 Fall 决策**——先前 sticky `JumpRejudgedGhost` 方案已整段回退,因它用 ghost 门控 Fall 违例且 firmware id 横跳致真人也被标 ghost 漏真摔)。改为**两道不读 ghost 的身份/守恒闸**:
(a) **id-swap 守恒闸** `hasOtherLiveTrackWithLogicID`:track_id 失锁但其 logic_id 仍活在另一条 track 上 = firmware 换 ID 游戏(数量守恒,无人真消失)→ 不入 pending。
(b) **空房账闸** `roomLedgerEmpty`:`lastExitMs > lastEnterMs`(**只信 EnterRoom/ExitRoom 过门空间证据,绝不信 np=0**——铁律 count=0≠离开,人摔倒丢锁也 np=0,用 np=0 会漏真摔)→ 房已空,失锁 track 是"人走后残影(冻住的反射)"→ 抑制。治 D5F7 残影在 ExitRoom 之后才消失、旧 ExitRoom-cancel 漏掉。
设计原则:**Fall 只与 pose**(firmware pose=5 走 qinglan 不受影响);lost_fall 仅"真人逻辑槽在守恒下真丢失(无替补+非过门离开)"才触发。**D5F7 fixture replay 实证 reported 1→0**(track1 id_swap 跳过 + track0 room_empty 跳过,全程无 ghost);oracle 全 13 案 真跌倒 confirm=true / FP confirm=false 不变。三服务已重启部署,待用户进 101 bathroom 复测。**显示层残影已修(2026-06-03)**:人走后 web 显示 track >60s(超前端残影上限故非残影=上游续命)。根因非 qinglan(88 必须发=device online 心跳),是 **cardagg monitor_handler 收 88 没按约定(monitor_buffer.go:13「88不入库」)清 track,反而照写 + 每1s 用新 ts_ms 重发旧位置**(writer.go:256)→ 前端按 ts_ms 永不过期。修:cardagg 收 88 → `MonitorBuffer.ClearDeviceTracks` 清该 device 全部 track + 不写 88 + 留 TouchLastSeen(88 仍维护 online)。前端 6s presence 窗(types.ts `DEFAULT_PRESENCE_SEC`,跌倒姿态 30s `FALL_PRESENCE_SEC_DEFAULT`)平滑帧间偶发 88 不闪;持续 88→≤6s 消失。受控注入验证通过。纯显示层不碰 Fall。详见 case README「显示层残影修复」。

**剩余**:① 完整 enter/exit 数量守恒 slot 层(本次聚焦实现:logic_id-swap 闸 + ExitRoom 账,非完整数据关联层);② np=0+门区位置合取的离开判定(D523 无 ExitRoom 类);③ release 模式 ghost verdict 显示过滤(当前 sandbox 仅 log)。replay 命令见 case README。

**How to apply**:lost_track 触发前加距离闸——丢轨点**平面**距 > d_fall(该支架实测,A=550) → 抑制或降级不发 CRITICAL;不可用 presence 半径(d_person,更远)当判据;不可拿 raw 斜距直接比(先投影成平面 d)。**此闸仍是 TODO 未落代码**(fall_rules_param.go 里 lost_track 无距离判据)。权威文档=owlBack/doc/AI_fall_detect.md §3.7(已按实测重写)。关联 [[track_coords_vs_grid_enter_frame_mismatch]] [[number_people_zero_exitroom_fallback]] [[fall_rules_three_classes]] [[fall_fp_roots_and_todo]]。
