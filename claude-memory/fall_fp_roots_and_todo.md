---
name: fall_fp_roots_and_todo
description: 2026-05-31 跌倒误报根源 + 止血→治本 TODO 开工清单(新会话解决);两 case fixture John.Y 9h person_silent + CABB lost_track
metadata: 
  node_type: memory
  type: project
  originSessionId: 7816c3f3-3c73-459e-bc46-7284c51b8a1b
---

2026-05-31 长 session 挖出的跌倒误报根源 + TODO,**新会话解决**。配 [[belief_state_rule_engine_reframe]](治本方向)。

## 两个 case fixture(验证集)
- **John.Y 9h person_silent**(card f...,room `fd00:0:3:111:3:100::/88` Bedroom,radar **D523**=`...100:a2ac:d523`): D523 **无床**(canvas Bed=0),人 05/31 05:32(UTC)走进床区(09E7+sleepad 那片 D523 看不到的盘),D523 把 track 冻在 (-390,30),census 不释放,9h 后 person_silent 在 ghost 上报;status active 永不 auto_resolve。sleepad/床层级与 D523 零交集,抑制使不上力。
- **CABB lost_track**(card yxmqgf,Bathroom,radar `4D8710D5CABB`=`...411:1:200:10d5:cabb`): event_id 1d7a162e。人在浴室 **Walking(pose=1×60+Standing×5,z 直立,从无 pose5/2)**,~16:24(UTC 05/30)走到门口 track 丢在 **(90,30)**=radar Enter 区内部(live query: Area1/2=Enter,x70-100/y0-80),**firmware 没发 ExitRoom**(门口中途丢 track 漏发,D 系老毛病),lost_track 等 5min 报 Fall(context `track_lost_no_exit_room_no_recovery`);status active 永不 auto_resolve。

## 根源(structural)
- **A. roomengine 不消费 device 级 layout(Enter/bed/boundary)** ← 最深根因。layout 全 /128 device 级、无 /88 room 级;engine_bootstrap JOIN `rvl.spatial_prefix=r.room_id(/88)` 不匹配 /128 → 这些房间**根本没加载 layout 几何**;grid cell 全靠学/衰减,门口=AreaActive 不是 Enter、床格衰减成 Unknown。→ 引擎无法用"track 丢在 Enter=离场""此 radar 无床 hasBed=false"去抑制。
- **B. 一房一雷达假设**。`RoomConfig.Radar` 单数(layout_parser.go:70 last-wins);`engine.mounts` 是 roomID→一个 RadarMount。多雷达房间(D523+09E7)第二台 boundary 静默丢。
- **C. 无统一 room 空间帧 / device-space 关联矩阵缺失**。每 device 一座 /128 孤岛。破局点(用户#2):**per-device 自描述矩阵**(device×{hasBed, enters+性质, 邻居 boundary 交集})**不需统一帧也能建**,但当前没建。
- **D. Fall 推断只信离散事件、不用 Track+layout 几何**。lost_track 离场判定被故意移出几何路径(track_manager.go:1106-1109 注释),100% 依赖 firmware ExitRoom 离散事件(中途丢 track 漏发);ExitRoom 取消(774-786)任意一条清全部 pending 不按 track_id—**处理侧没 bug,是 firmware 没发/没到**;person_silent 用 census LastActiveMs 不看丢失前 pose/z 也不看丢失位置;**pending 建时 cellArea 算出来了(1113-1117)却扔了**。
- **E. lost_track 结构上无法 auto_resolve**。报在已消失 track 上,没持续信号报 recovery、没 ExitRoom → 永远 active。需独立清除路径。
- **F. bathroom 专属标准 `StandingContinuousMin`(静止站立分钟)被绕过**。bathroom fall 应由静止站立统辖,走动(pose=1)不该进 fall 判定;lost_track 凭"track 消失"对走动的人报。
- **G. `iot-timeseries-group` 持久化消费者卡死(~05/17 起)** ← 独立审计事故。event_log 停记 EnterRoom/ExitRoom/InBed/LeftBed/activity(只剩 deviceStatus);流健康、roomengine/cardagg 等消费组 lag=0 正常,唯独 timeseries persister lag=full backlog。很可能 05/30 19:31 `owlback.iot Failed with result 'signal'` 崩溃后该组没恢复。**不是** CABB fall 的原因(引擎照常拿事件),但是生产事故。
- **H. gate-list 规则引擎结构性漏风** = 上述全是症状,根本缺隐状态信念估计层,详 [[belief_state_rule_engine_reframe]]。

## TODO(止血→治本)
**止血(小改快、不需统一帧):**
1. 修 `iot-timeseries-group` 卡死 → 恢复 event_log 持久化(查 consumer 没重连/panic 循环;独立生产事故,优先)。
2. **lost_fall fire 点加 Track+layout 几何门**(用户#2,改动最小止血最快):fire 前 `NearestEntryDist(LastX,LastY) <= ExitDistMinCm → 判离场取消`;且把 radar 自己的 **Enter 多边形喂进去**(轻量 point-in-polygon,不是全程 grid 重分类)。积木已有:`grid.NearestEntryDist`/`ExitDistMinCm`/pending 已记的 cellArea,只是没接到 fire 点。
3. lost_track auto_resolve 清除路径:房内新 track / 超时过期 / 延迟 ExitRoom / 门区确认 → 自动 expired。
4. bathroom fall 归 `StandingContinuousMin` 标准:走动(pose=1)不进 fall 判定。

**中期(per-device 自描述,不需统一帧):**
5. ~~roomengine 加载 /128 device layout~~ **【2026-05-31 DONE + 实测验证】** 采用 **room-owns-{grid, mounts[], TrackManager}** 模型(非 per-device 拆 grid):房间是空间整体=一 grid + 多 mount + 一 TM(保 sleepad/radar 融合)。实现:① `RoomConfig.Radar`→ 加 `Radars []RadarMount`+`RadarAddrs []string`;② engine `mounts map[roomID]`→`deviceMounts map[deviceAddr]`(删 last-wins 单 mount = 多雷达 bug 根源)+ 删死读 + `MountForRoom`→`MountForDevice`(feedback.go caller 改 deviceAddr);③ RegisterRoom 逐台 `StampRadar` 进同一 grid;④ 新建 `internal/roomengine/layout_load.go`(`LoadRoomCanvases` 按 /128 查 + `BuildRoomConfigFromCanvases` 按 /88 聚合,registerAllRooms + runDailyLayoutReload 共用免双写);⑤ bootstrap SQL 去掉 broken `rvl.spatial_prefix=r.room_id` JOIN;⑥ handleMessage mount 按 deviceAddr 取;⑦ LayoutHash 纳入 Radars 全集;persist 无需改(grid 仍 per-room,key /88)。**实测验证(2026-05-31 重启后,逐值核实非估计)**:placeholder_count **17→9**(8 房新获 layout 几何);CABB(`fd00:0:3:411:1:200::/88` Bathroom)**enters=1**→ 几何门 NearestEntryDist 不再返 9999,② 几何门激活;D523 房(`fd00:0:3:111:3:100::/88` 双雷达)**enters=3 beds=1** 双 canvas 几何合并;2 条 multi-radar warn:`111:3:100::/88`(radar_count=2)+`411:2:100::/88`(radar_count=3,3 台全 parse 成功);0 parse failure;build+vet 绿,NRestarts=0 active stable。多雷达房 device-local 帧未配准仍 best-effort(=⑧ seam)。**关键 bug(踩 2 次才修对)**:registerAllRooms 原把 LoadRoomCanvases 放在主 rooms 查询 rows cursor **未关时**调第二查询 → database/sql/lib/pq 同连接 open-rows footgun → canvasesByRoom 空 → 全 placeholder(enters=0);**正解 = LoadRoomCanvases 移到主查询之前**(已落)。次踩:SQL 初版 `host(set_masklen(..,88))` 保留全 host bits 不等于 room_id,改 `network(...)` 才匹配。**工具踩坑**:`go build ./...` 只编译不写 `.bin/`(systemd restart 时才 rebuild);本机 bash/Read 间歇 glitch 返回伪造 prose,**曾据此误写"DONE+enters=2/7"实为 enters=0 未加载**——必须 temp file + 单值数字逐项核实,严禁凭工具回显断言完成态([[feedback_no_unverified_claims]])。feedback.go MountForRoom caller 漏改致首次 crash-loop(systemd 留旧 binary 无 outage,已修)。plan 存 `.claude/plans/cozy-giggling-kazoo.md`。
   - **数据实证(owl_v2)**:11 房有 /128 layout 全 masklen=128;多雷达房 2 个 = `fd00:0:3:111:3:100::/88`(D523+9e7 2 台)、`fd00:0:3:411:2:100::/88`(3 台),余单雷达。两台 canvas 帧互不相通(D523 wall 500×500 vs 9e7 360×350,各 device-local)→ 实锤 per-device 自描述非 /88 统一帧。
   - **根因 A 实锤**:[engine_bootstrap.go:253] `JOIN ON rvl.spatial_prefix=r.room_id`(/88=/128) 对所有房 0 命中→layout 全没加载。② 几何门 checkLostFall(track_manager.go:2502, AreaEnter+NearestEntryDist≤30 放行) **已存在**,grid.go:175 `len(Enters)==0→9999` 是门永不触发直接因。**⑤做完②自动止血**。
   - **架构链路**:engine 全 roomID-key(~15 处 `e.rooms/grids/mounts[roomID]`)、`e.mounts[roomID]`=last-wins 单 mount(engine.go:1253)、路由 handleMessage:1658 `deviceAddr→deviceRoom[addr]→roomID→(tm,grid,mount)`。多雷达房第二台用错 mount 转坐标→必错乱(=当前 bug)。**关键降险点**:底层 SuiteCensus 已是 **suiteID-key**(`UpdatePersonFromTrack(suiteID,...)`,engine.go:911),per-device 拆 grid 后 census 跨 device 聚合天然成立(同房共享 suiteID/roomType/residentID)。
   - **改动面**:① bootstrap SQL 改按 /128 device 加载(每 device 一份 canvas+room 属性);② engine rooms/grids/mounts/trackLastSeen 全改 per-deviceAddr key;③ layout_parser.go:70 cfg.Radar last-wins 天然消解(每 canvas 各 1 radar);④ handleMessage/OnDeviceUnfit/publishTrackStatuses 路由+key 按 deviceAddr;⑤ bathroomGates/bathroomFall/bedroomFall 当前 roomID-key,需决定 per-device 还是保 room 级(census 已 suiteID-key 不动)。engine.go 健康 2173 行 build=0(前述"损坏"是工具 glitch 撤回)。
6. 建 per-device 自描述矩阵 device×{hasBed, enters+性质, 邻居 boundary 交集}→ 喂 fall fire 门。
7. 补 `enter_target`(outside/bathroom/inside_enter)authoring + 邻居 boundary 机制(canvas 画邻居 / 房间邻接声明 / 学 track 交接)。

**治本:**
8. /88 room 级统一空间帧(/128 若不可叠合则需 device mount pose 变换)。
9. belief 状态机替代 gate-list:见 [[belief_state_rule_engine_reframe]](S/A/观测模型 + fixture oracle + shadow mode + 确定性角点复现现状)。

关联:[[number_people_zero_exitroom_fallback]] [[cabb_ghost_fall_cases]] [[fall_rules_three_classes]] [[longsurvival_anchor_ghost_gap]] [[room_engine]] [[layout_scope_by_entry_point]]。
