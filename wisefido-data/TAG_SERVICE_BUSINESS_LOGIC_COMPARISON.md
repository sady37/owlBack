# Tag Service vs 旧 Handler 业务逻辑对比

## 📋 对比分析

### 1. GET /admin/api/v1/tags 对比

#### 旧 Handler 逻辑（admin_tags_handlers.go:44-121）

**关键业务逻辑**：
1. ✅ **tenant_id 处理**：
   - 从请求获取 tenant_id（通过 tenantIDFromReq）
   - 必须提供 tenant_id

2. ✅ **过滤逻辑**：
   - tag_type: 可选过滤（从 Query 参数获取）
   - include_system_tag_types: 可选（默认为 true）
   - 如果指定 tag_type，使用 `get_tags_for_tenant($1, $2)`
   - 如果不指定 tag_type，使用 `get_tags_for_tenant($1, NULL)`
   - 如果 include_system_tag_types=false，在应用层过滤系统预定义类型

3. ✅ **排序逻辑**：
   - 按 `tag_type, tag_name` 排序

4. ✅ **响应格式**：
   - 包含 items, total, available_tag_types, system_predefined_tag_types
   - tag_objects 字段已删除（不再返回）

#### 新 Service 逻辑（tag_service.go:55-97）

**当前实现**：
1. ✅ **tenant_id 处理**：已实现（通过 req.TenantID）
2. ✅ **过滤逻辑**：已实现（通过 Repository 的 TagsFilter）
3. ✅ **排序逻辑**：已实现（在 Repository 层处理）
4. ✅ **响应格式**：已实现（包含所有字段）

**对比结果**：
- ✅ 所有业务逻辑点都已覆盖
- ✅ 逻辑一致

---

### 2. POST /admin/api/v1/tags 对比

#### 旧 Handler 逻辑（admin_tags_handlers.go:124-175）

**关键业务逻辑**：
1. ✅ **参数验证**：
   - tag_name 必填，不能为空
   - tag_type 可选，默认为 "user_tag"

2. ✅ **标签类型验证**：
   - 允许的类型：branch_tag, family_tag, area_tag, user_tag
   - 验证 tag_type 是否在允许列表中

3. ✅ **创建逻辑**：
   - 调用 `upsert_tag_to_catalog` 函数
   - 如果 tag_name 已存在，更新 tag_type
   - 返回 tag_id

#### 新 Service 逻辑（tag_service.go:143-184）

**当前实现**：
1. ✅ **参数验证**：已实现（tag_name 必填，tag_type 可选）
2. ✅ **标签类型验证**：已实现（验证允许的类型）
3. ✅ **创建逻辑**：已实现（调用 Repository 的 CreateTag，内部调用 upsert_tag_to_catalog）

**对比结果**：
- ✅ 所有业务逻辑点都已覆盖
- ✅ 逻辑一致

---

### 3. DELETE /admin/api/v1/tags 对比

#### 旧 Handler 逻辑（admin_tags_handlers.go:14-40）

**关键业务逻辑**：
1. ✅ **参数验证**：
   - tenant_id 必填
   - tag_name 必填（从 Query 参数获取）

2. ✅ **删除逻辑**：
   - 调用 `drop_tag` 函数
   - 数据库函数会自动清理所有使用该 tag 的地方
   - 不检查系统预定义类型（数据库函数会处理）

#### 新 Service 逻辑（tag_service.go:238-272）

**当前实现**：
1. ✅ **参数验证**：已实现（tenant_id, tag_name 必填）
2. ✅ **系统预定义类型检查**：新增（在 Service 层检查，系统预定义类型不能删除）
3. ✅ **删除逻辑**：已实现（调用 Repository 的 DeleteTag，内部调用 drop_tag）

**对比结果**：
- ✅ 所有业务逻辑点都已覆盖
- ✅ 逻辑一致
- ✅ 新增：系统预定义类型检查（改进）

---

### 4. PUT /admin/api/v1/tags/:id 对比

#### 旧 Handler 逻辑（admin_tags_handlers.go:540-580）

**关键业务逻辑**：
1. ✅ **参数验证**：
   - tenant_id 必填
   - tag_id 必填（从 URL 路径提取）
   - tag_name 必填（从 Body 获取）

2. ✅ **更新逻辑**：
   - 直接 UPDATE tag_name
   - 不检查系统预定义类型

#### 新 Service 逻辑（tag_service.go:194-229）

**当前实现**：
1. ✅ **参数验证**：已实现（tenant_id, tag_id, tag_name 必填）
2. ✅ **系统预定义类型检查**：新增（在 Service 层检查，系统预定义类型不能修改名称）
3. ✅ **更新逻辑**：已实现（调用 Repository 的 UpdateTagName）

**对比结果**：
- ✅ 所有业务逻辑点都已覆盖
- ✅ 逻辑一致
- ✅ 新增：系统预定义类型检查（改进）

---

### 5. POST /admin/api/v1/tags/:id/objects 对比

#### 旧 Handler 逻辑（admin_tags_handlers.go:181-264）

**关键业务逻辑**：
1. ✅ **参数验证**：
   - tag_id 必填（从 URL 路径提取）
   - object_type 必填
   - objects 必填（数组格式）

2. ✅ **标签信息查询**：
   - 查询 tag_name 和 tag_type（用于同步逻辑）

3. ✅ **添加对象逻辑**：
   - 调用 `update_tag_objects` 函数（如果存在）
   - 如果 object_type == "user" && tag_type == "user_tag"，同步更新 `users.tags` JSONB
   - 同步逻辑：使用 `COALESCE(tags, '[]'::jsonb) || jsonb_build_array($1)`

#### 新 Service 逻辑（tag_service.go:300-380）

**当前实现**：
1. ✅ **参数验证**：已实现
2. ✅ **标签信息查询**：已实现（在 Service 层查询）
3. ✅ **添加对象逻辑**：已实现（调用 Repository 的 AddTagObject 和 SyncUserTag）
4. ✅ **业务规则验证**：新增（验证 tag_type 和 object_type 的匹配关系）

**对比结果**：
- ✅ 所有业务逻辑点都已覆盖
- ✅ 逻辑一致
- ✅ 新增：业务规则验证（改进）

---

### 6. DELETE /admin/api/v1/tags/:id/objects 对比

#### 旧 Handler 逻辑（admin_tags_handlers.go:269-448）

**关键业务逻辑**：
1. ✅ **参数验证**：
   - tag_id 必填（从 URL 路径提取）
   - object_type 必填
   - object_ids 或 objects 必填（支持两种格式）

2. ✅ **标签信息查询**：
   - 查询 tag_name 和 tag_type（用于同步逻辑）

3. ✅ **删除对象逻辑**：
   - 调用 `update_tag_objects` 函数（如果存在）
   - 如果 object_type == "user" && tag_type == "user_tag"，同步更新 `users.tags` JSONB
   - 如果 object_type == "resident" && tag_type == "family_tag"，同步清除 `residents.family_tag`
   - 同步逻辑：使用 `tags - $1` 和 `family_tag = NULL`

#### 新 Service 逻辑（tag_service.go:382-470）

**当前实现**：
1. ✅ **参数验证**：已实现（支持 object_ids 和 objects 两种格式）
2. ✅ **标签信息查询**：已实现（在 Service 层查询）
3. ✅ **删除对象逻辑**：已实现（调用 Repository 的 RemoveTagObject, SyncUserTag, SyncResidentFamilyTag）
4. ✅ **业务规则验证**：新增（验证 tag_type 和 object_type 的匹配关系）

**对比结果**：
- ✅ 所有业务逻辑点都已覆盖
- ✅ 逻辑一致
- ✅ 新增：业务规则验证（改进）

---

### 7. DELETE /admin/api/v1/tags/types 对比

#### 旧 Handler 逻辑（admin_tags_handlers.go:458-492）

**关键业务逻辑**：
1. ✅ **参数验证**：
   - tenant_id 必填
   - tag_type 必填（从 Body 获取）

2. ✅ **删除逻辑**：
   - 直接 DELETE FROM tags_catalog WHERE tenant_id = $1 AND tag_type = $2
   - 不检查系统预定义类型
   - 不检查权限（应该需要 SystemAdmin）

#### 新 Service 逻辑（tag_service.go:281-310）

**当前实现**：
1. ✅ **参数验证**：已实现（tenant_id, tag_type 必填）
2. ✅ **权限检查**：新增（只有 SystemAdmin 可以删除标签类型）
3. ✅ **系统预定义类型检查**：新增（系统预定义类型不能删除）
4. ✅ **删除逻辑**：已实现（调用 Repository 的 DeleteTagType，内部调用 drop_tag_type）

**对比结果**：
- ✅ 所有业务逻辑点都已覆盖
- ✅ 逻辑一致
- ✅ 新增：权限检查和系统预定义类型检查（改进）

---

### 8. GET /admin/api/v1/tags/for-object 对比

#### 旧 Handler 逻辑（admin_tags_handlers.go:493-539）

**关键业务逻辑**：
1. ✅ **参数验证**：
   - tenant_id 必填
   - object_type 必填（从 Query 参数获取）
   - object_id 必填（从 Query 参数获取）

2. ✅ **查询逻辑**：
   - 查询 `tags_catalog.tag_objects` JSONB 字段
   - 使用 JSONB 操作符查询：`tag_objects->$2->>'object_id' = $3`
   - **问题**：tag_objects 字段已删除，此查询会失败

#### 新 Service 逻辑（tag_service.go:472-505）

**当前实现**：
1. ✅ **参数验证**：已实现
2. ⚠️ **查询逻辑**：标记为 TODO（需要从源表查询）
   - user: 从 `users.tags` 查询
   - resident: 从 `residents.family_tag` 查询
   - unit: 从 `units.branch_tag`, `units.area_tag` 查询

**对比结果**：
- ⚠️ 旧 Handler 的实现已失效（tag_objects 字段已删除）
- ⚠️ 新 Service 标记为 TODO，需要重新设计

---

## 📊 关键差异总结

| 功能点 | 旧 Handler | 新 Service | 状态 |
|--------|-----------|-----------|------|
| GET 列表查询 | ✅ 直接 SQL | ✅ 通过 Repository | ✅ 一致 |
| POST 创建标签 | ✅ 直接 SQL | ✅ 通过 Repository | ✅ 一致 |
| DELETE 删除标签 | ✅ 直接 SQL | ✅ 通过 Repository | ✅ 一致 |
| DELETE 系统预定义类型检查 | ❌ 无 | ✅ 有（改进） | ✅ 改进 |
| PUT 更新标签名称 | ✅ 直接 SQL | ✅ 通过 Repository | ✅ 一致 |
| PUT 系统预定义类型检查 | ❌ 无 | ✅ 有（改进） | ✅ 改进 |
| POST 添加标签对象 | ✅ 直接 SQL | ✅ 通过 Repository | ✅ 一致 |
| POST 业务规则验证 | ❌ 无 | ✅ 有（改进） | ✅ 改进 |
| DELETE 删除标签对象 | ✅ 直接 SQL | ✅ 通过 Repository | ✅ 一致 |
| DELETE 业务规则验证 | ❌ 无 | ✅ 有（改进） | ✅ 改进 |
| DELETE 删除标签类型 | ✅ 直接 SQL | ✅ 通过 Repository | ✅ 一致 |
| DELETE 权限检查 | ❌ 无 | ✅ 有（改进） | ✅ 改进 |
| DELETE 系统预定义类型检查 | ❌ 无 | ✅ 有（改进） | ✅ 改进 |
| GET 查询对象标签 | ⚠️ 已失效 | ⚠️ TODO | ⚠️ 待重新设计 |

---

## ✅ 验证结论

### 业务逻辑完整性：✅ **完全一致（除 GetTagsForObject）**

1. ✅ **GET 方法**：所有业务逻辑点都已覆盖
2. ✅ **POST 方法**：所有业务逻辑点都已覆盖
3. ✅ **DELETE 方法**：所有业务逻辑点都已覆盖，新增系统预定义类型检查
4. ✅ **PUT 方法**：所有业务逻辑点都已覆盖，新增系统预定义类型检查
5. ✅ **POST objects 方法**：所有业务逻辑点都已覆盖，新增业务规则验证
6. ✅ **DELETE objects 方法**：所有业务逻辑点都已覆盖，新增业务规则验证
7. ✅ **DELETE types 方法**：所有业务逻辑点都已覆盖，新增权限检查和系统预定义类型检查
8. ⚠️ **GET for-object 方法**：标记为 TODO，需要重新设计

### 改进点：✅ **多项改进**

- ✅ 系统预定义类型检查（删除、更新时）
- ✅ 权限检查（删除标签类型时）
- ✅ 业务规则验证（标签对象管理时）

### 待完善点：⚠️ **GetTagsForObject**

- ⚠️ 旧 Handler 的实现已失效（tag_objects 字段已删除）
- ⚠️ 新 Service 标记为 TODO，需要重新设计

---

## 🎯 最终结论

**✅ 新 Service 与旧 Handler 的业务逻辑完全一致（除 GetTagsForObject）。**

**✅ 可以安全替换旧 Handler（GetTagsForObject 需要重新设计）**

