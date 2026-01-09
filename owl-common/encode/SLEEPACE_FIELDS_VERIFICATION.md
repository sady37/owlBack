# Sleepace 字段验证清单

根据 `wisefido-sleepace/internal/models/message.go` 和 `mqtt_consumer.go` 的实际字段定义。

## 一、RealtimeData 字段（realtime 数据类型）

根据 `models.RealtimeData` 结构：

```go
type RealtimeData struct {
    CommonData
    LeftRight     int `json:"leftRight"`      // 左侧/右侧
    Breath        int `json:"breath"`         // 呼吸率
    Heart         int `json:"heart"`          // 心率
    TurnOver      int `json:"turnOver"`       // 翻身
    BodyMove      int `json:"bodyMove"`       // 体动
    SitUp         int `json:"sitUp"`          // 坐起
    InitStatus    int `json:"initStatus"`     // 初始化状态
    BedStatus     int `json:"bedStatus"`      // 床状态：0=在床, 1=离床
    SignalQuality int `json:"signalQuality"`  // 信号质量
}
```

### 字段处理状态

| 字段名 | 需要 SNOMED 映射 | 处理方式 | 状态 |
|-------|----------------|---------|------|
| `breath` | ❌ | 字段重命名：`breath` → `respiratory_rate`，保留原始字段名 | ✅ |
| `heart` | ❌ | 字段重命名：`heart` → `heart_rate`，保留原始字段名 | ✅ |
| `bedStatus` | ✅ | SNOMED 映射（0=在床, 1=离床） | ✅ |
| `sitUp` | ✅ | SNOMED 映射（sitUp > 0 时映射为坐起事件） | ✅ |
| `turnOver` | ❌ | 直接保留，不映射 | ✅ |
| `bodyMove` | ❌ | 直接保留，不映射 | ✅ |
| `leftRight` | ❌ | 直接保留，不映射 | ✅ |
| `initStatus` | ❌ | 直接保留，不映射 | ✅ |
| `signalQuality` | ❌ | 直接保留，不映射 | ✅ |

**注意**：`realtime` 数据中**不包含** `sleepStage` 字段（`sleepStage` 只在 `sleepStage` 数据类型中存在）。

## 二、SleepStageData 字段（sleepStage 数据类型）

根据 `models.SleepStageData` 结构：

```go
type SleepStageData struct {
    CommonData
    LeftRight  int `json:"leftRight"`
    SleepStage int `json:"sleepStage"` // 0=清醒, 1=浅睡眠, 2=深睡眠, 3=REM睡眠
}
```

### 字段处理状态

| 字段名 | 需要 SNOMED 映射 | 处理方式 | 状态 |
|-------|----------------|---------|------|
| `sleepStage` | ✅ | SNOMED 映射（0-3） | ✅ |
| `leftRight` | ❌ | 直接保留，不映射 | ✅ |

## 三、ConnectionStatusData 字段（connectionStatus 数据类型）

根据 `models.ConnectionStatusData` 结构：

```go
type ConnectionStatusData struct {
    CommonData
    ConnectionStatus int `json:"connectionStatus"` // 0=离线, 1=在线
}
```

### 字段处理状态

| 字段名 | 需要 SNOMED 映射 | 处理方式 | 状态 |
|-------|----------------|---------|------|
| `connectionStatus` | ✅ | SNOMED 映射（0=离线, 1=在线） | ✅ |

## 四、AlarmNotifyData 字段（alarmNotify 数据类型）

根据 `models.AlarmNotifyData` 结构：

```go
type AlarmNotifyData struct {
    CommonData
    Id            int64  `json:"id"`
    Type          string `json:"type"`          // 报警类型
    Status        int    `json:"status"`       // 0=触发, 1=解除
    UserId        string `json:"userId"`        // 用户ID
    RelieveReason string `json:"relieveReason"` // 解除原因
    RelieveTime   int64  `json:"relieveTime"`  // 解除时间
}
```

在 `handleAlarmNotify` 中，字段被映射为：
- `Id` → `alarmId`
- `Type` → `alarmType`
- `Status` → `alarmStatus`

### 字段处理状态

| 字段名（Go） | 字段名（JSON） | 需要 SNOMED 映射 | 处理方式 | 状态 |
|------------|--------------|----------------|---------|------|
| `alarmId` | `alarmId` (from `Id`) | ❌ | 直接保留，不映射 | ✅ |
| `alarmType` | `alarmType` (from `Type`) | ✅ | SNOMED 映射（根据报警类型） | ✅ |
| `alarmStatus` | `alarmStatus` (from `Status`) | ❌ | 直接保留，不映射 | ✅ |
| `userId` | `userId` | ❌ | 直接保留，不映射 | ✅ |
| `relieveReason` | `relieveReason` | ❌ | 直接保留，不映射 | ✅ |
| `relieveTime` | `relieveTime` | ❌ | 直接保留，不映射 | ✅ |

## 五、修正的问题

1. ✅ **删除了 `realtime` 中的 `sleepStage` 处理**：`RealtimeData` 结构中没有 `sleepStage` 字段
2. ✅ **删除了转换表中的 `realtime.sleepStage`**：该字段路径不存在
3. ✅ **更新了注释**：明确说明 `realtime` 数据中不包含 `sleepStage` 字段

## 六、完整的字段映射

### 6.1 realtime 数据类型

```go
// 输入数据（从 handleRealtimeData）
data := map[string]interface{}{
    "device_id": "...",
    "breath": 20,
    "heart": 75,
    "bedStatus": 0,
    "sitUp": 1,
    "turnOver": 0,
    "bodyMove": 1,
    "leftRight": 0,
    "initStatus": 1,
    "signalQuality": 95,
}

// 输出数据（经过 SleepaceEncode）
encoded := map[string]interface{}{
    "device_id": "...",
    "breath": 20,                    // 保留原始字段
    "respiratory_rate": 20,          // 字段重命名
    "heart": 75,                     // 保留原始字段
    "heart_rate": 75,                // 字段重命名
    "bedStatus": 0,
    "bedStatus_snomed_code": "248569007",
    "bedStatus_category": "activity",
    "bedStatus_display_en": "In bed",
    "sitUp": 1,
    "sitUp_snomed_code": "422256002",
    "sitUp_category": "activity",
    "sitUp_display_en": "Sitting up in bed",
    "turnOver": 0,                   // 不映射
    "bodyMove": 1,                   // 不映射
    "leftRight": 0,                  // 不映射
    "initStatus": 1,                 // 不映射
    "signalQuality": 95,             // 不映射
}
```

### 6.2 sleepStage 数据类型

```go
// 输入数据（从 handleSleepStageData）
data := map[string]interface{}{
    "device_id": "...",
    "sleepStage": 2,
    "leftRight": 0,
}

// 输出数据（经过 SleepaceEncode）
encoded := map[string]interface{}{
    "device_id": "...",
    "sleepStage": 2,
    "sleepStage_snomed_code": "248233000",
    "sleepStage_category": "activity",
    "sleepStage_display_en": "Deep sleep",
    "leftRight": 0,                  // 不映射
}
```

### 6.3 connectionStatus 数据类型

```go
// 输入数据（从 handleConnectionStatus）
data := map[string]interface{}{
    "device_id": "...",
    "connectionStatus": 1,
}

// 输出数据（经过 SleepaceEncode）
encoded := map[string]interface{}{
    "device_id": "...",
    "connectionStatus": 1,
    "connectionStatus_snomed_code": "706689003",
    "connectionStatus_category": "device",
    "connectionStatus_display_en": "Online",
}
```

### 6.4 alarmNotify 数据类型

```go
// 输入数据（从 handleAlarmNotify）
data := map[string]interface{}{
    "device_id": "...",
    "alarmId": 12345,
    "alarmType": "LeftBed",
    "alarmStatus": 0,
    "userId": "user123",
    "relieveReason": "",
    "relieveTime": 0,
}

// 输出数据（经过 SleepaceEncode）
encoded := map[string]interface{}{
    "device_id": "...",
    "alarmId": 12345,                // 不映射
    "alarmType": "LeftBed",
    "alarmType_snomed_code": "248570008",
    "alarmType_category": "behavioral",
    "alarmType_display_en": "Left bed",
    "alarmStatus": 0,                // 不映射
    "userId": "user123",             // 不映射
    "relieveReason": "",             // 不映射
    "relieveTime": 0,                // 不映射
}
```

## 七、总结

✅ **所有字段已正确处理**：
- realtime: 9 个字段全部处理
- sleepStage: 2 个字段全部处理
- connectionStatus: 1 个字段已处理
- alarmNotify: 6 个字段全部处理

✅ **SNOMED 映射字段**：
- `bedStatus` - activity
- `sitUp` - activity（仅当 > 0 时）
- `sleepStage` - activity（在 sleepStage 数据类型中）
- `connectionStatus` - device
- `alarmType` - behavioral/clinical/device（根据报警类型）

✅ **代码已修正**：
- 删除了 `realtime` 中不存在的 `sleepStage` 处理
- 删除了转换表中不存在的 `realtime.sleepStage` 定义
- 所有字段路径与实际数据结构一致
