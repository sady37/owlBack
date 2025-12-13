# wisefido-data-transformer 服务的目的和作用

## 📋 背景

### v1.0 vs v1.5 的变化

**v1.0**:
- 雷达设备 → TCP Socket → wisefido-radar 服务 → Redis Hash
- 雷达设备 → 厂家 MQTT 中转 → wisefido-radar 服务 → Redis Hash

**v1.5**:
- 雷达设备 → MQTT Broker（直接）→ wisefido-radar 服务 → Redis Streams
- **变化**: 通讯协议由 TCP Socket 改为 MQTT，**但数据格式没有改变**
- SleepPad 没有变化

---

## 🎯 wisefido-data-transformer 的目的

### 核心目的：**数据标准化和统一格式**

将**不同厂家、不同设备的原始数据**统一转换为**标准格式**，确保：

1. **格式一致性**: 未来各厂家的数据格式统一
2. **标准编码**: 使用 SNOMED CT 编码和 FHIR Category
3. **数据持久化**: 存储到 PostgreSQL TimescaleDB
4. **下游服务兼容**: 为 sensor-fusion、alarm 等服务提供统一的数据格式

---

## 🔍 为什么需要数据标准化？

### 1. 多厂家设备兼容

**现状**:
- 雷达设备：可能有多个厂家（如清澜、其他厂家）
- 每个厂家的数据格式可能不同：
  - 姿态值：厂家A 用 `0,1,2`，厂家B 用 `standing,sitting,lying`
  - 事件类型：厂家A 用 `type:13`，厂家B 用 `event:tracking`
  - 单位：厂家A 用 `dm`，厂家B 用 `cm`

**标准化后**:
- 统一使用 SNOMED CT 编码：`383370001` (Standing position)
- 统一使用标准事件类型：`ENTER_ROOM`, `LEFT_BED`
- 统一单位：`cm`（厘米）

### 2. 未来扩展性

**场景**: 未来可能接入新的设备类型或厂家

**好处**:
- 新设备只需要实现数据转换逻辑
- 下游服务（sensor-fusion、alarm）无需修改
- 数据库结构统一，查询和分析更方便

### 3. 医疗标准对接

**需求**: 与外部 HIS/EPIC 系统对接

**标准化后**:
- 使用 SNOMED CT 编码（国际医疗标准）
- 使用 FHIR Category（FHIR 标准）
- 便于与外部系统对接和数据交换

---

## 📊 数据转换流程

### 输入：原始数据（Redis Streams）

```json
{
    "device_id": "uuid",
    "tenant_id": "uuid",
    "serial_number": "xxx",
    "device_type": "Radar",
    "raw_data": {
        // 厂家原始数据格式（可能不同）
        "posture": 3,           // 厂家原始值
        "event": 13,            // 厂家原始事件类型
        "position_x": 100,      // 单位可能是 dm
        "heart_rate": 75,
        "breath_rate": 18
    },
    "timestamp": 1234567890
}
```

### 处理：数据标准化

1. **SNOMED CT 映射**
   - `posture: 3` → `posture_snomed_code: "109030009"` (Lying position)
   - `event: 13` → `event_type: "ENTER_ROOM"`

2. **FHIR Category 分类**
   - 生命体征 → `category: "vital-signs"`
   - 姿态/运动 → `category: "activity"`
   - 告警事件 → `category: "safety"` / `"clinical"` / `"behavioral"`

3. **单位转换**
   - `position_x: 100` (dm) → `radar_pos_x: 1000` (cm)

4. **数据验证和清洗**
   - 验证数据完整性
   - 过滤无效数据
   - 补充缺失字段

### 输出：标准化数据（PostgreSQL）

```sql
INSERT INTO iot_timeseries (
    tenant_id,
    device_id,
    timestamp,
    data_type,              -- 'observation' 或 'alarm'
    category,               -- 'vital-signs', 'activity', 'safety', ...
    posture_snomed_code,    -- "109030009" (标准编码)
    posture_display,        -- "Lying position" (标准显示名称)
    radar_pos_x,            -- 1000 (cm，统一单位)
    heart_rate,             -- 75 (bpm)
    breath_rate,            -- 18 (次/分)
    raw_original            -- 原始数据（JSONB，用于追溯）
) VALUES (...)
```

---

## 🔄 数据流（完整）

```
[已实现] IoT 设备 → MQTT Broker
    ├─ Radar → wisefido-radar → Redis Streams (radar:data:stream)
    └─ SleepPad → wisefido-sleepace → Redis Streams (sleepace:data:stream)

[待实现] Redis Streams → wisefido-data-transformer
    ├─ 读取原始数据（不同厂家格式）
    ├─ SNOMED CT 映射（查询 snomed_mapping 表）
    ├─ FHIR Category 分类
    ├─ 单位转换（dm → cm）
    ├─ 数据验证和清洗
    └─→ PostgreSQL TimescaleDB (iot_timeseries 表)
    └─→ Redis Streams (iot:data:stream) - 触发下游服务

[待实现] PostgreSQL → wisefido-sensor-fusion
    ├─ 读取标准化数据（统一格式）
    ├─ 多传感器融合
    └─→ Redis (vital-focus:card:{card_id}:realtime)

[待实现] Redis → wisefido-alarm
    ├─ 读取标准化数据（统一格式）
    ├─ 应用报警规则
    └─→ PostgreSQL (alarm_events) + Redis (alarms缓存)
```

---

## ✅ 数据标准化的好处

### 1. 统一数据格式

**问题**: 不同厂家的数据格式不同

**解决**: 统一转换为标准格式
- 姿态：统一使用 SNOMED CT 编码
- 事件：统一使用标准事件类型
- 单位：统一使用 `cm`（厘米）

### 2. 便于查询和分析

**问题**: 不同格式的数据难以统一查询

**解决**: 标准化后可以：
- 统一查询所有设备的数据
- 使用标准编码进行过滤和统计
- 便于数据分析和报表生成

### 3. 下游服务兼容

**问题**: 下游服务需要处理不同格式的数据

**解决**: 标准化后：
- sensor-fusion 服务只需要处理标准格式
- alarm 服务只需要处理标准格式
- 新增设备类型时，下游服务无需修改

### 4. 医疗标准对接

**问题**: 与外部 HIS/EPIC 系统对接需要标准格式

**解决**: 使用 SNOMED CT 和 FHIR 标准
- SNOMED CT 编码：国际医疗标准
- FHIR Category：FHIR 标准分类
- 便于与外部系统对接

### 5. 数据追溯

**问题**: 需要保留原始数据用于追溯

**解决**: 双存储策略
- 标准值：存储在标准字段中（用于查询和处理）
- 原始数据：存储在 `raw_original` 字段（JSONB，用于追溯）

---

## 📝 总结

### wisefido-data-transformer 的核心价值

1. **格式统一**: 将不同厂家的原始数据转换为统一的标准格式
2. **标准编码**: 使用 SNOMED CT 编码和 FHIR Category
3. **数据持久化**: 存储到 PostgreSQL TimescaleDB
4. **下游兼容**: 为下游服务提供统一的数据格式
5. **未来扩展**: 便于接入新设备类型和厂家

### 关键点

- **不是改变数据内容**，而是**统一数据格式**
- **不是丢弃原始数据**，而是**双存储**（标准值 + 原始数据）
- **不是一次性转换**，而是**实时转换**（数据流处理）

---

## 🔗 相关文档

- [iot_timeseries 表结构](../owlRD/db/14_iot_timeseries.sql)
- [snomed_mapping 表结构](../owlRD/db/19_snomed_mapping.sql)
- [FHIR 转换指南](../owlRD/docs/06_FHIR_Simple_Conversion_Guide.md)

