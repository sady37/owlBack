# Engine 订阅 iot:event/alarm:stream + 床边跌倒（R4）设计

记录日期：2026-04-26
触发讨论：BM87224700978 在 04-26 01:58 短促 InBed/LeftBed (12s)，
radar 同时也给过 InBed/LeftBed (01:37:32-01:39:14, 102s)，
两者物理一致，但**无 Fall 报警** —— 揭示当前 engine 的两个 gap。

---

## Gap 1：未订阅 iot:alarm:stream + radar 事件未利用

### 现状
Engine 当前消费：
- `iot:monitor:stream`：radar + sleepad track 帧（pose / position / bed_status）
- `iot:event:stream`：**仅 sleepad InBed/LeftBed**

完全忽略：
- `iot:alarm:stream`：**radar Fall 报警**（7天 46 个！firmware 多帧聚合，可信度高）
- `iot:event:stream` 中 radar 部分：EnterRoom (1751/7d) / ExitRoom (1679/7d) / InBed (90/7d) / LeftBed (85/7d)

### 应做
1. **alarm:stream 订阅**：radar Fall 直接转 `AnomalyFall + Source=radar_direct`，关联当前 track；
   并关联到 silent fall pending（如同 track 后续真消失，升级 critical）
2. **event:stream 加 radar 路径**：
   - EnterRoom/ExitRoom：硬证据"人数变化"，校准 segment 1 的 realCount 推断
   - InBed/LeftBed：与 sleepad 同类事件 cross-check，不一致即 sleepad/radar 误触发降权
3. **回看前 30s 轨迹**：alarm/event 到达时，从 history buffer 回看消失/触发前 30s 的 track 状态，做事件解释（如 Fall 报警 + 前 30s 在 AreaBed → bed-fall；前 30s 在 walkway → 普通跌倒）

---

## Gap 2：床边短停跌倒（R4）—— 新规则

### 场景
夜起老人下床去卫生间途中**在床边滑倒/晕倒**，常见模式：
1. LeftBed 事件触发（radar 或 sleepad 任一）
2. 应预期：人走开 → radar track 移到走廊/卫生间
3. 实际：人在床边 cell 1m 范围内**持续静止 > 15 min**，pose 可能仍是 Stand（雷达失锁前最后值）

### 触发条件（R4: Bedside Long-Still After LeftBed）
1. 风险时段（IsNightTime，23:30-07:30）
2. **过去 N 分钟内**有 LeftBed 事件（来源任一：radar event / sleepad event / sleepad bed_status 转换）
3. radar track 距离最近 AreaBed cell ≤ 100cm（"床边"判定）
4. 持续静止 > 15 min（stillSec 阈值，或 track 消失后 60s pending 直接报）
5. → `AnomalyBedsideFall`，Risk=100

### 与现有规则关系
- 不重复 R2 (bed-fall)：R2 是"radar 仍认为人在床 + sleepad 离床" 的物理矛盾；R4 是"人离床后到不了远方"
- 不重复 LongStill (toilet/shower)：R4 触发**仅在 LeftBed 后 N 分钟窗口内**（1-3 min？需调），LongStill 是任意时刻

### 实现要点
- TrackManager 加 `lastLeftBedAt int64`：任何来源的 LeftBed 都更新（radar event / sleepad event / sleepad monitor 状态转换）
- scoreMovement 静止超时检查里加新分支：
  ```go
  isBedside := nearestAreaBedDistCm(x, y) <= 100
  inLeftBedWindow := nowMs - tm.lastLeftBedAt < bedsideWindowMs  // 例 3 min
  if isBedside && inLeftBedWindow && stillSec > 15*60 && IsNightTime(nowMs) {
      anomaly = AnomalyBedsideFall
  }
  ```
- 双源时间容差：radar/sleepad LeftBed 差 8-15s 是正常的（不同传感器响应延迟），任一触发即开窗

### 取舍
- **15 min 静止阈值**：跟 LongStill toilet/shower 一致；若太长可能错过急救窗口，但 5-10 min 老人正常床边小坐也常见
- **100 cm 床边判定**：层窗、扶手桌可能在范围内；如果误报多调到 80cm

---

## 待办（按时间）

1. [ ] 加 iot:alarm:stream 订阅（处理 radar Fall）
2. [ ] iot:event:stream 加 radar 路径（EnterRoom/ExitRoom/InBed/LeftBed）
3. [ ] 实现 lastLeftBedAt 多源更新
4. [ ] 实现 R4 (AnomalyBedsideFall)
5. [ ] alarm/event 触发时回看前 30s track history（需要 ts.History 已有 30 帧窗口，但判定逻辑要写）
6. [ ] 单元测试覆盖

不立即实现，等 R1/R2/R3 在 prod 跑出真实数据再决定。
