# wisefido-card-aggregator 卡片管理服务实现总结

## ✅ 已完成的工作

### 1. 项目结构 ✅
- ✅ 创建项目基础结构
- ✅ 配置管理 (`internal/config/config.go`)
- ✅ Repository 层 (`internal/repository/card.go`)
- ✅ 卡片创建逻辑 (`internal/aggregator/card_creator.go`)
- ✅ 服务主逻辑 (`internal/service/aggregator.go`)
- ✅ 主程序 (`cmd/wisefido-card-aggregator/main.go`)

### 2. 核心功能实现 ✅

#### 2.1 ActiveBed 判断逻辑 ✅
- ✅ 根据 `beds.bound_device_count > 0` 判断 ActiveBed
- ✅ 查询时只返回有激活监护设备的床位

#### 2.2 卡片创建场景 ✅
- ✅ **场景 A**：门牌下只有 1 个 ActiveBed
  - 创建 1 个 ActiveBed 卡片
  - 绑定该 bed 的设备 + 该 unit 下未绑床的设备
- ✅ **场景 B**：门牌下有多个 ActiveBed（≥2）
  - 为每个 ActiveBed 创建 1 个 ActiveBed 卡片
  - 如果有未绑床的设备，创建 1 个 UnitCard
- ✅ **场景 C**：门牌下无 ActiveBed
  - 如果有未绑床的设备，创建 1 个 UnitCard

#### 2.3 卡片名称计算 ✅
- ✅ **ActiveBed 卡片名称**：
  - 如果 bed 绑定了住户 → 使用该住户的 nickname
  - 如果 bed 未绑住户：
    - 非多人房间 → 显示该 unit 下第一个住户的 nickname
    - 多人房间 → 显示 'disable monitor'
- ✅ **UnitCard 名称**：
  - 优先级1：`is_public_space = TRUE` → `unit_name`
  - 优先级2：`is_multi_person_room = TRUE` → `unit_name`
  - 优先级3/4：该 unit 下有住户绑定 → 第一个住户的 nickname

#### 2.4 卡片地址计算 ✅
- ✅ 规则：`branch_tag + "-" + building + "-" + unit_name`
- ✅ 跳过空值或默认值 "-"
- ✅ 至少包含 `unit_name`（必填字段）

#### 2.5 设备绑定规则 ✅
- ✅ 按最小地址优先原则（床 > 门牌号）
- ✅ ActiveBed 卡片：绑定该 bed 的设备（direct）+ 该 unit 下未绑床的设备（indirect，仅场景 A）
- ✅ UnitCard：绑定该 unit 下未绑床的设备（indirect）

#### 2.6 卡片去重规则 ✅
- ✅ 删除旧卡片，创建新卡片（不保留历史记录）
- ✅ 基于 `unit_id` 删除，确保重新创建时数据一致

#### 2.7 触发机制 ✅
- ✅ 轮询模式（polling）：定时轮询所有 unit，重新创建卡片
- ✅ 默认轮询间隔：60 秒
- ⏳ 事件驱动模式（events）：待实现

### 3. Repository 层 ✅

#### 3.1 查询功能
- ✅ `GetActiveBedsByUnit` - 获取指定 unit 下的所有 ActiveBed
- ✅ `GetUnitInfo` - 获取 Unit 信息
- ✅ `GetDevicesByBed` - 获取指定 bed 绑定的设备
- ✅ `GetUnboundDevicesByUnit` - 获取指定 unit 下未绑床的设备
- ✅ `GetResidentByBed` - 获取指定 bed 绑定的住户
- ✅ `GetResidentsByUnit` - 获取指定 unit 下的所有住户
- ✅ `GetAllUnits` - 获取所有 unit（用于全量卡片创建）

#### 3.2 卡片操作
- ✅ `DeleteCardsByUnit` - 删除指定 unit 下的所有卡片
- ✅ `CreateCard` - 创建卡片
- ✅ `ConvertDevicesToJSON` - 将设备列表转换为 JSON（用于 cards.devices JSONB 字段）
- ✅ `ConvertResidentsToJSON` - 将住户列表转换为 JSON（用于 cards.residents JSONB 字段）

## 📊 数据流

```
定时轮询（默认 60 秒）
    │
    ▼
获取所有 unit
    │
    ▼
为每个 unit 创建卡片
    ├─ 获取 ActiveBed 列表
    ├─ 根据 ActiveBed 数量判断场景（A/B/C）
    ├─ 删除旧卡片
    ├─ 创建新卡片
    │   ├─ 计算卡片名称
    │   ├─ 计算卡片地址
    │   ├─ 获取设备列表
    │   ├─ 获取住户列表
    │   └─ 写入 cards 表
    └─ 完成
```

## 🔧 配置

### 环境变量

```bash
# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=owlrd
DB_SSLMODE=disable

# 租户 ID（必需）
TENANT_ID=your-tenant-id

# 触发模式
CARD_TRIGGER_MODE=polling  # "polling" 或 "events"

# 轮询配置
CARD_POLLING_INTERVAL=60  # 轮询间隔（秒）

# 日志
LOG_LEVEL=info
LOG_FORMAT=json
```

## 🚀 运行

### 启动服务

```bash
cd wisefido-card-aggregator
go run cmd/wisefido-card-aggregator/main.go
```

### 构建

```bash
cd wisefido-card-aggregator
go build -o bin/wisefido-card-aggregator cmd/wisefido-card-aggregator/main.go
```

## ⚠️ 待完善

### 1. 事件驱动模式 ⏳
- 当前只实现了轮询模式
- 需要实现事件驱动模式（监听设备/住户/床位绑定关系变化）

### 2. 多租户支持 ⏳
- 当前需要手动配置 TENANT_ID
- 可以扩展为支持多个租户

### 3. 增量更新优化 ⏳
- 当前是全量重新创建所有卡片
- 可以优化为增量更新（只更新变化的 unit）

### 4. 错误处理和重试 ⏳
- 需要更完善的错误处理
- 需要重试机制

### 5. 监控和日志 ⏳
- 需要更详细的监控指标
- 需要性能指标

## 📝 相关文档

- [卡片创建规则](../../owlRD/docs/20_Card_Creation_Rules_Final.md)
- [cards 表结构](../../owlRD/db/21_cards.sql)

## 🎯 下一步

1. **测试**：测试卡片创建逻辑，确保符合业务规则
2. **实现事件驱动模式**：监听设备/住户/床位绑定关系变化，实时更新卡片
3. **优化性能**：实现增量更新，减少数据库操作
4. **完善错误处理**：添加重试机制和错误恢复

