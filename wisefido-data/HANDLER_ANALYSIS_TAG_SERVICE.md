# TagService Handler 重构分析

## 📋 第一步：当前 Handler 业务功能点分析

### 1.1 Handler 基本信息

```
Handler 名称：AdminTags (StubHandler 方法)
文件路径：internal/http/admin_tags_handlers.go
当前行数：583 行
业务领域：标签管理
```

### 1.2 业务功能点列表

| 功能点 | HTTP 方法 | 路径 | 功能描述 | 复杂度 | 当前实现行数 |
|--------|----------|------|----------|--------|-------------|
| 查询标签列表 | GET | `/admin/api/v1/tags` | 支持 tag_type 过滤、include_system_tag_types 过滤 | 中 | ~80 |
| 创建标签 | POST | `/admin/api/v1/tags` | 创建标签，调用 upsert_tag_to_catalog，默认 user_tag | 中 | ~50 |
| 删除标签 | DELETE | `/admin/api/v1/tags` | 删除标签，调用 drop_tag 函数（自动清理所有关联） | 高 | ~30 |
| 更新标签名称 | PUT | `/admin/api/v1/tags/:id` | 更新标签名称（tag_id 不变） | 低 | ~30 |
| 添加标签对象 | POST | `/admin/api/v1/tags/:id/objects` | 添加标签成员（user/resident/unit），同步 users.tags | 高 | ~80 |
| 删除标签对象 | DELETE | `/admin/api/v1/tags/:id/objects` | 删除标签成员，同步 users.tags 和 residents.family_tag | 高 | ~180 |
| 删除标签类型 | DELETE | `/admin/api/v1/tags/types` | 删除所有指定类型的标签 | 中 | ~30 |
| 查询对象标签 | GET | `/admin/api/v1/tags/for-object` | 查询指定对象的所有标签 | 中 | ~40 |

**总计**：8 个功能点，583 行代码

### 1.3 业务规则分析

#### 权限检查
- ✅ 所有操作都需要权限检查（R/C/U/D）
- ✅ 删除标签类型需要 SystemAdmin 权限

#### 业务规则验证
1. **标签类型验证**
   - 允许的类型：`branch_tag`, `family_tag`, `area_tag`, `user_tag`
   - 系统预定义类型（`branch_tag`, `family_tag`, `area_tag`）不能删除
   - 创建时默认 `user_tag`

2. **标签名称唯一性**
   - `tag_name` 在同一 `tenant_id` 下全局唯一（跨所有 `tag_type`）
   - `tag_id` 基于 `tag_name` 确定性生成（UUID v5），即使改名也不变

3. **标签删除规则**
   - 系统预定义类型不能删除
   - 如果 tag 还在源表中使用，不能删除（由 `drop_tag` 函数检查）
   - 删除时自动清理所有关联（users.tags, residents.family_tag, units.*, etc.）

4. **标签对象管理**
   - 添加 user 到 user_tag 时，同步更新 `users.tags` JSONB
   - 删除 user 从 user_tag 时，同步更新 `users.tags` JSONB
   - 删除 resident 从 family_tag 时，同步清除 `residents.family_tag`

#### 数据转换
- ✅ 前端格式 ↔ 领域模型（Tag）
- ✅ 标签类型过滤（应用层过滤系统预定义类型）

#### 业务编排
- ✅ 标签对象管理（添加/删除成员）
- ✅ 同步 users.tags（user_tag 类型）
- ✅ 同步 residents.family_tag（family_tag 类型）

---

## 📐 第二步：Service 方法拆解

### 2.1 Service 接口设计

```go
type TagService interface {
    // 查询
    ListTags(ctx context.Context, req ListTagsRequest) (*ListTagsResponse, error)
    GetTag(ctx context.Context, req GetTagRequest) (*TagItem, error)
    GetTagsForObject(ctx context.Context, req GetTagsForObjectRequest) (*GetTagsForObjectResponse, error)
    
    // 创建
    CreateTag(ctx context.Context, req CreateTagRequest) (*CreateTagResponse, error)
    
    // 更新
    UpdateTag(ctx context.Context, req UpdateTagRequest) error
    
    // 删除
    DeleteTag(ctx context.Context, req DeleteTagRequest) error
    DeleteTagType(ctx context.Context, req DeleteTagTypeRequest) error
    
    // 标签对象管理
    AddTagObjects(ctx context.Context, req AddTagObjectsRequest) error
    RemoveTagObjects(ctx context.Context, req RemoveTagObjectsRequest) error
}
```

### 2.2 Service 方法详细设计

| Service 方法 | 对应 Handler 功能点 | 职责 | 复杂度 |
|-------------|-------------------|------|--------|
| `ListTags` | 查询标签列表 | 权限检查、参数验证、数据转换、调用 Repository | 中 |
| `GetTag` | 查询标签详情 | 权限检查、调用 Repository | 低 |
| `GetTagsForObject` | 查询对象标签 | 权限检查、调用 Repository | 中 |
| `CreateTag` | 创建标签 | 权限检查、业务规则验证（标签类型）、数据转换、调用 Repository | 中 |
| `UpdateTag` | 更新标签名称 | 权限检查、业务规则验证、调用 Repository | 低 |
| `DeleteTag` | 删除标签 | 权限检查、业务规则验证（系统预定义类型不能删除）、调用 Repository（调用 drop_tag 函数） | 高 |
| `DeleteTagType` | 删除标签类型 | 权限检查（SystemAdmin）、业务规则验证（系统预定义类型不能删除）、调用 Repository | 中 |
| `AddTagObjects` | 添加标签对象 | 权限检查、业务规则验证、业务编排（同步 users.tags）、调用 Repository | 高 |
| `RemoveTagObjects` | 删除标签对象 | 权限检查、业务规则验证、业务编排（同步 users.tags, residents.family_tag）、调用 Repository | 高 |

### 2.3 Service 请求/响应结构

```go
// ListTagsRequest 查询标签列表请求
type ListTagsRequest struct {
    TenantID          string
    UserRole          string
    TagType           string  // 可选，按 tag_type 过滤
    IncludeSystemTags bool    // 是否包含系统预定义类型
    Page              int
    Size              int
}

// ListTagsResponse 查询标签列表响应
type ListTagsResponse struct {
    Items                     []TagItem `json:"items"`
    Total                     int       `json:"total"`
    AvailableTagTypes         []string  `json:"available_tag_types"`
    SystemPredefinedTagTypes  []string  `json:"system_predefined_tag_types"`
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
    TenantID string
    UserRole string
    TagName  string
    TagType  string  // 可选，默认为 "user_tag"
}

// CreateTagResponse 创建标签响应
type CreateTagResponse struct {
    TagID string `json:"tag_id"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
    TenantID string
    UserRole string
    TagID    string
    TagName  string
}

// DeleteTagRequest 删除标签请求
type DeleteTagRequest struct {
    TenantID string
    UserRole string
    TagName  string  // 使用 tag_name（全局唯一）
}

// DeleteTagTypeRequest 删除标签类型请求
type DeleteTagTypeRequest struct {
    TenantID string
    UserRole string
    TagType  string
}

// AddTagObjectsRequest 添加标签对象请求
type AddTagObjectsRequest struct {
    TenantID   string
    UserRole   string
    TagID      string
    ObjectType string  // "user", "resident", "unit"
    Objects    []TagObject
}

// RemoveTagObjectsRequest 删除标签对象请求
type RemoveTagObjectsRequest struct {
    TenantID   string
    UserRole   string
    TagID      string
    ObjectType string
    ObjectIDs  []string  // 支持 object_ids 格式
    Objects    []TagObject  // 支持 objects 格式
}

// GetTagsForObjectRequest 查询对象标签请求
type GetTagsForObjectRequest struct {
    TenantID   string
    ObjectType string
    ObjectID   string
}

// TagObject 标签对象
type TagObject struct {
    ObjectID   string `json:"object_id"`
    ObjectName string `json:"object_name"`
}
```

---

## 🔧 第三步：Handler 方法拆解

### 3.1 Handler 结构设计

```go
type TagsHandler struct {
    tagService *service.TagService
    logger     *zap.Logger
}

func (h *TagsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 路由分发
}
```

### 3.2 Handler 方法详细设计

| Handler 方法 | 对应 Service 方法 | 职责 | 复杂度 |
|------------|------------------|------|--------|
| `ListTags` | `TagService.ListTags` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `CreateTag` | `TagService.CreateTag` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `DeleteTag` | `TagService.DeleteTag` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `UpdateTag` | `TagService.UpdateTag` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `AddTagObjects` | `TagService.AddTagObjects` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `RemoveTagObjects` | `TagService.RemoveTagObjects` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `DeleteTagType` | `TagService.DeleteTagType` | HTTP 参数解析、调用 Service、返回响应 | 低 |
| `GetTagsForObject` | `TagService.GetTagsForObject` | HTTP 参数解析、调用 Service、返回响应 | 低 |

---

## ✅ 第四步：职责边界确认

### 4.1 Handler 职责

**只负责**：
- ✅ HTTP 请求/响应处理
- ✅ 参数解析和验证（HTTP 层面）
- ✅ 调用 Service
- ✅ 错误处理和日志记录

### 4.2 Service 职责

**负责**：
- ✅ 权限检查（基于 role_permissions 表）
- ✅ 业务规则验证（标签类型、系统预定义类型不能删除）
- ✅ 数据转换（前端格式 ↔ 领域模型）
- ✅ 业务编排（同步 users.tags, residents.family_tag）
- ✅ 调用 Repository

### 4.3 Repository 职责

**负责**：
- ✅ 数据访问（CRUD 操作）
- ✅ 调用数据库函数（`upsert_tag_to_catalog`, `drop_tag`, `update_tag_objects`）
- ✅ 数据完整性验证

---

## 📋 第五步：重构计划

### 5.1 实施步骤

1. **创建 Service 接口和实现**
   - [ ] 定义 Service 接口（`tag_service.go`）
   - [ ] 实现所有 Service 方法
   - [ ] 编写 Service 单元测试

2. **创建 Handler**
   - [ ] 定义 Handler 结构（`admin_tags_handler.go`）
   - [ ] 实现所有 Handler 方法
   - [ ] 编写 Handler 单元测试

3. **集成测试**
   - [ ] 编写 Service + Repository 集成测试
   - [ ] 编写 Handler + Service 集成测试
   - [ ] 运行所有测试

4. **路由注册**
   - [ ] 在 `router.go` 中添加注册方法
   - [ ] 在 `main.go` 中集成 Service 和 Handler

5. **验证和清理**
   - [ ] 手动测试 API 端点
   - [ ] 前端功能验证
   - [ ] 清理旧代码（可选）

### 5.2 预估工作量

| 任务 | 预估时间 | 优先级 |
|------|---------|--------|
| Service 实现 | 6-8 小时 | 高 |
| Handler 实现 | 3-4 小时 | 高 |
| 测试编写 | 4-5 小时 | 高 |
| 集成和验证 | 3-4 小时 | 中 |
| **总计** | **16-21 小时** | |

