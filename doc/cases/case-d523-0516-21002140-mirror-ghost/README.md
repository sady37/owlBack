# d523-mirror-ghost-0526

Mirror ghost fixture from measure-mode testing. **P0 持续 38+ 分钟没断**，跨多个 real-person 出现/消失，明显是 mirror reflection 锁死。

## Scene

- **device_uid**: `E598A2ACD523`
- **device_addr**: `fd00:0:3:111:3:100:a2ac:d523/128`
- **time window**: 2026-05-26 21:00:00 → 21:40:00 PDT (40 min)
- **scenario**: user 走 room perimeter 做 Wall 测量；radar 出 3 个 concurrent tracks，其中 P0 是 phantom never dies。

## Track 精确时间表（ms epoch UTC, 字符串 PDT）

| track_id | first_ts (PDT) | last_ts (PDT) | first_ms | last_ms | samples | duration (s) | label |
|---|---|---|---|---|---|---|---|
| **P0** | 21:00:00.515 | 21:38:27.126 | 1779854400515 | 1779856707126 | 1938 | 2306.6 | **ghost** (mirror) |
| **P1** | 21:02:45.216 | 21:38:27.126 | 1779854565216 | 1779856707126 |  438 | 2141.9 | **flicker** (gaps 见下) |
| **P2** | 21:30:28.493 | 21:38:27.126 | 1779856228493 | 1779856707126 |  484 |  478.6 | **real** (current scene) |
| 88 | — | — | — | — | 22 | — | "no person" sentinel |

## P1 间断分析（不是 stable real person，疑似次级 ghost / transient false detection）

P1 有 4 个 >30s 的 gap：
```
21:10:10   gap=430s 前
21:22:09   gap=668s 前 (~11 min)
21:30:17   gap=120s 前
21:38:19   gap=467s 前 (~7.8 min)
```

P1 持续在线时段碎片化，跟稳定的 P0 (持续 38min 不断) 和 P2 (持续 8 min 不断) 都不同。可能是另一个反射方向的 intermittent ghost。

## 用户标注（canvas 直观观察）

- **P0 = ghost** (mirror reflection)
- **P2 = real** (current physical person)
- **P1 = 待分析**（你只看到 2 track 时，P1 应该正处于 gap 期）

## Ghost 特征信号（P0 vs P2 对比）

| 维度 | P0 (ghost) | P2 (real) |
|---|---|---|
| 持续时长 | **2306s ≈ 38 min** | 478s ≈ 8 min |
| 样本数 | 1938 | 484 |
| x 范围 | [-630, 0] | [-430, -190] |
| y 范围 | [-40, 610] | [70, 500] |
| track_confidence | 80 | 80 |

关键鉴别点：
1. **长寿命**：ghost 几乎不消失（这里 38min），real person 不会无停顿存在那么久（要么动到 FOV 外、要么 pose 切到 Lying）
2. **空间范围更广**：ghost x 跨 630cm，real 跨 240cm — 反射路径几何会拉远
3. **confidence 无用**：两者都 80
4. **跨 real 持续**：P0 同时跟 P1（早段）和 P2（后段）共存，real 不会"霸占"track ID 这么久

## 文件

- `monitor_stream_21-00_to_21-40_PDT.json` — 2873 rows (`radar.track` 2835 + `radar.heart` 38)
- `sensor.log` — wisefido-sensor zap-json 行，时间窗 04:01-04:36 UTC (= 21:01-21:36 PDT)，20 行 (sensor 服务该设备 chatter 不多，留作回头排查)
- 每行 schema:
  ```
  { ts (ISO UTC), device_addr, device_type, stream_type, payload, severity, tags }
  ```
- `radar.track` payload 是 target 数组：
  ```
  [{ track_id, pose, position_x, position_y, position_z, area_id, track_count, remaining_time, track_confidence }, ...]
  ```

## 用途

- **ghost 检测算法回归**：feed 全 sequence，期望 P0 被 flagged ghost
- **mirror-reflection 模型 ground truth**：P0 trajectory 完整记录 wall-reflected path
- **P1 二次分析**：判断 P1 是 ghost / real / 别的（gap pattern 启发新规则）
- **正负样本对**：P0(ghost) vs P2(real) 同 device 同时间，最佳对比
