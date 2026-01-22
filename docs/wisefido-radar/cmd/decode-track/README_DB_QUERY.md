# 数据库查询说明

## 为什么数据库中没有记录？

### 可能的原因

1. **数据还没有被写入数据库**
   - `iot-timeseries` 服务可能刚启动，还没有消费 Redis Stream 中的数据
   - 或者数据流中断（MQTT → Redis → Database）

2. **查询条件不匹配**
   - `raw_original` 字段可能为 NULL
   - `topic_type` 可能不在 `raw_original` 中
   - 或者查询的 `topic_type` 与实际数据不匹配

3. **数据写入失败**
   - 数据库插入可能失败（检查服务日志）
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
-- 检查是否有任何记录
SELECT COUNT(*) FROM iot_timeseries;

-- 检查最近的记录
SELECT 
    id,
    timestamp,
    device_id,
    CASE 
        WHEN raw_original IS NOT NULL 
        THEN convert_from(raw_original, 'UTF8')::jsonb->>'topic_type'
        ELSE 'NULL'
    END as topic_type
FROM iot_timeseries
ORDER BY timestamp DESC
LIMIT 10;

-- 检查 raw_original 内容
SELECT 
    id,
    timestamp,
    convert_from(raw_original, 'UTF8')::jsonb as raw_original_json
FROM iot_timeseries
WHERE raw_original IS NOT NULL
ORDER BY timestamp DESC
LIMIT 5;
```

### 4. 检查 Redis Stream 是否有数据
```bash
# 使用 decode-track 工具从 Redis 读取
go run ./cmd/decode-track --source redis --stream iot:monitor:stream --count 1
```

## 重要说明

### raw_original 字段内容

根据设计，`raw_original` 字段**只包含最小审计数据**，不包含完整原始数据：

```json
{
  "device_type": "Radar",
  "topic_type": "monitor",
  "data_type": "observation",
  "category": "activity",
  "timestamp": 1704960000
}
```

**不包含**：
- `track` 字段（base64 编码的轨迹数据）
- `bh` 字段（base64 编码的呼吸心率数据）
- 其他原始数据字段

这是**设计决定**，根据 HIPAA/FDA 要求，只保存必要的审计追溯信息。

### 建议

1. **优先使用 Redis Stream**：Redis Stream 包含完整的原始数据（包括 track 和 bh 字段）
   ```bash
   go run ./cmd/decode-track --source redis --stream iot:monitor:stream --count 1
   ```

2. **数据库用于标准值查询**：数据库存储的是转换后的标准值，用于业务查询和分析

3. **如果需要原始数据**：从 Redis Stream 读取，而不是从数据库读取

## 修复后的查询逻辑

修复后的查询逻辑：
- 处理 `raw_original` 为 NULL 的情况
- 从数据库字段获取 `device_id`（如果 `raw_original` 中没有）
- 更健壮的错误处理

如果数据库中有记录但查询失败，现在应该能够正确返回结果（虽然 track 字段为空）。
