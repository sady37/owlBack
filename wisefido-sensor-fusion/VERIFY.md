# 传感器融合功能验证指南

## 📋 验证前准备

### 1. 环境要求
- ✅ PostgreSQL 数据库运行中（包含 `cards` 表）
- ✅ Redis 运行中
- ✅ `wisefido-card-aggregator` 已运行并创建了卡片
- ✅ 设备数据已写入 `iot_timeseries` 表

### 2. 检查清单

#### 步骤 1：检查卡片数据
```sql
-- 连接到数据库
psql -h localhost -U postgres -d owlrd

-- 检查 cards 表是否有数据
SELECT 
    card_id, 
    card_type, 
    bed_id, 
    unit_id, 
    card_name,
    jsonb_array_length(devices) as device_count
FROM cards 
LIMIT 10;

-- 检查卡片绑定的设备
SELECT 
    card_id,
    card_type,
    devices
FROM cards
WHERE jsonb_array_length(devices) > 0
LIMIT 5;
```

#### 步骤 2：检查设备绑定关系
```sql
-- 检查设备是否绑定到卡片
SELECT 
    d.device_id, 
    d.device_type, 
    d.bound_bed_id, 
    d.bound_room_id,
    d.unit_id,
    d.monitoring_enabled
FROM devices d
WHERE d.monitoring_enabled = TRUE
LIMIT 10;

-- 检查设备是否有对应的卡片
SELECT 
    d.device_id,
    d.device_type,
    d.bound_bed_id,
    d.bound_room_id,
    c.card_id,
    c.card_type
FROM devices d
LEFT JOIN cards c ON (
    (c.bed_id = d.bound_bed_id AND c.card_type = 'ActiveBed')
    OR 
    (c.unit_id = (
        SELECT r.unit_id FROM rooms r 
        WHERE r.room_id = d.bound_room_id
    ) AND c.card_type = 'Location' AND d.bound_bed_id IS NULL)
)
WHERE d.monitoring_enabled = TRUE
LIMIT 10;
```

#### 步骤 3：检查设备数据
```sql
-- 检查 iot_timeseries 表是否有数据
SELECT 
    device_id,
    device_type,
    COUNT(*) as data_count,
    MAX(timestamp) as latest_timestamp
FROM iot_timeseries
GROUP BY device_id, device_type
ORDER BY latest_timestamp DESC
LIMIT 10;

-- 检查特定设备的最新数据
SELECT 
    device_id,
    device_type,
    timestamp,
    data
FROM iot_timeseries
WHERE device_id = 'your-device-id'
ORDER BY timestamp DESC
LIMIT 5;
```

## 🚀 运行验证

### 步骤 1：启动服务

```bash
cd /Users/sady3721/project/owlBack/wisefido-sensor-fusion

# 设置环境变量（如果需要）
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export REDIS_ADDR=localhost:6379

# 启动服务
go run cmd/wisefido-sensor-fusion/main.go
```

### 步骤 2：检查服务日志

服务启动后，应该看到：
```
{"level":"info","msg":"Starting wisefido-sensor-fusion service","version":"1.5.0","input_stream":"iot:data:stream",...}
{"level":"info","msg":"Stream consumer started","consumer_group":"sensor-fusion-group",...}
```

### 步骤 3：发送测试数据

如果 `iot:data:stream` 中没有数据，可以手动发送测试消息：

```bash
# 使用 redis-cli 发送测试消息
redis-cli XADD iot:data:stream * data '{"device_id":"test-device-1","device_type":"Radar","tenant_id":"test-tenant","timestamp":"2024-01-01T00:00:00Z","data":{"heart_rate":72,"respiration_rate":18}}'
```

### 步骤 4：检查 Redis 缓存

```bash
# 检查缓存键
redis-cli KEYS "vital-focus:card:*:realtime"

# 查看特定卡片的缓存
redis-cli GET "vital-focus:card:{card_id}:realtime"

# 检查 TTL
redis-cli TTL "vital-focus:card:{card_id}:realtime"
```

## ✅ 验证检查点

### 1. 卡片查询验证
- [ ] 服务能成功查询到卡片（日志中无 "Card not found" 错误）
- [ ] `GetCardByDeviceID` 能根据设备ID找到关联的卡片
- [ ] `GetCardDevices` 能获取卡片绑定的所有设备

### 2. 融合逻辑验证
- [ ] HR/RR 融合：优先 Sleepace，无数据则 Radar
- [ ] 床状态/睡眠状态融合：优先 Sleepace
- [ ] 姿态数据：使用所有 Radar 数据
- [ ] 融合条件：同时有 Radar 和 Sleepace 时进行融合

### 3. 缓存更新验证
- [ ] Redis 缓存键格式正确：`vital-focus:card:{card_id}:realtime`
- [ ] 缓存数据格式正确（JSON）
- [ ] TTL 设置正确（300秒 = 5分钟）
- [ ] 缓存数据包含融合后的实时数据

### 4. 完整数据流验证
- [ ] 设备数据能正确发送到 `iot:data:stream`
- [ ] `wisefido-sensor-fusion` 能消费数据
- [ ] 查询卡片 → 融合数据 → 更新缓存的流程正常

## 🐛 常见问题排查

### 问题 1：cards 表为空
**症状**：日志中出现 "Card not found for device"
**解决**：
```bash
# 运行 wisefido-card-aggregator 创建卡片
cd /Users/sady3721/project/owlBack/wisefido-card-aggregator
go run cmd/wisefido-card-aggregator/main.go
```

### 问题 2：设备未绑定到卡片
**症状**：设备数据到达，但找不到关联的卡片
**解决**：
1. 检查设备绑定关系（`devices.bound_bed_id` 或 `devices.bound_room_id`）
2. 确保设备绑定到床位或房间
3. 运行 `wisefido-card-aggregator` 重新创建卡片

### 问题 3：融合数据为空
**症状**：卡片存在，但融合后的数据为空
**解决**：
1. 检查 `iot_timeseries` 表是否有设备数据
2. 检查设备类型是否为 Radar、Sleepace 或 SleepPad
3. 检查数据时间戳是否在合理范围内

### 问题 4：Redis 连接失败
**症状**：日志中出现 "Failed to connect to Redis"
**解决**：
1. 检查 Redis 是否运行：`redis-cli ping`
2. 检查环境变量 `REDIS_ADDR` 是否正确
3. 检查 Redis 密码配置

### 问题 5：数据库连接失败
**症状**：日志中出现 "Failed to connect to database"
**解决**：
1. 检查 PostgreSQL 是否运行
2. 检查环境变量（`DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`）
3. 检查数据库连接权限

## 📊 验证结果记录

### 测试日期：___________

### 测试环境
- PostgreSQL 版本：___________
- Redis 版本：___________
- Go 版本：___________

### 测试结果
- [ ] 卡片查询：✅ / ❌
- [ ] 融合逻辑：✅ / ❌
- [ ] 缓存更新：✅ / ❌
- [ ] 完整数据流：✅ / ❌

### 发现的问题
1. ___________
2. ___________

### 备注
___________

