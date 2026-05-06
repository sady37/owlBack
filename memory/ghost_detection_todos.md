---
name: Ghost detection 待做项（mirror + track-id 重新关联）
description: 用户决定先做 birth-position + speed 过滤，以下两条留待未来 PR
type: project
originSessionId: c27ea410-6e58-493c-bbdc-8e863b87c54e
---
## TODO 1: 镜面反射 ghost 检测

**Why**：浴室 / 玄关常见镜子，雷达光程多径——人在镜前 D，雷达同时报告"另一个人"在镜后 D 处。
当前 engine 仅在镜子格子本身（`AreaDeny`）birthScore -30，**不检测几何对称的镜像 ghost**。

**How to apply**：
- 对每对活 track A, B，求 B 关于镜面轴的镜像 B'
- 若 dist(A, B') < 50cm → B 是 A 的镜像 → 杀 B 保 A
- 镜面定义：`cfg.Interferes` rect 的长边 + height 信息
- D5F7 浴室是典型测试场

**优先级**：低（不阻塞主路径），先记下来。等 birth-filter 上线后跑 case B 看是否 mirror 是主因。

## TODO 2: Track-ID 重新关联（数据关联架构改造）

**Why**：当前 engine 直接用 `tm.tracks[radar_track_id]` 做 key，100% 信任雷达 firmware 的 ID 分配。
firmware 经常错——同一人 split 成多 ID（D5F7 case A），不同人复用同一 ID。

**How to apply（标准 multi-target tracking）**：
- Engine 维护内部 track ID 空间，与 radar ID 解耦
- 每帧 Hungarian / NN 匹配：observation → existing internal track（cost = Kalman 残差 + pose 一致性 + 时间）
- radar_track_id 只作辅助提示，不是身份
- 涉及：processFrameAt 段 1 重写 / TrackState 加 RadarTrackID 字段 / outputs map 改 internal id / 大量单测重写

**优先级**：中（架构性改造，但影响面大）。先靠"出生 + 速度过滤"挡住大部分 ghost，等数据证明 ID 错乱仍是大量误报源再做。

**用户原话**：「track_id 有时可能有错，但相信厂家也尽力处理过，我们处理也未必更佳」——倾向相信 firmware ID 直到证据反过来。

## 当前优先做

**Birth-position + 速度过滤（elder care 1.0-1.5 m/s 阈值）+ EnterRoom 配对**：
- dEntry > 150cm 直接 Ghost
- dEntry ≤ 150cm 但近 2s 内无 EnterRoom/enter2out 事件 → Ghost
- dEntry ≤ 50cm + EnterRoom 配对 → Real
- 单元测试：D5F7 case A track 1（dEntry=100cm 无 Enter pair）应该被判 Ghost
