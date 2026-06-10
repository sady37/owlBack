---
name: mm_relationship_matrix
description: "2026-06-04 MM per-unit 空间关系方阵(#6)——DBN 静态结构底座；owl-common/spatial/relmatrix.go 已建+测试，sensor 当 maintainer"
metadata: 
  node_type: memory
  type: project
  originSessionId: e8750e7d-da7e-421d-98f1-fc1e56c285ed
---

MM = per-unit 空间关系**方阵**，做 DBN 的**静态结构层**（与时变证据/CPD 干净分离）。代码=`owl-common/spatial/relmatrix.go`（已建+测试，**未接生产**）。

**格式（用户拍板，几经简化）**：`M[i][j] ∈ [0,1]` 单 float。L[]=C[]=Obj[]（行列同一实体序）。
- **类型由 kind 对隐含**（任意有序 kind 对≤1 种关系）：`room→bed=contains` / `device→bed=onbed` / `radar→bed=covers` / `radar→sleepad=samebed`。格子只存**置信强度**不存类型 → 不用 bitmask。
- `1=确定`（前缀类 contains/onbed，或几何无歧义 covers）；`(0,1)=候选`（几何有歧义 covers，值=几何先验如 0.5）；`0=无`。
- **方向给不对称**：`M[room][bed]=1` 而 `M[bed][room]=0`；`M[sleepad][bed]=M[bed][sleepad]=1`（onbed + 反向 contains 都确定）。
- **关键洞察（用户）**：candidate(bit) + geomPrior(旁表) 两概念**塌进单 float**——"是候选"=值在(0,1)，"先验数值"=格子本身。bitmask+旁表全删。
- **不确定性沿派生传播**：`SameBedConf(r,s)=max_bed min(covers(r,b),onbed(s,b))`；候选 covers 0.5 ⊗ onbed 1 → samebed 0.5，正好作 DBN 先验。float 能表达，bitmask 不能。
- API：`Conf(a,b)` / `SameBedConf` / `FuseGroup(bed)→[]Source{Prefix,Conf}`（带权重的融合源，bed_bayesian/DBN 直接用）/ `BedsCoveredBy(radar)`。`RelOf(a,b,covers)` 是算单格的外部 helper（covers 几何谓词注入，spatial 不依赖 radar 包）。

**放哪**：类型/纯逻辑=**owl-common/spatial**（前缀免费 ContainsAddr，covers 几何注入）；管理/实例/失效=**sensor**（maintainer，几何 layout + spatial 绑定都在它手里，热路径消费方 zoneengine/bed_bayesian 也在）。**不开微服务**（RPC 进滤波内循环=性能+稳定性+学习落点全坏）、**不落库**（派生数据，从 layout 重建，落库=第二真相源违规则#1.3）。

**纪律**：矩阵**只存静态关系**（从 layout/前缀重建）。**运行时学习不回写矩阵**——15s 内 R 与 S 上下床同步把候选 0.5 推向 1，这是 DBN **后验**（belief），不动 MM 几何层。

**Q1（未知前 R 候选盖 A/B）**：`(R,A)=(R,B)=0.5`（"R 边界内一张床"→互斥候选，和=1）。不是 0（R 确实盖某床）不是 1（不知是 A）。
**Q2（确定后改不改）**：运行时**只动 DBN 后验、不动 MM**；长期收敛了才**固化进独立 learned 层**（同 float 格式，去抖+审计+可撤销，**跨重启暖启动**免冷学：DBN 开机先验=learned 层有则用否则几何层）。
**learned 层现状**：关系级**不存在**；cell 级假报学习**已有**（`feedback.go`→cell history→`engine.go SaveDaily`+retention，persist.go v3 FakeAlarmCount），要做照此范式做 `learned_covers` 旁表。

为 #6；服务 #2(Layer1 置信进 zone BedConfidence)/#5(多床房 covers 几何)/#7(DBN 映射)。与 [[belief_state_rule_engine_reframe]]（DBN 治本）、[[bed_fusion_authority_model]]（床态融合 Layer0/1/2）、[[layout_authority_ai_correction_model]]（learned 范式）同盘。

**已落地(2026-06-04，未 commit)**：① covers 真几何=`engine.RadarBedReachCount(deviceAddr)`(用 engine 自己 deviceRoom 定位，取床矩形中心+床高判 SignalReachable 数 m，避房前缀字符串匹配脆弱性)；MatrixCache.makeCovers=`m/n`(n=同房 bed 前缀数，同房判定用 netip /88 掩码)→单床房 m=n=1 得 1.0 确定 / 你例子 m=1,n=2 得 0.5 候选。**硬约束**：canvas 的 bed 对象**不带 /96 前缀**，几何无法贴 A/B 标签，只能数(故 m/n)。**long-sofa 坑**：long-sofa 在 layout 以 typeName "Bed"/"MonitorBed" 进 cfg.Beds(几何分不出真床)，但 DB beds 表无 long-sofa → m(canvas含sofa) 可 > n(DB真床) → 必须 `min(m/n,1)` cap；sofa 污染只会 saturate 到并列 1.0 → ResolveBed 支配规则判不出 → 留空退回 radar /96，**不误归床**(安全)；彻底分需前端给 long-sofa 独立 typeName。② maintainer=`wiring.MatrixCache`(main.go wire，从 SpatialCache 前缀+engine 几何 Build)，失效=启动 BuildAll + **config:card commit**(layout/bind 的 DB 提交，**不是 fw-ack**——covers 源自 server room_visual_layout 非 firmware 状态；config_card_consumer 加 SetMatrixInvalidator)InvalidateAll lazy 重建 + 每天本地 1AM 全刷。③ **#5 真消费已落地**：BedResolver 接口 `ResolveSingleBed(room)`→`ResolveBed(radarAddr,room)`，MatrixCache 实现：BedsCoveredBy 取最高置信床，`≥阈(bedResolveConf=0.8) 且严格压过次高`→解析(单床 1.0/DBN 后 0.95)，多床候选 0.5<阈→""(退回 radar /96 待 DBN)，无覆盖→兜底 SpatialCache.ResolveSingleBed(单床 parity 不回归)。**0.8 阈卡在 0.5 与 1.0 间→今天多床房永远 ""(零行为变化零风险)，DBN 把候选推到 0.95 时同段代码自动点亮多床房路由,无需再改 wiring**。main.go SetBedResolver 必在 zone.Start 前(避 adapter goroutine 竞态)。
④ **#2 模型 A 已落地**：FuseGroup covers 权重进 bed_bayesian——`ResolveBed` 返回 (床, covers置信)，经 `SignalEvidence.RadarCoversWeight` 串到 BedBayesianScorer.radarCoversWeight(默认1.0)，`radarContrib = radarLR · radarCoversWeight · γ`(covers 两向打折=空间结构权重，与 γ 时间退让正交)。候选(0.5)半权/确定(1.0)全权。**今天零行为变化**(单床房只 covers=1.0 路由→权重1.0)，纯为 DBN 连续 covers 铺垫;On* 直调默认全权向后兼容。未选模型 B(软扇出改 routing+融合核,风险大)。test=TestBayesian_RadarCoversWeight。
owl-common/spatial relmatrix.go + wiring unit_matrix.go + engine.go(RadarBedReachCount/roomBeds) + adapter_radar(接口 ResolveBed 返 (床,conf)) + setup.go + bed_bayesian_scorer.go(radarCoversWeight) + types.go(SignalEvidence.RadarCoversWeight) 均 build/vet/test 全绿，未 commit。**剩 #7**(DBN 把 covers 候选 0.5→0.95，自动越 0.8 阈点亮多床房路由 + 加权融合)。

**2026-06-05 关键修正(用户洞察)**：关系图**天生在 room_visual_layout(wisefido-data 维护)+ spatial /96 前缀里**，不需要 sensor 触发更新/周期计算派生。① **radar→bed covers 是显式声明**，不必只靠 RadarBedReachCount 几何 m/n 推断：`room_visual_layout` 的 radar 对象 `device.iot.radar.areas[]` 每条带 `{areaId, areaType, objectType, vertices}`——**areaType==2=床区**(本例 9e7 areaId=2 objectType=Bed),用户画 bed object 转 declare_area 下发,wisefido-data 存。② **samebed/onbed 在 /96 前缀**：同床 radar+sleepad 同在 bed /96(如 9e7 `...3263:9e7` 与 978 `...2470:978` 同 `fd00:0:3:111:3:101::/96`),devices 表天然带,前缀相等即 samebed。③ 故 MM 应**按需从 layout+前缀读**(wisefido-data 唯一真相源),不另维护周期失效的 MatrixCache(与规则#1.3 单源一致,之前那套 sensor-maintainer 偏重)。**多床房 /96-A/B 仍需补**:radar.areas 的 bed objectId 不带 /96(同 canvas bed 无前缀的老坑),仍需 m/n 或前端给 objectId↔/96 映射。**bed-presence 应用**:床区 areaId 直接从 radar.areas 取 areaType==2,"track 在床区"=monitor_stream 的 area_id==该值(反向查最近几分钟,Track.AreaID 已通);**撤掉 InBed 抓 area_id 那套(inbedAreaID)**,详 [[bed_stale_leftbed_vetoes_radar_inbed]]。

**2026-06-05 失效机制改造(已实施,build/vet/restart 全绿)**:① **删 1AM 定时全刷**(`Run()`+`durationUntilHour`+main.go `go matrixCache.Run` 全删)——改纯事件驱动。② **config:card 改 scoped 失效**:消息已带 device_addrs/cards,consumer 不再 `InvalidateAll`,改 `MatrixCache.InvalidateScope([]string)`——每个 device/room/card → `maskToUnit` /80 → 只删那些 unit(懒重建);scope 空才兜底 InvalidateAll。`MatrixInvalidator` 接口加 InvalidateScope。粒度=per-unit(matrix 本就 per-unit;"device 只更 device"=重建其 /80 unit,逐 cell 更新不值)。③ **补上缺失的触发**:实测 `SaveRoomLayout` 之前**根本没发 config:card**(只直调 feedback-veto),engine 几何靠 22:00 批量 reload、MM 靠 1AM——所以删 1AM 若不补这条=回归。已让 `RadarInstall.SaveRoomLayout` 在 canvas 变更后 `configPublisher.PublishConfigChanged("update",[spatialPrefix],...)`(RadarInstall 加 configPublisher 字段+SetConfigPublisher,main.go 注入)。**链路闭环**:FE trajectory save→layout DB+UpdateConfig→config:card→sensor `ReloadRooms`(registerAllRooms→LoadRoomCanvases 刷 engine canvas 几何)+`InvalidateScope`(重建该 unit MM)。monitor-setting 走 config:alarmDevice 不碰 MM(✓ 用户要的)。改:unit_matrix.go/config_card_consumer.go/sensor main.go + radar_install_service.go/data main.go。**注**:covers 几何仍用 canvas(RadarBedReachCount),未切 radar.boundary——radar.* 已落 spatial_config(见 [[radar_config_snapshot_spatial_config]])但 MM 暂未消费,切换待定(会引入 UpdateConfig async-verify 与 layout-save config:card 的时序竞争)。
