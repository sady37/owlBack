# wisefido-qinglan 架构重构待办事项

## 一、整体设计结构

### 架构分层（接口-实现分离模式）
```
wisefido-qinglan/
├── cmd/wisefido-qinglan/
│   └── main.go                    # 程序入口
├── internal/
│   ├── domain/                    # 领域模型（与 wisefido-data 对齐）
│   │   └── device.go              # Device, DeviceLocationInfo 结构
│   ├── repository/                # 数据访问层
│   │   ├── device_repo.go         # DeviceRepository 接口定义
│   │   └── postgres_device.go     # PostgreSQL 实现（直接访问 owlRD 数据库）
│   ├── service/                   # 业务逻辑层
│   │   └── radar_service.go       # 雷达设备业务逻辑
│   ├── consumer/                  # 消息消费者
│   │   ├── mqtt_consumer.go       # MQTT 消息处理（已存在，需重构）
│   │   └── stream_publisher.go    # Redis Stream 发布（已存在，需重构）
│   ├── http/                      # HTTP API 层
│   │   ├── api.go                 # API 处理器（已存在，需重构）
│   │   └── server.go              # HTTP 服务器（已存在，需重构）
│   ├── mqtt/                      # MQTT 客户端
│   │   └── client.go              # MQTT 客户端（已存在）
│   └── config/                    # 配置管理
│       └── config.go              # 配置结构（已存在，需检查）
└── db/                            # 数据库相关（仅保留连接配置）
    └── connection.go              # 数据库连接配置
```

### 数据库访问原则
- **直接访问**：直接访问 `owlRD` PostgreSQL 数据库（与 `wisefido-data` 共享）
- **不建表**：数据库表已由 `owlRD/db/` schema 创建
- **只操作数据**：仅进行查询、插入、更新、删除操作

## 二、待完成工作清单

### 1. 文件创建与重构

#### 1.1 Domain 层 ✅ 已完成
- [x] `internal/domain/device.go`
  - 功能：定义设备领域模型
  - 参考：`wisefido-data/internal/domain/device.go` + `owlRD/db/17_devices.sql`
  - 状态：已完成，基于实际 schema 更新

#### 1.2 Repository 层
- [x] `internal/repository/device_repo.go` ✅ 已完成
  - 功能：定义 DeviceRepository 接口
  - 方法清单：
    - `GetDeviceByUID(ctx, uid)` - 根据UID获取设备
    - `GetDeviceLocationInfo(ctx, uid)` - 获取设备位置信息
    - `GetDevicesByTenant(ctx, tenantID)` - 根据租户获取设备列表
    - `UpdateDeviceStatus(ctx, uid, status)` - 更新设备状态
    - `UpdateDeviceMonitoring(ctx, uid, enabled)` - 更新监控状态
    - `GetDeviceProperties(ctx, uid, keys)` - 获取设备属性
    - `SetDeviceProperties(ctx, uid, properties)` - 设置设备属性
    - `CreateDevice(ctx, device)` - 创建设备
    - `UpdateDevice(ctx, device)` - 更新设备
    - `DeleteDevice(ctx, uid)` - 删除设备（软删除）
    - `SearchDevices(ctx, criteria)` - 搜索设备
    - `CountDevicesByStatus(ctx, tenantID)` - 按状态统计

- [ ] `internal/repository/postgres_device.go` ⚠️ 待创建
  - 功能：PostgreSQL 实现
  - 参考：`wisefido-data/internal/repository/postgres_devices.go`
  - 关键：基于 `owlRD/db/17_devices.sql` 实际 schema 实现

#### 1.3 Service 层
- [ ] `internal/service/radar_service.go` ⚠️ 待创建
  - 功能：雷达设备业务逻辑
  - 方法清单：
    - `GetDeviceInfo(ctx, uid)` - 获取设备信息（调用 Repository）
    - `GetDeviceLocationInfo(ctx, uid)` - 获取位置信息
    - `GetDeviceProperties(ctx, uid, keys)` - 获取属性
    - `SetDeviceProperties(ctx, uid, properties)` - 设置属性
    - `SubscribeRealtimeData(ctx, uid, content, duration)` - 订阅实时数据
    - `CallDeviceFunction(ctx, uid, dev)` - 调用设备功能
    - `GetDevicesByTenant(ctx, tenantID)` - 获取租户设备列表
    - `HandleMQTTMessage(ctx, topic, payload)` - 处理MQTT消息（协调 Repository 和 Consumer）

#### 1.4 数据库连接
- [ ] `db/connection.go` ⚠️ 待创建
  - 功能：数据库连接配置
  - 参考：`wisefido-data` 的数据库连接方式
  - 注意：删除 `db/init.sql`（已清空）

### 2. 现有文件重构

#### 2.1 MQTT Consumer 重构
- [ ] `internal/consumer/mqtt_consumer.go` ⚠️ 待重构
  - 问题：调用 `GetDeviceByUID` 参数不匹配
  - 修复：使用新的 Repository 接口
  - 依赖：需要注入 `DeviceRepository` 和 `RadarService`

#### 2.2 Stream Publisher 重构
- [ ] `internal/consumer/stream_publisher.go` ⚠️ 待重构
  - 问题：未使用导入已修复
  - 修复：使用新的 Domain 模型
  - 依赖：需要注入 `DeviceRepository`

#### 2.3 HTTP API 重构
- [ ] `internal/http/api.go` ⚠️ 待重构
  - 问题：语法错误已修复
  - 修复：使用新的 Service 层
  - 依赖：需要注入 `RadarService`

#### 2.4 HTTP Server 重构
- [ ] `internal/http/server.go` ⚠️ 待重构
  - 问题：`config.HTTPConfig` 未定义已修复
  - 修复：使用 `commonconfig.HTTPConfig`
  - 依赖：需要注入 `RadarService`

#### 2.5 清理重复文件
- [x] `internal/repository/device_repository.go` ✅ 待删除
  - 原因：将被新的接口-实现模式替代
- [x] `internal/repository/device.go` ✅ 已清空
  - 原因：内容已合并到 Domain 和 Repository

### 3. 主程序重构
- [ ] `cmd/wisefido-qinglan/main.go` ⚠️ 待重构
  - 功能：初始化所有组件
  - 步骤：
    1. 加载配置
    2. 初始化数据库连接
    3. 创建 Repository
    4. 创建 Service
    5. 创建 Consumer
    6. 创建 HTTP Server
    7. 启动所有服务

## 三、实现顺序建议

### 第一阶段：基础架构（1-2天）
1. ✅ 创建 Domain 模型（已完成）
2. ✅ 创建 Repository 接口（已完成）
3. ⚠️ 创建 PostgreSQL Repository 实现
4. ⚠️ 创建数据库连接配置
5. ⚠️ 创建 Service 层

### 第二阶段：组件重构（1-2天）
6. ⚠️ 重构 MQTT Consumer
7. ⚠️ 重构 Stream Publisher
8. ⚠️ 重构主程序

### 第三阶段：HTTP API 重构（1天）
9. ⚠️ 重构 HTTP API
10. ⚠️ 重构 HTTP Server

### 第四阶段：集成测试（1天）
11. ⚠️ 编译测试
12. ⚠️ 功能验证
13. ⚠️ 性能测试

## 四、参考源清单

### 1. 数据库 Schema（权威来源）
- `/Users/sady3721/project/owlRD/db/16_device_store.sql` - 设备库存表
- `/Users/sady3721/project/owlRD/db/17_devices.sql` - 设备表
- 其他相关表：`branches`, `buildings`, `units`, `rooms`, `beds` 等

### 2. wisefido-data 参考（架构模式）
- `wisefido-data/internal/domain/device.go` - 领域模型定义
- `wisefido-data/internal/domain/device_store.go` - 设备库存模型
- `wisefido-data/internal/repository/devices_repo.go` - Repository 接口
- `wisefido-data/internal/repository/postgres_devices.go` - PostgreSQL 实现
- `wisefido-data/internal/repository/alarm_cloud_repo.go` - 接口定义示例
- `wisefido-data/internal/repository/postgres_alarm_cloud.go` - 实现示例

### 3. wisefido-radar 参考（类似功能）
- `wisefido-radar/internal/repository/device.go` - 设备 Repository
- `wisefido-radar/internal/service/` - 业务逻辑层参考

### 4. 现有代码（需重构）
- `wisefido-qinglan/internal/consumer/mqtt_consumer.go` - MQTT 处理逻辑
- `wisefido-qinglan/internal/consumer/stream_publisher.go` - Redis Stream 发布
- `wisefido-qinglan/internal/http/api.go` - HTTP API 处理器
- `wisefido-qinglan/internal/http/server.go` - HTTP 服务器

## 五、注意事项

### 1. 数据库访问
- 使用相同的 PostgreSQL 数据库（`owlRD`）
- 不创建表，只操作数据
- 注意外键约束和业务规则

### 2. 性能考虑
- MQTT 消息高频到达，Repository 需要高效
- Redis Stream 发布需要低延迟
- 数据库连接需要连接池

### 3. 错误处理
- Repository 层：数据库错误处理
- Service 层：业务错误处理
- Handler 层：HTTP 错误响应

### 4. 测试策略
- Repository：单元测试（可 mock 数据库）
- Service：单元测试（可 mock Repository）
- 集成测试：实际数据库和 MQTT

## 六、当前状态检查清单

### 已完成 ✅
- [x] 分析数据库 schema
- [x] 确定架构模式（接口-实现分离）
- [x] 创建 Domain 模型
- [x] 创建 Repository 接口
- [x] 清理不必要的初始化文件
- [x] 修复部分编译错误

### 进行中 ⚠️
- [ ] 创建 PostgreSQL Repository 实现
- [ ] 创建 Service 层
- [ ] 重构现有组件

### 待开始 ❌
- [ ] 集成测试
- [ ] 性能优化
- [ ] 文档更新

---

**最后更新**：2024-01-15  
**负责人**：AI Assistant  
**目标完成时间**：3-5天