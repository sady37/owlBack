# Service 层需求总结（基于前端和后端实际需求）

## 📋 前端页面和 API 需求

### 页面列表

1. **AlarmCloud.vue** - 报警策略配置页面
   - 功能：查看和编辑报警策略配置
   - 权限：需要 `canEdit` 权限检查

2. **AlarmRecord.vue** - 报警记录页面
   - 功能：查看报警记录（Pending/Resolved两个tab）

3. **AlarmRecordList.vue** - 报警记录列表组件
   - 功能：显示报警列表，处理报警
   - 权限：处理报警需要权限检查（Facility vs Home）

---

## 🔌 API 端点需求

### 1. Alarm Cloud API

| 端点 | 方法 | 功能 | 权限检查 | 业务规则 | 数据转换 |
|------|------|------|---------|---------|---------|
| `/admin/api/v1/alarm-cloud` | GET | 获取配置 | ✅ 需要 | ❌ | ✅ JSONB |
| `/admin/api/v1/alarm-cloud` | PUT | 更新配置 | ✅ 需要 | ✅ 需要 | ✅ JSONB |

### 2. Alarm Events API

| 端点 | 方法 | 功能 | 权限检查 | 权限过滤 | 业务规则 | 复杂查询 |
|------|------|------|---------|---------|---------|---------|
| `/admin/api/v1/alarm-events` | GET | 获取列表 | ✅ 需要 | ✅ 需要 | ❌ | ✅ 多表JOIN |
| `/admin/api/v1/alarm-events/:id/handle` | PUT | 处理报警 | ✅ 需要（Facility vs Home） | ❌ | ✅ 需要 | ✅ 跨表查询 |

---

## 🎯 Service 层设计决策（修正版）

### 需要 Service 的 Repository

| Repository | API 端点 | 需要 Service 的原因 |
|-----------|---------|-------------------|
| **AlarmCloudRepository** | GET /admin/api/v1/alarm-cloud | ✅ 权限检查、数据转换 |
| **AlarmCloudRepository** | PUT /admin/api/v1/alarm-cloud | ✅ 权限检查、业务规则验证、数据转换 |
| **AlarmEventsRepository** | GET /admin/api/v1/alarm-events | ✅ 权限过滤、复杂查询、数据转换 |
| **AlarmEventsRepository** | PUT /admin/api/v1/alarm-events/:id/handle | ✅ 权限检查（Facility vs Home）、业务规则验证、状态管理、跨表查询 |

### 不需要 Service 的 Repository

| Repository | 原因 |
|-----------|------|
| AlarmDeviceRepository | 后台服务使用，无 HTTP API |
| CardRepository | 后台服务使用，无 HTTP API |
| DeviceRepository | 后台服务使用，无 HTTP API |
| RoomRepository | 后台服务使用，无 HTTP API |

---

## 📊 最终决策矩阵

| Repository | HTTP API | 后台服务 | 是否需要 Service |
|-----------|---------|---------|----------------|
| **AlarmCloudRepository** | ✅ 有（GET, PUT） | ✅ 有 | ✅ **需要**（HTTP API 场景） |
| **AlarmEventsRepository** | ✅ 有（GET, PUT） | ✅ 有 | ✅ **需要**（HTTP API 场景） |
| AlarmDeviceRepository | ❌ 无 | ✅ 有 | ❌ **不需要** |
| CardRepository | ❌ 无 | ✅ 有 | ❌ **不需要** |
| DeviceRepository | ❌ 无 | ✅ 有 | ❌ **不需要** |
| RoomRepository | ❌ 无 | ✅ 有 | ❌ **不需要** |

---

## 🏗️ Service 层架构

### HTTP API 场景

```
HTTP Handler
  ↓
AlarmCloudService（需要）
  ↓
AlarmCloudRepository

HTTP Handler
  ↓
AlarmEventService（需要）
  ↓
AlarmEventsRepository
```

### 后台服务场景

```
Evaluator
  ↓
直接使用所有 Repository
  - AlarmCloudRepository
  - AlarmEventsRepository
  - AlarmDeviceRepository
  - CardRepository
  - DeviceRepository
  - RoomRepository
```

---

## ✅ 结论

**需要实现的 Service**：
1. ✅ **AlarmCloudService** - 用于 HTTP API（GET, PUT /admin/api/v1/alarm-cloud）
2. ✅ **AlarmEventService** - 用于 HTTP API（GET, PUT /admin/api/v1/alarm-events）

**不需要实现的 Service**：
- AlarmDeviceRepository - 后台服务使用
- CardRepository - 后台服务使用
- DeviceRepository - 后台服务使用
- RoomRepository - 后台服务使用

---

## 📚 参考文档

- `SERVICE_DESIGN_BASED_ON_REQUIREMENTS.md` - 基于实际需求的详细设计
- `SERVICE_DECISION_MATRIX.md` - 决策矩阵表
- `owlFront/docs/Alarm_event.md` - 前端 API 设计文档
- `owlFront/src/api/alarm/alarm.ts` - 前端 API 实现
- `owlBack/wisefido-data/internal/http/admin_alarm_handlers.go` - 后端 Handler 实现

