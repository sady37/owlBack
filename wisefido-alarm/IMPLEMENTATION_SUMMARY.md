# wisefido-alarm 实现总结

## ✅ 已完成

### 1. Repository 层
- ✅ `alarm_cloud.go` - 报警策略仓库
- ✅ `alarm_device.go` - 设备报警配置仓库
- ✅ `alarm_events.go` - 报警事件仓库
- ✅ `card.go` - 卡片仓库（包含 `GetAllCards` 方法）
- ✅ `device.go` - 设备仓库
- ✅ `room.go` - 房间仓库（包含 `IsBathroom` 方法）

### 2. Consumer 层
- ✅ `cache_manager.go` - Redis 缓存管理器
  - `GetRealtimeData` - 读取实时数据
  - `UpdateAlarmCache` - 更新报警缓存
  - `GetAllCardIDs` - 获取所有卡片ID（通过扫描 Redis 键）
- ✅ `state_manager.go` - 报警状态管理器
  - `SetState` / `GetState` / `DeleteState` - 状态管理
  - `Event1State` / `Event2State` / `Event3State` / `Event4State` - 事件状态结构
- ✅ `cache_consumer.go` - 缓存消费者（轮询模式）
  - `Start` - 启动消费者（定期轮询）
  - `evaluateAllCards` - 评估所有卡片
  - `evaluateBatch` - 批量评估卡片

### 3. Evaluator 层
- ✅ `evaluator.go` - 主评估器（整合所有事件评估器）
- ✅ `event1_bed_fall.go` - 事件1：床上跌落检测（基础框架）
- ✅ `event2_sleepad_reliability.go` - 事件2：Sleepad可靠性判断（基础框架）
- ✅ `event3_bathroom_fall.go` - 事件3：Bathroom可疑跌倒检测（基础框架）
- ✅ `event4_sudden_disappear.go` - 事件4：雷达检测到人突然消失（基础框架）

### 4. Service 层
- ✅ `alarm.go` - 报警服务（整合各层）
  - `NewAlarmService` - 创建服务
  - `Start` - 启动服务
  - `Stop` - 停止服务

### 5. Main 入口
- ✅ `cmd/wisefido-alarm/main.go` - 主程序入口
  - 配置加载
  - 日志初始化
  - 服务启动
  - 优雅关闭

## 📝 待完善

### 1. Evaluator 层的事件评估逻辑

当前实现是**基础框架**，事件1-4的评估逻辑需要进一步完善：

#### 事件1：床上跌落检测
- ⚠️ 需要实现完整的状态管理（lying基线、离床时间、track_id状态）
- ⚠️ 需要实现定时器（T0+5秒、T0+60秒、T0+120秒）
- ⚠️ 需要实现退出条件检查（持续检查）
- ⚠️ 需要实现危险情况检测（track_id突然消失）

#### 事件2：Sleepad可靠性判断
- ⚠️ 需要实现核查1（前置条件检查）
- ⚠️ 需要实现分支判断（分支A和分支B）
- ⚠️ 需要实现核查2和核查3

#### 事件3：Bathroom可疑跌倒检测
- ⚠️ 需要实现站立状态检测
- ⚠️ 需要实现位置变化检测（位置变化小于10cm，超过10分钟）
- ⚠️ 需要实现单人检测（雷达检测范围内仅1人）

#### 事件4：雷达检测到人突然消失
- ⚠️ 需要实现 track_id 历史状态管理
- ⚠️ 需要实现质心降低检测（高度降低超过60cm）
- ⚠️ 需要实现5分钟无活动检测

### 2. 报警事件创建 ✅ **已完成**

- ✅ 在 `Evaluator.Evaluate` 中实现报警事件的创建逻辑
- ✅ 调用 `alarmEventsRepo.CreateAlarmEvent` 写入数据库
- ✅ 实现报警去重检查（`CheckDuplicate` 方法）
- ✅ 实现报警事件构建器（`AlarmEventBuilder`）
- ✅ 实现触发数据序列化（`BuildTriggerData`）
- ✅ 更新报警缓存（只更新活跃的报警）

### 3. 配置和优化

- ⚠️ 添加更多配置项（事件1-4的阈值、时间窗口等）
- ⚠️ 优化 `GetAllCardIDs` 方法（当前通过扫描 Redis 键，效率较低，建议改为从 PostgreSQL 查询）
- ⚠️ 添加指标监控（处理速度、报警数量等）

## 🚀 使用方式

### 1. 环境变量

```bash
# 必需
export TENANT_ID="your-tenant-id"

# 数据库配置
export DB_HOST="localhost"
export DB_PORT=5432
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="owlrd"
export DB_SSLMODE="disable"

# Redis 配置
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD=""

# 日志配置
export LOG_LEVEL="info"
export LOG_FORMAT="json"
```

### 2. 运行服务

```bash
cd /Users/sady3721/project/owlBack/wisefido-alarm
go run cmd/wisefido-alarm/main.go
```

### 3. 编译

```bash
cd /Users/sady3721/project/owlBack/wisefido-alarm
go build -o wisefido-alarm cmd/wisefido-alarm/main.go
```

## 📊 数据流

```
PostgreSQL (cards 表)
    ↓
CacheConsumer.GetAllCards()
    ↓
CacheManager.GetRealtimeData() (从 Redis 读取)
    ↓
Evaluator.Evaluate() (评估事件1-4)
    ↓
CacheManager.UpdateAlarmCache() (更新 Redis 报警缓存)
    ↓
(待实现) AlarmEventsRepository.CreateAlarmEvent() (写入 PostgreSQL)
```

## 🔗 相关文档

- `owlBack/docs/alarm_rule.md` - 报警规则详细说明
- `owlBack/wisefido-alarm/REPOSITORY_LAYER_SUMMARY.md` - Repository 层总结
- `owlBack/wisefido-alarm/REQUIREMENTS_ANALYSIS.md` - 需求分析
- `owlBack/wisefido-alarm/ROOM_NAME_USAGE.md` - room_name 使用说明

