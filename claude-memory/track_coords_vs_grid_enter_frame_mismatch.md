---
name: track_coords_vs_grid_enter_frame_mismatch
description: "2026-06-02 【已更正:生产无bug,是replay harness漏转坐标;production track_parse.go:65已RadarToCanvas】真发现=CABB-FP与D5F7真跌倒雷达上几乎不可分(np=0相同/73vs100cm)→单雷达P2无法分离门口走出vs门口跌倒"
metadata: 
  node_type: memory
  type: project
  originSessionId: 21aa5953-dac0-4bc6-a011-15ffda8936ad
---

> **【2026-06-02 重大更正——下方"生产门区闸失效/根因"结论是错的】** 实查 `track_parse.go:65` 生产**已**做 `radarutils.RadarToCanvas(RadarPoint{H,V,Z},mount)` → `TrackFrame.X/Y`=画布坐标(留 RawH/RawV);`TrackStatusBase.X/Y`(track_status.go:108)=canvas;belief_shadow 用 b.X/b.Y=canvas。**生产坐标正确、门区闸正常**。跨帧 2-3m 只是 **belief_replay_test.go harness 自己漏调 RadarToCanvas**(读 monitor_stream 原始雷达坐标直接喂 canvas-grid)。**已修 harness**(buildGridFromLayout 返回 mount + replay ingest 处转画布)。修后真实画布距离:CABB 末位 **73cm**(最初 73cm 其实对)、D5F7-1031 真跌倒末位 100cm、MoM 出门 20cm。
>
> **真正的硬发现(更有价值)**:CABB(hunzi 浴室 FP) vs D5F7-1031(101 实跌倒>15min)在雷达上**几乎不可分**:① **np=0 完全相同**(都末帧后+0.9s → 实证铁律:倒地者也读 np=0,不可作判别);② 距离仅差 27cm(73 vs 100);③ 都"移动中消失";④ 观测走速也浑(CABB 末分钟 walk 2m/16s+stand 40s=多在站)。**单雷达 P2(距离+reachability+np=0)无法分离门口走出 vs 门口跌倒**;判别只能靠 (a) 跨房重现(中耦合/suite,超 P2)或 (b) 接受不确定(模糊门口消失→不当 live fire→升级/落档,=belief §8 不创造可观测性)。oracle:CABB confirm=false 但 maxP=0.959(脆,未 assert);D5F7-1031 仍 confirm=true;**Hunzi-0530 因 harness 修复从误确认→不报(白赚)**;np=0 弱化(#14)经此**证明对**(真跌倒 np=0 相同,强信会漏报)。

---
**【以下为更正前的误判记录,结论已被上方推翻——保留作教训:探针用错坐标帧会得出"生产坏了"的假根因】**

2026-06-02 实现 DBN P2 时撞出的**重大根因**（confirm 了 [[fall_fp_roots_and_todo]] 的"C 无统一 room 空间帧"假设，并钉到精确 locus）。

**铁证**（belief_replay_test 探针，doc/cases 真机帧）：人**确证走出门**（MoM 有 ExitRoom / CABB np=0）到 Enter 的最近距离——
- **原始雷达局部帧**（track 坐标 = 布局顶点 h/v 帧）：MoM **10cm** / CABB **0cm**（人就在门里）。
- **grid 变换后帧**（`g.Enters` 存的）：MoM **199cm** / CABB **280–391cm**。
跨帧差 2–3 米。**4 个 fixture（含 angle-0 的 MoM 和 angle-90 的 CABB）全部 InEnter_hits=0、minDist 190–280cm**——angle 无关，系统性。

**机制**：`layout_parser.go:214 parseRadarMount`（rotation+mount position）把 Enter/areas **变换到 room/canvas 帧**存进 `grid.Enters`（CABB raw `[70,100]×[0,80]`→存 `[-50,30]×[320,350]`，90°旋转+平移）；**但 track 坐标全程是雷达局部帧，没人把它一起变换**。`grid.NearestEntryDist(x,y)`(grid.go:174,世界坐标 DistTo) 和 `geomFromGrid`(belief_adapter.go:47) 拿**原始 track** 比**变换后 Enter** = 跨帧。

**影响面（双杀）**：
1. **生产 gate-list 门区闸**：`track_manager.go:2548 NearestEntryDist≤ExitDistMinCm(30)` **永远不触发**（track 离 grid-Enter 恒 ≥190cm）→ 门区正常退场无法被几何抑制 → **这是 CABB/09E7/MoM/D5F7 门区 lost_track FP 的真根因**（不是我之前说的"73cm>30 阈值太紧"——73cm 是原始帧、grid 算成 391cm）。
2. **belief geomFromGrid**：所有 angled/offset mount 的 geom 分类错（门口的人判 Unknown）；P2 `reachableExit` 的 `f_dist` 喂到错距离 → CABB 本应 e≈1 强抑制，实际 e≈0 无效。

**修复方向（架构级，待用户拍——动共享 layout/grid + 影响生产 gate-list）**：track 坐标与 Enter 必须同帧。① **areas 不变换(全雷达局部)**——单雷达房(绝大多数)天然正确,多雷达房各 device 用自己局部帧对自己 track(per-device 自描述,对齐 [[belief_state_rule_engine_reframe]] 破局点);或 ② **track 坐标也做 mount 变换到 room 帧**(多雷达统一帧需要,但要给 track 补 mount.position+angle 变换)。当前=变换了 area 没变换 track=两头错。room-frame 变换很可能是 layout-load(room-owns-mounts) 时为多雷达加的,但 track 没跟上。

**P2 现状**：代码已实现(observation.go ObsReachableExit + likelihood.go 软门 + np=0 弱化#14 + belief_adapter.reachableExitObs + shadow/replay wire),build/vet/test 全绿,**oracle 零回归**(真跌倒仍 fire/FP 仍不报),但 **reachableExit 因喂错距离对 oracle 无可见效果**,CABB **未** promote 到 assert。**shadow-only 未部署**——治本要先修坐标帧。修帧后 CABB d≈0→e≈1→强抑制→可 assert,且生产门区闸同时复活。

关联 [[fall_fp_roots_and_todo]](C/根因A)、[[belief_state_rule_engine_reframe]](空间关联前置/step-0 一房一雷达)、[[radar_geometry_point_invariant]]、[[radar_layout_device_invariant]]、[[belief_p2_absence_emission 文档]]。
