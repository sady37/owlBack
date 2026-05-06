---
name: D5F7 bathroom ghost cases (track-ID split + 6.5min 2-ghost)
description: Radar_D5F7 (E598A2ACD5F7) 浴室两次典型 ghost 场景，未来 ghost 检测算法的回归测试 fixture
type: project
originSessionId: c27ea410-6e58-493c-bbdc-8e863b87c54e
---
设备：E598A2ACD5F7 (Radar_D5F7)，房间 265665a6-1c48-4d55-9ced-665bec73e9a2 (Denver-LakeW-101 bathroom)
房间布局：300×150cm 小浴室，有 bath/curtain/mirror，雷达天花板装在 (0, 240, 270)，FOV 边界 leftH=190 rightH=120 frontV=70 rearV=90，墙壁 h=-120~180 v=-80~70

## Case A: 04-25 17:26:17-17:28:05 — firmware track-ID 切换 ghost

**timeline (本地 Denver / MDT)**：
```
17:26:17-23  track 0 在 Enter (-50, 80) 进房，pose=4，z=99→0
17:26:26-37  track 0 走过房间，(-70,50)→(160,-20), z 0-83 真实路径
17:26:39-41  track 0 到右侧 (170,-30)→(160,0)
17:26:42 ★  track 0 突然冻结在 (190, -50)，z=0，pose=4，**完全不动 90+ 秒**
17:26:42 ★  track 1 同时出现在房间中央 (40, 0)
17:26:42-59  track 1 走出连贯路径 (40,0)→(0,50)→(-50,60)，z 33-88，pose=4→1→3
17:27:08+    track 1 消失（出门）
17:27:08-17:28:05+  仅 track 0 还在 (190, -50) 假冻结
```

**物理矛盾**：1 个真人，被 firmware split 成 2 个 track。track 0 卡在 (190, -50)（**右墙 h=180 之外**，物理不可能）。

**当前 engine 检测能力**：未实测。需要 playback API 跑一遍看 verdict。

## Case B: 04-26 12:41:17-12:47:51 — 6.5min 持续 2-ghost

**timeline**：
```
12:37:05  EnterRoom，np=1
12:37:12-41:10  activity tc=1，walk_distance/duration 有真实数字
12:41:17 ★  np: 1→2，第二个 track 出现
12:42-47   连续 6 分钟 tc=2，**所有 walk_duration/stand_duration=0**
12:47:51   ExitRoom + np: 2→1
12:48:07   activity tc=1, stand=17（人还在）
12:48:27   np: 1→0
```

**物理矛盾**：6 分钟"两个人"零位移 + 零站立 + 零行走时间累计。真双人不可能 6 分钟两人都不动；ghost 才会这样。

**Why（这两个 case 重要）**：
- 都是浴室小空间多径反射 + firmware track 管理算法弱点
- A 是"track 切换 ghost"——同一帧两个 track ID 共存
- B 是"长时间共存 ghost"——firmware 无法淘汰错误 track

**How to apply**：
- 未来 ghost 检测算法以这两个 case 为回归测试 fixture
- A 的关键信号：track 0 冻结同时 track 1 出生（时间反相关）+ 房间几何越界
- B 的关键信号：activity 事件 walk_duration=0 持续 N 分钟 + 多 track 共存
- 测试：拉这两段 monitor 数据，喂 TrackManager.ProcessFrame，看 verdict 是否 Ghost
