# 传感器融合功能验证计划

## 📋 验证目标

验证 `wisefido-sensor-fusion` 服务能够：
1. ✅ 正确从 Redis Streams 消费设备数据
2. ✅ 正确查询卡片信息（根据设备ID）
3. ✅ 正确融合多设备数据（Radar + Sleepace）
4. ✅ 正确更新 Redis 缓存

## 🔍 验证方法

### 方法 1：代码逻辑验证（单元测试）

**优点**：快速、不依赖外部服务
**缺点**：无法验证完整数据流

**步骤**：
1. 运行现有单元测试（如果有）
2. 创建关键函数的单元测试
3. 验证融合逻辑正确性

### 方法 2：集成测试（Mock 外部依赖）

**优点**：验证完整流程，不依赖真实环境
**缺点**：需要 Mock 数据库和 Redis

**步骤**：
1. 使用 `go-sqlmock` Mock 数据库
2. 使用 `miniredis` 或 Mock Redis 客户端
3. 测试完整的数据处理流程

### 方法 3：端到端测试（真实环境）

**优点**：最接近生产环境
**缺点**：需要完整的环境配置

**步骤**：
1. 确保 PostgreSQL 和 Redis 运行
2. 运行 `wisefido-card-aggregator` 创建卡片
3. 启动 `wisefido-sensor-fusion` 服务
4. 发送测试数据到 `iot:data:stream`
5. 验证 Redis 缓存更新

## 📝 验证步骤

### 阶段 1：代码逻辑验证 ✅ **推荐先做**

#### 1.1 检查代码编译
```bash
cd /Users/sady3721/project/owlBack/wisefido-sensor-fusion
go build ./...
```

#### 1.2 运行现有测试（如果有）
```bash
go test ./... -v
```

#### 1.3 检查关键函数
- [ ] `GetCardByDeviceID` - 卡片查询逻辑
- [ ] `GetCardDevices` - 设备列表解析
- [ ] `FuseCardData` - 融合逻辑
- [ ] `fuseVitalSigns` - HR/RR 融合
- [ ] `fuseBedAndSleepStatus` - 床状态融合
- [ ] `useRadarPostures` - 姿态数据处理

### 阶段 2：集成测试（可选）

#### 2.1 创建测试数据库 Mock
```go
// 使用 go-sqlmock 创建 Mock 数据库
db, mock, err := sqlmock.New()
```

#### 2.2 创建测试 Redis Mock
```go
// 使用 miniredis 创建 Mock Redis
mr := miniredis.RunT(t)
```

#### 2.3 测试完整流程
- [ ] 测试消息解析
- [ ] 测试卡片查询
- [ ] 测试数据融合
- [ ] 测试缓存更新

### 阶段 3：端到端测试（需要真实环境）

#### 3.1 环境准备
```bash
# 运行环境检查脚本
bash scripts/verify_setup.sh
```

#### 3.2 创建测试数据
```sql
-- 1. 确保有卡片数据
-- 运行 wisefido-card-aggregator 创建卡片

-- 2. 确保有设备数据
INSERT INTO iot_timeseries (device_id, device_type, tenant_id, timestamp, data)
VALUES ('test-device-1', 'Radar', 'test-tenant', NOW(), '{"heart_rate": 72, "respiration_rate": 18}');
```

#### 3.3 启动服务
```bash
cd /Users/sady3721/project/owlBack/wisefido-sensor-fusion
go run cmd/wisefido-sensor-fusion/main.go
```

#### 3.4 发送测试数据
```bash
# 使用 redis-cli 发送测试消息
redis-cli XADD iot:data:stream * data '{"device_id":"test-device-1","device_type":"Radar","tenant_id":"test-tenant","timestamp":1704067200,"data_type":"observation","category":"vital-signs"}'
```

#### 3.5 验证结果
```bash
# 检查 Redis 缓存
redis-cli GET "vital-focus:card:{card_id}:realtime"
```

## ✅ 验证检查清单

### 代码逻辑验证
- [ ] 代码编译通过
- [ ] 单元测试通过（如果有）
- [ ] 关键函数逻辑正确

### 集成测试验证
- [ ] Mock 测试通过
- [ ] 完整流程测试通过

### 端到端测试验证
- [ ] 环境检查通过
- [ ] 服务启动成功
- [ ] 能正确消费数据
- [ ] 能正确查询卡片
- [ ] 能正确融合数据
- [ ] 能正确更新缓存

## 🐛 问题排查

### 问题 1：编译失败
**检查**：
- Go 版本是否 >= 1.18
- 依赖是否完整：`go mod tidy`

### 问题 2：卡片查询失败
**检查**：
- `cards` 表是否有数据
- 设备绑定关系是否正确
- SQL 查询逻辑是否正确

### 问题 3：融合数据为空
**检查**：
- `iot_timeseries` 表是否有数据
- 设备类型是否匹配（Radar/Sleepace/SleepPad）
- 融合逻辑是否正确

### 问题 4：缓存更新失败
**检查**：
- Redis 连接是否正常
- 缓存键格式是否正确
- TTL 设置是否正确

## 📊 验证结果记录

### 测试日期：___________

### 测试方法
- [ ] 代码逻辑验证
- [ ] 集成测试
- [ ] 端到端测试

### 测试结果
- [ ] 代码编译：✅ / ❌
- [ ] 单元测试：✅ / ❌
- [ ] 集成测试：✅ / ❌
- [ ] 端到端测试：✅ / ❌

### 发现的问题
1. ___________
2. ___________

### 备注
___________

## 🚀 下一步

验证通过后，下一步应该是：
1. **实现报警评估层**（wisefido-alarm）
   - 读取融合后的实时数据
   - 评估报警规则
   - 生成报警事件

## 🔗 相关文档

- `VERIFY.md` - 详细验证指南
- `scripts/verify_setup.sh` - 环境检查脚本
- `CHECK_SUMMARY.md` - 代码检查总结
- `docs/system_architecture_complete.md` - 系统架构文档

