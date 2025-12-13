# OwlBack - v1.5 后台服务

## 项目概述

OwlBack 是 WiseFido 平台 v1.5 版本的后台服务集合，基于 Go 语言开发，使用 PostgreSQL + TimescaleDB 作为数据库，通过 MQTT 与 IoT 设备通信。

## 架构设计

### 服务列表

1. **wisefido-radar** - 雷达服务（MQTT + OTA，已实现）
2. **wisefido-sleepace** - 睡眠垫服务（**v1.5 格式，统一数据流**）
   - Sleepad → Sleepace 厂家服务（第三方，有独立数据库和 HTTP API）
   - Sleepace 厂家服务 → MQTT Broker（厂家提供）
   - wisefido-sleepace 服务 → 订阅厂家 MQTT → 发布到 Redis Streams
   - **注意**：数据发布到 Redis Streams，由 wisefido-data-transformer 统一处理
3. **wisefido-data-transformer** - 数据转换服务（SNOMED CT映射、数据标准化）
4. **wisefido-sensor-fusion** - 传感器融合服务（多传感器数据融合）
5. **wisefido-alarm** - 报警处理服务（规则评估 + AI智能评估）
6. **wisefido-card-aggregator** - 卡片聚合服务（VitalFocusCard聚合）
7. **wisefido-data** - API服务（HTTP API接口）

### 数据流

```
Radar 设备 → MQTT Broker（直接）
    └─ wisefido-radar → Redis Streams (radar:data:stream)

Sleepace 设备 → Sleepace 厂家服务（第三方，独立 DB + HTTP API + MQTT）
    └─ Sleepace 厂家 MQTT Broker
        └─ wisefido-sleepace 服务（v1.5 格式）
            └─→ Redis Streams (sleepace:data:stream)

Redis Streams → wisefido-data-transformer
    ├─ 消费 radar:data:stream
    ├─ 消费 sleepace:data:stream
    ├─ 数据标准化（SNOMED CT映射）
    └─→ PostgreSQL TimescaleDB (iot_timeseries)
    └─→ Redis Streams (iot:data:stream)

Redis Streams (iot:data:stream) → wisefido-sensor-fusion
    ├─ 消费标准化数据
    ├─ 多传感器融合（HR/RR 优先 Sleepace，姿态合并 Radar）
    └─→ Redis (vital-focus:card:{card_id}:realtime)

Redis → wisefido-alarm
    ├─ 传统规则评估
    ├─ AI智能评估
    └─→ PostgreSQL (alarm_events) + Redis (alarms缓存)

Redis → wisefido-card-aggregator
    ├─ 聚合卡片数据
    └─→ Redis (vital-focus:card:{card_id}:full)

Redis → wisefido-data (API)
    └─→ HTTP Response (前端)
```

## 技术栈

- **语言**: Go 1.21+
- **数据库**: PostgreSQL 15+ with TimescaleDB
- **缓存**: Redis 7+
- **消息队列**: MQTT (Eclipse Mosquitto) + Redis Streams
- **API框架**: Gin / Echo (待定)

## 项目结构

```
owlBack/
├── owl-common/              # 共享库
│   ├── database/           # 数据库连接和工具
│   ├── redis/              # Redis客户端
│   ├── mqtt/               # MQTT客户端
│   └── logger/             # 日志工具
├── wisefido-radar/          # 雷达服务
├── wisefido-sleepace/       # 睡眠垫服务
├── wisefido-data-transformer/  # 数据转换服务
├── wisefido-sensor-fusion/  # 传感器融合服务
├── wisefido-alarm/          # 报警处理服务
├── wisefido-card-aggregator/ # 卡片聚合服务
├── wisefido-data/           # API服务
├── docker-compose.yml       # Docker Compose配置
└── README.md               # 本文档
```

## 数据库

数据库设计参考：`../owlRD/db/`

**重要**：owlRD是项目的核心参考，包含所有数据库表结构定义和需求文档。

### 主要表结构

所有表结构定义在 `owlRD/db/` 目录下，按顺序执行：

1. `00_extensions.sql` - TimescaleDB扩展
2. `01_tenants.sql` - 租户表
3. `02_roles.sql` - 角色表
4. `03_role_permissions.sql` - 权限表
5. `04_users.sql` - 用户表
6. `05_units.sql` - 单元表
7. `06_rooms.sql` - 房间表
8. `07_beds.sql` - 床位表
9. `08_residents.sql` - 住户表
10. `09_resident_phi.sql` - 住户PHI表
11. `10_resident_contacts.sql` - 住户联系人表
12. `11_resident_caregivers.sql` - 护工分配表
13. `12_devices.sql` - 设备表
14. `13_device_store.sql` - 设备库存表
15. `14_iot_timeseries.sql` - 时序数据表（TimescaleDB超表）
16. `15_alarm_events.sql` - 报警事件表
17. `16_alarm_device.sql` - 设备报警配置表
18. `17_alarm_cloud.sql` - 云端报警策略表
19. `18_config_versions.sql` - 配置版本表
20. `19_snomed_mapping.sql` - SNOMED CT映射表
21. `20_service_levels.sql` - 服务级别表
22. `21_cards.sql` - 卡片表
23. `22_tags_catalog.sql` - 标签目录表
24. `23_rounds.sql` - 巡房表
25. `24_round_details.sql` - 巡房详情表
26. `25_integrity_constraints.sql` - 完整性约束补充（跨租户验证、关系一致性验证）

### 快速初始化

使用提供的脚本：
```bash
./scripts/init-db.sh
```

或手动执行：
```bash
cd ../owlRD/db
for file in $(ls -1 *.sql | sort); do
  psql -U postgres -d owlrd -f "$file"
done
```

## 开发指南

### 环境要求

- Go 1.21+
- PostgreSQL 15+ with TimescaleDB extension
- Redis 7+
- MQTT Broker (Eclipse Mosquitto)

### 本地开发

1. 启动依赖服务（PostgreSQL, Redis, MQTT）
   ```bash
   docker-compose up -d postgresql redis mqtt
   ```

2. 初始化数据库
   ```bash
   # 执行owlRD/db/目录下的SQL文件（按顺序）
   cd ../owlRD/db
   for file in $(ls -1 *.sql | sort); do
     echo "Executing $file..."
     psql -U postgres -d owlrd -f "$file"
   done
   ```
   
   或者使用提供的初始化脚本：
   ```bash
   ./scripts/init-db.sh
   ```

3. 运行服务
   ```bash
   # 每个服务独立运行
   cd wisefido-radar && go run cmd/wisefido-radar/main.go
   ```

### 构建

```bash
# 构建所有服务
./build.sh

# 构建单个服务
cd wisefido-radar && go build -o bin/wisefido-radar cmd/wisefido-radar/main.go
```

## 部署

### Docker部署

```bash
docker-compose up -d
```

### 配置

每个服务使用独立的配置文件 `config/config.yaml`，支持环境变量覆盖。

## 项目结构

```
project/
├── owlRD/              # v1.5 需求及数据库设计（核心参考）
│   ├── db/            # 数据库表结构SQL文件
│   ├── docs/          # 需求文档和技术文档
│   └── README.md      # owlRD项目说明
├── owlFront/          # v1.5 前端项目
├── owlBack/           # v1.5 后台服务（当前项目）
└── wisefido-backend/  # v1.0 后端（参考）
```

## 相关项目

- [owlRD](../owlRD) - v1.5 需求及数据库设计（**核心参考**）
  - 数据库表结构：`owlRD/db/*.sql`
  - 需求文档：`owlRD/docs/*.md`
  - FHIR转换指南：`owlRD/docs/06_FHIR_Simple_Conversion_Guide.md`
  - 卡片创建规则：`owlRD/docs/20_Card_Creation_Rules_Final.md`
- [owlFront](../owlFront) - v1.5 前端项目
- [wisefido-backend](../wisefido-backend) - v1.0 后端（参考）

## 迁移计划

从 v1.0 (wisefido-backend) 迁移到 v1.5 (owlBack)：

1. ✅ 项目结构创建
2. ⏳ 共享库开发
3. ⏳ 服务逐个迁移
4. ⏳ 数据库迁移（MySQL → PostgreSQL）
5. ⏳ 集成测试

## 许可证

私有项目

# owlBack
