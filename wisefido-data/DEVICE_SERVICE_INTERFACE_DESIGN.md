# Device Service 接口设计

## 📋 设计概述

基于阶段 1 的分析，设计 `DeviceService` 接口，将业务逻辑从 Handler 层迁移到 Service 层。

---

## 🎯 设计原则

1. **职责分离**：
   - Handler 层：HTTP 请求/响应处理、参数解析、数据格式转换
   - Service 层：业务逻辑、业务规则验证、业务编排
   - Repository 层：数据访问、数据持久化

2. **强类型**：
   - 使用 `domain.Device` 而不是 `map[string]any`
   - 使用明确的请求/响应结构体

3. **错误处理**：
   - Service 层返回明确的错误信息
   - Handler 层负责错误响应格式化

---

## 📐 Service 接口设计

### 1. 接口定义

```go
package service

import (
    "context"
    "wisefido-data/internal/domain"
    "wisefido-data/internal/repository"
)

// DeviceService 设备管理 Service
type DeviceService interface {
    // 查询
    ListDevices(ctx context.Context, req ListDevicesRequest) (*ListDevicesResponse, error)
    GetDevice(ctx context.Context, req GetDeviceRequest) (*GetDeviceResponse, error)
    
    // 更新
    UpdateDevice(ctx context.Context, req UpdateDeviceRequest) (*UpdateDeviceResponse, error)
    
    // 删除
    DeleteDevice(ctx context.Context, req DeleteDeviceRequest) (*DeleteDeviceResponse, error)
}
```

---

### 2. 请求/响应结构体

#### 2.1 ListDevicesRequest

```go
type ListDevicesRequest struct {
    TenantID       string   // 必填
    Status         []string // 可选：设备状态过滤（online, offline, error）
    BusinessAccess string   // 可选：业务访问权限（pending, approved, rejected）
    DeviceType     string   // 可选：设备类型
    SearchType     string   // 可选：搜索类型（device_name, serial_number, uid）
    SearchKeyword  string   // 可选：搜索关键词
    Page           int      // 可选，默认 1
    Size           int      // 可选，默认 20
}
```

#### 2.2 ListDevicesResponse

```go
type ListDevicesResponse struct {
    Items []*domain.Device // 设备列表
    Total int                // 总数量
}
```

#### 2.3 GetDeviceRequest

```go
type GetDeviceRequest struct {
    TenantID string // 必填
    DeviceID string // 必填
}
```

#### 2.4 GetDeviceResponse

```go
type GetDeviceResponse struct {
    Device *domain.Device // 设备信息
}
```

#### 2.5 UpdateDeviceRequest

```go
type UpdateDeviceRequest struct {
    TenantID       string // 必填
    DeviceID       string // 必填
    Device         *domain.Device // 设备信息（部分更新）
}
```

#### 2.6 UpdateDeviceResponse

```go
type UpdateDeviceResponse struct {
    Success bool // 更新成功
}
```

#### 2.7 DeleteDeviceRequest

```go
type DeleteDeviceRequest struct {
    TenantID string // 必填
    DeviceID string // 必填
}
```

#### 2.8 DeleteDeviceResponse

```go
type DeleteDeviceResponse struct {
    Success bool // 删除成功
}
```

---

## 🔍 方法详细设计

### 1. ListDevices - 查询设备列表

#### 1.1 职责

- ✅ 参数验证（tenant_id 必填）
- ✅ 构建 DeviceFilters
- ✅ 调用 Repository.ListDevices
- ✅ 返回设备列表和总数

#### 1.2 业务规则

- ✅ `tenant_id` 必填
- ✅ `page` 默认 1，`size` 默认 20
- ✅ `status` 支持多个值
- ✅ 自动过滤 `status='disabled'` 的设备（Repository 层处理）

#### 1.3 错误处理

- ✅ `tenant_id` 为空：返回错误 "tenant_id is required"
- ✅ Repository 查询失败：返回错误 "failed to list devices"

---

### 2. GetDevice - 查询设备详情

#### 2.1 职责

- ✅ 参数验证（tenant_id, device_id 必填）
- ✅ 调用 Repository.GetDevice
- ✅ 返回设备信息

#### 2.2 业务规则

- ✅ `tenant_id` 必填
- ✅ `device_id` 必填

#### 2.3 错误处理

- ✅ `tenant_id` 为空：返回错误 "tenant_id is required"
- ✅ `device_id` 为空：返回错误 "device_id is required"
- ✅ 设备不存在：返回错误 "device not found"
- ✅ Repository 查询失败：返回错误 "failed to get device"

---

### 3. UpdateDevice - 更新设备

#### 3.1 职责

- ✅ 参数验证（tenant_id, device_id 必填）
- ✅ 业务规则验证（设备绑定规则）
- ✅ 调用 Repository.UpdateDevice
- ✅ 返回更新结果

#### 3.2 业务规则

- ✅ `tenant_id` 必填
- ✅ `device_id` 必填
- ✅ **设备绑定验证**：
  - 如果提供了 `unit_id`（通过 bound_room_id 或 bound_bed_id 推断），必须同时提供 `bound_room_id` 或 `bound_bed_id`
  - 验证失败返回：`"invalid binding: unit_id provided but bound_room_id/bound_bed_id missing"`

**注意**：当前 Handler 中的验证逻辑：
```go
unitID, _ := payload["unit_id"].(string)
if unitID != "" {
    roomVal, hasRoom := payload["bound_room_id"]
    bedVal, hasBed := payload["bound_bed_id"]
    roomEmpty := !hasRoom || roomVal == nil || roomVal == ""
    bedEmpty := !hasBed || bedVal == nil || bedVal == ""
    if roomEmpty && bedEmpty {
        writeJSON(w, http.StatusOK, Fail("invalid binding: unit_id provided but bound_room_id/bound_bed_id missing"))
        return
    }
}
```

**Service 层实现**：
- 由于 `domain.Device` 中没有 `unit_id` 字段，需要从 `bound_room_id` 或 `bound_bed_id` 推断
- 或者，如果前端传递了 `unit_id`，需要在 Handler 层转换为 `bound_room_id`/`bound_bed_id`
- **建议**：在 Handler 层处理 `unit_id` 转换，Service 层只验证 `bound_room_id`/`bound_bed_id` 的逻辑

#### 3.3 错误处理

- ✅ `tenant_id` 为空：返回错误 "tenant_id is required"
- ✅ `device_id` 为空：返回错误 "device_id is required"
- ✅ 设备绑定验证失败：返回错误 "invalid binding: unit_id provided but bound_room_id/bound_bed_id missing"
- ✅ Repository 更新失败：返回错误 "failed to update device"

---

### 4. DeleteDevice - 删除设备

#### 4.1 职责

- ✅ 参数验证（tenant_id, device_id 必填）
- ✅ 调用 Repository.DisableDevice（软删除）
- ✅ 返回删除结果

#### 4.2 业务规则

- ✅ `tenant_id` 必填
- ✅ `device_id` 必填
- ✅ **软删除**：调用 `DisableDevice` 而不是 `DeleteDevice`

#### 4.3 错误处理

- ✅ `tenant_id` 为空：返回错误 "tenant_id is required"
- ✅ `device_id` 为空：返回错误 "device_id is required"
- ✅ Repository 删除失败：返回错误 "failed to delete device"

---

## 🔍 职责边界

### Handler 层职责

- ✅ HTTP 路由分发
- ✅ 参数解析（Query、Path、Body）
- ✅ 租户 ID 获取
- ✅ 数据格式转换（map ↔ domain）
- ✅ HTTP 响应构建
- ✅ 错误响应格式化
- ✅ **特殊处理**：`unit_id` 转换（如果前端传递了 `unit_id`，需要转换为 `bound_room_id`/`bound_bed_id`）

### Service 层职责

- ✅ 业务规则验证
- ✅ 参数验证
- ✅ 调用 Repository
- ✅ 错误处理
- ✅ 日志记录（可选）

### Repository 层职责

- ✅ 数据库查询
- ✅ 数据过滤
- ✅ 分页处理
- ✅ 数据持久化

---

## 📊 对比旧 Handler 逻辑

| 功能点 | 旧 Handler | 新 Service | 状态 |
|--------|-----------|-----------|------|
| 查询设备列表 | ✅ Handler 层 | ✅ Service 层 | ✅ 迁移 |
| 查询设备详情 | ✅ Handler 层 | ✅ Service 层 | ✅ 迁移 |
| 更新设备 | ✅ Handler 层 | ✅ Service 层 | ✅ 迁移 |
| 删除设备 | ✅ Handler 层 | ✅ Service 层 | ✅ 迁移 |
| 参数解析 | ✅ Handler 层 | ✅ Handler 层 | ✅ 保留 |
| 数据转换 | ✅ Handler 层 | ✅ Handler 层 | ✅ 保留 |
| 业务规则验证 | ✅ Handler 层 | ✅ Service 层 | ✅ 迁移 |
| 错误处理 | ✅ Handler 层 | ✅ Service 层 | ✅ 迁移 |

---

## ✅ 接口设计确认

### 设计原则

- ✅ 职责边界清晰
- ✅ 使用强类型（domain.Device）
- ✅ 错误处理明确
- ✅ 与旧 Handler 逻辑一致

### 待确认问题

1. **设备创建**：当前 Handler 中没有创建设备的端点，是否需要添加？
   - **建议**：暂时不添加，保持与旧 Handler 一致

2. **unit_id 处理**：前端可能传递 `unit_id`，但 `domain.Device` 中没有该字段
   - **建议**：在 Handler 层处理 `unit_id` 转换，Service 层只处理 `bound_room_id`/`bound_bed_id`

3. **权限检查**：是否需要添加权限检查逻辑？
   - **建议**：暂时不添加，保持与旧 Handler 一致（后续可以添加）

---

## 🎯 下一步

**阶段 2 完成**：Service 接口设计已完成。

**下一步**：进入阶段 3，实现 Service。

