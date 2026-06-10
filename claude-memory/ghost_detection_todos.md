---
name: Ghost detection 待做项
description: L1 镜面对称 ghost 2026-05-19 commit c360bbb 已完工（无 layout 先验，自学习 + 持久化）；L2 track-id 重新关联留 v3
type: project
originSessionId: c27ea410-6e58-493c-bbdc-8e863b87c54e
---

## ✅ L1: 镜面反射 ghost 检测 + 自学习（2026-05-19 commit `c360bbb`）

**与原计划不同**：不依赖 cfg.Interferes 镜面坐标先验，纯几何三不变量识别 mirror pair。

### 算法

- I1: v_t = (B-A) 方向 5 帧恒定（镜面线方向固定）
- I2: 中点 M_t = (A+B)/2 共线且 ⊥ v（落在镜面线上）
- I3: 同步等速 |ΔA| ≈ |ΔB|（独立行人不可能同步）
- Tiebreaker: 距 radar mount 远者 = ghost（反射路径必长于直达）

### 自学习（不写 layout cfg.Interferes）

- 每帧 bounce point M_t = 线段 R'→A_t ∩ midline → grid.MarkMirrorBounce 单 10cm cell +=1（commit `9efde6a` 简化，原 2×2 微块涂抹会破坏"≥3 独立配对"语义）
- ≥3 hits → Cell.Belief[0] = AreaDeny + SourceLearned（SourceHuman cell 永不被覆盖）
- snapshot v8 持久化（roomengine_grid_snapshot.payload）— FE 可读出 learned mirrors 给用户 confirm/手画

### 命中后效

- ghost 端 GhostPenalty += 50（cap 100）→ 喂既有 verdict + fall_verify 管线
- 单 pair 60s cooldown 防刷分
- 最小累计位移 30cm 防静态场景误判

### 文件

`mirror_detect.go` / `cell.go` 加 MirrorBounceCount + LastMirrorMs / `grid.go` MarkMirrorBounce / `persist.go` v7→v8 / `track_manager.go` 段 4d hook + SetRadarMount / `engine.go` RegisterRoom 注入 radar mount

### 后续 FE wire（不在本 PR）

- FE 读 roomengine_grid_snapshot.payload 显示带 AreaDeny+MBC>0 的 cell
- 用户接受/手工画多边形 → 通过 wisefido-data API 写 room_visual_layout.cfg.Interferes
- "回写 layout" 是 FE 接受用户操作后做的，runtime 不直接改 layout 真相

---

## L2: Track-ID 重新关联（v3 backlog，详 [v3_ToDo.md](../../../owl/owlBack/doc/v3_ToDo.md) §L2）

**Why**：engine 直接用 `tm.tracks[radar_track_id]` 100% 信任 firmware ID；firmware 有时 split 同一人为多 ID（D5F7 case A），有时不同人复用同一 ID。

**How**：multi-target tracking 架构改造（Hungarian / NN 匹配 observation→internal track），engine 维护内部 ID 空间，radar ID 只作辅助提示。涉及面大：processFrameAt 段 1 重写 / TrackState 加 RadarTrackID 字段 / outputs map 改 internal id / 大量单测重写。

**用户原话**："track_id 有时可能有错，但相信厂家也尽力处理过，我们处理也未必更佳"—— 倾向相信 firmware ID 直到证据反过来。

**触发时机**：先做完 L1 看 ghost 残余率；ghost 主因若仍是 ID 错乱再启 v3 周期。
