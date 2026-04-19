# Bug 记录

## BUG-002: Redis stream 残留旧 consumer group 导致 cardagg 事件重复消费

**发现日期**: 2026-04-19  
**严重程度**: P2  
**状态**: 已修复  

### 现象
每条 InBed/LeftBed 事件被 cardagg 消费两次。

### 根因
`iot:event:stream` 等 6 个 stream 上同时存在两个 consumer group：
- `$cons-cardagg`（当前代码使用）
- `cardagg-group`（旧代码残留）

两个 group 各消费一次 → cardagg 收到重复消息。

### 修复
`XGROUP DESTROY` 删除所有 stream 上的 `cardagg-group`。

---

## BUG-001: BedState.StartTime 被 monitor 数据覆盖，导致 LeftBed pending 5分钟判断失效

**发现日期**: 2026-04-19  
**严重程度**: P1  
**状态**: 修复中  

### 现象

LeftBed 离床报警配置了 `duration_sec=2700`（45分钟 pending），但 pending 从未被加入。
日志显示 `inBedMs=2812`（2.8秒），实际在床超过 6 分钟。

### 根因

`BedState.StartTime` 设计为"当前 BedStatus 开始时间"，InBed 时正确设置为上床时刻。
但 `DeriveBedStateFromRealtime`（monitor 实时数据推导）和 `PublishBedStateSleepStage` 多处创建新的 `BedState` 时用 `StartTime: now`，
覆盖了原始上床时间。每次 monitor 数据到达都重置 StartTime → LeftBed 检查时 `inBedMs` 只有几秒 → `pass5min=false` → 不进 pending。

### 受影响的代码位置

`wisefido-cardagg/internal/service/state_service.go`:

| 行号 | 函数 | 说明 | 是否需修 |
|------|------|------|----------|
| 724 | `EnsureCardStatePrepared` | 首次初始化，默认 `StartTime: now` | ✅ 初始化后被 729 行覆盖，但 curr 为 nil 时走默认值，OK |
| 736 | `EnsureCardStatePrepared` | `StartTime == 0` 时 fallback 到 now | ⚠️ OK（仅首次） |
| 763 | `PublishBedStateSleepStage` | 新建 BedState 时 `StartTime: now` | ❌ 应继承 `prev.StartTime` |
| 850 | `DeriveBedStateFromRealtime` | vital derive 恢复在床：`StartTime: now` | ❌ 应继承 `curr.BedState.StartTime`，仅状态从离床→在床时用 now |
| 916 | `DeriveBedStateFromRealtime` | 无事件在床推导：`StartTime: now` | ❌ 同上 |
| 950 | `DeriveBedStateFromRealtime` | 无事件离床推导：`StartTime: now` | ❌ 同上 |
| 986 | `DeriveBedStateFromRealtime` | Sleepad 在床推导：`StartTime: now` | ❌ 同上 |
| 1020 | `DeriveBedStateFromRealtime` | Sleepad 离床推导：`StartTime: now` | ❌ 同上 |

### 修复原则

`StartTime` 仅在 BedStatus **发生变化**时更新为当前时间，状态不变时**继承前值** `curr.BedState.StartTime`。

### 验证方式

1. 上床 6 分钟 → 下床 → 日志 `LeftBed.pending.check` 应显示 `inBedMs ≈ 360000, pass5min=true`
2. 确认 `pending.added` 日志出现
3. 8 分钟内上床 → 确认 `pending.removed` 日志出现
