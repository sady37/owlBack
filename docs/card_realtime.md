# RealtimeData 架构设计文档

┌─ Handler 层 ────────────────────────────────────┐
│ if msg.ts + 5s > realtime.ts?                    │
│   YES → 进入 service 层（设备级检查）            │
│   NO  → 直接拒绝（古老数据）                     │
└─────────────────────────────────────────────────┘
                    ↓
┌─ Service 层 ────────────────────────────────────┐
│ （仅在 5s 窗口内执行）                           │
│                                                  │
│ if device.ts > existing.ts?                      │
│   YES → 更新该设备数据                           │
│   NO  → 丢弃（设备级排队问题）                   │
└─────────────────────────────────────────────────┘

## 1. 数据结构设计

### 1.1 Vital/Posture 的组织方式

**目标**：清晰的设备级时间戳管理，避免排队导致的数据丢失

#### VitalSimplified（生命体征）
```go
type VitalSimplified struct {
	DeviceID        string  `json:"device_id"`
	Timestamp       int64   `json:"timestamp"`      // 该设备本次MQTT的时间戳
	RespiratoryRate *int    `json:"respiratory_rate,omitempty"`
	HeartRate       *int    `json:"heart_rate,omitempty"`
	SleepStatus     *string `json:"sleep_status,omitempty"`
	Stability       *string `json:"stability,omitempty"`
}
```

#### DevicePosture（姿态）
```go
type DevicePosture struct {
	DeviceID  string `json:"device_id"`
	Timestamp int64  `json:"timestamp"`     // 该设备本次MQTT的时间戳
	Postures  []int  `json:"postures"`      // [pose1, pose2, pose3, ...]（每人一个）
}
```

#### RealtimeData 顶层
```go
type RealtimeData struct {
	CardID    string            `json:"card_id"`
	Timestamp int64             `json:"timestamp"`    // 整张卡片最后一次更新的时间戳（用于时序检查）
	
	Vital     []VitalSimplified `json:"vital"`       // 数组，每个元素是一个设备的vital
	Postures  []DevicePosture   `json:"postures"`    // 数组（不用map），便于遍历和时间戳管理
	
	BedState  *BedState        `json:"bed_state,omitempty"`
	RoomState *RoomState       `json:"room_state,omitempty"`
	ActiveAlarms map[string]int `json:"alarms,omitempty"`   // key: EMERG/ALERT/CRIT/ERR/WARNING/NOTICE
}
```

**改进点**：
- ✅ `Vital` 和 `Postures` 都是结构体数组，每个元素自带设备级 `Timestamp`
- ✅ 避免嵌套，JSON 更清晰
- ✅ 便于按设备维度做时间检查
- ✅ `RealtimeData.Timestamp` 用于全局时序，设备自己的 `Timestamp` 用于设备级时序

---

## 2. 多设备时间刷新问题分析

### 2.1 场景描述

```
Timeline:
  realtimeData.Timestamp = 100

消息1 (mqtt.ts=101, device_A)
  → 101 > 100? YES
  → 更新 realtimeData.Timestamp=101
  → 更新 device_A.vital.timestamp=101
  
消息2 (mqtt.ts=100, device_B，因排队迟到)
  → 100 > 101? NO（如果直接检查会丢弃！❌）
  → 但 device_B.vital.timestamp=? 可能是 95
  → 如果 100 > 95，应该更新 device_B ✅

消息3 (mqtt.ts=99, device_C，报警数据)
  → 99 > 101? NO
  → device_C.vital.timestamp=? 可能是 98
  → 如果 99 > 98，应该更新 device_C ✅
```

### 2.2 问题根源

- **问题**：单一全局时间戳检查会导致排队中的消息被丢弃
- **原因**：网络延迟、处理队列、MQTT 集群负载不均导致消息到达顺序不一致
- **影响**：设备级数据无法及时更新

### 2.3 解决方案：双重时序检查

#### 检查流程

```
Handler 层（全局时序检查）
  ↓
if msg.Timestamp < realtimeData.Timestamp {
    warn("Message older than realtime, but checking device-level timestamp")
    // 不直接丢弃，继续到 service 层
}
  ↓
Service 层（设备级时序检查）
  ↓
existingVital := findVitalByDeviceID(realtimeData.Vital, msg.DeviceID)
if existingVital != nil && msg.Timestamp <= existingVital.Timestamp {
    debug("Device vital already newer, skip")
    return  // 丢弃
}
  ↓
// 允许更新
vital := &VitalSimplified{
    DeviceID:  msg.DeviceID,
    Timestamp: msg.Timestamp,  // 设备级时间戳
    ...
}
realtimeData.Vital = append/replace(vital)

// 只在全局时间戳需要更新时才更新
if msg.Timestamp > realtimeData.Timestamp {
    realtimeData.Timestamp = msg.Timestamp
}
```

#### 伪代码实现

```go
func (s *EventAlarmService) handleVital(ctx context.Context, msg *redis.IoTStreamMessage, realtimeData *card.RealtimeData) {
	// 1. 全局时序检查（信息日志）
	if msg.Timestamp < realtimeData.Timestamp {
		s.logger.Info("Message older than realtime timestamp, checking device-level",
			zap.Int64("msg_ts", msg.Timestamp),
			zap.Int64("realtime_ts", realtimeData.Timestamp),
			zap.String("device_id", msg.DeviceID))
	}

	// 2. 设备级时序检查
	existingVital := findVitalByDeviceID(realtimeData.Vital, msg.DeviceID)
	if existingVital != nil && msg.Timestamp <= existingVital.Timestamp {
		s.logger.Debug("Device vital already newer, skip",
			zap.Int64("msg_ts", msg.Timestamp),
			zap.Int64("existing_ts", existingVital.Timestamp))
		return
	}

	// 3. 更新设备的 vital
	newVital := &card.VitalSimplified{
		DeviceID:  msg.DeviceID,
		Timestamp: msg.Timestamp,
		// ... 其他字段
	}

	// Replace or append
	if existingVital != nil {
		// 找到并替换
		for i, v := range realtimeData.Vital {
			if v.DeviceID == msg.DeviceID {
				realtimeData.Vital[i] = *newVital
				break
			}
		}
	} else {
		// 追加
		realtimeData.Vital = append(realtimeData.Vital, *newVital)
	}

	// 4. 更新全局时间戳（只有在更新消息时戳时）
	if msg.Timestamp > realtimeData.Timestamp {
		realtimeData.Timestamp = msg.Timestamp
	}
}
```

---

## 3. BedState/RoomState/ActiveAlarms 的持久化策略

### 3.1 现状问题

**问题**：
- 高频更新（每秒）导致 Redis 写入压力大
- 前端轮询时如果没有新事件，状态会消失
- 前端误以为"设备掉线"或"事件清除"

### 3.2 分离设计方案

#### A. 低频状态字段（BedState/RoomState）

**特点**：
- 只在**状态变化时更新**（不是每秒）
- 更新时必须 `Timestamp > 旧值.Timestamp`
- Redis TTL：12小时（长期持久化）

**处理逻辑**：
```
消息1: BedState=in_bed, stateSince=100
  → 初始化 realtimeData.BedState = {in_bed, 100}
  → 保存到 Redis

消息2: BedState=in_bed, stateSince=101
  → 101 > 100? YES
  → 更新 realtimeData.BedState.Timestamp = 101

消息3: BedState=out_of_bed, stateSince=102
  → 102 > 101? YES
  → 更新状态 + 时间戳

消息4: BedState=in_bed, stateSince=98（排队）
  → 98 > 102? NO
  → 丢弃（保留 out_of_bed）
```

#### B. 高频告警字段（ActiveAlarms）

**特点**：
- 告警到达时立即记录
- 永不因过期而自动清除
- 需要前端或管理员手动确认/清除

**Redis 存储策略**：
```
Key: iot:alarms:{CardID}
Structure: {
  "EMERG": {
    "count": 2,
    "last_alarm": "alarm_json_object",
    "timestamp": 1234567890
  },
  "ALERT": {
    "count": 1,
    ...
  },
  ...
}
```

**更新时机**：
- 告警到达 → 立即 HSET 更新对应 level 的计数 + 最新告警信息
- 不与 vitals/postures 混在一起（避免被 5秒 TTL 清理）
- 独立 TTL：24小时或更长（由业务决定）

### 3.3 前端交互流程

#### 读取流程
```
前端轮询（2秒一次）
  ↓
GET /api/v1/card/{cardID}/realtime
  ↓
后端执行 HGETALL iot:realtime:{CardID}
返回：{
  vitals: [...],           // 高频，可能刚更新
  postures: [...],         // 高频，可能刚更新
  bed_state: {...},        // 低频，可能是 5 分钟前的值 ✅（表示还在床上）
  room_state: {...},       // 低频，可能是 N 分钟前的值 ✅（表示房间有人）
  alarms: {EMERG: 2, ...}  // 告警计数，永久保存 ✅
}
```

#### 前端自增确认流程
```
1. 用户看到 EMERG 告警计数 = 2
   
2. 用户点击"确认" → activeCount[EMERG]--
   → 本地显示 activeCount[EMERG] = 1

3. 同时发送 API:
   POST /api/v1/card/{cardID}/alarms/acknowledge
   {
     "level": "EMERG",
     "count": 1
   }
   
4. 后端更新 Redis:
   HSET iot:alarms:{CardID} EMERG.count 1

5. 前端下次轮询时获取最新计数
```

### 3.4 后端定时清理（可选）

**目的**：清理超期的告警记录（防止 Redis 无限增长）

**策略**：
```
定时任务（5分钟一次）
  ↓
遍历所有 iot:alarms:{CardID}
  ↓
对每个告警：
  if (now - last_timestamp) > 24小时 {
    删除该告警记录
  } else if (count == 0 && last_timestamp > 1小时) {
    删除已确认的告警
  }
```

---

## 4. Redis 存储结构优化（建议）

### 4.1 当前方案（Object）
```
Key: iot:realtime:{CardID}
Value: JSON(RealtimeData)  // 整个对象，5秒 TTL

问题：
❌ 任何字段变化都要序列化整个对象
❌ 高频字段（vitals）和低频字段（bed_state）混在一起
❌ 不利于单字段查询和更新
```

### 4.2 改进方案（Hash）
```
Key: iot:realtime:{CardID}
Fields:
  - vitals       (高频，每秒可能变)
  - postures     (高频，每秒可能变)
  - bed_state    (低频，仅状态变化时)
  - room_state   (低频，仅状态变化时)
  - timestamp    (全局版本号)
  
TTL: 12小时

优点：
✅ HSET iot:realtime:{CardID} vitals "<json>" → 只更新一个字段
✅ HGET iot:realtime:{CardID} bed_state → 只获取需要的字段
✅ 低频字段不会因为高频字段更新而重新序列化
✅ 支持 HGETALL 一次获取所有字段
```

### 4.3 告警单独存储
```
Key: iot:alarms:{CardID}
Fields:
  - EMERG
  - ALERT
  - CRIT
  - ERR
  - WARNING
  - NOTICE
  
TTL: 24小时或更长

结构（每个 field）:
{
  "count": 2,
  "last_alarm_type": "Fall",
  "timestamp": 1234567890
}
```

---

## 5. 实施优先级

### Phase 1（立即）
- [x] EventAlarmService 中实现设备级时序检查
- [x] BedState/RoomState 只在状态变化时更新
- [x] 建立 category = "AlarmLevel.AlarmType" 格式

### Phase 2（优化）
- [ ] Redis 改为 Hash 结构（避免整对象序列化）
- [ ] 告警单独存储（独立 TTL）
- [ ] 前端轮询优化（选择性字段获取）

### Phase 3（后续）
- [ ] 前端告警计数自增
- [ ] 后端定时告警清理任务
- [ ] 告警队列（处理高并发场景）

---

## 7. 周期性清理任务（5分钟周期）

### 7.1 目的

防止以下数据过期污染实时数据，同时保持 BedState/RoomState/ActiveAlarms 的持久化：

1. **清理过期的 Vital/Posture** - 移除超过 5 秒未更新的设备数据
2. **刷新持续状态** - 更新 BedState/RoomState 的 StayTime（防止被误认为过期）
3. **同步告警数据** - 重新从 DB 查询最新的告警使能配置，更新 ActiveAlarms 计数

### 7.2 清理流程

#### Step 1: 清理过期 Vital/Posture（5秒阈值）

```go
func (s *EventAlarmService) CleanupExpiredDeviceData(ctx context.Context, cardID string) error {
    realtimeData, _ := s.cacheRepo.GetRealtimeData(ctx, cardID)
    if realtimeData == nil {
        return nil
    }
    
    threshold := realtimeData.Timestamp - 5  // 5秒阈值
    
    // 清理过期 Vital
    cleanedVital := []card.VitalSimplified{}
    for _, vital := range realtimeData.Vital {
        if vital.Timestamp >= threshold {
            cleanedVital = append(cleanedVital, vital)
        } else {
            s.logger.Debug("Removing expired vital",
                zap.String("device_id", vital.DeviceID),
                zap.Int64("device_ts", vital.Timestamp),
                zap.Int64("threshold_ts", threshold))
            // 同时删除 Redis 中的设备级缓存
            _ = s.cacheRepo.DeleteVitalSimplified(ctx, cardID, vital.DeviceID)
        }
    }
    realtimeData.Vital = cleanedVital
    
    // 清理过期 Posture
    cleanedPostures := []card.DevicePosture{}
    for _, posture := range realtimeData.Postures {
        if posture.Timestamp >= threshold {
            cleanedPostures = append(cleanedPostures, posture)
        } else {
            s.logger.Debug("Removing expired posture",
                zap.String("device_id", posture.DeviceID),
                zap.Int64("device_ts", posture.Timestamp),
                zap.Int64("threshold_ts", threshold))
            // 同时删除 Redis 中的设备级缓存
            _ = s.cacheRepo.DeleteDevicePosture(ctx, cardID, posture.DeviceID)
        }
    }
    realtimeData.Postures = cleanedPostures
    
    // 更新 Redis
    if err := s.cacheRepo.SetRealtimeData(ctx, cardID, realtimeData, 12*time.Hour); err != nil {
        s.logger.Warn("Failed to update realtime data after cleanup", zap.Error(err))
        return err
    }
    
    return nil
}
```

#### Step 2: 刷新持续状态（BedState/RoomState）

```go
func (s *EventAlarmService) RefreshPersistentState(ctx context.Context, cardID string) error {
    realtimeData, _ := s.cacheRepo.GetRealtimeData(ctx, cardID)
    if realtimeData == nil {
        return nil
    }
    
    now := time.Now().Unix()
    
    // 刷新 BedState 的 StayTime（防止被认为过期）
    if realtimeData.BedState != nil {
        stayTimeMinutes := int((now - int64(realtimeData.BedState.Timestamp)) / 60)
        realtimeData.BedState.Timestamp = int(now)
        s.logger.Debug("Refreshed BedState",
            zap.String("bed_id", realtimeData.BedState.BedID),
            zap.String("state", realtimeData.BedState.CurrentState),
            zap.Int("stay_time_min", stayTimeMinutes))
    }
    
    // 刷新 RoomState 的 StayTime
    if realtimeData.RoomState != nil {
        stayTimeMinutes := int((now - int64(realtimeData.RoomState.Timestamp)) / 60)
        realtimeData.RoomState.Timestamp = int(now)
        s.logger.Debug("Refreshed RoomState",
            zap.String("room_id", realtimeData.RoomState.RoomID),
            zap.String("room_name", realtimeData.RoomState.RoomName),
            zap.Int("stay_time_min", stayTimeMinutes))
    }
    
    // 更新 Redis
    if err := s.cacheRepo.SetRealtimeData(ctx, cardID, realtimeData, 12*time.Hour); err != nil {
        s.logger.Warn("Failed to update realtime data after refresh", zap.Error(err))
        return err
    }
    
    return nil
}
```

#### Step 3: 同步告警数据（从 DB 重新查询）

```go
func (s *EventAlarmService) SyncActiveAlarms(ctx context.Context, cardID string, tenantID string, deviceUID string) error {
    realtimeData, _ := s.cacheRepo.GetRealtimeData(ctx, cardID)
    if realtimeData == nil {
        return nil
    }
    
    // 从 DB 查询最新的告警使能配置
    enablementItems, err := s.deviceRepo.GetAlarmEnablement(ctx, tenantID, deviceUID)
    if err != nil {
        s.logger.Warn("Failed to get alarm enablement", zap.Error(err))
        return err
    }
    
    // 构建告警统计（按 level 统计）
    if realtimeData.ActiveAlarms == nil {
        realtimeData.ActiveAlarms = make(map[string]int)
    }
    
    // 初始化告警计数为 0（后续从 Redis alarms cache 读取真实数据）
    for _, item := range enablementItems {
        if item.IsEnabled == 1 {
            // 从 Redis alarms cache 查询该 level 的告警计数
            alarmKey := fmt.Sprintf("iot:alarms:%s", cardID)
            count, err := s.cacheRepo.GetAlarmCount(ctx, alarmKey, item.AlarmLevel)
            if err == nil && count > 0 {
                realtimeData.ActiveAlarms[item.AlarmLevel] = count
            } else if err == nil {
                realtimeData.ActiveAlarms[item.AlarmLevel] = 0
            }
        }
    }
    
    s.logger.Debug("Synced active alarms",
        zap.String("card_id", cardID),
        zap.Any("alarm_counts", realtimeData.ActiveAlarms))
    
    // 更新 Redis
    if err := s.cacheRepo.SetRealtimeData(ctx, cardID, realtimeData, 12*time.Hour); err != nil {
        s.logger.Warn("Failed to update realtime data after sync", zap.Error(err))
        return err
    }
    
    return nil
}
```

### 7.3 定时调度（5分钟周期）

在 main.go 中添加周期性任务：

```go
// 启动 5 分钟周期的清理任务
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        // 遍历所有活跃卡片
        cards := getAllActiveCards(ctx)  // 从 Redis 或 DB 获取活跃卡片列表
        for _, card := range cards {
            // 清理过期设备数据
            if err := eventAlarmSvc.CleanupExpiredDeviceData(ctx, card.CardID); err != nil {
                logger.Warn("Failed to cleanup expired data", zap.String("card_id", card.CardID))
            }
            
            // 刷新持续状态
            if err := eventAlarmSvc.RefreshPersistentState(ctx, card.CardID); err != nil {
                logger.Warn("Failed to refresh persistent state", zap.String("card_id", card.CardID))
            }
            
            // 同步告警数据
            if err := eventAlarmSvc.SyncActiveAlarms(ctx, card.CardID, card.TenantID, card.DeviceUID); err != nil {
                logger.Warn("Failed to sync active alarms", zap.String("card_id", card.CardID))
            }
            
            logger.Debug("Completed periodic cleanup",
                zap.String("card_id", card.CardID),
                zap.Time("timestamp", time.Now()))
        }
    }
}()
```

### 7.4 清理任务的关键点

| 任务 | 周期 | 触发条件 | 操作 |
|------|------|---------|------|
| **清理过期 Vital/Posture** | 5分钟 | 永远执行 | 删除 `timestamp < realtimeData.ts - 5s` 的元素 |
| **刷新 BedState/RoomState** | 5分钟 | 状态存在时 | 更新 `Timestamp = now`（防止被认为过期） |
| **同步 ActiveAlarms** | 5分钟 | 永远执行 | 重新查询 DB 的告警配置，更新计数 |
| **删除设备级缓存** | 5分钟 | 清理时 | 删除 Redis 中对应设备的 vital/posture key |

### 7.5 时间流的完整图景

```
初始化（T=0）
  ↓ mqtt vitals/postures 到达
  → Handler 检查：ts + 5s > realtime.ts? YES
  → Service 设备级检查：device.ts > existing.ts? YES
  → 更新 realtimeData，保存 Redis
  
5 分钟后（T=300s）
  → 清理任务启动
  → 删除 timestamp < realtimeData.ts - 5s 的 vital/posture
  → 刷新 BedState/RoomState 的 timestamp
  → 重新查询 DB 同步 ActiveAlarms 计数
  → 结果保存回 Redis
  
10 分钟后（T=600s）
  → 新的 mqtt 消息继续到达
  → 循环...
```

---

## 8. 相关文件

- [card_types.go](../owl-common/card/card_types.go) - RealtimeData 定义
- [event_alarm_service.go](../wisefido-cardagg/internal/service/event_alarm_service.go) - 事件/告警处理 + 清理任务
- [event_alarm_handler.go](../wisefido-cardagg/internal/consumer/event_alarm_handler.go) - 5秒窗口检查
- [mqtt_consumer.go](../wisefido-qinglan/internal/consumer/mqtt_consumer.go) - MQTT 消费逻辑

