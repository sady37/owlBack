# wisefido-card-aggregator 服务

## 概述

`wisefido-card-aggregator` 是负责**自动生成和管理 Cards（卡片）**的后端服务。

### 主要功能

1. **Card 自动生成**
   - 根据 unit/room/bed/device/resident 的绑定关系自动生成 cards
   - 支持 ActiveBed Card 和 Location Card 两种类型
   - 自动维护 card 与设备、住户的关联关系

2. **数据聚合**（可选）
   - 聚合 card 的实时数据（vital signs）
   - 聚合 card 的报警数据
   - 缓存到 Redis 供前端快速访问

## 服务位置

```
/Users/sady3721/project/owlBack/wisefido-card-aggregator/
```

## Card 生成规则

### ActiveBed 判断条件

一个 bed 被认为是 ActiveBed，必须同时满足：
1. ✅ 该 bed 绑定的设备数量 > 0
2. ✅ **设备的 `monitoring_enabled = TRUE`**（关键条件）
3. ✅ 设备的 `status <> 'disabled'`

### Card 创建场景

- **场景 A**：unit 下只有 1 个 ActiveBed → 创建 1 个 ActiveBed Card
- **场景 B**：unit 下有多个 ActiveBed（≥2）→ 创建 N 个 ActiveBed Cards + 0 或 1 个 Location Card
- **场景 C**：unit 下无 ActiveBed → 创建 0 或 1 个 Location Card

## 运行方式

### 1. 环境变量配置

```bash
# 必填：租户 ID
export TENANT_ID="bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c"  # demo tenant ID

# 数据库配置（从环境变量读取，或使用默认值）
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_PASSWORD="postgres"
export DB_NAME="owlrd"

# Redis 配置（从环境变量读取，或使用默认值）
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
export REDIS_PASSWORD=""

# Card 生成触发模式
export CARD_TRIGGER_MODE="polling"  # 或 "events"
export CARD_POLLING_INTERVAL="3600"   # 轮询间隔（秒），默认 60 分钟（3600 秒）

# 数据聚合配置（可选）
export CARD_AGGREGATION_ENABLED="true"  # 是否启用数据聚合
export CARD_AGGREGATION_INTERVAL="10"   # 聚合间隔（秒），默认 10 秒
```

### 2. 启动服务

```bash
cd /Users/sady3721/project/owlBack/wisefido-card-aggregator
go run cmd/wisefido-card-aggregator/main.go
```

### 3. 服务行为

#### 轮询模式（polling，当前默认）

- **启动时**：立即执行一次全量 card 创建
- **定时轮询**：每 60 分钟（可配置，默认 3600 秒）全量重新创建所有 cards
- **特点**：
  - ✅ 简单可靠，不依赖外部事件
  - ✅ 作为保底机制：即使事件驱动失败，也能保证数据最终一致性
  - ⚠️ 延迟：最多等待一个轮询周期（如 60 分钟）才能反映变化
  - ⚠️ 资源消耗：即使没有变化，也会全量重新计算
- **设计理由**：
  - 有事件驱动触发更新机制（主要更新方式）
  - 即使触发更新失败，新入户第1小时没有生效也很正常
  - 配置更新需要时间，病人可能也没有这么快即时入住

#### 事件驱动模式（events，待完善）

- **启动时**：立即执行一次全量 card 创建
- **事件监听**：监听设备/住户/床位绑定关系变化事件（Redis Streams）
- **定时任务**：每天上午 9 点全量更新
- **特点**：
  - ✅ 实时响应：数据变化后立即更新
  - ✅ 资源高效：只在有变化时才执行
  - ⚠️ 需要 `wisefido-data` 服务发布事件（当前未实现）

## 日志输出示例

```
{"level":"info","msg":"Starting wisefido-card-aggregator service"}
{"level":"info","msg":"Starting card aggregator service","trigger_mode":"polling","aggregation_enabled":true}
{"level":"info","msg":"Starting polling mode","interval":"60s"}
{"level":"info","msg":"Starting to create cards for all units"}
{"level":"info","msg":"Found units to process","unit_count":3}
{"level":"info","msg":"Completed creating cards","success_count":3,"error_count":0}
```

## 验证 Card 生成

### 1. 检查数据库中的 cards

```sql
SELECT 
    c.card_id,
    c.card_type,
    c.card_name,
    c.card_address,
    c.bed_id,
    c.unit_id,
    c.resident_id,
    b.bed_name,
    u.unit_name,
    r.nickname AS resident_nickname
FROM cards c
LEFT JOIN beds b ON c.bed_id = b.bed_id
LEFT JOIN units u ON c.unit_id = u.unit_id
LEFT JOIN residents r ON c.resident_id = r.resident_id
WHERE c.tenant_id = 'bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c'
ORDER BY u.unit_name, b.bed_name;
```

### 2. 检查 Redis 缓存（如果启用了数据聚合）

```bash
# 检查完整卡片缓存
redis-cli KEYS "vital-focus:card:*:full"

# 查看特定卡片的缓存
redis-cli GET "vital-focus:card:{card_id}:full"
```

## 常见问题

### Q: 为什么没有 card 生成？

**A: 检查以下几点：**

1. **设备的 `monitoring_enabled` 是否为 `TRUE`**
   ```sql
   SELECT device_id, device_name, monitoring_enabled, bound_bed_id, bound_room_id
   FROM devices
   WHERE tenant_id = 'your-tenant-id';
   ```
   如果 `monitoring_enabled = FALSE`，需要启用：
   ```sql
   UPDATE devices
   SET monitoring_enabled = TRUE
   WHERE tenant_id = 'your-tenant-id'
     AND (bound_bed_id IS NOT NULL OR bound_room_id IS NOT NULL);
   ```

2. **服务是否正在运行**
   ```bash
   ps aux | grep wisefido-card-aggregator
   ```

3. **是否有满足条件的 ActiveBed**
   ```sql
   -- 检查 ActiveBed（满足 card 生成条件）
   SELECT DISTINCT
       b.bed_id,
       b.bed_name,
       r.room_name,
       u.unit_name,
       COUNT(DISTINCT d.device_id)::int AS bound_device_count
   FROM beds b
   INNER JOIN rooms r ON b.room_id = r.room_id
   INNER JOIN devices d ON d.bound_bed_id = b.bed_id
   INNER JOIN units u ON r.unit_id = u.unit_id
   WHERE b.tenant_id = 'your-tenant-id'
     AND d.monitoring_enabled = TRUE
     AND d.status <> 'disabled'
   GROUP BY b.bed_id, b.bed_name, r.room_name, u.unit_name
   HAVING COUNT(DISTINCT d.device_id) > 0;
   ```

### Q: Card 生成延迟多久？

**A: 轮询模式下，最多延迟一个轮询周期（默认 60 分钟）。**

- 服务启动时会立即执行一次全量创建
- 之后每 60 分钟自动重新创建（作为保底机制）
- 主要更新方式是通过事件驱动（实时响应）
- 可以通过 `CARD_POLLING_INTERVAL` 环境变量调整间隔

### Q: 如何手动触发 Card 生成？

**A: 重启服务即可立即触发一次全量创建。**

或者等待下次轮询（最多 60 秒）。

## 相关文档

- `docs/IMPLEMENTATION_SUMMARY.md` - 实现总结
- `docs/CARD_UPDATE_STRATEGIES.md` - Card 更新策略
- `docs/EVENT_TRIGGER_MECHANISM.md` - 事件触发机制
- `docs/DATA_AGGREGATION_IMPLEMENTATION.md` - 数据聚合实现

## 代码结构

```
wisefido-card-aggregator/
├── cmd/
│   └── wisefido-card-aggregator/
│       └── main.go                    # 服务入口
├── internal/
│   ├── aggregator/
│   │   ├── card_creator.go           # Card 创建逻辑
│   │   ├── data_aggregator.go        # 数据聚合逻辑
│   │   └── cache_manager.go          # 缓存管理
│   ├── repository/
│   │   └── card.go                   # 数据库访问
│   ├── service/
│   │   └── aggregator.go             # 服务生命周期管理
│   └── config/
│       └── config.go                 # 配置加载
└── docs/                              # 文档目录
```




Facility 型 Unit (非 publicSpace)
│
├─ SharedUnit (多人共享)
│  └─ ActiveBed Card
│     ├─ 床上有人 → resident nickname
│     └─ 床上无人 → "Unoccupied"
│
└─ 独享型 (非 SharedUnit)
   ├─ ActiveBed Card
   │  ├─ unit 内有人 → 第一个 resident nickname
   │  └─ unit 内无人 → "Unoccupied"
   │
   └─ UnitCard
      ├─ unit 内有人 → 第一个 resident nickname
      └─ unit 内无人 → "Unoccupied"