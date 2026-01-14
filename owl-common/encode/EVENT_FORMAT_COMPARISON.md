# 新旧两版事件数据格式对比

## 一、事件类型映射关系

| 旧版 (Redis Stream) | 新版 (HTTP接口) | 说明 |
|-------------------|----------------|------|
| type=1 (进出事件) | event="4" (进出门) | 部分映射 |
| type=1 (进出事件) | event="6" (进出床) | 部分映射 |
| type=1 (进出事件) | event="10" (进出告警区域) | 部分映射 |
| type=2 (姿态变化) | event="2" (跌倒事件) | 部分映射 (pose=5) |
| type=2 (姿态变化) | event="9" (其他告警) | 部分映射 (pose=7/8/10/11) |
| type=3 (人数变化) | event="1" (人数变化) | 完全映射 |
| type=4 (设备状态) | event="5" (上线/离线) | 转换为旧版格式 |
| type=4 (设备状态) | event="7" (信号差) | 转换为旧版格式 |
| type=4 (设备状态) | event="8" (倾角异常) | 转换为旧版格式 |
| - | event="3" (呼吸心率) | 新版独有 |

## 二、相同类型事件的字段差异

### 2.1 进出事件对比

#### 旧版 (type=1)
```json
{
  "type": 1,
  "data": [
    {
      "track-id": 0,        // 人员轨迹 ID
      "event": 1,            // 1-进入房间 2-离开房间 3-进入区域 4-离开区域 5-进入监护模式 6-退出监护模式
      "area_type": 2         // 2-普通床 5-监护床 6-感应区
    }
  ]
}
```

#### 新版对应的事件

**event="4" (进出门)** - 对应旧版的 event=1/2 (进入/离开房间)
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "4",
  "eventName": "用户1离开房间",
  "uid": "F59D3E873F5B",
  "params": {
    "entry2Exit": "1"       // "0"=进入 "1"=离开
  }
}
```

**差异**：
- ❌ 旧版有 `track-id`，新版没有
- ❌ 旧版有 `area_type`，新版没有
- ✅ 新版有 `uid`（设备ID），旧版没有
- ✅ 新版有 `eventName`（事件描述），旧版没有
- ✅ 新版使用 `entry2Exit` 字符串（"0"/"1"），旧版使用 `event` 数值（1/2）

**event="6" (进出床)** - 对应旧版的 event=5/6 (进入/退出监护模式) + area_type=5 (监护床)
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "6",
  "eventName": "进入监护床",
  "uid": "F59D3E873F5B",
  "params": {
    "entry2Exit": "0"       // "0"=进入 "1"=离开
  }
}
```

**差异**：
- ❌ 旧版有 `track-id`，新版没有
- ❌ 旧版有 `area_type` 区分床类型，新版没有（只区分进出床）
- ✅ 新版有 `uid`，旧版没有
- ✅ 新版有 `eventName`，旧版没有
- ✅ 新版使用 `entry2Exit` 字符串，旧版使用 `event` 数值

**event="10" (进出告警区域)** - 对应旧版的 event=3/4 (进入/离开区域) + area_type=6 (感应区)
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "10",
  "eventName": "进入告警区域",
  "uid": "F59D3E873F5B",
  "params": {
    "entry2Exit": "0"       // "0"=进入 "1"=离开
  }
}
```

**差异**：
- ❌ 旧版有 `track-id`，新版没有
- ❌ 旧版有 `area_type` 区分区域类型，新版没有（只区分进出告警区域）
- ✅ 新版有 `uid`，旧版没有
- ✅ 新版有 `eventName`，旧版没有
- ✅ 新版使用 `entry2Exit` 字符串，旧版使用 `event` 数值

### 2.2 姿态变化事件对比

#### 旧版 (type=2)
```json
{
  "type": 2,
  "data": [
    {
      "track-id": 0,        // 人员轨迹 ID
      "pose": 1              // 2-疑似跌倒 5-确认跌倒 7-疑似坐地 8-确认坐地 10-疑似床上坐起 11-确认床上坐起
    }
  ]
}
```

#### 新版对应的事件

**event="2" (跌倒事件)** - 对应旧版的 pose=5 (确认跌倒)
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "2",
  "eventName": "设备检测出跌倒",
  "uid": "F59D3E873F5B"
  // 无 params
}
```

**差异**：
- ❌ 旧版有 `track-id`，新版没有
- ❌ 旧版区分疑似/确认（pose=2/5），新版不区分（只报告确认跌倒）
- ❌ 旧版有 `pose` 字段，新版没有
- ✅ 新版有 `uid`，旧版没有
- ✅ 新版有 `eventName`，旧版没有
- ✅ 新版无 `params`，旧版有 `pose` 字段

**event="9" (其他告警)** - 对应旧版的 pose=7/8/10/11 (坐地、床上坐起)
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "9",
  "eventName": "设备检测出滞留",
  "uid": "F59D3E873F5B",
  "params": {
    "alarmType": "2"        // "4"=坐地告警 "5"=床上坐起告警
  }
}
```

**差异**：
- ❌ 旧版有 `track-id`，新版没有
- ❌ 旧版区分疑似/确认（pose=7/8/10/11），新版不区分（只报告确认告警）
- ❌ 旧版使用 `pose` 数值字段，新版使用 `alarmType` 字符串字段
- ✅ 新版有 `uid`，旧版没有
- ✅ 新版有 `eventName`，旧版没有
- ✅ 新版使用 `alarmType` 字符串，旧版使用 `pose` 数值

### 2.3 人数变化事件对比

#### 旧版 (type=3)
```json
{
  "type": 3,
  "data": {
    "number-people": 1
  }
}
```

#### 新版 (event="1")
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "1",
  "eventName": "人数变化",
  "count": 3,
  "uid": "F59D3E873F5B"
}
```

**差异**：
- ❌ 旧版字段名是 `number-people`，新版是 `count`
- ❌ 旧版在 `data` 对象中，新版在顶层
- ✅ 新版有 `uid`，旧版没有
- ✅ 新版有 `eventName`，旧版没有
- ✅ 新版有 `cmd` 字段，旧版没有

## 三、新版独有的事件类型

### 3.1 event="3" (呼吸心率)
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "3",
  "eventName": "设备检测出心率过高",
  "uid": "F59D3E873F5B",
  "params": {
    "breath": 40,
    "heartbeat": 110,
    "alarmType": 14
  }
}
```
**说明**：旧版没有此事件类型，需要从 monitor/stat 数据中计算

### 3.2 event="5" (上线/离线) - 转换为旧版格式

#### 新版格式
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "5",
  "eventName": "下线",
  "uid": "F59D3E873F5B",
  "params": {
    "isOnline": "1"         // "0"=在线 "1"=离线
  }
}
```

#### 旧版 (Redis Stream) 格式
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "event",
  "timestamp": 1234567890,
  "type": 4,                // 新增类型：设备状态事件
  "data": {
    "device_status": "offline",  // "online"=在线 "offline"=离线
    "device_uid": "F59D3E873F5B"
  }
}
```

**转换规则**：
- `type` = 4（设备状态事件，新增类型）
- `data.device_status` = `params.isOnline === "0" ? "online" : "offline"`
- `data.device_uid` = `uid`
- `eventName` → 不保留（旧版无此字段）

### 3.3 event="7" (信号差) - 转换为旧版格式

#### 新版格式
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "7",
  "eventName": "信号阈值触发",
  "uid": "F59D3E873F5B",
  "params": {
    "recovery": "0"         // "0"=触发 "1"=恢复
  }
}
```

#### 旧版 (Redis Stream) 格式
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "event",
  "timestamp": 1234567890,
  "type": 4,                // 设备状态事件
  "data": {
    "device_status": "signal_poor",  // "signal_poor"=信号差 "signal_normal"=信号正常
    "device_uid": "F59D3E873F5B",
    "recovery": 0            // 0=触发 1=恢复
  }
}
```

**转换规则**：
- `type` = 4（设备状态事件）
- `data.device_status` = `params.recovery === "0" ? "signal_poor" : "signal_normal"`
- `data.device_uid` = `uid`
- `data.recovery` = `parseInt(params.recovery)`
- `eventName` → 不保留（旧版无此字段）

### 3.4 event="8" (倾角异常) - 转换为旧版格式

#### 新版格式
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "8",
  "eventName": "倾角阈值触发",
  "uid": "F59D3E873F5B",
  "params": {
    "recovery": "0"         // "0"=触发 "1"=恢复
  }
}
```

#### 旧版 (Redis Stream) 格式
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "event",
  "timestamp": 1234567890,
  "type": 4,                // 设备状态事件
  "data": {
    "device_status": "angle_abnormal",  // "angle_abnormal"=倾角异常 "angle_normal"=倾角正常
    "device_uid": "F59D3E873F5B",
    "recovery": 0            // 0=触发 1=恢复
  }
}
```

**转换规则**：
- `type` = 4（设备状态事件）
- `data.device_status` = `params.recovery === "0" ? "angle_abnormal" : "angle_normal"`
- `data.device_uid` = `uid`
- `data.recovery` = `parseInt(params.recovery)`
- `eventName` → 不保留（旧版无此字段）

### 3.5 event="9" (其他告警 - 部分)
```json
{
  "cmd": "DEVICE_EVENT",
  "event": "9",
  "eventName": "设备检测出滞留",
  "uid": "F59D3E873F5B",
  "params": {
    "alarmType": "2"        // "1"=离床未归 "2"=滞留 "3"=长时间无人活动
  }
}
```
**说明**：旧版没有这些告警类型（alarmType=1/2/3），需要从其他数据计算

## 四、旧版新增类型：type=4 (设备状态事件)

### 4.1 格式定义

为了支持新版独有的设备状态事件，旧版（Redis Stream）格式新增 `type=4`（设备状态事件）：

```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "event",
  "timestamp": 1234567890,
  "type": 4,                // 设备状态事件（新增）
  "data": {
    "device_status": "...",  // 状态值（见下表）
    "device_uid": "F59D3E873F5B",
    "recovery": 0            // 可选，0=触发 1=恢复（仅信号差和倾角异常）
  }
}
```

### 4.2 device_status 值定义

| 新版事件 | device_status 值 | 说明 |
|---------|-----------------|------|
| event="5" (isOnline="0") | "online" | 设备上线 |
| event="5" (isOnline="1") | "offline" | 设备离线 |
| event="7" (recovery="0") | "signal_poor" | 信号差（触发） |
| event="7" (recovery="1") | "signal_normal" | 信号正常（恢复） |
| event="8" (recovery="0") | "angle_abnormal" | 倾角异常（触发） |
| event="8" (recovery="1") | "angle_normal" | 倾角正常（恢复） |

### 4.3 完整示例

**上线/离线事件**：
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "event",
  "timestamp": 1234567890,
  "type": 4,
  "data": {
    "device_status": "offline",
    "device_uid": "F59D3E873F5B"
  }
}
```

**信号差事件**：
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "event",
  "timestamp": 1234567890,
  "type": 4,
  "data": {
    "device_status": "signal_poor",
    "device_uid": "F59D3E873F5B",
    "recovery": 0
  }
}
```

**倾角异常事件**：
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "event",
  "timestamp": 1234567890,
  "type": 4,
  "data": {
    "device_status": "angle_abnormal",
    "device_uid": "F59D3E873F5B",
    "recovery": 0
  }
}
```

## 五、关键差异总结

### 4.1 字段差异

| 维度 | 旧版 | 新版 |
|------|------|------|
| **人员标识** | `track-id` (数值) | ❌ 无（但有 `uid` 设备ID） |
| **事件类型字段** | `type` (数值: 1/2/3) | `event` (字符串: "1"-"10") |
| **命令字段** | `cmd: "event"` | `cmd: "DEVICE_EVENT"` |
| **事件描述** | ❌ 无 | ✅ `eventName` (字符串) |
| **设备标识** | ❌ 无 | ✅ `uid` (设备UID) |
| **参数结构** | `data` (数组/对象) | `params` (对象，部分事件无) |
| **数值类型** | 数值 (int) | 字符串 (string) 或数值 (int) |

### 4.2 数据粒度差异

| 维度 | 旧版 | 新版 |
|------|------|------|
| **事件分类** | 4种基础类型（type=1/2/3/4） | 10种细化类型 |
| **告警判断** | ❌ 无（需平台计算） | ✅ 有（区分普通事件/告警） |
| **疑似/确认** | ✅ 区分（pose=2/5, 7/8, 10/11） | ❌ 不区分（只报告确认） |
| **区域类型** | ✅ 有（area_type: 2/5/6） | ❌ 无（通过不同event区分） |
| **设备状态** | ✅ 有（type=4，新增） | ✅ 有（上线/离线、信号差、倾角异常） |

### 4.3 相同类型事件的主要差异

1. **进出事件**：
   - 旧版：一个 type=1 包含所有进出类型（房间/区域/床），通过 `event` 和 `area_type` 区分
   - 新版：拆分为 3 个独立事件（event="4"/"6"/"10"），通过不同 event 区分

2. **姿态变化**：
   - 旧版：一个 type=2 包含所有姿态变化，通过 `pose` 区分疑似/确认
   - 新版：拆分为 2 个独立事件（event="2"跌倒，event="9"其他），不区分疑似/确认

3. **人数变化**：
   - 旧版：字段名 `number-people`，在 `data` 对象中
   - 新版：字段名 `count`，在顶层，增加了 `uid` 和 `eventName`

## 六、兼容性建议

如果需要同时支持新旧两版格式：

1. **字段映射**：
   - `track-id` → 无法映射（新版没有）
   - `type` → `event`（需要映射表）
   - `number-people` → `count`
   - `pose` → `alarmType`（部分映射）

2. **事件拆分**：
   - 旧版 type=1 需要根据 `event` 和 `area_type` 映射到新版的 event="4"/"6"/"10"
   - 旧版 type=2 需要根据 `pose` 映射到新版的 event="2" 或 event="9"

3. **缺失数据**：
   - 新版的 `uid` 和 `eventName` 在旧版中不存在，需要从其他数据源获取或生成
   - 旧版的 `track-id` 在新版中不存在，无法映射

4. **设备状态事件转换**：
   - 新版 event="5"/"7"/"8" → 旧版 type=4（设备状态事件）
   - `params.isOnline` → `data.device_status`（"online"/"offline"）
   - `params.recovery` → `data.device_status` + `data.recovery`（"signal_poor"/"signal_normal" 或 "angle_abnormal"/"angle_normal"）
   - `uid` → `data.device_uid`
