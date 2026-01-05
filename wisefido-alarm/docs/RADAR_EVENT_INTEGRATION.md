# 雷达事件集成设计

## 📋 雷达设备事件类型

根据用户提供的文档，雷达设备支持以下事件类型：

### 1. 进出事件（type=1）

**MQTT 主题**：`/prefix productId/UID/post`

**消息格式**：
```json
{
  "cmd": "event",
  "type": 1,
  "data": [{
    "track-id": 0,        // 人员轨迹 ID
    "event": 1,           // 1-进入房间, 2-离开房间
                          // 3-进入区域, 4-离开区域
                          // 5-进入监护模式, 6-退出监护模式
    "area_type": 2        // 2-普通床, 5-监护床, 6-感应区
  }]
}
```

**事件类型**：
- `event=1`：进入房间
- `event=2`：离开房间
- `event=3`：进入区域
- `event=4`：离开区域
- `event=5`：进入监护模式
- `event=6`：退出监护模式

### 2. 姿态变化事件（type=2）

**消息格式**：
```json
{
  "cmd": "event",
  "type": 2,
  "data": [{
    "track-id": 0,        // 人员轨迹 ID
    "pose": 1             // 2-疑似跌倒, 5-确认跌倒
                          // 7-疑似坐地, 8-确认坐地
                          // 10-疑似床上坐起, 11-确认床上坐起
  }]
}
```

**姿态类型**：
- `pose=2`：疑似跌倒
- `pose=5`：确认跌倒
- `pose=7`：疑似坐地
- `pose=8`：确认坐地
- `pose=10`：疑似床上坐起
- `pose=11`：确认床上坐起

### 3. 人数变化事件（type=3）

**消息格式**：
```json
{
  "cmd": "event",
  "type": 3,
  "data": {
    // 人数变化信息
  }
}
```

---

## 🎯 Event 3 (Bathroom 跌倒检测) 集成方案

### 方案1：事件驱动（推荐）

**触发机制**：
1. **进入事件**：订阅 `type=1, event=1`（进入房间）事件
2. **人数变化**：订阅 `type=3`（人数变化）事件
3. **姿态变化**：订阅 `type=2, pose=2/5`（跌倒）事件

**流程**：
```
收到进入房间事件（type=1, event=1）
    ↓
检查房间是否是 bathroom
    ↓
检查 track_id 数量（必须 == 1）
    ↓
初始化监测状态（T0 = now）
    ↓
订阅后续数据（每10秒轮询一次）
    ↓
检查是否仍在 bathroom
    ↓
检查 track_id 数量（中途有新 track_id 进入时退出）
    ↓
检查姿态（必须是站立）
    ↓
检查位置变化（>= 10cm 重置计时）
    ↓
检查时间阈值（>= 10分钟/30分钟）
```

### 方案2：混合模式（事件 + 轮询）

**触发机制**：
1. **进入事件**：订阅 `type=1, event=1`（进入房间）事件 → 开始监测
2. **持续监测**：每10秒轮询一次实时数据
3. **退出事件**：订阅 `type=1, event=2`（离开房间）事件 → 停止监测

**优势**：
- 进入/离开事件准确
- 持续监测使用轮询（更可靠）

---

## 🔄 事件处理流程

### 1. 在 wisefido-data-transformer 中处理雷达事件

**位置**：`wisefido-data-transformer/internal/transformer/radar.go`

**处理逻辑**：
```go
// transformEvent 转换事件数据
func (t *RadarTransformer) transformEvent(rawData map[string]interface{}, stdData *models.StandardizedData) error {
    // 检查是否是事件消息
    if cmd, ok := rawData["cmd"]; ok && cmd == "event" {
        eventType, ok := rawData["type"]
        if !ok {
            return nil
        }
        
        switch eventType {
        case 1: // 进出事件
            return t.transformEnterExitEvent(rawData, stdData)
        case 2: // 姿态变化事件
            return t.transformPostureChangeEvent(rawData, stdData)
        case 3: // 人数变化事件
            return t.transformPersonCountEvent(rawData, stdData)
        }
    }
    
    return nil
}

// transformEnterExitEvent 转换进出事件
func (t *RadarTransformer) transformEnterExitEvent(rawData map[string]interface{}, stdData *models.StandardizedData) error {
    data, ok := rawData["data"].([]interface{})
    if !ok {
        return nil
    }
    
    for _, item := range data {
        itemMap, ok := item.(map[string]interface{})
        if !ok {
            continue
        }
        
        event, ok := itemMap["event"].(int)
        if !ok {
            continue
        }
        
        trackID, ok := itemMap["track-id"].(int)
        if ok {
            trackIDPtr := int(trackID)
            stdData.TrackingID = &trackIDPtr
        }
        
        // 设置事件类型
        switch event {
        case 1: // 进入房间
            eventType := "ENTER_ROOM"
            stdData.EventType = &eventType
        case 2: // 离开房间
            eventType := "LEFT_ROOM"
            stdData.EventType = &eventType
        case 3: // 进入区域
            eventType := "ENTER_AREA"
            stdData.EventType = &eventType
        case 4: // 离开区域
            eventType := "LEFT_AREA"
            stdData.EventType = &eventType
        }
    }
    
    return nil
}
```

### 2. 在 wisefido-alarm 中订阅事件

**位置**：`wisefido-alarm/internal/consumer/event_consumer.go`

**处理逻辑**：
```go
// processMessage 处理单条消息
func (c *EventConsumer) processMessage(ctx context.Context, msg rediscommon.StreamMessage, evaluator Evaluator) error {
    // 解析消息数据
    var iotData models.IoTDataMessage
    // ...
    
    // 过滤：处理 ENTER_ROOM 事件
    if iotData.EventType != nil && *iotData.EventType == "ENTER_ROOM" {
        return c.handleENTER_ROOM_Event(ctx, iotData, evaluator)
    }
    
    // 过滤：处理 LEFT_ROOM 事件
    if iotData.EventType != nil && *iotData.EventType == "LEFT_ROOM" {
        return c.handleLEFT_ROOM_Event(ctx, iotData, evaluator)
    }
    
    return nil
}

// handleENTER_ROOM_Event 处理进入房间事件
func (c *EventConsumer) handleENTER_ROOM_Event(
    ctx context.Context,
    iotData models.IoTDataMessage,
    evaluator Evaluator,
) error {
    // 1. 通过 device_id 查询 card_id
    cardInfo, err := c.cardRepo.GetCardByDeviceID(iotData.TenantID, iotData.DeviceID)
    if err != nil {
        return nil // 设备可能未绑定到卡片
    }
    
    // 2. 检查是否是 bathroom
    if !c.checkBathroom(cardInfo) {
        return nil // 不是 bathroom，退出
    }
    
    // 3. 从 Redis 缓存读取实时数据
    realtimeData, err := c.cache.GetRealtimeData(cardInfo.CardID)
    if err != nil {
        return nil // 实时数据不存在
    }
    
    // 4. 检查 track_id 数量（必须 == 1）
    if len(realtimeData.Postures) != 1 {
        return nil // 不是1个 track_id，退出
    }
    
    // 5. 处理 Event 3 监测
    alarms, err := evaluator.Evaluate(iotData.TenantID, *cardInfo, realtimeData)
    // ...
    
    return nil
}
```

---

## 📊 事件映射表

| 雷达事件 | 标准事件类型 | 用途 |
|---------|------------|------|
| `type=1, event=1` | `ENTER_ROOM` | 进入房间（触发 Event 3 监测） |
| `type=1, event=2` | `LEFT_ROOM` | 离开房间（停止 Event 3 监测） |
| `type=1, event=3` | `ENTER_AREA` | 进入区域（可用于 Event 1） |
| `type=1, event=4` | `LEFT_AREA` | 离开区域（可用于 Event 1） |
| `type=2, pose=2` | `SUSPECTED_FALL` | 疑似跌倒 |
| `type=2, pose=5` | `FALL` | 确认跌倒 |
| `type=2, pose=7` | `SUSPECTED_SIT_GROUND` | 疑似坐地 |
| `type=2, pose=8` | `SIT_GROUND` | 确认坐地 |
| `type=3` | `PERSON_COUNT_CHANGED` | 人数变化（检测多人进入） |

---

## ✅ 实现步骤

### 阶段1：在 wisefido-data-transformer 中处理雷达事件

1. ✅ 更新 `transformEvent` 方法
2. ✅ 实现 `transformEnterExitEvent` 方法
3. ✅ 实现 `transformPostureChangeEvent` 方法
4. ✅ 实现 `transformPersonCountEvent` 方法
5. ✅ 设置 `EventType` 字段

### 阶段2：在 wisefido-alarm 中订阅事件

1. ✅ 更新 `EventConsumer` 处理 `ENTER_ROOM` 事件
2. ✅ 更新 `EventConsumer` 处理 `LEFT_ROOM` 事件
3. ✅ 更新 `EventConsumer` 处理 `PERSON_COUNT_CHANGED` 事件
4. ✅ 在 `handleENTER_ROOM_Event` 中触发 Event 3 监测

### 阶段3：更新 Event 3 实现

1. ✅ 支持事件驱动触发
2. ✅ 保持轮询监测（每10秒）
3. ✅ 处理 `LEFT_ROOM` 事件（清除状态）
4. ✅ 处理 `PERSON_COUNT_CHANGED` 事件（多人进入时退出）

---

## 🔍 关键点

1. **事件驱动触发**：使用 `ENTER_ROOM` 事件触发监测，而不是轮询检测
2. **事件驱动退出**：使用 `LEFT_ROOM` 事件停止监测
3. **人数变化检测**：使用 `PERSON_COUNT_CHANGED` 事件检测多人进入
4. **混合模式**：事件触发 + 轮询监测（更可靠）

---

## 📌 待确认问题

1. **事件数据格式**：确认雷达设备上报的事件数据格式是否与文档一致
2. **房间识别**：如何从 `ENTER_ROOM` 事件中识别是 bathroom？
   - 方案1：通过 `device_id` 查询 `room_id`，再检查是否是 bathroom
   - 方案2：事件中包含 `room_id` 或 `area_id`
3. **track_id 识别**：如何从事件中获取 `track_id`？
   - 方案1：从 `data[].track-id` 获取
   - 方案2：从实时数据中获取

