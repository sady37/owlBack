# RadarDecoder 输入/输出分析

## 一、函数签名

```go
func RadarDecoder(data map[string]interface{}, topicType string) (interface{}, error)
```

## 二、输入参数

### 1. `data map[string]interface{}`
- **来源**：从 MQTT 报文解析的原始数据
- **包含字段**：
  - 原始数据字段（如 `track`, `bh`, `sleep` base64 字符串）
  - 事件数据字段（如 `type`, `data`）
  - 其他 MQTT 报文中的字段

### 2. `topicType string`
- **可能值**：`"monitor"`, `"stat"`, `"event"`, `"alarm"`
- **说明**：主题类型，用于选择不同的解码函数

## 三、输出返回值

### 返回值类型：`interface{}`
- **可能的值**：
  - `map[string]interface{}`：单个对象（当只有一个 category 时）
  - `[]map[string]interface{}`：数组（当有多个 category 时）
  - `map[string]interface{}{}`：空对象（当没有数据时）

### 返回值内容：`data_value`
- **只包含解码后的数据值**，不包含元数据（`device_id`, `tenant_id`, `topic_type`, `timestamp` 等）
- **格式**：符合 `RADAR_REDIS_STREAM_FORMAT_STANDARD.md` 中定义的 `data_value` 格式

## 四、Topic Type 和 Category 对应关系检查

### ✅ Monitor (topic_type = "monitor")

| 标准文档要求 | 代码实现 | 状态 |
|------------|---------|------|
| category: "track" | ✅ `decodeRadarMonitor` 生成 `category: "track"` | ✅ 正确 |
| category: "vital" | ✅ `decodeRadarMonitor` 生成 `category: "vital"` | ✅ 正确 |

### ✅ Stat/Statistics (topic_type = "stat" → "statistics")

| 标准文档要求 | 代码实现 | 状态 |
|------------|---------|------|
| category: "track" | ✅ `decodeRadarStat` 生成 `category: "track"` | ✅ 正确 |
| category: "sleep" | ✅ `decodeRadarStat` 生成 `category: "sleep"` | ✅ 正确 |

### ✅ Event (topic_type = "event")

| 标准文档要求 | 代码实现 | 状态 |
|------------|---------|------|
| type=1 → category: "enter2out" | ✅ `decodeRadarEvent` 中 `eventType == 1` 生成 `category: "enter2out"` | ✅ 正确 |
| type=2 → category: "pose" | ✅ `decodeRadarEvent` 中 `eventType == 2` 生成 `category: "pose"` | ✅ 正确 |
| type=3 → category: "number-people" | ✅ `decodeRadarEvent` 中 `eventType == 3` 生成 `category: "number-people"` | ✅ 正确 |
| type=5 → category: "isOnline" | ✅ `decodeRadarEvent` 中 `eventType == 5` 生成 `category: "isOnline"` | ✅ 正确 |
| type=7 → category: "signal_poor" | ✅ `decodeRadarEvent` 中 `eventType == 7` 生成 `category: "signal_poor"` | ✅ 正确 |
| type=8 → category: "angle_abnormal" | ✅ `decodeRadarEvent` 中 `eventType == 8` 生成 `category: "angle_abnormal"` | ✅ 正确 |
| type=9 → category: "other" | ✅ `decodeRadarEvent` 中 `eventType == 9` 生成 `category: "other"` | ✅ 正确 |

### ⚠️ Alarm (topic_type = "alarm")

| 标准文档要求 | 代码实现 | 状态 |
|------------|---------|------|
| category: "fall_alarm" | ❌ `decodeRadarAlarm` 返回空数组 | ⚠️ 未实现（告警在 Consumer 层生成） |
| category: "device_offline" | ❌ `decodeRadarAlarm` 返回空数组 | ⚠️ 未实现（告警在 Consumer 层生成） |
| category: "signal_poor_alarm" | ❌ `decodeRadarAlarm` 返回空数组 | ⚠️ 未实现（告警在 Consumer 层生成） |
| category: "angle_abnormal_alarm" | ❌ `decodeRadarAlarm` 返回空数组 | ⚠️ 未实现（告警在 Consumer 层生成） |
| category: "vital_signs_weak" | ❌ `decodeRadarAlarm` 返回空数组 | ⚠️ 未实现（告警在 Consumer 层生成） |

**说明**：`decodeRadarAlarm` 当前为空函数是**正确的**，因为：
- 告警数据是从 Event 和 Stat 数据中**抽取**生成的
- 告警判断和生成应该在 MQTT Consumer 层进行
- 解码器只负责数据格式转换，不处理告警判断

## 五、总结

### ✅ 正确的部分
1. **Monitor**: 正确实现了 `track` 和 `vital` category
2. **Stat**: 正确实现了 `track` 和 `sleep` category
3. **Event**: 正确实现了所有 7 种 event type 对应的 category

### ⚠️ 需要注意的部分
1. **Alarm**: `decodeRadarAlarm` 返回空数组是**设计如此**，告警应该在 Consumer 层从 Event/Stat 数据中生成
2. **Topic Type 映射**: 代码中 `stat` → `statistics` 的映射在 Consumer 层处理，不在解码器中

### 📝 建议
- 当前实现符合设计：解码器只负责解码和格式转换，告警生成在 Consumer 层
- 如果需要，可以在 Consumer 层实现告警生成逻辑，从 Event/Stat 的 `data_value` 中提取告警条件并生成 Alarm 数据
