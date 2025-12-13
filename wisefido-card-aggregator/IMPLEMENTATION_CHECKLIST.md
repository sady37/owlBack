# wisefido-card-aggregator 实现检查清单

## 文件结构检查

### ✅ 已实现文件

1. **配置层**
   - ✅ `internal/config/config.go` - 配置加载（数据库、Redis、租户ID、轮询间隔等）

2. **数据访问层（Repository）**
   - ✅ `internal/repository/card.go` - 卡片数据访问
     - ✅ `GetActiveBedsByUnit` - 获取 ActiveBed 列表
     - ✅ `GetUnitInfo` - 获取 Unit 信息（包含 groupList 和 userList）
     - ✅ `GetDevicesByBed` - 获取床位绑定的设备
     - ✅ `GetUnboundDevicesByUnit` - 获取未绑床的设备
     - ✅ `GetResidentsByBed` - 获取床位关联的住户
     - ✅ `GetResidentsByUnit` - 获取单元关联的住户
     - ✅ `GetAllUnits` - 获取所有单元ID
   - ✅ `internal/repository/routing.go` - 报警路由转换
     - ✅ `ConvertUserListToUUIDArray` - 将 units.userList (JSONB) 转换为 UUID[]
     - ✅ `ConvertGroupListToStringArray` - 将 units.groupList (JSONB) 转换为 VARCHAR[]

3. **业务逻辑层（Aggregator）**
   - ✅ `internal/aggregator/card_creator.go` - 卡片创建逻辑
     - ✅ `CreateCardsForUnit` - 为指定 unit 创建卡片
     - ✅ 场景 A：门牌下只有 1 个 ActiveBed
     - ✅ 场景 B：门牌下有多个 ActiveBed（≥2）
     - ✅ 场景 C：门牌下无 ActiveBed
     - ✅ 卡片名称计算（ActiveBed 和 UnitCard）
     - ✅ 卡片地址计算（branch_tag + building + unit_name）
     - ✅ 设备绑定规则（按最小地址优先原则）
     - ✅ 报警路由配置转换（从 units.groupList/userList 转换）

4. **服务层（Service）**
   - ✅ `internal/service/aggregator.go` - 服务生命周期管理
     - ✅ `NewAggregatorService` - 创建服务
     - ✅ `Start` - 启动服务（支持轮询模式）
     - ✅ `Stop` - 停止服务
     - ✅ `startPollingMode` - 轮询模式实现
     - ✅ `createAllCards` - 为所有 unit 创建卡片

5. **入口层**
   - ✅ `cmd/wisefido-card-aggregator/main.go` - 主程序入口
     - ✅ 配置加载
     - ✅ 日志初始化
     - ✅ 服务启动和优雅关闭

6. **依赖管理**
   - ✅ `go.mod` - Go 模块定义
   - ✅ `go.sum` - 依赖校验和

## 功能实现检查

### ✅ 核心功能

1. **ActiveBed 判断**
   - ✅ `beds.bound_device_count > 0`
   - ✅ 设备 `monitoring_enabled = TRUE`（在查询设备时过滤）
   - ⚠️ 同一Bed下只能绑定1*radar+sleepad（需要在设备查询时验证）

2. **卡片创建场景**
   - ✅ 场景 A：门牌下只有 1 个 ActiveBed
     - ✅ 创建 1 个 ActiveBed 卡片
     - ✅ 绑定该门牌内所有 monitoring_enabled = TRUE 的设备
   - ✅ 场景 B：门牌下有多个 ActiveBed（≥2）
     - ✅ 为每个 ActiveBed 创建 1 个 ActiveBed 卡片
     - ✅ 创建 UnitCard（仅当有未绑床设备时）
   - ✅ 场景 C：门牌下无 ActiveBed
     - ✅ 创建 UnitCard（仅当有未绑床设备时）

3. **卡片名称计算**
   - ✅ ActiveBed 卡片：使用住户 nickname
   - ✅ UnitCard：根据规则计算（公共空间、多人房间、单人/伴侣房）

4. **卡片地址计算**
   - ✅ 格式：`branch_tag + "-" + building + "-" + unit_name`
   - ✅ 跳过空值或默认值 "-"

5. **设备绑定规则**
   - ✅ 按最小地址优先原则（bed < room < unit）
   - ✅ 区分 direct 和 indirect 绑定类型

6. **报警路由配置**
   - ✅ 从 `units.groupList` 转换为 `cards.routing_alarm_tags`
   - ✅ 从 `units.userList` 转换为 `cards.routing_alarm_user_ids`
   - ✅ 支持 JSONB 格式解析（简单数组和对象数组）

7. **数据库操作**
   - ✅ `CreateCard` - 创建卡片（包含所有必需字段）
   - ✅ `DeleteCardsByUnit` - 删除指定 unit 下的所有卡片
   - ✅ 支持 PostgreSQL UUID[] 和 VARCHAR[] 数组类型

## 待实现功能

### ⚠️ 待优化

1. **事件驱动模式**
   - ⚠️ 当前仅实现轮询模式
   - ⚠️ 事件驱动模式（监听设备/住户/床位绑定关系变化）待实现

2. **增量更新**
   - ⚠️ 当前实现为全量重新创建（删除后重建）
   - ⚠️ 增量更新逻辑待实现（仅更新变化的卡片）

3. **错误处理和重试**
   - ⚠️ 当前错误处理较简单
   - ⚠️ 重试机制待实现

4. **设备绑定验证**
   - ⚠️ 同一Bed下只能绑定1*radar+sleepad 的验证逻辑待完善

## 代码质量检查

### ✅ 已通过

1. **编译检查**
   - ✅ `go build` 成功
   - ✅ 无编译错误

2. **依赖管理**
   - ✅ `go.mod` 包含所有必需依赖
   - ✅ `github.com/lib/pq` 已添加

3. **代码结构**
   - ✅ 分层清晰（config → repository → aggregator → service → main）
   - ✅ 职责分离明确

### ⚠️ 待检查

1. **单元测试**
   - ⚠️ 暂无单元测试

2. **集成测试**
   - ⚠️ 暂无集成测试

3. **文档**
   - ⚠️ 代码注释较完整，但缺少 README.md

## 数据库表结构一致性检查

### ✅ 已确认

1. **cards 表字段**
   - ✅ `tenant_id` - 已插入
   - ✅ `card_type` - 已插入
   - ✅ `bed_id` - 已插入（ActiveBed 必需，Location 为 NULL）
   - ✅ `unit_id` - 已插入
   - ✅ `card_name` - 已插入
   - ✅ `card_address` - 已插入
   - ✅ `resident_id` - 已插入（ActiveBed 可选，Location 为 NULL）
   - ✅ `devices` - 已插入（JSONB）
   - ✅ `residents` - 已插入（JSONB）
   - ✅ `routing_alarm_user_ids` - 已插入（从 units.userList 转换）
   - ✅ `routing_alarm_tags` - 已插入（从 units.groupList 转换）
   - ✅ `unhandled_alarm_0-4` - 使用默认值 0
   - ✅ `icon_alarm_level` - 使用默认值 3
   - ✅ `pop_alarm_emerge` - 使用默认值 0

2. **约束检查**
   - ✅ `chk_card_type_binding` - ActiveBed 时 bed_id IS NOT NULL，Location 时 unit_id IS NOT NULL AND bed_id IS NULL
   - ✅ `UNIQUE(tenant_id, bed_id)` WHERE `card_type = 'ActiveBed'`
   - ✅ `UNIQUE(tenant_id, unit_id)` WHERE `card_type = 'Location'`

## 环境变量配置

### 必需环境变量

- `TENANT_ID` - 租户ID（必需）
- `DB_HOST` - 数据库主机（默认：localhost）
- `DB_USER` - 数据库用户（默认：postgres）
- `DB_PASSWORD` - 数据库密码（默认：postgres）
- `DB_NAME` - 数据库名称（默认：owlrd）
- `DB_SSLMODE` - SSL模式（默认：disable）
- `REDIS_ADDR` - Redis地址（默认：localhost:6379）
- `REDIS_PASSWORD` - Redis密码（默认：空）
- `CARD_TRIGGER_MODE` - 触发模式（默认：polling）
- `LOG_LEVEL` - 日志级别（默认：info）

## 总结

### ✅ 已完成
- 核心卡片创建功能已实现
- 数据库操作完整
- 报警路由配置转换已实现
- 代码编译通过

### ⚠️ 待完善
- 事件驱动模式
- 增量更新逻辑
- 错误处理和重试机制
- 单元测试和集成测试
- README 文档

