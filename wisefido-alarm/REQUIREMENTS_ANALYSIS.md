# wisefido-alarm 需求分析

## 📋 检查结果

### 1. 现有代码状态 ✅

**项目结构**：
- ✅ 项目目录已创建
- ✅ `internal/config/config.go` - 配置已定义
- ✅ `internal/models/` - 模型已定义（realtime_data, alarm_event, alarm_config）

**待实现**：
- ❌ Repository 层（数据库操作）
- ❌ Consumer 层（Redis 缓存读取）
- ❌ Evaluator 层（报警规则评估）
- ❌ Service 层（主逻辑）
- ❌ Main 入口

### 2. alarm_rule.md 需求分析

根据 `owlBack/docs/alarm_rule.md`，需要实现 **4 个复杂事件**：

#### 事件1：防止雷达漏报 - 床上跌落检测 ⭐ **最复杂**

**触发条件**：
- T0：sleepad检测到离床事件

**阶段1：建立lying基线（准备阶段）**
- 检测到**睡眠状态**时，记录lying值
- 记录内容：`lying_height`, `lying_position`, `lying_time`
- 存储：Redis `alarm:state:{card_id}:track_{track_id}:lying_height`

**阶段2：跌落检测（T0触发）**
- **退出条件**（持续检查）：
  - sleepad有HR/RR → 退出
  - sleepad有上床事件 → 退出
  - radar检测到在移动 → 退出

- **T0+5秒**：危险情况检测（提前触发）
  - track_id突然消失
  - 30秒后未出现新的可移动track
  - → 即刻报跌倒（AI-fall，ALERT级别）

- **T0+60秒**：可疑跌倒
  - 睡眠板没有呼吸心率
  - 雷达仍显示相同track在床的区域
  - → 报可疑跌倒（SuspectedFall，WARNING级别）

- **T0+120秒**：确认跌倒（AI-fall）
  - 身高值较床或之前的lying高度降低
  - 人未移动（位置未变化）
  - track_id仍存在
  - → 立即报：AI-fall（Fall，ALERT级别）

**关键点**：
- 需要维护状态（T0时间、lying基线、track_id状态）
- 需要定时器（T0+5秒、T0+60秒、T0+120秒）
- 需要持续检查退出条件

#### 事件2：Sleepad可靠性判断 - 避免电磁或振动干扰

**触发条件**：
- 当sleepad检测到呼吸心率时，进行事件核查

**执行流程**：
1. **核查1**（前置条件）：
   - sleepad检测到呼吸心率
   - 但sleepad未检测到上床事件
   - → 进入分支判断

2. **分支判断**：
   - **分支A**：床上绑radar
     - 绑到床上的雷达，未检测到区域ID-床有人存在
     - 且睡眠板也未检测到有上床事件
     - → error: `Environmental interference`
   
   - **分支B**：床上未绑但房间内有Radar
     - radar与sleepad均绑在同一room_ID
     - 且房间内仅有一床
     - 床在雷达的检测区域
     - 但雷达检测区域内无人
     - 且sleepad有呼吸心率
     - → error: `AI-Environmental interference`

3. **如果分支A和分支B都不满足**：
   - → warning: `AI-Environmental interference`

**关键点**：
- 需要查询设备绑定关系（床、房间）
- 需要查询雷达检测区域数据

#### 事件3：Bathroom可疑跌倒检测

**触发条件**：
- 在bathroom（卫生间）房间内

**房间识别**：
- 通过 `room_name` 或 `unit_name` 中是否包含：bathroom, restroom, toilet（不区分大小写）

**条件检查**：
1. 在bathroom房间内
2. 雷达检测范围内仅1人
3. 1个人处于**站立状态**（不是坐着）
4. 位置未有变化（位置变化小于10cm，超过10分钟）
5. 房间内仅有1个track_id

**如果满足**：
- → 在当前位置报：AI-suspected fall（SuspectedFall，WARNING级别）

**关键点**：
- 需要识别bathroom房间
- 需要跟踪位置变化（10cm阈值）
- 需要时间阈值（10分钟）

#### 事件4：雷达检测到人突然消失

**触发条件**：
- 雷达检测到质心降低（高度降低）
- 第2秒人消失（track_id在2秒内消失）

**条件检查**：
1. 质心降低（高度降低超过60cm，可配置）
2. 第2秒人消失（track_id在2秒内突然消失）
3. 5分钟内无人员活动（未检测到任何人类活动）

**如果满足**：
- → 在消失位置报：AI-suspected fall（SuspectedFall，WARNING级别）

**关键点**：
- 需要跟踪高度变化
- 需要检测track_id消失
- 需要检测5分钟内无活动

## 🎯 实现挑战

### 1. 状态管理
- **事件1**需要维护复杂的状态（T0时间、lying基线、track_id状态、定时器）
- 需要使用 Redis 存储状态：`alarm:state:{card_id}:{event_type}:{track_id}`

### 2. 定时器/延迟任务
- **事件1**需要多个定时检查点（T0+5秒、T0+60秒、T0+120秒）
- 可以使用 Redis 的延迟队列或 Go 的定时器

### 3. 数据查询
- 需要查询设备绑定关系（床、房间、unit）
- 需要查询雷达检测区域数据
- 需要查询房间信息（识别bathroom）

### 4. 位置/高度跟踪
- **事件1**：跟踪lying高度、位置变化
- **事件3**：跟踪位置变化（10cm阈值）
- **事件4**：跟踪高度变化（60cm阈值）

### 5. 多传感器数据融合
- 需要同时使用 Radar 和 Sleepace 数据
- 需要判断数据来源（Radar vs Sleepace）

## 📊 数据流设计

```
Redis (vital-focus:card:{card_id}:realtime)
    ↓ 读取融合后的实时数据
wisefido-alarm 服务
    ├─ 读取报警规则（alarm_cloud, alarm_device）
    ├─ 读取卡片信息（cards, devices, rooms）
    ├─ 读取状态缓存（alarm:state:*）
    ├─ 评估4个事件：
    │   ├─ 事件1：床上跌落检测（复杂状态机）
    │   ├─ 事件2：Sleepad可靠性判断
    │   ├─ 事件3：Bathroom可疑跌倒检测
    │   └─ 事件4：雷达检测到人突然消失
    ├─ 生成报警事件
    ├─→ PostgreSQL (alarm_events)
    ├─→ Redis (vital-focus:card:{card_id}:alarms)
    └─→ Redis (alarm:state:* - 更新状态)
```

## 🏗️ 架构设计建议

### 1. Repository 层
- `alarm_cloud.go` - 读取租户级别报警策略
- `alarm_device.go` - 读取设备级别报警配置
- `alarm_events.go` - 写入报警事件
- `card.go` - 读取卡片信息（复用 wisefido-card-aggregator 的逻辑）
- `device.go` - 读取设备绑定关系
- `room.go` - 读取房间信息

### 2. Consumer 层
- `cache_consumer.go` - 轮询 Redis 缓存（`vital-focus:card:{card_id}:realtime`）
- `cache_manager.go` - 更新报警缓存（`vital-focus:card:{card_id}:alarms`）
- `state_manager.go` - 管理报警状态（`alarm:state:*`）

### 3. Evaluator 层
- `event1_bed_fall.go` - 事件1：床上跌落检测（最复杂）
- `event2_sleepad_reliability.go` - 事件2：Sleepad可靠性判断
- `event3_bathroom_fall.go` - 事件3：Bathroom可疑跌倒检测
- `event4_sudden_disappear.go` - 事件4：雷达检测到人突然消失
- `common.go` - 公共评估逻辑（位置跟踪、高度跟踪等）

### 4. Service 层
- `alarm.go` - 报警服务主逻辑，整合各层
- `scheduler.go` - 定时任务管理（事件1的定时检查）

### 5. Main 入口
- `main.go` - 服务启动入口

## 📝 实现顺序建议

1. **Repository 层**（数据访问基础）
2. **Consumer 层**（数据读取）
3. **Evaluator 层 - 事件2/3/4**（相对简单的事件）
4. **Evaluator 层 - 事件1**（最复杂的事件）
5. **Service 层**（整合）
6. **Main 入口**（启动服务）

## 🔗 相关文档

- `owlBack/docs/alarm_rule.md` - 报警规则详细说明
- `owlBack/docs/13_Alarm_Fusion_Implementation.md` - 详细设计文档
- `owlBack/docs/system_architecture_complete.md` - 系统架构文档
- `owlBack/wisefido-alarm/IMPLEMENTATION_PLAN.md` - 实现计划

