# Device Handler 深度分析

## 📋 Handler 基本信息

```
Handler 名称：AdminAPI.DevicesHandler
文件路径：internal/http/admin_units_devices_handlers.go
实现文件：internal/http/admin_units_devices_impl.go
当前行数：约 150 行（Device 相关）
业务领域：设备管理
```

---

## 🔍 端点分析

### 端点列表

| 端点 | HTTP 方法 | 路径 | 功能描述 | 复杂度 | 实现行数 |
|------|----------|------|----------|--------|---------|
| 查询设备列表 | GET | `/admin/api/v1/devices` | 支持状态、业务访问权限、设备类型、搜索过滤，分页 | 中 | ~35 |
| 查询设备详情 | GET | `/admin/api/v1/devices/:id` | 获取单个设备信息 | 低 | ~15 |
| 更新设备 | PUT | `/admin/api/v1/devices/:id` | 更新设备信息，包含绑定验证 | 中 | ~45 |
| 删除设备 | DELETE | `/admin/api/v1/devices/:id` | 禁用设备（软删除） | 低 | ~10 |

**总计**：4 个端点，约 150 行代码

---

## 📝 详细业务逻辑分析

### 1. GET /admin/api/v1/devices - 查询设备列表

#### 1.1 路由分发

**位置**：`admin_units_devices_handlers.go:250-283`

```go
func (a *AdminAPI) DevicesHandler(w http.ResponseWriter, r *http.Request) {
    if a.Devices == nil {
        a.Stub.AdminDevices(w, r)
        return
    }
    if r.URL.Path == "/admin/api/v1/devices" {
        switch r.Method {
        case http.MethodGet:
            a.getDevices(w, r)
        default:
            w.WriteHeader(http.StatusMethodNotAllowed)
        }
        return
    }
    // ... 其他路由
}
```

**逻辑**：
- ✅ 检查 `Devices` Repository 是否存在，不存在则 fallback 到 Stub
- ✅ 路径匹配 `/admin/api/v1/devices`
- ✅ 方法匹配 `GET`

---

#### 1.2 参数解析

**位置**：`admin_units_devices_impl.go:293-326`

```go
func (a *AdminAPI) getDevices(w http.ResponseWriter, r *http.Request) {
    tenantID, ok := a.tenantIDFromReq(w, r)
    if !ok {
        return
    }
    // status can be repeated ?status=online&status=offline or status[]=...
    statuses := r.URL.Query()["status"]
    // Some frontend uses status as array directly; if it's comma-separated, split
    if len(statuses) == 1 && strings.Contains(statuses[0], ",") {
        statuses = strings.Split(statuses[0], ",")
    }
    filters := repository.DeviceFilters{
        Status:         statuses,
        BusinessAccess: r.URL.Query().Get("business_access"),
        DeviceType:     r.URL.Query().Get("device_type"),
        SearchType:     r.URL.Query().Get("search_type"),
        SearchKeyword:  r.URL.Query().Get("search_keyword"),
    }
    page := parseInt(r.URL.Query().Get("page"), 1)
    size := parseInt(r.URL.Query().Get("size"), 20)
    // ...
}
```

**参数列表**：
- ✅ `tenant_id` - 租户 ID（必填，从 Query 或 Header 获取）
- ✅ `status` - 设备状态（可选，支持多个值，支持逗号分隔或数组格式）
- ✅ `business_access` - 业务访问权限（可选：pending, approved, rejected）
- ✅ `device_type` - 设备类型（可选）
- ✅ `search_type` - 搜索类型（可选：device_name, serial_number, uid）
- ✅ `search_keyword` - 搜索关键词（可选）
- ✅ `page` - 页码（可选，默认 1）
- ✅ `size` - 每页数量（可选，默认 20）

**特殊处理**：
- ✅ `status` 参数支持多种格式：
  - 多个值：`?status=online&status=offline`
  - 逗号分隔：`?status=online,offline`
  - 数组格式：`?status[]=online&status[]=offline`

---

#### 1.3 业务逻辑

**位置**：`admin_units_devices_impl.go:313-325`

```go
items, total, err := a.Devices.ListDevices(r.Context(), tenantID, filters, page, size)
if err != nil {
    writeJSON(w, http.StatusOK, Fail("failed to list devices"))
    return
}
out := make([]any, 0, len(items))
for _, d := range items {
    out = append(out, d.ToJSON())
}
writeJSON(w, http.StatusOK, Ok(map[string]any{
    "items": out,
    "total": total,
}))
```

**业务逻辑**：
1. ✅ 调用 `DevicesRepository.ListDevices` 查询设备列表
2. ✅ 将每个设备转换为 JSON 格式（使用 `domain.Device.ToJSON()`）
3. ✅ 返回分页结果：`{items: [...], total: ...}`

**错误处理**：
- ✅ 查询失败：返回 `"failed to list devices"`

---

### 2. GET /admin/api/v1/devices/:id - 查询设备详情

#### 2.1 路由分发

**位置**：`admin_units_devices_handlers.go:264-280`

```go
if strings.HasPrefix(r.URL.Path, "/admin/api/v1/devices/") {
    id := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/devices/")
    if id == "" || strings.Contains(id, "/") {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    switch r.Method {
    case http.MethodGet:
        a.getDeviceDetail(w, r, id)
    // ...
    }
}
```

**逻辑**：
- ✅ 提取设备 ID
- ✅ 验证 ID 格式（不能为空，不能包含 `/`）
- ✅ 方法匹配 `GET`

---

#### 2.2 参数解析

**位置**：`admin_units_devices_impl.go:328-343`

```go
func (a *AdminAPI) getDeviceDetail(w http.ResponseWriter, r *http.Request, deviceID string) {
    tenantID, ok := a.tenantIDFromReq(w, r)
    if !ok {
        return
    }
    d, err := a.Devices.GetDevice(r.Context(), tenantID, deviceID)
    // ...
}
```

**参数列表**：
- ✅ `tenant_id` - 租户 ID（必填）
- ✅ `device_id` - 设备 ID（从路径获取）

---

#### 2.3 业务逻辑

```go
d, err := a.Devices.GetDevice(r.Context(), tenantID, deviceID)
if err != nil {
    if err == sql.ErrNoRows {
        writeJSON(w, http.StatusOK, Fail("device not found"))
        return
    }
    writeJSON(w, http.StatusOK, Fail("failed to get device"))
    return
}
writeJSON(w, http.StatusOK, Ok(d.ToJSON()))
```

**业务逻辑**：
1. ✅ 调用 `DevicesRepository.GetDevice` 查询设备详情
2. ✅ 将设备转换为 JSON 格式
3. ✅ 返回设备信息

**错误处理**：
- ✅ 设备不存在：返回 `"device not found"`
- ✅ 查询失败：返回 `"failed to get device"`

---

### 3. PUT /admin/api/v1/devices/:id - 更新设备

#### 3.1 路由分发

**位置**：`admin_units_devices_handlers.go:270-274`

```go
case http.MethodPut:
    a.updateDevice(w, r, id)
```

---

#### 3.2 参数解析

**位置**：`admin_units_devices_impl.go:345-376`

```go
func (a *AdminAPI) updateDevice(w http.ResponseWriter, r *http.Request, deviceID string) {
    tenantID, ok := a.tenantIDFromReq(w, r)
    if !ok {
        return
    }
    var payload map[string]any
    if err := readBodyJSON(r, 1<<20, &payload); err != nil {
        writeJSON(w, http.StatusOK, Fail("invalid body"))
        return
    }
    // 关键对齐：前端不会"只传 unit_id"，它会先 ensureUnitRoom 再传 bound_room_id
    // 因此这里收紧：如果请求里携带了 unit_id，但 bound_room_id/bound_bed_id 都为空/缺失，直接报错，避免后端兜底掩盖问题
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
    // ...
}
```

**参数列表**：
- ✅ `tenant_id` - 租户 ID（必填）
- ✅ `device_id` - 设备 ID（从路径获取）
- ✅ Body 参数（可选）：
  - `device_name` - 设备名称
  - `device_store_id` - 设备库存 ID
  - `serial_number` - 序列号
  - `uid` - UID
  - `bound_room_id` - 绑定的房间 ID
  - `bound_bed_id` - 绑定的床位 ID
  - `unit_id` - 单元 ID（如果提供，必须同时提供 bound_room_id 或 bound_bed_id）
  - `status` - 设备状态
  - `business_access` - 业务访问权限
  - `monitoring_enabled` - 是否启用监控
  - `metadata` - 元数据（JSON 字符串）

**业务规则验证**：
- ✅ 如果提供了 `unit_id`，必须同时提供 `bound_room_id` 或 `bound_bed_id`
- ✅ 验证失败返回：`"invalid binding: unit_id provided but bound_room_id/bound_bed_id missing"`

---

#### 3.3 数据转换

**位置**：`admin_units_devices_impl.go:390-434`

```go
// payloadToDevice 将map[string]any转换为domain.Device
func payloadToDevice(payload map[string]any) *domain.Device {
    device := &domain.Device{}
    
    if v, ok := payload["device_name"].(string); ok {
        device.DeviceName = v
    }
    if v, ok := payload["device_store_id"].(string); ok && v != "" {
        device.DeviceStoreID = sql.NullString{String: v, Valid: true}
    }
    // ... 其他字段转换
    return device
}
```

**转换逻辑**：
- ✅ 将 `map[string]any` 转换为 `domain.Device`
- ✅ 处理可选字段（使用 `sql.NullString`）
- ✅ 处理空值（设置为 `Valid: false`）

---

#### 3.4 业务逻辑

```go
// 转换为domain.Device
device := payloadToDevice(payload)
if err := a.Devices.UpdateDevice(r.Context(), tenantID, deviceID, device); err != nil {
    writeJSON(w, http.StatusOK, Fail("failed to update device"))
    return
}
writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
```

**业务逻辑**：
1. ✅ 将 payload 转换为 `domain.Device`
2. ✅ 调用 `DevicesRepository.UpdateDevice` 更新设备
3. ✅ 返回成功响应：`{success: true}`

**错误处理**：
- ✅ 更新失败：返回 `"failed to update device"`

---

### 4. DELETE /admin/api/v1/devices/:id - 删除设备

#### 4.1 路由分发

**位置**：`admin_units_devices_handlers.go:275-276`

```go
case http.MethodDelete:
    a.deleteDevice(w, r, id)
```

---

#### 4.2 参数解析

**位置**：`admin_units_devices_impl.go:378-388`

```go
func (a *AdminAPI) deleteDevice(w http.ResponseWriter, r *http.Request, deviceID string) {
    tenantID, ok := a.tenantIDFromReq(w, r)
    if !ok {
        return
    }
    if err := a.Devices.DisableDevice(r.Context(), tenantID, deviceID); err != nil {
        writeJSON(w, http.StatusOK, Fail("failed to delete device"))
        return
    }
    writeJSON(w, http.StatusOK, Ok(map[string]any{"success": true}))
}
```

**参数列表**：
- ✅ `tenant_id` - 租户 ID（必填）
- ✅ `device_id` - 设备 ID（从路径获取）

---

#### 4.3 业务逻辑

**注意**：虽然端点是 DELETE，但实际调用的是 `DisableDevice`（软删除）

**业务逻辑**：
1. ✅ 调用 `DevicesRepository.DisableDevice` 禁用设备
2. ✅ 返回成功响应：`{success: true}`

**错误处理**：
- ✅ 删除失败：返回 `"failed to delete device"`

---

## 📊 业务规则总结

### 1. 租户验证

- ✅ 所有端点都需要 `tenant_id`
- ✅ 从 Query 参数或 Header (`X-Tenant-Id`) 获取
- ✅ 如果无法获取，返回错误：`"tenant_id is required"`

### 2. 设备绑定验证

- ✅ 更新设备时，如果提供了 `unit_id`，必须同时提供 `bound_room_id` 或 `bound_bed_id`
- ✅ 验证失败返回：`"invalid binding: unit_id provided but bound_room_id/bound_bed_id missing"`

### 3. 状态过滤

- ✅ `status` 参数支持多种格式：
  - 多个值：`?status=online&status=offline`
  - 逗号分隔：`?status=online,offline`
  - 数组格式：`?status[]=online&status[]=offline`

### 4. 分页

- ✅ 默认页码：1
- ✅ 默认每页数量：20

### 5. 删除操作

- ✅ DELETE 端点实际执行的是软删除（禁用设备）
- ✅ 调用 `DisableDevice` 而不是 `DeleteDevice`

---

## 🔍 数据转换

### 1. 请求到领域模型

- ✅ `map[string]any` → `domain.Device`
- ✅ 处理可选字段（`sql.NullString`）
- ✅ 处理空值

### 2. 领域模型到响应

- ✅ `domain.Device` → JSON（使用 `ToJSON()` 方法）
- ✅ 处理可选字段（null 值）

---

## 📝 错误处理

### 错误响应格式

所有错误都使用统一格式：
```json
{
  "code": -1,
  "type": "error",
  "message": "错误消息",
  "result": null
}
```

### 错误消息列表

| 错误场景 | 错误消息 |
|---------|---------|
| 缺少 tenant_id | "tenant_id is required" |
| 查询设备列表失败 | "failed to list devices" |
| 设备不存在 | "device not found" |
| 查询设备详情失败 | "failed to get device" |
| 无效的绑定 | "invalid binding: unit_id provided but bound_room_id/bound_bed_id missing" |
| 更新设备失败 | "failed to update device" |
| 删除设备失败 | "failed to delete device" |
| 无效的请求体 | "invalid body" |

---

## ✅ 业务逻辑清单

### 查询设备列表

1. ✅ 获取 tenant_id（Query 或 Header）
2. ✅ 解析 status 参数（支持多种格式）
3. ✅ 构建 DeviceFilters
4. ✅ 解析分页参数（page, size）
5. ✅ 调用 Repository.ListDevices
6. ✅ 转换设备列表为 JSON
7. ✅ 返回分页结果

### 查询设备详情

1. ✅ 获取 tenant_id
2. ✅ 提取 device_id（从路径）
3. ✅ 调用 Repository.GetDevice
4. ✅ 处理设备不存在的情况
5. ✅ 转换设备为 JSON
6. ✅ 返回设备信息

### 更新设备

1. ✅ 获取 tenant_id
2. ✅ 提取 device_id（从路径）
3. ✅ 解析请求体
4. ✅ 验证设备绑定规则（unit_id + bound_room_id/bound_bed_id）
5. ✅ 转换 payload 为 domain.Device
6. ✅ 调用 Repository.UpdateDevice
7. ✅ 返回成功响应

### 删除设备

1. ✅ 获取 tenant_id
2. ✅ 提取 device_id（从路径）
3. ✅ 调用 Repository.DisableDevice（软删除）
4. ✅ 返回成功响应

---

## 🎯 职责边界

### Handler 层职责

- ✅ HTTP 路由分发
- ✅ 参数解析（Query、Path、Body）
- ✅ 租户 ID 获取
- ✅ 数据格式转换（map ↔ domain）
- ✅ HTTP 响应构建
- ✅ 错误响应格式化

### Repository 层职责

- ✅ 数据库查询
- ✅ 数据过滤
- ✅ 分页处理
- ✅ 数据持久化

### Service 层职责（待设计）

- ✅ 业务规则验证
- ✅ 业务逻辑编排
- ✅ 错误处理
- ✅ 日志记录

---

## 📋 待确认问题

1. **设备创建**：当前 Handler 中没有创建设备的端点，是否需要添加？
2. **权限检查**：是否需要添加权限检查逻辑？
3. **设备状态**：设备状态有哪些有效值？
4. **业务访问权限**：business_access 有哪些有效值？
5. **设备类型**：device_type 有哪些有效值？

---

## ✅ 分析完成

**阶段 1 完成**：已提取所有业务逻辑，创建了完整的业务逻辑清单。

**下一步**：进入阶段 2，设计 Service 接口。

