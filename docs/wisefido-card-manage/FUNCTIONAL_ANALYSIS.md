# wisefido-card-manage 功能分析

## 📋 服务概述

`wisefido-card-manage` 是一个独立的微服务，专门负责**卡片创建和维护**功能。它是从 `wisefido-card-aggregator` 中拆分出来的，专注于低频的卡片管理操作。

## 🏗️ 服务架构

### 分层架构

```
┌─────────────────────────────────────┐
│      HTTP API Layer                 │
│  (handler.go, router.go)            │
│  - POST /api/v1/cards/create        │
│  - POST /api/v1/cards/create-all    │
│  - GET  /health                     │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      Service Layer                  │
│  (card_service.go)                  │
│  - CreateCardsForUnit()             │
│  - CreateAllCards()                 │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      Repository Layer               │
│  (card.go, card_info.go, routing.go)│
│  - 数据库访问抽象                    │
│  - 实现 card.RepositoryInterface    │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      Core Logic                     │
│  (owl-common/card/creator.go)       │
│  - CardCreator.CreateCardsForUnit() │
│  - 卡片创建规则和逻辑                │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      PostgreSQL Database             │
│  - cards 表                          │
│  - devices 表                        │
│  - residents 表                      │
│  - beds 表                           │
│  - units 表                          │
└─────────────────────────────────────┘
```

## 🎯 核心功能

### 1. 卡片创建（Card Creation）

#### 1.1 为指定单元创建卡片
- **方法**：`CardService.CreateCardsForUnit(ctx, tenantID, unitID)`
- **API 端点**：`POST /api/v1/cards/create`
- **功能**：为指定的 `unit_id` 创建或更新所有相关卡片
- **触发场景**：
  - 设备绑定/解绑到床位或房间
  - 住户绑定/解绑到床位
  - 单元信息更新
  - 由 `wisefido-data` 服务通过 HTTP API 调用

#### 1.2 全量创建卡片
- **方法**：`CardService.CreateAllCards(ctx)`
- **API 端点**：`POST /api/v1/cards/create-all`
- **功能**：为所有单元创建或更新卡片
- **触发场景**：
  - **服务启动时**：自动执行一次全量创建
  - **定时任务**：通过 crontab 每天早上 8:00 执行（保底机制）

### 2. 卡片创建规则

卡片创建遵循以下规则（由 `owl-common/card/creator.go` 实现）：

#### 场景 A：单元只有 1 个 ActiveBed
- 创建 1 个 `ActiveBed` 卡片
- 绑定该床上的所有设备
- 绑定未绑床的设备（`bed_id` 为 NULL）

#### 场景 B：单元有多个 ActiveBed（≥2）
- 为每个 ActiveBed 创建 1 个 `ActiveBed` 卡片
- 每个卡片只绑定对应床上的设备
- 未绑床的设备不绑定到任何 ActiveBed 卡片

#### 场景 C：单元没有 ActiveBed
- 创建 1 个 `Location` 卡片
- 绑定该单元下所有未绑床的设备

### 3. 卡片更新优化

- **智能比较**：创建前比较现有卡片与预期卡片
- **增量更新**：只更新变化的卡片，保留未变化的卡片
- **统计信息**：返回详细的更新统计（创建、更新、删除、未变化数量）

## 📡 API 接口

### 1. POST /api/v1/cards/create

为指定单元创建/更新卡片。

**请求体**：
```json
{
  "tenant_id": "string",
  "unit_id": "string"
}
```

**响应**：
```json
{
  "success": true,
  "stats": {
    "existing_count": 0,
    "created_count": 2,
    "updated_count": 1,
    "deleted_count": 0,
    "unchanged_count": 0
  }
}
```

**使用场景**：
- `wisefido-data` 服务在设备/住户/单元变化时调用
- 实时响应数据变化

### 2. POST /api/v1/cards/create-all

为所有单元创建/更新卡片。

**请求体**：
```json
{
  "tenant_id": "string"
}
```

**响应**：
```json
{
  "success": true,
  "message": "All cards created/updated successfully"
}
```

**使用场景**：
- 服务启动时自动调用
- crontab 定时任务（每天早上 8:00）
- 手动触发全量更新

### 3. GET /health

健康检查端点。

**响应**：
```
OK
```

## 🔄 数据流

### 实时更新流程（API 触发）

```
wisefido-data (设备/住户/单元变化)
    │
    ├─ 更新数据库 (devices/residents/beds/units 表)
    │
    └─ HTTP POST /api/v1/cards/create
       │
       ▼
wisefido-card-manage
    │
    ├─ CardService.CreateCardsForUnit()
    │
    ├─ CardCreator.CreateCardsForUnit()
    │   ├─ 查询单元信息
    │   ├─ 查询床位信息
    │   ├─ 查询设备信息
    │   ├─ 查询住户信息
    │   ├─ 计算预期卡片
    │   ├─ 比较现有卡片
    │   ├─ 创建新卡片
    │   ├─ 更新变化卡片
    │   └─ 删除多余卡片
    │
    └─ 更新 PostgreSQL cards 表
```

### 全量更新流程（启动时/定时任务）

```
服务启动 / crontab (每天 8:00)
    │
    ▼
wisefido-card-manage
    │
    ├─ CardService.CreateAllCards()
    │   ├─ 获取所有 unit IDs
    │   └─ 循环调用 CreateCardsForUnit()
    │
    └─ 更新 PostgreSQL cards 表
```

## 📦 依赖关系

### 内部依赖

1. **Repository Layer** (`internal/repository/`)
   - `CardRepository`：实现 `card.RepositoryInterface`
   - 提供数据库访问方法：
     - `GetUnitInfo()` - 获取单元信息
     - `GetActiveBedsByUnit()` - 获取单元下的床位
     - `GetDevicesByBed()` - 获取床位绑定的设备
     - `GetUnboundDevicesByUnit()` - 获取未绑床的设备
     - `GetResidentByBed()` - 获取床位绑定的住户
     - `GetResidentsByUnit()` - 获取单元下的住户
     - `CreateCard()` - 创建卡片
     - `UpdateCard()` - 更新卡片
     - `DeleteCard()` - 删除卡片
     - `GetAllUnits()` - 获取所有单元 ID

2. **Service Layer** (`internal/service/`)
   - `CardService`：业务逻辑封装
   - 调用 `owl-common/card` 包的 `CardCreator`

### 外部依赖

1. **owl-common/card 包**
   - `CardCreator`：核心卡片创建逻辑
   - `RepositoryInterface`：Repository 接口定义
   - `CardUpdateStats`：更新统计信息

2. **owl-common/database 包**
   - PostgreSQL 数据库连接

3. **owl-common/logger 包**
   - 日志记录

### 被依赖关系

1. **wisefido-data 服务**
   - 通过 HTTP API 调用 `POST /api/v1/cards/create`
   - 在设备/住户/单元变化时触发卡片更新

## ⚙️ 配置项

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `DB_HOST` | `localhost` | 数据库主机 |
| `DB_USER` | `postgres` | 数据库用户 |
| `DB_PASSWORD` | `postgres` | 数据库密码 |
| `DB_NAME` | `owlrd` | 数据库名称 |
| `DB_SSLMODE` | `disable` | SSL 模式 |
| `REDIS_ADDR` | `127.0.0.1:6379` | Redis 地址（当前未使用） |
| `CARD_MANAGE_PORT` | `8082` | HTTP 服务器端口 |
| `TENANT_ID` | `""` | 租户 ID（必需） |
| `CARD_TRIGGER_MODE` | `api` | 触发模式（当前只支持 `api`） |
| `LOG_LEVEL` | `info` | 日志级别 |
| `LOG_FORMAT` | `json` | 日志格式 |

### 配置结构

```go
type Config struct {
    Database    DatabaseConfig  // 数据库配置
    Redis       RedisConfig     // Redis 配置（预留）
    Server      ServerConfig    // HTTP 服务器配置
    CardManage  CardManageConfig // 卡片管理配置
    Log         LogConfig       // 日志配置
}
```

## 🚀 启动流程

1. **加载配置**：从环境变量加载配置
2. **初始化日志**：创建 logger
3. **连接数据库**：建立 PostgreSQL 连接
4. **创建服务**：
   - 创建 `CardRepository`
   - 创建 `CardCreator`（使用 `owl-common/card` 包）
   - 创建 `CardService`
5. **启动时全量创建**：调用 `CreateAllCards()` 为所有单元创建卡片
6. **启动 HTTP 服务器**：监听 `:8082` 端口
7. **等待信号**：监听系统信号（SIGTERM, SIGINT）进行优雅关闭

## 📊 卡片数据结构

### cards 表结构（关键字段）

- `card_id`：卡片唯一标识
- `tenant_id`：租户 ID
- `card_type`：卡片类型（`ActiveBed` 或 `Location`）
- `bed_id`：床位 ID（ActiveBed 卡片）
- `unit_id`：单元 ID（Location 卡片）
- `card_name`：卡片名称
- `card_address`：卡片地址
- `devices`：JSONB - 绑定的设备列表
- `residents`：JSONB - 关联的住户列表
- `routing_alarm_user_ids`：JSONB - 报警路由用户 ID 列表
- `routing_alarm_tags`：JSONB - 报警路由标签列表

## 🔍 关键特性

### 1. 职责单一
- **专注卡片管理**：只负责卡片的创建和维护
- **不处理数据聚合**：数据聚合由 `wisefido-card-aggregator` 负责

### 2. 实时响应
- **API 触发**：`wisefido-data` 服务变化时立即更新卡片
- **同步调用**：确保卡片更新与数据变化同步

### 3. 保底机制
- **启动时创建**：服务启动时自动创建所有卡片
- **定时任务**：crontab 每天早上 8:00 执行全量更新

### 4. 智能更新
- **增量更新**：只更新变化的卡片
- **保留 card_id**：未变化的卡片保留原有 `card_id`
- **统计信息**：返回详细的更新统计

## 🔗 与其他服务的关系

### wisefido-data（API 服务）

#### 调用方式
- **HTTP 客户端**：`CardManageClient`（`wisefido-data/internal/service/card_manage_client.go`）
- **调用接口**：`POST /api/v1/cards/create`
- **超时设置**：30 秒
- **配置项**：`CARD_MANAGE_INTERNAL_API_BASE_URL`（默认：`http://localhost:8082`）

#### 调用时机
1. **设备更新**（`device_service.go`）：
   - 设备绑定/解绑到床位或房间（`UpdateBoundRoomID` 或 `UpdateBoundBedID`）
   - 设备监控状态变化（`monitoring_enabled`）
   - 调用：`cardManageClient.CreateCardsForUnit(tenantID, unitID)`

2. **住户更新**（`resident_service.go`）：
   - 住户绑定/解绑到床位
   - 住户护理人员分配变化
   - 调用：`cardManageClient.CreateCardsForUnit(tenantID, unitID)`

3. **单元更新**（`unit_service.go`）：
   - 单元信息更新
   - 调用：`cardManageClient.CreateCardsForUnit(tenantID, unitID)`

#### 错误处理
- **非阻塞**：卡片更新失败不会影响主业务逻辑
- **日志记录**：失败时记录警告日志，不中断 API 响应
- **容错机制**：如果 `cardManageClient` 为 `nil`，跳过卡片更新

### wisefido-card-aggregator（数据聚合服务）
- **关系**：独立运行，不直接通信
- **数据共享**：通过 PostgreSQL `cards` 表共享数据
- **职责分离**：
  - `wisefido-card-manage`：卡片创建和维护
  - `wisefido-card-aggregator`：卡片数据聚合和融合

## 📝 使用示例

### 1. 通过 API 创建指定单元的卡片

```bash
curl -X POST http://localhost:8082/api/v1/cards/create \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "your-tenant-id",
    "unit_id": "unit-123"
  }'
```

### 2. 通过 API 全量创建卡片

```bash
curl -X POST http://localhost:8082/api/v1/cards/create-all \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "your-tenant-id"
  }'
```

### 3. 通过脚本全量创建（crontab）

```bash
# 设置环境变量
export CARD_MANAGE_URL="http://localhost:8082"
export TENANT_ID="your-tenant-id"

# 执行脚本
/path/to/wisefido-card-manage/scripts/create-all-cards.sh
```

### 4. 配置 crontab（每天早上 8:00）

```bash
# 编辑 crontab
crontab -e

# 添加以下行
0 8 * * * /path/to/wisefido-card-manage/scripts/create-all-cards.sh
```

## ⚠️ 注意事项

1. **TENANT_ID 必需**：服务启动时和全量创建时需要设置 `TENANT_ID` 环境变量
2. **数据库连接**：确保 PostgreSQL 数据库可访问
3. **服务依赖**：`wisefido-data` 服务需要知道 `wisefido-card-manage` 的地址（通过 `CARD_MANAGE_INTERNAL_API_BASE_URL` 配置）
4. **错误处理**：启动时创建卡片失败不会阻止服务启动，只记录警告日志

## 🎯 设计优势

1. **职责清晰**：卡片管理独立成服务，职责单一
2. **易于维护**：卡片创建逻辑集中在一个服务中
3. **可扩展性**：可以独立扩展和部署
4. **实时性**：通过 API 调用实现实时更新
5. **可靠性**：启动时和定时任务双重保底机制

