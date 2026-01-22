# OwlBack Services Startup Guide

## 服务说明

### 1. wisefido-data 服务
- **功能**: 提供 HTTP API，处理前端请求
- **端口**: **8080**（必需，HTTP 服务器）
- **职责**: 
  - 提供 REST API 接口
  - 从 Redis 读取卡片缓存数据（由 wisefido-card-aggregator 写入）
  - **不负责**卡片生成和更新

### 2. wisefido-card-aggregator 服务
- **功能**: 自动生成和管理 Cards（卡片）
- **端口**: **不需要端口**（纯后台服务，只连接数据库和 Redis）
- **职责**:
  - 根据 unit/room/bed/device/resident 的绑定关系自动生成 cards
  - 定期检查并更新卡片（轮询模式）
  - 将卡片数据聚合后写入 Redis 缓存

## 服务关系

### 当前实现（轮询模式）
```
wisefido-data (独立运行)
    ↓ (读取)
Redis Cache (vital-focus:card:*:full)
    ↑ (写入)
wisefido-card-aggregator (独立运行，每60分钟轮询一次)
```

**重要**: 
- `wisefido-data` **不会触发** `wisefido-card-aggregator`
- `wisefido-card-aggregator` 使用**轮询模式**，每60分钟自动检查并更新卡片
- 两个服务是**独立运行**的，通过 Redis 共享数据

### 未来实现（事件驱动模式）
```
wisefido-data (发布事件)
    ↓ (发布到 Redis Streams)
Redis Streams (card:events)
    ↑ (消费事件)
wisefido-card-aggregator (实时响应事件)
```

## 启动和停止服务

### 方式1: 使用统一脚本（推荐）

#### 启动服务
```bash
cd /Users/sady3721/project/owlBack

# 启动两个服务（带日志输出）
./start_all_services.sh

# 清理旧日志并启动
./start_all_services.sh --clean
```

#### 停止服务
```bash
cd /Users/sady3721/project/owlBack

# 停止所有服务
./stop_all_services.sh
```

**注意**: 如果使用 `./start_all_services.sh` 启动，按 `Ctrl+C` 也会自动停止所有服务。

**日志文件位置**:
- `wisefido-data`: `/tmp/owlBack_logs/wisefido-data.log`
- `wisefido-card-aggregator`: `/tmp/owlBack_logs/wisefido-card-aggregator.log`
- 合并日志: `/tmp/owlBack_logs/combined.log`

### 方式2: 分别启动

#### 启动 wisefido-data
```bash
cd /Users/sady3721/project/owlBack/wisefido-data
go run cmd/wisefido-data/main.go
```

#### 启动 wisefido-card-aggregator
```bash
cd /Users/sady3721/project/owlBack/wisefido-card-aggregator

# 设置环境变量
export TENANT_ID="bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c"
export DB_HOST="localhost"
export DB_NAME="owlrd"
export REDIS_ADDR="localhost:6379"
export CARD_TRIGGER_MODE="polling"
export CARD_POLLING_INTERVAL="3600"
export LOG_LEVEL="info"

# 启动服务
go run cmd/wisefido-card-aggregator/main.go
```

或使用启动脚本：
```bash
cd /Users/sady3721/project/owlBack/wisefido-card-aggregator
./start_service.sh
```

## 环境变量

### 通用环境变量
```bash
export DB_HOST="localhost"
export DB_NAME="owlrd"
export REDIS_ADDR="localhost:6379"
export LOG_LEVEL="info"
```

### wisefido-card-aggregator 专用
```bash
export TENANT_ID="bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c"
export CARD_TRIGGER_MODE="polling"  # 或 "events"
export CARD_POLLING_INTERVAL="3600"  # 秒（默认60分钟）
export CARD_AGGREGATION_ENABLED="true"
export CARD_AGGREGATION_INTERVAL="10"  # 秒
```

## 查看日志

### 实时查看 wisefido-data 日志
```bash
tail -f /tmp/owlBack_logs/wisefido-data.log
```

### 实时查看 wisefido-card-aggregator 日志
```bash
tail -f /tmp/owlBack_logs/wisefido-card-aggregator.log
```

### 查看卡片统计信息
```bash
grep "Card Check/Update Statistics" /tmp/owlBack_logs/wisefido-card-aggregator.log
```

## 停止服务

### 方式1: 使用统一停止脚本（推荐）

```bash
cd /Users/sady3721/project/owlBack
./stop_all_services.sh
```

### 方式2: 使用 Ctrl+C（如果使用统一启动脚本）

如果使用 `./start_all_services.sh` 启动，按 `Ctrl+C` 会自动停止所有服务。

### 方式3: 手动停止

```bash
pkill -f "go run.*wisefido-data"
pkill -f "go run.*wisefido-card-aggregator"
```

## 常见问题

### Q: wisefido-data 启动后，卡片没有更新？
**A**: `wisefido-data` 不负责卡片更新。需要启动 `wisefido-card-aggregator` 服务，它会每60分钟自动检查并更新卡片。

### Q: 卡片统计信息在哪里查看？
**A**: 在 `wisefido-card-aggregator` 服务的标准输出中，会显示：
```
=== Card Check/Update Statistics ===
Original card count: 4
Updated card count: 0 (deleted: 0, created: 0, content updated: 0)
Unchanged cards: 4
Final card count: 4
Units processed: 4 (success: 4, failed: 0)
===================================
```

### Q: 如何让卡片实时更新？
**A**: 目前使用轮询模式（每60分钟）。未来可以实现事件驱动模式，需要：
1. `wisefido-data` 在数据变化时发布事件到 Redis Streams
2. `wisefido-card-aggregator` 配置为 `CARD_TRIGGER_MODE=events` 模式

