# wisefido-alarm - 报警评估服务

## 📋 简介

`wisefido-alarm` 是报警评估层服务，负责**云端事件报警评估**：

### 处理的报警类型（云端事件）

- **事件1**：床上跌落检测
  - 触发：sleepad 离床事件 + radar 存在
  - 方式：事件驱动（sleepad 离床事件触发）
  
- **事件3**：Bathroom可疑跌倒检测
  - 触发：bathroom + 无出门事件
  - 方式：基于卡片数据更新触发（2秒1次）
  
- **事件4**：雷达检测到人突然消失
  - 触发：无出门事件 + 事件距雷达边界
  - 方式：基于卡片数据更新触发（2秒1次）

### 不处理的报警类型

- **设备直接报警**（由 `wisefido-sensor-fusion` 处理）：
  - Fall, SuspectedFall, OfflineAlarm, LowBattery, DeviceFailure 等
  - 设备直接上报的报警事件，在数据流中立即处理

### 服务职责

- 读取融合后的实时数据（`vital-focus:card:{card_id}:realtime`）
- 应用报警规则评估（事件1, 3, 4）
- 生成报警事件
- 写入 PostgreSQL（`alarm_events` 表）
- 更新 Redis 缓存（`vital-focus:card:{card_id}:alarms`）

## 🚀 快速开始

### 1. 环境要求

- Go 1.21+
- PostgreSQL（包含 `cards`, `alarm_cloud`, `alarm_device`, `alarm_events` 表）
- Redis
- 前置服务：
  - `wisefido-card-aggregator` - 创建卡片
  - `wisefido-sensor-fusion` - 生成实时数据

### 2. 环境验证

```bash
cd /Users/sady3721/project/owlBack/wisefido-alarm
bash scripts/verify_setup.sh
```

### 3. 设置环境变量

```bash
# 必需
export TENANT_ID="your-tenant-id"

# 可选（有默认值）
export DB_HOST="localhost"
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="owlrd"
export REDIS_ADDR="localhost:6379"
export LOG_LEVEL="info"
```

### 4. 运行服务

```bash
# 方式1：使用测试脚本
bash scripts/run_test.sh

# 方式2：直接运行
go run cmd/wisefido-alarm/main.go

# 方式3：编译后运行
go build -o wisefido-alarm cmd/wisefido-alarm/main.go
./wisefido-alarm
```

## 📊 服务行为

### 轮询模式
- 每 **10秒** 轮询一次所有卡片
- 批量评估（每批 **10** 张卡片）
- 读取 Redis 实时数据缓存
- 评估报警事件（事件1-4）
- 写入报警事件到 PostgreSQL
- 更新报警缓存到 Redis

### 日志输出

```json
{"level":"info","msg":"Starting alarm service","tenant_id":"your-tenant-id"}
{"level":"info","msg":"Cache consumer started","tenant_id":"your-tenant-id","poll_interval":10}
{"level":"debug","msg":"Evaluating cards","card_count":10}
{"level":"info","msg":"Alarm event created","event_id":"...","event_type":"Fall","alarm_level":"ALERT"}
```

## ✅ 验证服务运行

### 1. 检查日志
- 确认服务启动成功
- 确认定期轮询（每10秒）
- 确认卡片评估过程
- 确认报警事件创建（如果有）

### 2. 检查数据库

```sql
-- 检查报警事件
SELECT 
    event_id,
    event_type,
    alarm_level,
    alarm_status,
    triggered_at,
    device_id
FROM alarm_events
ORDER BY triggered_at DESC
LIMIT 10;
```

### 3. 检查 Redis 缓存

```bash
# 检查报警缓存
redis-cli KEYS "vital-focus:card:*:alarms"
redis-cli GET "vital-focus:card:{card_id}:alarms"

# 检查状态缓存
redis-cli KEYS "alarm:state:*"
```

## 📁 项目结构

```
wisefido-alarm/
├── cmd/
│   └── wisefido-alarm/
│       └── main.go              # 主程序入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置加载
│   ├── consumer/
│   │   ├── cache_manager.go     # Redis 缓存管理器
│   │   ├── cache_consumer.go    # 缓存消费者（轮询模式）
│   │   └── state_manager.go     # 报警状态管理器
│   ├── evaluator/
│   │   ├── evaluator.go         # 主评估器
│   │   ├── alarm_event_builder.go # 报警事件构建器
│   │   ├── event1_bed_fall.go  # 事件1：床上跌落检测
│   │   ├── event2_sleepad_reliability.go # 事件2：Sleepad可靠性判断
│   │   ├── event3_bathroom_fall.go # 事件3：Bathroom可疑跌倒检测
│   │   └── event4_sudden_disappear.go # 事件4：人突然消失
│   ├── models/
│   │   ├── alarm_event.go       # 报警事件模型
│   │   ├── alarm_config.go      # 报警配置模型
│   │   └── realtime_data.go     # 实时数据模型
│   ├── repository/
│   │   ├── alarm_cloud.go       # 报警策略仓库
│   │   ├── alarm_device.go      # 设备报警配置仓库
│   │   ├── alarm_events.go      # 报警事件仓库
│   │   ├── card.go              # 卡片仓库
│   │   ├── device.go            # 设备仓库
│   │   └── room.go              # 房间仓库
│   └── service/
│       └── alarm.go             # 报警服务（整合各层）
├── scripts/
│   ├── verify_setup.sh          # 环境验证脚本
│   └── run_test.sh              # 运行测试脚本
└── docs/
    ├── QUICK_START.md           # 快速启动指南
    ├── VERIFY.md                # 详细验证指南
    ├── TESTING_GUIDE.md         # 测试指南
    ├── RUN_TEST.md              # 运行测试指南
    └── IMPLEMENTATION_SUMMARY.md # 实现总结
```

## 📝 当前状态

### ✅ 已完成
- Repository 层（数据库操作）
- Consumer 层（Redis 缓存读取）
- Evaluator 层（基础框架）
- Service 层（整合各层）
- Main 入口
- 报警事件写入功能

### ⏳ 待完善
- 事件1-4的完整评估逻辑（当前为简化版本）
- 报警去重逻辑（在事件评估器中调用）
- 性能优化（从 PostgreSQL 查询卡片，而非扫描 Redis 键）

## 🔗 相关文档

- `QUICK_START.md` - 快速启动指南
- `VERIFY.md` - 详细验证指南
- `TESTING_GUIDE.md` - 测试指南
- `RUN_TEST.md` - 运行测试指南
- `IMPLEMENTATION_SUMMARY.md` - 实现总结
- `ALARM_EVENT_WRITE.md` - 报警事件写入说明
- `REPOSITORY_LAYER_SUMMARY.md` - Repository 层总结
- `REQUIREMENTS_ANALYSIS.md` - 需求分析

## 🐛 问题排查

参考 `RUN_TEST.md` 中的问题排查部分。

## 📄 许可证

（根据项目许可证）

