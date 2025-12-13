# wisefido-alarm 测试指南

## 📋 测试准备

### 1. 环境要求
- PostgreSQL 运行中
- Redis 运行中
- `wisefido-card-aggregator` 已创建卡片
- `wisefido-sensor-fusion` 已生成实时数据

### 2. 运行环境验证

```bash
cd /Users/sady3721/project/owlBack/wisefido-alarm
bash scripts/verify_setup.sh
```

## 🧪 测试场景

### 测试 1：服务启动测试

**目标**：验证服务能正常启动

**步骤**：
```bash
export TENANT_ID="test-tenant"
go run cmd/wisefido-alarm/main.go
```

**预期结果**：
- ✅ 服务成功启动
- ✅ 无错误日志
- ✅ 显示 "Cache consumer started"

### 测试 2：卡片查询测试

**目标**：验证服务能正确查询卡片

**前置条件**：
- `cards` 表有数据

**步骤**：
1. 启动服务
2. 观察日志中的 `card_count`

**预期结果**：
- ✅ 日志显示正确的卡片数量
- ✅ 能成功读取卡片信息

**验证SQL**：
```sql
SELECT COUNT(*) FROM cards WHERE tenant_id = 'test-tenant';
```

### 测试 3：实时数据读取测试

**目标**：验证服务能正确读取实时数据

**前置条件**：
- Redis 中有实时数据缓存

**步骤**：
1. 确保 Redis 中有 `vital-focus:card:*:realtime` 键
2. 启动服务
3. 观察日志

**预期结果**：
- ✅ 能成功读取实时数据
- ✅ 如果卡片没有实时数据，服务能正确处理（跳过）

**验证命令**：
```bash
redis-cli KEYS "vital-focus:card:*:realtime"
redis-cli GET "vital-focus:card:{card_id}:realtime"
```

### 测试 4：报警评估测试（基础）

**目标**：验证评估器能正常执行

**前置条件**：
- 卡片有实时数据

**步骤**：
1. 准备测试数据（卡片 + 实时数据）
2. 启动服务
3. 观察评估结果

**预期结果**：
- ✅ 评估器能正常执行
- ✅ 当前返回空列表（待完善评估逻辑）
- ✅ 无异常错误

### 测试 5：状态管理测试

**目标**：验证状态管理器能正常工作

**步骤**：
1. 启动服务
2. 检查 Redis 状态键

**验证命令**：
```bash
# 检查状态键（如果有状态生成）
redis-cli KEYS "alarm:state:*"
```

**预期结果**：
- ✅ 状态管理器能正常设置/获取状态
- ✅ 状态键格式正确

### 测试 6：报警缓存更新测试

**目标**：验证报警缓存能正确更新

**前置条件**：
- 评估器生成报警（待完善）

**步骤**：
1. 启动服务
2. 等待评估完成
3. 检查报警缓存

**验证命令**：
```bash
redis-cli KEYS "vital-focus:card:*:alarms"
redis-cli GET "vital-focus:card:{card_id}:alarms"
redis-cli TTL "vital-focus:card:{card_id}:alarms"
```

**预期结果**：
- ✅ 报警缓存键格式正确
- ✅ TTL 设置正确（30秒）
- ✅ 数据格式正确（JSON）

## 🔍 调试技巧

### 1. 启用调试日志

```bash
export LOG_LEVEL="debug"
go run cmd/wisefido-alarm/main.go
```

### 2. 检查服务状态

```bash
# 检查进程
ps aux | grep wisefido-alarm

# 检查端口（如果有HTTP接口）
netstat -an | grep LISTEN
```

### 3. 查看详细日志

```bash
# 使用 jq 格式化 JSON 日志
go run cmd/wisefido-alarm/main.go | jq .
```

### 4. 数据库调试

```sql
-- 检查卡片数据
SELECT * FROM cards WHERE tenant_id = 'test-tenant' LIMIT 5;

-- 检查设备绑定
SELECT device_id, bound_bed_id, bound_room_id, unit_id 
FROM devices 
WHERE tenant_id = 'test-tenant' 
LIMIT 10;
```

### 5. Redis 调试

```bash
# 监控 Redis 命令
redis-cli MONITOR

# 检查键
redis-cli KEYS "*"

# 检查特定键的值
redis-cli GET "vital-focus:card:{card_id}:realtime"
```

## 🐛 常见问题

### 问题 1：服务启动失败

**症状**：服务启动后立即退出

**排查**：
1. 检查环境变量（特别是 `TENANT_ID`）
2. 检查数据库连接
3. 检查 Redis 连接
4. 查看错误日志

### 问题 2：没有卡片数据

**症状**：日志显示 `card_count: 0`

**排查**：
1. 检查 `cards` 表是否有数据
2. 检查租户ID是否正确
3. 确认是否运行了 `wisefido-card-aggregator`

### 问题 3：无法读取实时数据

**症状**：日志显示 "Realtime data not found"

**排查**：
1. 检查 Redis 中是否有实时数据缓存
2. 检查缓存键格式是否正确
3. 确认是否运行了 `wisefido-sensor-fusion`

### 问题 4：评估器返回空列表

**症状**：评估器执行但没有报警

**说明**：
- 这是**正常现象**，因为当前事件1-4的评估逻辑是简化版本
- 需要完善评估逻辑才能生成报警

## 📊 性能测试

### 1. 批量评估性能

**测试**：评估大量卡片（100+ 张）

**观察指标**：
- 评估时间
- 内存使用
- CPU 使用

### 2. 轮询性能

**测试**：长时间运行（1小时+）

**观察指标**：
- 内存泄漏
- 连接池状态
- 错误率

## 🔗 相关文档

- `QUICK_START.md` - 快速启动指南
- `VERIFY.md` - 详细验证指南
- `IMPLEMENTATION_SUMMARY.md` - 实现总结

