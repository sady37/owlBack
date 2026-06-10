---
name: bed_stale_leftbed_vetoes_radar_inbed
description: 2026-06-05 陈旧sleepad LeftBed(sticky-2.94无衰减)永久否决fresh radar InBed;治本=L指数leak+event年龄门控+bed_status=8待机;sim已验证未port
metadata: 
  node_type: memory
  type: project
  originSessionId: a821a2af-8448-4576-87bb-f1a826e65a17
---

2026-06-05 test#3 实测 production bug + 治本设计(sim 已验证,Go 未写)。

**根因**:`bed_bayesian_scorer` 的 explicit event(InBed/LeftBed)是 sticky 且**无时间衰减**——`sleepadClusterLRLocked` 里 `if leftBedHappened: return -2.94` 恒定,每分钟重复累加把 L 压到 floor −5。sleepad "离床即 noreport"(离床后不再发任何帧),所以那条 LeftBed **只能被一条新 sleepad InBed 清**,sleepad 一哑就永远清不掉。后果:37 分钟前(test#2 19:18:26)那条 sleepad LeftBed 到 19:55 仍 `sleepad_lr=−2.94`,**压过 fresh radar InBed 的 +1.39**(净 −1.55/min)→ P≈0.03 → radar 明明发了 InBed(19:55:58),融合 bed_state 仍判 Vacant/未在床。production trace 实证(bed_decision_trace,bed=fd00:0:3:111:3:101::/96)。

**纠正旧误判**:之前说"sleepad 哑=返回0中性不否定radar"只在 sleepad 从未发过任何事件时成立;一旦发过 LeftBed,走 `leftBedHappened` 分支=−2.94 活跃负证据,**确实在否定 radar**(陈旧 sticky)。这是"他可否定"原则碰上"无新鲜度"的恶果。

**治本设计(用户拍板,对称,5min)**:① event LR 按年龄门控:控制性 bed event 超陈旧 → 整簇贡献归 0(**卡 event 年龄,忽略 vital**——空床期 radar.heart 每分钟发但无效);② L 指数遗忘:无 fresh 贡献的分钟 `L *= ~0.63`(从 floor leak 到 |L|<0.5 约 **5min**,镜像上床);③ 待机带 `|L|<0.5 → bed_status=8`(待机,`observation.FieldBedStatus` 已有 8=未改变/待机),不硬判 1。从 8 中性基线 fresh radar InBed 干净翻 InBed。

**refresh 判据(关键,防误衰减在床人)**:radar 无有效 vital(本 install vital_flag=0 全无效,radar.heart 空床/在床同 payload),**所以靠"radar track 在床区出现"refresh**。实测:**pose=6(躺)单独不行**(在床期 gap 达 114s/56s,人坐起 pose=4 时断 refresh);改"**床区任意 track(area_id==inbedAreaID,任意 pose)**"→ 两组在床数据 max gap 仅 **2s,0 次超 35s**,稳。机制:InBed 时记 `inbedAreaID`(Enter2Out 带 area_type=2 普通床 + area_id;本数据 area_id=1),只要任一 track 仍在该 area_id → 刷新 `radarInBedLastTs`;该 area_id 不再出现+无 LeftBed → 超 5min → leak → standby。**协议**:Enter2Out event 1进房/2离房/3进区/4离区/5进监护模式/6退监护;area_type 2普通床/5监护床/6感应区;本系统 InBed=event3+area_type2,LeftBed=event4+area_type2(非监护模式5/6,本 install 未用)。

**接线点**:zoneengine 床 scorer 当前**事件驱动,不订 radar.track 1Hz**;实现 area_id 刷新需给 scorer 喂"track 仍在 inbedAreaID"信号——倾向复用现成 `OnRadarPoseLying` 通道,改判据 area_id 匹配即刷新(不卡 pose=6);adapter_radar 已把含 area_id 的 fields 作 TriggerData 传入,area_id 可得。

**2026-06-05 presence 接线改向(用户)**:不用 InBed 事件抓 area_id(inbedAreaID 那套作废,下次撤)。床区是**静态声明**,在 `room_visual_layout` 的 radar 对象 `device.iot.radar.areas[]` 里 areaType==2 那条(本例 9e7 areaId=2),wisefido-data 维护——按需/缓存读床区 areaId,"track 在床区"=**反向查 monitor_stream 最近 N 分钟有无 track 的 area_id==床区id**(每分钟或算衰减时算),不需实时逐帧喂 scorer。Track.AreaID(已通)是反向查的依据。关系图(covers/samebed)天生在 layout+/96 前缀,详 [[mm_relationship_matrix]]。**待撤**:scorer `OnRadarInBed(areaID)` 参数 + `inbedAreaID` 字段 + `OnRadarBedAreaPresence` 按 inbedAreaID 匹配 → 改成查 layout 床区 + 反向查。

**2026-06-05 observation.Track 对齐 fields.go(全 7 模块 build+test 绿,已部署)**:Track 之前三字段都没按 fields.go——①**AreaID 原本缺失**,补 `int`+From/ToFieldMap 映射(omit-on-0)+qinglan monitor(radarTrackToData)与 event(buildEnter2OutTrack)两路填充;raw monitor radar.track **只有 area_id 没 area_type**(area_type 仅在 Enter2Out 事件,int 2床/4门/5监护床/6感应区)。②**BedStatus `*int`→`int`**(用户拍板不用指针,对齐 fields.go TInt 0/1/8):加 `observation.BedStatusInBed=0/LeftBed=1/Unchanged=8` 常量;**指针零值坑**=去指针后 Go 零值 0=在床会泄漏,故 10 个 `observation.Track{}` 构造点(qinglan/sleepace 心跳+radarTrackToData、roomengine 各 fall/alarm 输出)全部显式置 Unchanged;ToFieldMap 省略 Unchanged、FromFieldMap 缺键默认 Unchanged;sleepad 实时床态走手搓 map 不经 Track 故不受影响。③**AreaType 死字段删除**(无读取点+类型错 string vs int+monitor 无数据+没映射)。card.BedState.BedStatus 是另一结构(int 已对,不动)。

**sim 验证**(export_1137_1146/sim_decay.py 喂 test#3 真实证据):19:18 floor→进 Standby8→19:55:58 radar InBed→P=0.80 InBed ✅,不再被陈旧 −2.94 否决。

**2026-06-05 Go shadow 已写+部署+实测**(bed_bayesian_scorer.go 加 shadowL/inbedAreaID/shadowDecayFactor=0.63/shadowStandbyBandL=0.5;OnRadarInBed(nowMs,areaID);OnRadarBedAreaPresence(未接线);engine.go bed_decision_trace 打 shadow_l/shadow_decision/inbed_area;build+vet+test 全绿,producer-first 不动现网 Decision)。实测(21:13 上下床)**shadow 抓到两个自身坑**:① **area_id 丢失(根在 canonical 类型)**:`observation.Track` 结构体缺 `AreaID`(只有 AreaType),且 `FromFieldMap`/`ToFieldMap` 也没映射 → 任何走 Track 的路径(尤其 MonitorBuffer presence 路径)都丢;qinglan `buildEnter2OutTrack` 事件 map 也没拷。**已修三处**:Track 加 `AreaID int json:"area_id"` + From/ToFieldMap 读写 FieldAreaID(根修,走 MonitorBuffer);qinglan buildEnter2OutTrack 加 `"area_id"`(事件路径,与监控路径不同不冗余)。全模块 build+observation test 绿,sensor+qinglan 重启生效。mqtt rx 日志 area_id=1 是 describe 从原始读的,解码输出原本没带。② **shadow 脉冲模型错**:用"事件加脉冲"导致上床+1.39/离床−1.39 **抵消成 0→Standby**(应判 LeftBed);LeftBed 物理上取代 InBed 非抵消。**待修**:`shadowTickLocked` 改成每分钟用 fresh 门控后的 cluster-LR 累加+双簇陈旧才 leak(即 sim_decay.py 模型),弃加性脉冲。③ presence 接线仍未做(radar.track→OnRadarBedAreaPresence)。下次:先修坑2 shadow 模型,再接 presence,再验 LeftBed 衰减+再上床不被否决。

**待办**:① 时间参数:LeftBed→Standby 当前 ≈9min(5阈+4leak),用户若要总 5min 则砍陈旧阈让 leak 立即起算;② port 到 `bed_bayesian_scorer` **shadow**(并行 shadowL + shadow decision 进 bed_decision_trace,producer-first 不动现网决策);③ 下游 cardagg/zonealarm 认 bed_status=8 单独一轮(段4c 跌床不在 8 态武装)。两个置信度别混:决策用贝叶斯 P(z.state.Score),**发到 card 的 BedConfidence 是源固定 sleepace=90/radar=60**(translator.go bedConfidenceForSource)。关联 [[bed_fusion_authority_model]] [[bed_status_default_not_in_bed]] [[silent_leftbed_fall_recovery_window_gap]]。
