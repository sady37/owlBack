# 数据库查询问题分析

## 问题描述

运行 `go run ./cmd/decode-track --source db --topic-type monitor --limit 1` 时，提示：
```
Warning: Database raw_original does not contain track data.
Please use --source redis to read from Redis Streams.
exit status 1
```

## 问题分析

### 1. 数据库查询逻辑

`ReadFromDatabase` 函数的查询逻辑：
```sql
SELECT 
    id,
    timestamp,
    convert_from(raw_original, 'UTF8')::jsonb as raw_original_json
FROM iot_timeseries
WHERE convert_from(raw_original, 'UTF8')::jsonb->>'topic_type' = $1
ORDER BY timestamp DESC
LIMIT $2
```

### 2. raw_original 字段内容

根据 `buildMinimalAuditData` 函数，`raw_original` 只包含：
- `device_type`
- `topic_type`
- `data_type`
- `category`
- `timestamp`

**不包含**：
- `track` 字段（base64 编码的轨迹数据）
- `bh` 字段（base64 编码的呼吸心率数据）
- 其他原始数据字段

这是**设计决定**，根据 HIPAA/FDA 要求，只保存必要的审计追溯信息，不保存完整原始数据。

### 3. 可能的原因

#### 原因 1：数据库中没有记录
- `iot-timeseries` 服务可能没有正常运行
- 或者数据还没有被消费和写入数据库
- 或者查询条件不匹配（topic_type 不匹配）

#### 原因 2：查询条件问题
- `raw_original` 字段可能为 NULL
- `convert_from` 转换可能失败
- JSON 解析可能失败

#### 原因 3：数据写入失败
- 数据库插入可能失败（但服务日志没有显示错误）
- 设备信息查询可能失败（tenant_id, device_id 缺失）

## 诊断步骤

### 1. 检查服务是否运行
```bash
ps aux | grep wisefido-iot-timeseries
```

### 2. 检查服务日志
```bash
tail -f /tmp/owlBack_logs/wisefido-iot-timeseries.log
```

### 3. 直接查询数据库
```sql
-- 检查是否有 monitor 类型的记录
SELECT COUNT(*) 
FROM iot_timeseries 
WHERE convert_from(raw_original, 'UTF8')::jsonb->>'topic_type' = 'monitor';

-- 检查最近的记录
SELECT 
    id,
    timestamp,
    device_id,
    convert_from(raw_original, 'UTF8')::jsonb as raw_original_json
FROM iot_timeseries
ORDER BY timestamp DESC
LIMIT 10;

-- 检查 raw_original 是否为 NULL
SELECT 
    COUNT(*) as total,
    COUNT(raw_original) as has_raw_original,
    COUNT(*) - COUNT(raw_original) as null_raw_original
FROM iot_timeseries;
```

### 4. 检查 Redis Stream 是否有数据
```bash
# 使用 decode-track 工具从 Redis 读取
go run ./cmd/decode-track --source redis --stream iot:monitor:stream --count 1
```

## 解决方案

### 方案 1：修复查询逻辑（如果确实需要从数据库读取）

如果数据库中有记录，但查询失败，可能需要：
1. 处理 `raw_original` 为 NULL 的情况
2. 处理 JSON 解析失败的情况
3. 添加更详细的错误日志

### 方案 2：确认数据流

数据流应该是：
```
MQTT → Radar Consumer → Redis Stream → IoT-Timeseries Consumer → Database
```

如果数据库中没有记录，检查：
1. Radar Consumer 是否正常发布到 Redis Stream
2. IoT-Timeseries Consumer 是否正常消费 Redis Stream
3. 数据库插入是否成功

### 方案 3：使用 Redis Stream 作为数据源

由于 `raw_original` 不包含 track 数据，**建议使用 Redis Stream 作为数据源**：
```bash
go run ./cmd/decode-track --source redis --stream iot:monitor:stream --count 1
```

## 建议

1. **优先使用 Redis Stream**：Redis Stream 包含完整的原始数据（包括 track 和 bh 字段）
2. **数据库用于标准值查询**：数据库存储的是转换后的标准值，用于业务查询和分析
3. **如果需要原始数据**：从 Redis Stream 读取，而不是从数据库读取
