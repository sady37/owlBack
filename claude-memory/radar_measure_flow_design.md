---
name: radar-measure-flow-design
description: 2026-05-25 拍板 radar measure 流程 — 4 intent 平等 / cell 着色 + 人画 / algo 是 best-effort 不背锅 / 零新表
metadata: 
  node_type: memory
  type: project
  originSessionId: c45f5823-b5a6-44ea-be33-528d9dab537c
---

2026-05-25 与用户协作收敛的 radar measure 边界/家具测量流程设计。

## 4 个 intent 平等独立 — 启动时锁定

| intent | color | 走法 | alarm 路径 |
|---|---|---|---|
| `wall`      | 灰 | 沿房间外圈走一圈（不强求闭合） | InRoom/InFOV 边界 |
| `bed`       | 蓝 | 绕床走（贴墙缺边 OK） | InBed/LeftBed/NightAbsence |
| `long_sofa` | 绿 | 绕长沙发走（lying area） | 仅 area_deny（lying ≠ fall） |
| `enter`     | 红 | 门口反复来回穿越 5-10 次 | ExitRoom 出入口 |

intent 是 alarm 路径分流的根，启动时锁定不允许事后改。`display_name` 让用户起 "主床"/"客厅西墙"，underlying intent 不变。

## 核心哲学："算法只忠实记录，人画最后一公里"

铁律：**算法不背锅，人画自家房间永远比算法准。**

- ❌ 不做矩形拟合自动落 layout（algo 给提议但人决定）
- ❌ 不做缺边补全自动闭合多边形
- ❌ 不做中断点派生 enter（enter 是单独 intent，门口反复走标红 cell）
- ✅ 只做：xy→cell hit 累积，按 intent 上色

精度匹配雷达 ±10-20cm 误差：cell=10cm / 饱和 5-10 hits / 不强制几何校验 / 吸附 10cm 步长。

## 关键洞察：detection rectangle = polygon ∩ FOV 最大内接矩形

firmware 只接受 axis-aligned 矩形作 detection boundary。完整逻辑链：

```
雷达 FOV (硬件能力,由 install_height/style 决定,固定)
        ∩
人画的 room 多边形 (基于 colored cell)
        ↓
最大内接矩形 (server 算)
        ↓
作为 rectangle 下发 firmware (替换 firmware 自测的 {-31,0;...})
```

firmware 自测的 `{-31,0;0,0;-31,49;0,49}` 只是临时凑数 — server 应基于人画 polygon 算出真正合适的 detection rect 下发。

## firmware measure 协议（实测）

- 下发 `rectangle={-127,-127;...}` → 启动 measure，60s 倒计时
- track 第 12 字节 `remaining_time` 60→0 递减（measure 模式天然标签）
- 下发 `rectangle={127,127;...}` 或 60s 超时 → 结束 measure
- firmware 主动回灌所有 prop（含自测 rectangle）

## 数据归宿（零新表）

| 数据 | 归宿 |
|---|---|
| measure 期间 raw track | `monitor_stream` 表（90d 保留，已有） |
| session 状态 (intent, t0) | FE state + localStorage（用户会话内） |
| cell 着色 | FE 实时算（不持久化） |
| 最终 layout 几何 | `room_visual_layout.canvas.objects[]`（已有） |

**Why**：单次 measure 是临时态，最终人画的 layout 才入库。曾起草 `radar_measure_session` 表，被用户驳回（over-engineering）—— 已删除。

## owlback 落地的 endpoint（wisefido-data, 2026-05-25 wire）

路径前缀 `/radar-device/api/v1/radar-device/device/{id}/measure/`：

- `GET  /tracks?from=&to=` — monitor_stream 反查（thin wrapper over `MonitorPlaybackRepository.GetMonitorRowsByAddr`）
- `POST /fit-polygon` — 凸包提议（Andrew's monotone chain），L 形房间需人工调整
- `POST /fit-rectangle` — AABB（bed/long_sofa）
- `POST /fit-entry-lines` — 8-邻接连通簇 → 线段集合（enter）
- `POST /detection-rectangle` — `MaxInscribedRect(polygon ∩ FOV)`；`apply=true` 时调 `UpdateConfig{rectangle}` 下发

算法包 = `wisefido-data/internal/measure/`（4 个无状态函数 + smoke test 用真实 20 点 track 数据）。复用现有 `PUT /config` rectangle endpoint 作 measure start/stop（不开新 endpoint）。

## FE 待做（按 [[feedback_producer_first]]）

- measure mode：选 intent → start → 实时画 colored cell（轮询 `GET /tracks`）→ finish
- canvas mode：colored cell 叠加 → 人工绘制工具（wall line / rectangle / entry line）→ 保存 canvas
- 形状编辑：select + handle + 旋转 + 平移 + cell 吸附 + 穿墙警示
- wall polygon 改时联动调 `/detection-rectangle apply=true`

## 关联

- 走法依赖 [[radar_layout_device_invariant]]：InstallMod/Height 必须先设
- 落地于 [[track_fusion_and_gate_cardid]] 之外的独立 layout 编辑流程
- cell 数据结构与 [[room_engine]] roomengine_grid 10cm cell 一致（未来可作 ground truth）
